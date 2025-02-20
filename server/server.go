package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"postfiles/protocol"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

type ServerInterface interface {
	Start() error
	HandleConnection(net.Conn)
	HandleSignals()
	Shutdown()
	IsShutdown() bool
	SendFilesQuantityAndInfomation(*bufio.Writer) error
	ReceiveClientConfirmation(*bufio.Reader) (bool, error)
	SendFilesData(*bufio.Writer) error
	GetFileStat(string) (string, int64, error)
}

type Server struct {
	ip            string
	port          int
	filelist      []string
	listlength    int
	listener      net.Listener
	connectionMap *sync.Map
	shutdown      chan struct{}
	wg            *sync.WaitGroup
	connCtxs      *sync.Map
}

func NewServer(ip string, port int, filelist []string) *Server {
	return &Server{
		ip:            ip,
		port:          port,
		filelist:      filelist,
		listlength:    len(filelist),
		connectionMap: new(sync.Map),
		shutdown:      make(chan struct{}),
		wg:            new(sync.WaitGroup),
		connCtxs:      new(sync.Map),
	}
}

func (s *Server) Start() error {
	address := fmt.Sprintf("%s:%d", s.ip, s.port)
	listener, listenErr := net.Listen("tcp", address)
	if listenErr != nil {
		return listenErr
	}
	s.listener = listener

	log.Info().Str("address", address).Msg("Starting server")

	go s.HandleSignals()

	for {
		select {
		case <-s.shutdown:
			return nil
		default:
			conn, connErr := listener.Accept()
			if connErr != nil {
				if ne, ok := connErr.(net.Error); ok && ne.Timeout() {
					continue
				}
				if s.IsShutdown() {
					return nil
				}
				continue
			}

			s.wg.Add(1)
			s.connectionMap.Store(conn.RemoteAddr(), conn)
			go s.HandleConnection(conn)
		}
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.connectionMap.Delete(conn.RemoteAddr())
		s.connCtxs.Delete(conn.RemoteAddr())
		s.wg.Done()
	}()

	reader := bufio.NewReaderSize(conn, 32*1024)
	writer := bufio.NewWriterSize(conn, 32*1024)
	if setErr := conn.SetDeadline(time.Now().Add(15 * time.Second)); setErr != nil {
		log.Error().Err(setErr).Msg("Failed to set connection deadline")
		return
	}

	if sendErr := s.SendFilesQuantityAndInfomation(writer); sendErr != nil {
		log.Error().Err(sendErr).Msg("Failed to send file quantity and information")
		return
	}
	isConfirm, recvErr := s.ReceiveClientConfirmation(reader)
	if recvErr != nil {
		if errors.Is(recvErr, io.EOF) {
			return
		}
		log.Error().Err(recvErr).Msg("Failed to receive client confirmation")
		return
	}
	if isConfirm {
		if sendErr := s.SendFilesData(writer); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send file data")
			return
		}
	}
}

func (s *Server) HandleSignals() {
	singleChan := make(chan os.Signal, 1)
	signal.Notify(singleChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-singleChan
	s.Shutdown()
}

func (s *Server) Shutdown() {
	log.Warn().Msg("Stop server...")
	close(s.shutdown)
	if s.listener != nil {
		if closeErr := s.listener.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close server listener")
		}
	}

	log.Warn().Msg("Stop connections...")
	s.connectionMap.Range(func(key, value interface{}) bool {
		if conn, ok := value.(net.Conn); ok {
			if closeErr := conn.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("Failed to close connection")
			}
		}
		return true
	})

	s.wg.Wait()
	log.Warn().Msg("Server stopped...")
}

func (s *Server) IsShutdown() bool {
	select {
	case <-s.shutdown:
		return true
	default:
		return false
	}
}

func (s *Server) SendFilesQuantityAndInfomation(writer *bufio.Writer) error {
	for i := 0; i < s.listlength; i++ {
		filename, filesize, statErr := s.GetFileStat(s.filelist[i])
		if statErr != nil {
			return statErr
		}
		quantityPkt := protocol.NewPacket(protocol.FileQuantity, filename, filesize)
		if _, quantityErr := quantityPkt.EnableAndWrite(writer); quantityErr != nil {
			return quantityErr
		}
	}
	endPkt := protocol.NewPacket(protocol.EndOfTransmission, "", 0)
	if _, endErr := endPkt.EnableAndWrite(writer); endErr != nil {
		return endErr
	}
	if flushErr := writer.Flush(); flushErr != nil {
		return flushErr
	}
	return nil
}

func (s *Server) ReceiveClientConfirmation(reader *bufio.Reader) (bool, error) {
	confirmPkt := new(protocol.Packet)
	if _, readErr := confirmPkt.ReadAndDecode(reader); readErr != nil {
		return false, readErr
	}
	return confirmPkt.DataType == protocol.Confirm, nil
}

func (s *Server) SendFilesData(writer *bufio.Writer) error {
	for i := 0; i < s.listlength; i++ {
		filename, filesize, statErr := s.GetFileStat(s.filelist[i])
		if statErr != nil {
			return statErr
		}
		metaPacket := protocol.NewPacket(protocol.FileMeta, filename, 0)
		if _, metaErr := metaPacket.EnableAndWrite(writer); metaErr != nil {
			return metaErr
		}

		openFile, openErr := os.OpenFile(s.filelist[i], os.O_RDONLY, 0644)
		if openErr != nil {
			return openErr
		}
		defer openFile.Close()

		if _, copyErr := io.CopyN(writer, openFile, filesize); copyErr != nil {
			return copyErr
		}
		if flushErr := writer.Flush(); flushErr != nil {
			return flushErr
		}
	}
	return nil
}

func (s *Server) GetFileStat(path string) (string, int64, error) {
	fp := filepath.Clean(path)
	filestat, statErr := os.Stat(fp)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return "", 0, fmt.Errorf("[%s] file does not exist", fp)
		}
		if os.IsPermission(statErr) {
			return "", 0, fmt.Errorf("[%s] permission denied", fp)
		}
		return "", 0, statErr
	}
	if filestat.IsDir() {
		return "", 0, fmt.Errorf("[%s] can not be a folder", fp)
	}
	return filepath.Base(filestat.Name()), filestat.Size(), nil
}

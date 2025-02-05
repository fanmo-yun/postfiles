package server

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"postfiles/protocol"
	"sync"
	"syscall"
	"time"
)

type ServerInterface interface {
	Start() error
	HandleConnection(conn net.Conn)
	HandleSignals()
	Shutdown()
	IsShutdown() bool
	SendFilesQuantityAndInfomation(writer *bufio.Writer) error
	ReceiveClientConfirmation(reader *bufio.Reader) (bool, error)
	SendFilesData(writer *bufio.Writer) error
	GetFileStat(path string) (string, int64, error)
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
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewServer(ip string, port int, filelist []string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		ip:            ip,
		port:          port,
		filelist:      filelist,
		listlength:    len(filelist),
		connectionMap: new(sync.Map),
		shutdown:      make(chan struct{}),
		wg:            new(sync.WaitGroup),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (s *Server) Start() error {
	listener, listenErr := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if listenErr != nil {
		return listenErr
	}
	s.listener = listener

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
		s.wg.Done()
	}()

	reader := bufio.NewReaderSize(conn, 16)
	writer := bufio.NewWriterSize(conn, 4*1024)
	conn.SetDeadline(time.Now().Add(15 * time.Second))

	if err := s.SendFilesQuantityAndInfomation(writer); err != nil {
		log.Printf("Failed to send file metadata: %s\n", err)
		return
	}
	isConfirm, recvErr := s.ReceiveClientConfirmation(reader)
	if recvErr != nil {
		log.Printf("Failed to receive client confirmation: %s\n", recvErr)
		return
	}
	if isConfirm {
		if err := s.SendFilesData(writer); err != nil {
			log.Printf("Failed to send files data: %s\n", err)
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
	close(s.shutdown)
	s.cancel()
	if s.listener != nil {
		s.listener.Close()
	}

	s.connectionMap.Range(func(key, value interface{}) bool {
		if conn, ok := value.(net.Conn); ok {
			if closeErr := conn.Close(); closeErr != nil {
				log.Printf("Failed to close connection: %s\n", closeErr)
			}
		}
		return true
	})

	s.wg.Wait()
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
		quantityPacket := protocol.NewPacket(protocol.FileQuantity, filename, filesize)
		if quantitErr := quantityPacket.EnableAndWrite(writer); quantitErr != nil {
			return quantitErr
		}
	}
	endPacket := protocol.NewPacket(protocol.EndOfTransmission, "", 0)
	if endErr := endPacket.EnableAndWrite(writer); endErr != nil {
		return endErr
	}
	if flushErr := writer.Flush(); flushErr != nil {
		return flushErr
	}
	return nil
}

func (s *Server) ReceiveClientConfirmation(reader *bufio.Reader) (bool, error) {
	var decLength uint32
	if readErr := binary.Read(reader, binary.LittleEndian, &decLength); readErr != nil {
		return false, readErr
	}
	confirmData := make([]byte, decLength)
	if _, readErr := io.ReadFull(reader, confirmData); readErr != nil {
		return false, readErr
	}

	receive := new(protocol.Packet)
	decodeErr := receive.Decode(confirmData)
	if decodeErr != nil {
		return false, decodeErr
	}
	return receive.DataType == protocol.Confirm, nil
}

func (s *Server) SendFilesData(writer *bufio.Writer) error {
	for i := 0; i < s.listlength; i++ {
		filename, _, statErr := s.GetFileStat(s.filelist[i])
		if statErr != nil {
			return statErr
		}
		metaPacket := protocol.NewPacket(protocol.FileMeta, filename, 0)
		if metaErr := metaPacket.EnableAndWrite(writer); metaErr != nil {
			return metaErr
		}

		openFile, openErr := os.OpenFile(s.filelist[i], os.O_RDONLY, 0644)
		if openErr != nil {
			return openErr
		}
		defer openFile.Close()

		fileCopyBuf := make([]byte, 64*1024)
		if _, copyErr := io.CopyBuffer(writer, openFile, fileCopyBuf); copyErr != nil {
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
			return "", 0, fmt.Errorf("file does not exist: %s", fp)
		}
		if os.IsPermission(statErr) {
			return "", 0, fmt.Errorf("permission denied: %s", fp)
		}
		return "", 0, statErr
	}
	if filestat.IsDir() {
		return "", 0, fmt.Errorf("file can not be a folder: %s", fp)
	}
	return filepath.Base(filestat.Name()), filestat.Size(), nil
}

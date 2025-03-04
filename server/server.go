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
	"postfiles/log"
	"postfiles/protocol"
	"sync"
	"syscall"
)

type ServerInterface interface {
	Start() error
	Shutdown() error
	IsShutdown() bool
	HandleConnection(net.Conn)
	SendFilesQuantityAndInformation(*bufio.Writer) error
	ReceiveClientConfirmation(*bufio.Reader) (bool, error)
	SendFilesData(*bufio.Reader, *bufio.Writer) error
	GetFileStat(string) (string, int64, error)
}

type Server struct {
	ip         string
	port       int
	filelist   []string
	listlength int
	listener   net.Listener
	shutdown   chan struct{}
	connMap    *sync.Map
	wg         *sync.WaitGroup
}

func NewServer(ip string, port int, filelist []string) *Server {
	return &Server{
		ip:         ip,
		port:       port,
		filelist:   filelist,
		listlength: len(filelist),
		shutdown:   make(chan struct{}),
		connMap:    new(sync.Map),
		wg:         new(sync.WaitGroup),
	}
}

func (s *Server) Start() error {
	address := net.JoinHostPort(s.ip, fmt.Sprintf("%d", s.port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.listener = listener
	log.PrintToOut("Server start at %s\n", address)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if s.IsShutdown() {
					return
				}
				continue
			}

			s.wg.Add(1)
			s.connMap.Store(conn.RemoteAddr(), conn)
			log.PrintToOut("New Connection Accessed: %s\n", conn.RemoteAddr().String())
			go s.HandleConnection(conn)
		}
	}()

	<-signalCh
	return s.Shutdown()
}

func (s *Server) Shutdown() error {
	log.PrintToErr("Stopping server...\n")
	close(s.shutdown)

	s.connMap.Range(func(key, value any) bool {
		if conn, ok := value.(net.Conn); ok {
			if closeErr := conn.Close(); closeErr != nil {
				log.PrintToErr("Failed to close connection: %s\n", closeErr)
			}
		}
		return true
	})

	if s.listener != nil {
		if closeErr := s.listener.Close(); closeErr != nil {
			return closeErr
		}
	}

	s.wg.Wait()

	log.PrintToErr("Server stopped\n")
	return nil
}

func (s *Server) IsShutdown() bool {
	select {
	case <-s.shutdown:
		return true
	default:
		return false
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
		s.connMap.Delete(conn.RemoteAddr())
		log.PrintToOut("Connection is closed: %s\n", conn.RemoteAddr().String())
	}()

	reader := bufio.NewReaderSize(conn, 32*1024)
	writer := bufio.NewWriterSize(conn, 32*1024)

	if sendErr := s.SendFilesQuantityAndInformation(writer); sendErr != nil {
		log.PrintToErr("Failed to send file's quantity or infomation: %s\n", sendErr)
		return
	}
	isConfirm, recvErr := s.ReceiveClientConfirmation(reader)
	if recvErr != nil {
		if errors.Is(recvErr, io.EOF) {
			return
		}
		log.PrintToErr("Failed to receive client confirmation: %s\n", recvErr)
		return
	}
	if isConfirm {
		if sendErr := s.SendFilesData(reader, writer); sendErr != nil {
			log.PrintToErr("Failed to send files data: %s\n", sendErr)
			return
		}
	}
}

func (s *Server) ReceiveClientConfirmation(reader *bufio.Reader) (bool, error) {
	confirmPkt := new(protocol.Packet)
	_, readErr := confirmPkt.ReadAndDecode(reader)
	return confirmPkt.Is(protocol.ConfirmAccept), readErr
}

func (s *Server) SendFilesQuantityAndInformation(writer *bufio.Writer) error {
	for i := range s.listlength {
		filename, filesize, statErr := s.GetFileStat(s.filelist[i])
		if statErr != nil {
			return statErr
		}
		quantityPkt := protocol.NewPacket(protocol.FileQuantity, filename, filesize)
		if _, quantityErr := quantityPkt.EncodeAndWrite(writer); quantityErr != nil {
			return quantityErr
		}
	}
	endPkt := protocol.NewPacket(protocol.EndOfTransmission, "", 0)
	_, endErr := endPkt.EncodeAndWrite(writer)
	return endErr
}

func (s *Server) SendFilesData(reader *bufio.Reader, writer *bufio.Writer) error {
	for i := range s.listlength {
		filename, filesize, statErr := s.GetFileStat(s.filelist[i])
		if statErr != nil {
			return statErr
		}

		metaPkt := protocol.NewPacket(protocol.FileMeta, filename, 0)
		if _, metaErr := metaPkt.EncodeAndWrite(writer); metaErr != nil {
			return metaErr
		}

		respPkt := new(protocol.Packet)
		if _, decErr := respPkt.ReadAndDecode(reader); decErr != nil {
			return decErr
		}

		if respPkt.Is(protocol.RejectFile) {
			log.PrintToOut("Client rejected file: %s\n", filename)
			continue
		} else if !respPkt.Is(protocol.AcceptFile) {
			return fmt.Errorf("invalid response type: %d", respPkt.DataType)
		}

		openFile, openErr := os.OpenFile(s.filelist[i], os.O_RDONLY, 0644)
		if openErr != nil {
			return openErr
		}

		if _, copyErr := io.CopyN(writer, openFile, filesize); copyErr != nil {
			if closeErr := openFile.Close(); closeErr != nil {
				return closeErr
			}
			return copyErr
		}
		if closeErr := openFile.Close(); closeErr != nil {
			return closeErr
		}

		if flushErr := writer.Flush(); flushErr != nil {
			return flushErr
		}
	}
	return nil
}

func (s *Server) GetFileStat(path string) (string, int64, error) {
	cleanPath := filepath.Clean(path)

	filestat, statErr := os.Stat(cleanPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return "", 0, fmt.Errorf("[%s] file does not exist", cleanPath)
		}
		if os.IsPermission(statErr) {
			return "", 0, fmt.Errorf("[%s] permission denied", cleanPath)
		}
		return "", 0, statErr
	}
	if filestat.IsDir() {
		return "", 0, fmt.Errorf("[%s] can not be a folder", cleanPath)
	}
	return filepath.Base(filestat.Name()), filestat.Size(), nil
}

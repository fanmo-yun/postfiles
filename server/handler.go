package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"postfiles/protocol"
	"time"
)

func (s *Server) runServer(
	ctx context.Context,
	addr string,
	shutdownTimeout time.Duration,
) error {
	listener, listenErr := net.Listen("tcp", addr)
	if listenErr != nil {
		return listenErr
	}

	slog.Info("Server start", "address", addr)

	go func() {
		for {
			conn, acceptErr := listener.Accept()
			if acceptErr != nil {
				select {
				case <-ctx.Done():
					return
				default:
					slog.Error("Listener Accept Failed", "Error", acceptErr)
					continue
				}
			}

			slog.Info("New Connection Accessed", "connection", conn.RemoteAddr().String())
			s.connMap.Store(conn.RemoteAddr(), conn)
			s.wg.Add(1)
			go s.HandleConnection(conn)
		}
	}()
	s.Shutdown(ctx, listener, shutdownTimeout)
	return nil
}

func (s *Server) HandleConnection(conn net.Conn) {
	connAddr := conn.RemoteAddr()
	defer func() {
		_ = conn.Close()
		s.connMap.Delete(connAddr)
		s.wg.Done()
		slog.Warn("Connection is closed", "remote_addr", connAddr)
	}()

	reader := bufio.NewReaderSize(conn, 32*1024)
	writer := bufio.NewWriterSize(conn, 32*1024)

	if sendErr := s.sendInfo(writer); sendErr != nil {
		slog.Error("Failed to send file's quantity or infomation", "Error", sendErr)
		return
	}
	isConfirm, recvErr := s.recvConfirm(reader)
	if recvErr != nil {
		if errors.Is(recvErr, io.EOF) {
			return
		}
		slog.Error("Failed to receive client confirmation", "Error", recvErr)
		return
	}
	if isConfirm {
		if sendErr := s.sendAll(reader, writer); sendErr != nil {
			slog.Error("Failed to send files data", "Error", sendErr)
			return
		}
	}
}

func (s *Server) recvConfirm(reader *bufio.Reader) (bool, error) {
	confirmPkt := new(protocol.Packet)
	_, readErr := confirmPkt.ReadAndDecode(reader)
	return confirmPkt.TypeIs(protocol.ConfirmAccept), readErr
}

func (s *Server) sendInfo(writer *bufio.Writer) error {
	for _, file := range s.filelist {
		filename, filesize, statErr := s.getFileStat(file)
		if statErr != nil {
			return statErr
		}
		quantityPkt := protocol.NewPacket(protocol.FileQuantity, filename, filesize)
		if _, quantityErr := quantityPkt.EncodeAndWrite(writer); quantityErr != nil {
			return quantityErr
		}
	}
	endPkt := protocol.NewPacket(protocol.EndOfTransmission, "", 0)
	if _, endErr := endPkt.EncodeAndWrite(writer); endErr != nil {
		return endErr
	}
	return writer.Flush()
}

func (s *Server) sendAll(reader *bufio.Reader, writer *bufio.Writer) error {
	for _, file := range s.filelist {
		filename, _, statErr := s.getFileStat(file)
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

		switch {
		case respPkt.TypeIs(protocol.RejectFile):
			slog.Warn("Client rejected file", "filename", filename)
			continue
		case respPkt.TypeIs(protocol.AcceptFile):
			if sendErr := s.sendOne(writer, file); sendErr != nil {
				return sendErr
			}
		default:
			return fmt.Errorf("invalid response type: %d", respPkt.DataType)
		}

		if flushErr := writer.Flush(); flushErr != nil {
			return flushErr
		}
	}
	return nil
}

func (s *Server) sendOne(writer *bufio.Writer, filename string) error {
	openFile, openErr := os.OpenFile(filename, os.O_RDONLY, 0644)
	if openErr != nil {
		return openErr
	}
	defer openFile.Close()

	_, copyErr := io.Copy(writer, openFile)
	return copyErr
}

func (s *Server) getFileStat(path string) (string, int64, error) {
	p := filepath.Clean(path)

	info, err := os.Stat(p)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return "", 0, fmt.Errorf("%s: not exist", p)
		case os.IsPermission(err):
			return "", 0, fmt.Errorf("%s: permission denied", p)
		default:
			return "", 0, err
		}
	}

	if info.IsDir() {
		return "", 0, fmt.Errorf("%s: is a directory", p)
	}

	return filepath.Base(p), info.Size(), nil
}

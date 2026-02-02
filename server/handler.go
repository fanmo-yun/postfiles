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
	shutdownTimeout time.Duration,
) error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	slog.Info("Server start", "address", s.address)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					slog.Error("listener Accept Failed", "err", err)
					continue
				}
			}

			slog.Info("new connection accessed", "connection", conn.RemoteAddr().String())
			s.connMap.Store(conn.RemoteAddr(), conn)
			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}()
	s.Shutdown(ctx, listener, shutdownTimeout)
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	connAddr := conn.RemoteAddr()
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("connection close error", "err", err)
		}
		s.connMap.Delete(connAddr)
		s.wg.Done()
		slog.Warn("connection is closed", "remote_addr", connAddr)
	}()

	reader := bufio.NewReaderSize(conn, 32*1024)
	writer := bufio.NewWriterSize(conn, 32*1024)

	if err := s.sendInfo(writer); err != nil {
		slog.Error("failed to send file's quantity or infomation", "err", err)
		return
	}
	isConfirm, err := s.recvConfirm(reader)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}
		slog.Error("failed to receive client confirmation", "err", err)
		return
	}
	if isConfirm {
		if err := s.sendAll(reader, writer); err != nil {
			slog.Error("failed to send files data", "err", err)
			return
		}
	}
}

func (s *Server) sendInfo(writer *bufio.Writer) error {
	for _, file := range s.fileList {
		filename, filesize, err := s.getFileStat(file)
		if err != nil {
			return err
		}
		quantityPkt := protocol.NewPacket(protocol.FileQuantity, filename, filesize)
		if err := quantityPkt.EncodeAndWrite(writer); err != nil {
			return err
		}
	}
	endPkt := protocol.NewPacket(protocol.EndOfTransmission, "", 0)
	if err := endPkt.EncodeAndWrite(writer); err != nil {
		return err
	}
	return writer.Flush()
}

func (s *Server) recvConfirm(reader *bufio.Reader) (bool, error) {
	confirmPkt := new(protocol.Packet)
	err := confirmPkt.ReadAndDecode(reader)
	return confirmPkt.TypeIs(protocol.ConfirmAccept), err
}

func (s *Server) sendAll(reader *bufio.Reader, writer *bufio.Writer) error {
	for _, file := range s.fileList {
		filename, _, err := s.getFileStat(file)
		if err != nil {
			return err
		}
		metaPkt := protocol.NewPacket(protocol.FileMeta, filename, 0)
		if err := metaPkt.EncodeAndWrite(writer); err != nil {
			return err
		}
		if err := writer.Flush(); err != nil {
			return err
		}

		respPkt := new(protocol.Packet)
		if err := respPkt.ReadAndDecode(reader); err != nil {
			return err
		}
		switch {
		case respPkt.TypeIs(protocol.RejectFile):
			slog.Warn("client rejected file", "filename", filename)
			continue
		case respPkt.TypeIs(protocol.AcceptFile):
			if err := s.sendOne(writer, file); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid response type: %d", respPkt.DataType)
		}
	}
	return nil
}

func (s *Server) sendOne(writer *bufio.Writer, filename string) error {
	openFile, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer openFile.Close()

	_, err = io.Copy(writer, openFile)
	return err
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

package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"postfiles/protocol"
	"postfiles/utils"
	"sync"
	"syscall"
)

type ServerInterface interface {
	ServerRun(files []string)
	ServerStop()
	AcceptConnection(listener net.Listener, files []string)
	HandleConnection(conn net.Conn, files []string)
	WriteFileInfo(writer *bufio.Writer, files []string) error
	ReceiveClientConfirmation(reader *bufio.Reader) (bool, error)
	SendFiles(writer *bufio.Writer, files []string)
	WriteFile(writer *bufio.Writer, file string) error
}

type Server struct {
	IP         string
	Port       int
	Wg         *sync.WaitGroup
	ServerQuit chan os.Signal
}

func NewServer(IP string, Port int) *Server {
	return &Server{
		IP,
		Port,
		new(sync.WaitGroup),
		make(chan os.Signal, 1),
	}
}

func (s *Server) ServerRun(files []string) {
	listener, listenErr := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if listenErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to start listener: %s\n", listenErr)
		os.Exit(utils.ErrServer)
	}

	signal.Notify(s.ServerQuit, os.Interrupt, syscall.SIGTERM)

	go s.ServerStop(listener)
	s.AcceptConnection(listener, files)
}

func (s *Server) ServerStop(listener net.Listener) {
	<-s.ServerQuit

	fmt.Fprintf(os.Stdout, "Stopping server...\n")
	s.Wg.Wait()
	close(s.ServerQuit)
	fmt.Fprintf(os.Stdout, "Server Stopped\n")
	if closErr := listener.Close(); closErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to close listener: %v\n", closErr)
		os.Exit(utils.ErrServerClose)
	}
}

func (s *Server) AcceptConnection(listener net.Listener, files []string) {
	for {
		conn, connErr := listener.Accept()
		if connErr != nil {
			if opErr, ok := connErr.(*net.OpError); ok && opErr.Op == "accept" {
				return
			}
			fmt.Fprintf(os.Stdout, "Failed to accept connection: %s\n", connErr)
			continue
		}
		fmt.Fprintf(os.Stdout, "Connection established from %s\n", conn.RemoteAddr().String())
		s.Wg.Add(1)
		go s.HandleConnection(conn, files)
	}
}

func (s *Server) HandleConnection(conn net.Conn, files []string) {
	defer func() {
		conn.Close()
		s.Wg.Done()
	}()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	if err := s.WriteFileInfo(writer, files); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write Info: %s\n", err)
		return
	}
	isConfirm, recvErr := s.ReceiveClientConfirmation(reader)
	if recvErr != nil || !isConfirm {
		return
	}
	s.SendFiles(writer, files)
}

func (s *Server) WriteFileInfo(writer *bufio.Writer, files []string) error {
	for _, fileName := range files {
		name, size := utils.FileStat(fileName)
		info := protocol.NewDataInfo(name, size, protocol.File_Count)
		encodedInfo, encodeErr := info.Encode()
		if encodeErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to encode info: %v\n", encodeErr)
			continue
		}

		if _, writeErr := writer.Write(encodedInfo); writeErr != nil {
			return fmt.Errorf("failed to write file info for %s: %s", fileName, writeErr)
		}
		if writeErr := writer.WriteByte('\n'); writeErr != nil {
			return fmt.Errorf("failed to write newline for %s: %s", fileName, writeErr)
		}
		if flushErr := writer.Flush(); flushErr != nil {
			return fmt.Errorf("failed to flush writer for %s: %s", fileName, flushErr)
		}
	}

	endInfo := protocol.NewDataInfo("End_Of_Transmission", 0, protocol.End_Of_Transmission)
	encodedEndInfo, encodeErr := endInfo.Encode()
	if encodeErr != nil {
		return encodeErr
	}
	if _, writeErr := writer.Write(encodedEndInfo); writeErr != nil {
		return fmt.Errorf("failed to write end of transmission info: %s", writeErr)
	}
	if writeErr := writer.WriteByte('\n'); writeErr != nil {
		return fmt.Errorf("failed to write end of transmission newline: %s", writeErr)
	}
	if flushErr := writer.Flush(); flushErr != nil {
		return fmt.Errorf("failed to flush end of transmission info: %s", flushErr)
	}
	return nil
}

func (s *Server) ReceiveClientConfirmation(reader *bufio.Reader) (bool, error) {
	confirmData, readErr := reader.ReadBytes('\n')
	if readErr != nil {
		return false, fmt.Errorf("failed to read confirm info: %s", readErr)
	}
	info := new(protocol.DataInfo)
	decodeErr := info.Decode(confirmData)
	if decodeErr != nil {
		return false, decodeErr
	}
	if info.Type == protocol.Confirm_Accept {
		return true, nil
	}
	return false, nil
}

func (s *Server) SendFiles(writer *bufio.Writer, files []string) {
	for _, fileName := range files {
		if err := s.WriteFile(writer, fileName); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write file %s: %s\n", fileName, err)
			continue
		}
	}
}

func (s *Server) WriteFile(writer *bufio.Writer, file string) error {
	filename, filesize := utils.FileStat(file)
	info := protocol.NewDataInfo(filename, filesize, protocol.File_Info_Data)
	encodedInfo, encodeErr := info.Encode()
	if encodeErr != nil {
		return encodeErr
	}

	if _, writeErr := writer.Write(encodedInfo); writeErr != nil {
		return fmt.Errorf("failed to write file info: %s", writeErr)
	}
	if writeErr := writer.WriteByte('\n'); writeErr != nil {
		return fmt.Errorf("failed to write newline after file info: %s", writeErr)
	}
	if flushErr := writer.Flush(); flushErr != nil {
		return fmt.Errorf("failed to flush writer after file info: %s", flushErr)
	}

	fp, openErr := os.Open(file)
	if openErr != nil {
		return fmt.Errorf("failed to open file %s: %s", file, openErr)
	}
	defer fp.Close()

	if _, copyErr := io.CopyN(writer, fp, filesize); copyErr != nil {
		return fmt.Errorf("failed to copy file data for %s: %s", file, copyErr)
	}
	if flushErr := writer.Flush(); flushErr != nil {
		return fmt.Errorf("failed to flush writer after file data: %s", flushErr)
	}
	return nil
}

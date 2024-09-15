package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"postfiles/datainfo"
	"postfiles/exitcodes"
	"postfiles/utils"
	"syscall"
)

type Server struct {
	IP   string
	Port int
}

func NewServer(IP string, Port int) *Server {
	return &Server{IP, Port}
}

func (s Server) ServerRun(files []string) {
	listener, listenErr := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if listenErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to start listener: %s\n", listenErr)
		os.Exit(exitcodes.ErrServer)
	}
	defer listener.Close()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	done := make(chan bool)

	go func() {
		<-quit
		fmt.Println("Shutting down server...")
		listener.Close()
		done <- true
	}()

	for {
		select {
		case <-done:
			fmt.Println("Server stopped")
			return
		default:
			s.acceptConnection(listener, files)
		}
	}
}

func (s Server) acceptConnection(listener net.Listener, files []string) {
	conn, connErr := listener.Accept()
	if connErr != nil {
		if opErr, ok := connErr.(*net.OpError); ok && opErr.Op == "accept" {
			return
		}
		fmt.Fprintf(os.Stdout, "Failed to accept connection: %s\n", connErr)
		return
	}
	fmt.Fprintf(os.Stdout, "Connection established from %s\n", conn.RemoteAddr().String())
	go s.serverHandler(conn, files)
}

func (s Server) serverHandler(conn net.Conn, fileList []string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	if err := s.serverWriteAllInfo(writer, fileList); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write Info: %s\n", err)
		return
	}
	isConfirm, recvErr := s.recvClientConfirm(reader)
	if recvErr != nil || !isConfirm {
		return
	}
	s.sendFilesToClient(writer, fileList)
}

func (s Server) serverWriteAllInfo(w *bufio.Writer, fileList []string) error {
	for _, fileName := range fileList {
		name, size := utils.FileStat(fileName)
		info := datainfo.NewInfo(name, size, datainfo.File_Count)
		encodedInfo := datainfo.EncodeJSON(info)

		if _, writeErr := w.Write(encodedInfo); writeErr != nil {
			return fmt.Errorf("failed to write file info for %s: %s", fileName, writeErr)
		}
		if writeErr := w.WriteByte('\n'); writeErr != nil {
			return fmt.Errorf("failed to write newline for %s: %s", fileName, writeErr)
		}
		if flushErr := w.Flush(); flushErr != nil {
			return fmt.Errorf("failed to flush writer for %s: %s", fileName, flushErr)
		}
	}

	endInfo := datainfo.NewInfo("End_Of_Transmission", 0, datainfo.End_Of_Transmission)
	encodedEndInfo := datainfo.EncodeJSON(endInfo)
	if _, writeErr := w.Write(encodedEndInfo); writeErr != nil {
		return fmt.Errorf("failed to write end of transmission info: %s", writeErr)
	}
	if writeErr := w.WriteByte('\n'); writeErr != nil {
		return fmt.Errorf("failed to write end of transmission newline: %s", writeErr)
	}
	if flushErr := w.Flush(); flushErr != nil {
		return fmt.Errorf("failed to flush end of transmission info: %s", flushErr)
	}
	return nil
}

func (s Server) recvClientConfirm(r *bufio.Reader) (bool, error) {
	confirmData, readErr := r.ReadBytes('\n')
	if readErr != nil {
		return false, fmt.Errorf("failed to read confirm info: %s", readErr)
	}
	info := datainfo.DecodeJSON(confirmData)
	if info.Type == datainfo.Confirm_Accept {
		return true, nil
	}
	return false, nil
}

func (s Server) sendFilesToClient(w *bufio.Writer, fileList []string) {
	for _, fileName := range fileList {
		if err := s.serverWriteHandler(w, fileName); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write file %s: %s\n", fileName, err)
			continue
		}
	}
}

func (s Server) serverWriteHandler(w *bufio.Writer, file string) error {
	filename, filesize := utils.FileStat(file)
	info := datainfo.NewInfo(filename, filesize, datainfo.File_Info_Data)
	encodedInfo := datainfo.EncodeJSON(info)

	if _, writeErr := w.Write(encodedInfo); writeErr != nil {
		return fmt.Errorf("failed to write file info: %s", writeErr)
	}
	if writeErr := w.WriteByte('\n'); writeErr != nil {
		return fmt.Errorf("failed to write newline after file info: %s", writeErr)
	}
	if flushErr := w.Flush(); flushErr != nil {
		return fmt.Errorf("failed to flush writer after file info: %s", flushErr)
	}

	fp, openErr := os.Open(file)
	if openErr != nil {
		return fmt.Errorf("failed to open file %s: %s", file, openErr)
	}
	defer fp.Close()

	if _, copyErr := io.CopyN(w, fp, filesize); copyErr != nil {
		return fmt.Errorf("failed to copy file data for %s: %s", file, copyErr)
	}
	if flushErr := w.Flush(); flushErr != nil {
		return fmt.Errorf("failed to flush writer after file data: %s", flushErr)
	}
	return nil
}

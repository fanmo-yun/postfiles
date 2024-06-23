package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"postfiles/fileinfo"
	"postfiles/utils"
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
		fmt.Fprintf(os.Stderr, "Failed to start listener: %v\n", listenErr)
		os.Exit(1)
	}

	for {
		conn, connErr := listener.Accept()
		if connErr != nil {
			fmt.Fprintf(os.Stdout, "Failed to accept connection: %v\n", connErr)
			continue
		}
		fmt.Fprintf(os.Stdout, "Connection established from %s\n", conn.RemoteAddr().String())
		go s.serverHandler(conn, files)
	}
}

func (s Server) serverHandler(conn net.Conn, fileList []string) {
	defer conn.Close()
	writer := bufio.NewWriter(conn)

	if err := s.serverWriteAllInfo(writer, fileList); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write Info: %v\n", err)
		return
	}
	s.sendFilesToClient(writer, fileList)
}

func (s Server) serverWriteAllInfo(w *bufio.Writer, fileList []string) error {
	if writeErr := w.WriteByte(fileinfo.File_Count); writeErr != nil {
		return fmt.Errorf("failed to write file count byte: %w", writeErr)
	}

	for _, fv := range fileList {
		fileInfo := fileinfo.NewInfo(utils.FileStat(fv))
		encodedInfo := fileinfo.EncodeJSON(fileInfo)
		if _, writeErr := w.Write(encodedInfo); writeErr != nil {
			return fmt.Errorf("failed to write file info for %s: %w", fv, writeErr)
		}
		if writeErr := w.WriteByte('\n'); writeErr != nil {
			return fmt.Errorf("failed to write newline for %s: %w", fv, writeErr)
		}
		if flushErr := w.Flush(); flushErr != nil {
			return fmt.Errorf("failed to flush writer for %s: %w", fv, flushErr)
		}
	}

	endInfo := fileinfo.NewInfo("END_OF_TRANSMISSION", -1)
	encodedEndInfo := fileinfo.EncodeJSON(endInfo)
	if _, writeErr := w.Write(encodedEndInfo); writeErr != nil {
		return fmt.Errorf("failed to write end of transmission info: %w", writeErr)
	}
	if writeErr := w.WriteByte('\n'); writeErr != nil {
		return fmt.Errorf("failed to write end of transmission newline: %w", writeErr)
	}
	if flushErr := w.Flush(); flushErr != nil {
		return fmt.Errorf("failed to flush end of transmission info: %w", flushErr)
	}
	return nil
}

func (s Server) sendFilesToClient(w *bufio.Writer, fileList []string) {
	for _, fv := range fileList {
		if err := s.serverWriteHandler(w, fv); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write file %s: %v\n", fv, err)
			continue
		}
	}
}

func (s Server) serverWriteHandler(w *bufio.Writer, file string) error {
	filename, filesize := utils.FileStat(file)
	fileInfo := fileinfo.NewInfo(filename, filesize)
	encodedInfo := fileinfo.EncodeJSON(fileInfo)

	if writeErr := w.WriteByte(fileinfo.File_Info_Data); writeErr != nil {
		return fmt.Errorf("failed to write file info byte: %w", writeErr)
	}
	if _, writeErr := w.Write(encodedInfo); writeErr != nil {
		return fmt.Errorf("failed to write file info: %w", writeErr)
	}
	if writeErr := w.WriteByte('\n'); writeErr != nil {
		return fmt.Errorf("failed to write newline after file info: %w", writeErr)
	}
	if flushErr := w.Flush(); flushErr != nil {
		return fmt.Errorf("failed to flush writer after file info: %w", flushErr)
	}

	fp, openErr := os.Open(file)
	if openErr != nil {
		return fmt.Errorf("failed to open file %s: %w", file, openErr)
	}
	defer fp.Close()

	if _, copyErr := io.CopyN(w, fp, filesize); copyErr != nil {
		return fmt.Errorf("failed to copy file data for %s: %w", file, copyErr)
	}
	if flushErr := w.Flush(); flushErr != nil {
		return fmt.Errorf("failed to flush writer after file data: %w", flushErr)
	}
	return nil
}

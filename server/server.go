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
		os.Exit(1)
	}
	s.sendFilesToClient(writer, fileList)
}

func (s Server) serverWriteAllInfo(w *bufio.Writer, fileList []string) error {
	if writeErr := w.WriteByte(fileinfo.File_Count); writeErr != nil {
		return writeErr
	}

	for _, fv := range fileList {
		if _, writeErr := w.Write(fileinfo.EncodeJSON(fileinfo.NewInfo(utils.FileStat(fv)))); writeErr != nil {
			return writeErr
		}

		if writeErr := w.WriteByte('\n'); writeErr != nil {
			return writeErr
		}
	}

	if _, writeErr := w.Write(fileinfo.EncodeJSON(fileinfo.NewInfo("END_OF_TRANSMISSION", -1))); writeErr != nil {
		return writeErr
	}

	if writeErr := w.WriteByte('\n'); writeErr != nil {
		return writeErr
	}

	if flushErr := w.Flush(); flushErr != nil {
		return flushErr
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

	if writeErr := w.WriteByte(fileinfo.File_Info); writeErr != nil {
		return writeErr
	}

	if _, writeErr := w.Write(fileinfo.EncodeJSON(fileinfo.NewInfo(filename, filesize))); writeErr != nil {
		return writeErr
	}

	if writeErr := w.WriteByte('\n'); writeErr != nil {
		return writeErr
	}

	if flushErr := w.Flush(); flushErr != nil {
		return flushErr
	}

	fp, openErr := os.Open(file)
	if openErr != nil {
		return openErr
	}
	defer fp.Close()

	if writeErr := w.WriteByte(fileinfo.File_Data); writeErr != nil {
		return writeErr
	}

	if flushErr := w.Flush(); flushErr != nil {
		return flushErr
	}

	if _, copyErr := io.CopyN(w, fp, filesize); copyErr != nil {
		return copyErr
	}

	if flushErr := w.Flush(); flushErr != nil {
		return flushErr
	}

	return nil
}

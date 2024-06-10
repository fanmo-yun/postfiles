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

func (server Server) ServerRun(files []string) {
	listener, listenErr := net.Listen("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port))
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
		go server.serverhandler(conn, &files)
	}
}

func (server Server) serverhandler(conn net.Conn, fileList *[]string) {
	defer conn.Close()

	writer := bufio.NewWriter(conn)

	server.sendfiles(writer, fileList)
}

func (server *Server) sendfiles(w *bufio.Writer, fileList *[]string) {
	for _, fv := range *fileList {
		if err := server.serverwritehandler(w, fv); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write file %s: %v\n", fv, err)
			continue
		}
	}
}

func (server *Server) serverwritehandler(writer *bufio.Writer, file string) error {
	filename, filesize := utils.FileStat(file)

	if writeErr := writer.WriteByte(fileinfo.File_Info); writeErr != nil {
		return writeErr
	}

	if _, writeErr := writer.Write(fileinfo.EncodeJSON(fileinfo.NewInfo(filename, filesize))); writeErr != nil {
		return writeErr
	}

	if writeErr := writer.WriteByte('\n'); writeErr != nil {
		return writeErr
	}

	if flushErr := writer.Flush(); flushErr != nil {
		return flushErr
	}

	fp, openErr := os.Open(file)
	if openErr != nil {
		return openErr
	}
	defer fp.Close()

	if writeErr := writer.WriteByte(fileinfo.File_Data); writeErr != nil {
		return writeErr
	}

	if flushErr := writer.Flush(); flushErr != nil {
		return flushErr
	}

	if _, copyErr := io.CopyN(writer, fp, filesize); copyErr != nil {
		return copyErr
	}

	if flushErr := writer.Flush(); flushErr != nil {
		return flushErr
	}

	return nil
}

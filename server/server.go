package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"postfiles/api"

	"github.com/sirupsen/logrus"
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
		logrus.Fatal(listenErr)
	}

	for {
		conn, connErr := listener.Accept()
		if connErr != nil {
			logrus.Warn(connErr)
			continue
		}
		logrus.Infof("%s is come", conn.RemoteAddr().String())
		go server.serverhandler(conn, &files)
	}
}

func (server Server) serverhandler(conn net.Conn, fileList *[]string) {
	defer conn.Close()

	writer := bufio.NewWriter(conn)

	for _, value := range *fileList {
		if err := server.serverwritehandler(writer, value); err != nil {
			logrus.Fatal(err)
		}
	}
}

func (server *Server) serverwritehandler(writer *bufio.Writer, file string) error {
	filename, filesize := api.FileStat(file)

	if writeErr := writer.WriteByte(0); writeErr != nil {
		return writeErr
	}

	if _, writeErr := writer.Write(api.EncodeJSON(api.NewInfo(filename, filesize))); writeErr != nil {
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

	if writeErr := writer.WriteByte(1); writeErr != nil {
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

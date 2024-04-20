package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"postfiles/api"
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
		log.Fatal(listenErr)
	}

	for {
		conn, connErr := listener.Accept()
		if connErr != nil {
			log.Panic(connErr)
			continue
		}
		go server.serverhandler(conn, &files)
	}
}

func (server Server) serverhandler(conn net.Conn, fileList *[]string) {
	defer conn.Close()

	writer := bufio.NewWriter(conn)

	for _, value := range *fileList {
		filename, filesize := api.FileStat(value)
		_, writeErr := writer.Write(api.EncodeJSON(api.NewInfo(filename, filesize)))
		if writeErr == io.EOF {
			break
		} else if writeErr != nil {
			log.Fatal(writeErr)
		}

		if flushErr := writer.Flush(); flushErr != nil {
			log.Fatal(flushErr)
		}

	}
}

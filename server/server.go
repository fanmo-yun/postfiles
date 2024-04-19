package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	IP   string
	Port int
}

func NewServer(IP string, Port int) *Server {
	return &Server{IP, Port}
}

func (server Server) ServerRun() {
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
		go server.ServerHandler(conn)
	}
}

func (server Server) ServerHandler(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		var buf [1024]byte
		size, readErr := reader.Read(buf[:])
		if readErr == io.EOF {
			break
		} else if readErr != nil {
			log.Panicln(readErr)
			break
		}
		fmt.Println(string(buf[:size]))
		writer.Write(buf[:size])
		writer.Flush()
	}
}

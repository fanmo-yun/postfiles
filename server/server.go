package server

import (
	"fmt"
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
		go server.ServerHandler(conn)
	}
}

func (server Server) ServerHandler(conn net.Conn) {
	defer conn.Close()

}

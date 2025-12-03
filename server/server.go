package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Server struct {
	ip       string
	port     int
	filelist []string
	connMap  *sync.Map
	wg       *sync.WaitGroup
}

func NewServer(ip string, port int, filelist []string) *Server {
	return &Server{
		ip:       ip,
		port:     port,
		filelist: filelist,
		connMap:  new(sync.Map),
		wg:       new(sync.WaitGroup),
	}
}

func (s *Server) Start() error {
	address := net.JoinHostPort(s.ip, fmt.Sprintf("%d", s.port))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	return s.runServer(ctx, address, time.Second*5)
}

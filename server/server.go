package server

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Server struct {
	address  string
	fileList []string
	connMap  *sync.Map
	wg       *sync.WaitGroup
}

func NewServer(address string, filelist []string) *Server {
	return &Server{
		address:  address,
		fileList: filelist,
		connMap:  new(sync.Map),
		wg:       new(sync.WaitGroup),
	}
}

func (s *Server) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	return s.runServer(ctx, time.Second*5)
}

package server

import (
	"context"
	"log/slog"
	"net"
	"time"
)

func (s *Server) Shutdown(
	ctx context.Context,
	listener net.Listener,
	shutdownTimeout time.Duration,
) {
	<-ctx.Done()
	slog.Warn("Shutting down server...")

	if err := listener.Close(); err != nil {
		slog.Error("server close error", "err", err)
	}
	slog.Warn("Stopping server...")

	s.connMap.Range(func(_, v any) bool {
		conn := v.(net.Conn)
		if err := conn.Close(); err != nil {
			slog.Error("conn close error", "err", err)
		}
		return true
	})

	c := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		slog.Warn("All connections closed, server stopped gracefully.")
	case <-time.After(shutdownTimeout):
		slog.Warn("Shutdown timeout reached, forcing exit.")
	}
}

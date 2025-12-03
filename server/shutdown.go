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

	_ = listener.Close()
	slog.Warn("Stopping server...")

	s.connMap.Range(func(_, v any) bool {
		conn := v.(net.Conn)
		_ = conn.Close()
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

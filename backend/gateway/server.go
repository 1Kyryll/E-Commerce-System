package gateway

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Server is a thin wrapper around *http.Server that owns the run loop and
// graceful-shutdown handshake. It is intentionally minimal — TLS, custom
// timeouts, and HTTP/2 tuning are deferred until they have a reason to exist.
type Server struct {
	httpServer *http.Server
}

// NewServer builds a Server that will serve handler over an externally-provided
// net.Listener (see Serve). ReadHeaderTimeout is set defensively to avoid
// Slowloris-style attacks.
func NewServer(handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

// Serve runs the HTTP server on ln until ctx is cancelled. When ctx is
// cancelled, it calls Shutdown with shutdownTimeout. Returns the first error
// encountered, or nil for a clean shutdown.
func (s *Server) Serve(ctx context.Context, ln net.Listener, shutdownTimeout time.Duration) error {
	serveErr := make(chan error, 1)
	go func() {
		slog.Info("gateway listening", "addr", ln.Addr().String())
		if err := s.httpServer.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		slog.Info("gateway shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	}
}

// ListenAndServe is a convenience that opens a TCP listener on addr and then
// calls Serve.
func (s *Server) ListenAndServe(ctx context.Context, addr string, shutdownTimeout time.Duration) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.Serve(ctx, ln, shutdownTimeout)
}

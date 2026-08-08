package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 5 * time.Minute
	shutdownTimeout   = 10 * time.Second
)

// Server publishes the compiled list over HTTP for AdGuard Home to fetch.
type Server struct {
	http *http.Server
}

// New serves the directory holding outputFile, so the list is reachable at
// http://<addr>/<basename>.
func New(addr, outputFile string) *Server {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(filepath.Dir(outputFile))))

	return &Server{
		http: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
		},
	}
}

// Start serves until ctx is done, then shuts down gracefully. It returns once
// the listener is closed.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		slog.LogAttrs(ctx, slog.LevelInfo, "serving compiled list", slog.String("addr", s.http.Addr))

		err := s.http.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}

		errCh <- err
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(stopCtx); err != nil {
		return fmt.Errorf("shutting down list server: %w", err)
	}

	return <-errCh
}

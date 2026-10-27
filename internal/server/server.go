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

// New serves outputFile alone at http://<addr>/<basename>. Nothing else in its
// directory, such as in-progress temp files, is exposed.
func New(addr, outputFile string) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /"+filepath.Base(outputFile), func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, outputFile)
	})

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

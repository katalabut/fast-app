// Package service provides built-in services for FastApp applications,
// including debug and monitoring capabilities.
package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/katalabut/fast-app/logger"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// DebugServer contains configuration for the debug server.
type DebugServer struct {
	// Port specifies the HTTP port for debug endpoints (metrics, pprof, etc.)
	Port int `default:"9090"`
}

// DefaultDebugService provides a debug HTTP server with metrics and profiling endpoints.
// It exposes Prometheus metrics at /metrics and Go pprof endpoints for debugging.
type DefaultDebugService struct {
	cfg    DebugServer
	server http.Server
}

// NewDefaultDebugService creates a new debug service with the given configuration.
// The service will start an HTTP server on the configured port with metrics
// and profiling endpoints.
func NewDefaultDebugService(cfg DebugServer) *DefaultDebugService {
	return &DefaultDebugService{
		cfg: cfg,
	}
}

// Shutdown gracefully stops the debug server within the given context timeout.
func (s *DefaultDebugService) Shutdown(ctx context.Context) error {
	logger.InfoKV(ctx, "Shutting down debug server")
	return s.server.Shutdown(ctx)
}

// Run starts the debug HTTP server and blocks until the context is cancelled.
// The server provides the following endpoints:
//   - /metrics - Prometheus metrics endpoint
//   - /debug/pprof/* - Go profiling endpoints (automatically registered)
func (s *DefaultDebugService) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	s.server = http.Server{
		Addr:        fmt.Sprintf(":%d", s.cfg.Port),
		ReadTimeout: 5 * time.Second,
		IdleTimeout: 120 * time.Second,
		Handler:     mux,
	}

	logger.InfoKV(ctx, "Debug server is running", "address", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "failed to start debug server")
	}

	return nil
}

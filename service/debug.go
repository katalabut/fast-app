// Package service provides built-in services for FastApp applications,
// including debug and monitoring capabilities.
package service

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/katalabut/fast-app/config"
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
	cfg          DebugServer
	server       *http.Server
	started      atomic.Bool
	pprofEnabled bool
}

// NewDefaultDebugService creates a new debug service with the given configuration.
// The service will start an HTTP server on the configured port with metrics
// and profiling endpoints.
func NewDefaultDebugService(cfg DebugServer) *DefaultDebugService {
	return &DefaultDebugService{
		cfg:    cfg,
		server: &http.Server{Addr: httpAddress("", cfg.Port), ReadTimeout: 5 * time.Second, IdleTimeout: 120 * time.Second},
	}
}

// Shutdown gracefully stops the debug server within the given context timeout.
func (s *DefaultDebugService) Shutdown(ctx context.Context) error {
	logger.InfoKV(ctx, "Shutting down debug server")
	return s.server.Shutdown(ctx)
}

// Run starts the debug HTTP server once and waits for Shutdown; an already cancelled context prevents startup.
// The server provides the following endpoints:
//   - /metrics - Prometheus metrics endpoint
//   - /debug/pprof/* - Go profiling endpoints (explicit WithPprof opt-in)
func (s *DefaultDebugService) Run(ctx context.Context) error {
	if ctx.Err() != nil {
		return nil
	}
	if !s.started.CompareAndSwap(false, true) {
		return errors.New("debug service already started")
	}
	if err := (config.Observability{Enabled: true, Port: s.cfg.Port}).Validate(); err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	if s.pprofEnabled {
		registerPprof(mux, "/debug")
	}
	s.server.Handler = mux

	logger.InfoKV(ctx, "Debug server is running", "address", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "failed to start debug server")
	}

	return nil
}

// WithListenAddress changes the bind address before Run; the default remains :port.
func (s *DefaultDebugService) WithListenAddress(address string) *DefaultDebugService {
	s.server.Addr = address
	return s
}

// WithPprof explicitly enables pprof on the private mux, with a 35-second write budget.
func (s *DefaultDebugService) WithPprof() *DefaultDebugService {
	s.pprofEnabled = true
	s.server.WriteTimeout = 35 * time.Second
	return s
}

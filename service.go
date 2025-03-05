package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/katalabut/fast-app/logger"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type DefaultDebugService struct {
	cfg    DebugServer
	server http.Server
}

func NewDefaultDebugService(cfg DebugServer) *DefaultDebugService {
	return &DefaultDebugService{
		cfg: cfg,
	}
}

func (s *DefaultDebugService) Shutdown(ctx context.Context) error {
	logger.InfoKV(ctx, "Shutting down debug server")
	return s.server.Shutdown(ctx)
}

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

// Package service provides built-in services for FastApp applications.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof" // Register pprof handlers
	"time"

	"github.com/katalabut/fast-app/config"
	"github.com/katalabut/fast-app/health"
	"github.com/katalabut/fast-app/logger"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ObservabilityService provides a unified HTTP server for metrics, health checks, and debugging.
// It combines all observability endpoints on a single port to reduce resource usage.
type ObservabilityService struct {
	config        config.Observability
	healthManager *health.Manager
	server        *http.Server
}

// NewObservabilityService creates a new observability service with the given configuration.
// The service provides:
//   - Prometheus metrics at /metrics
//   - Health check endpoints at /health/*
//   - Go pprof debugging endpoints at /debug/pprof/*
func NewObservabilityService(cfg config.Observability, healthManager *health.Manager) *ObservabilityService {
	return &ObservabilityService{
		config:        cfg,
		healthManager: healthManager,
	}
}

// Run starts the observability HTTP server and blocks until the context is cancelled.
func (s *ObservabilityService) Run(ctx context.Context) error {
	if !s.config.Enabled {
		logger.Info(ctx, "Observability server is disabled")
		<-ctx.Done()
		return nil
	}

	mux := http.NewServeMux()

	// Register metrics endpoint
	if s.config.Metrics.Enabled {
		mux.Handle(s.config.Metrics.Path, promhttp.Handler())
		logger.InfoKV(ctx, "Registered metrics endpoint", "path", s.config.Metrics.Path)
	}

	// Register health check endpoints
	if s.config.Health.Enabled {
		s.registerHealthEndpoints(mux)
	}

	// Register debug endpoints (pprof is automatically registered via import)
	if s.config.Debug.Enabled {
		s.registerDebugEndpoints(mux)
	}

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.InfoKV(ctx, "Starting observability server",
		"address", s.server.Addr,
		"metrics_enabled", s.config.Metrics.Enabled,
		"health_enabled", s.config.Health.Enabled,
		"debug_enabled", s.config.Debug.Enabled,
	)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "failed to start observability server")
	}

	return nil
}

// Shutdown gracefully stops the observability server within the given context timeout.
func (s *ObservabilityService) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	logger.InfoKV(ctx, "Shutting down observability server")
	return s.server.Shutdown(ctx)
}

// registerHealthEndpoints registers all health check endpoints.
func (s *ObservabilityService) registerHealthEndpoints(mux *http.ServeMux) {
	// Liveness endpoint - always returns 200 if the process is alive
	mux.HandleFunc(s.config.Health.LivePath, s.handleLiveness)

	// Readiness endpoint - returns 200 if the application is ready to serve traffic
	mux.HandleFunc(s.config.Health.ReadyPath, s.handleReadiness)

	// Detailed health checks endpoint
	mux.HandleFunc(s.config.Health.CheckPath, s.handleHealthChecks)

	logger.InfoKV(context.Background(), "Registered health endpoints",
		"live_path", s.config.Health.LivePath,
		"ready_path", s.config.Health.ReadyPath,
		"check_path", s.config.Health.CheckPath,
	)
}

// registerDebugEndpoints registers debug and profiling endpoints.
func (s *ObservabilityService) registerDebugEndpoints(mux *http.ServeMux) {
	// pprof endpoints are automatically registered via the import
	// We just need to proxy them under our debug prefix if it's not the default
	if s.config.Debug.PathPrefix != "/debug" {
		// Redirect pprof endpoints to our custom prefix
		mux.HandleFunc(s.config.Debug.PathPrefix+"/pprof/", func(w http.ResponseWriter, r *http.Request) {
			// Remove our prefix and forward to default pprof handler
			r.URL.Path = "/debug" + r.URL.Path[len(s.config.Debug.PathPrefix):]
			http.DefaultServeMux.ServeHTTP(w, r)
		})
	}

	logger.InfoKV(context.Background(), "Registered debug endpoints",
		"path_prefix", s.config.Debug.PathPrefix,
	)
}

// handleLiveness handles liveness probe requests.
func (s *ObservabilityService) handleLiveness(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleReadiness handles readiness probe requests.
func (s *ObservabilityService) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.config.Health.Timeout)
	defer cancel()

	ready := s.healthManager.IsReady()
	overallStatus := s.healthManager.GetOverallStatus(ctx)

	response := map[string]interface{}{
		"status":         overallStatus,
		"ready":          ready && overallStatus == health.StatusHealthy,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"manager_ready":  ready,
		"overall_status": overallStatus,
	}

	statusCode := http.StatusOK
	if !ready || overallStatus != health.StatusHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// handleHealthChecks handles detailed health check requests.
func (s *ObservabilityService) handleHealthChecks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.config.Health.Timeout)
	defer cancel()

	results := s.healthManager.CheckAll(ctx)
	overallStatus := s.healthManager.GetOverallStatus(ctx)

	response := map[string]interface{}{
		"status":     overallStatus,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"checks":     results,
		"ready":      s.healthManager.IsReady(),
		"check_count": len(results),
	}

	statusCode := http.StatusOK
	if overallStatus != health.StatusHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

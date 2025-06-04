package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/katalabut/fast-app/health"
	"github.com/katalabut/fast-app/logger"
)

// Config contains configuration for the health server
type Config struct {
	Enabled   bool          `default:"true"`
	Port      int           `default:"8080"`
	LivePath  string        `default:"/health/live"`
	ReadyPath string        `default:"/health/ready"`
	CheckPath string        `default:"/health/checks"`
	Timeout   time.Duration `default:"30s"`
}

// Server provides HTTP endpoints for health checks
type Server struct {
	config  Config
	manager *health.Manager
	server  *http.Server
}

// NewServer creates a new health server
func NewServer(config Config, manager *health.Manager) *Server {
	return &Server{
		config:  config,
		manager: manager,
	}
}

// Start starts the health server
func (s *Server) Start(ctx context.Context) error {
	if !s.config.Enabled {
		logger.Info(ctx, "Health server is disabled")
		return nil
	}

	mux := http.NewServeMux()
	
	// Register health endpoints
	mux.HandleFunc(s.config.LivePath, s.handleLiveness)
	mux.HandleFunc(s.config.ReadyPath, s.handleReadiness)
	mux.HandleFunc(s.config.CheckPath, s.handleChecks)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      mux,
		ReadTimeout:  s.config.Timeout,
		WriteTimeout: s.config.Timeout,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info(ctx, "Starting health server", 
		"port", s.config.Port,
		"live_path", s.config.LivePath,
		"ready_path", s.config.ReadyPath,
		"checks_path", s.config.CheckPath)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("health server failed: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the health server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	logger.Info(ctx, "Shutting down health server")
	return s.server.Shutdown(ctx)
}

// handleLiveness handles liveness probe requests
func (s *Server) handleLiveness(w http.ResponseWriter, r *http.Request) {
	// Liveness probe should always return 200 if the process is running
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleReadiness handles readiness probe requests
func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.config.Timeout)
	defer cancel()

	isReady := s.manager.IsReady()
	overallStatus := s.manager.GetOverallStatus(ctx)

	// Ready if manager says ready AND overall status is not unhealthy
	ready := isReady && overallStatus != health.StatusUnhealthy

	response := map[string]interface{}{
		"status":         overallStatus,
		"ready":          ready,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"manager_ready":  isReady,
		"overall_status": overallStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	
	if ready {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	json.NewEncoder(w).Encode(response)
}

// handleChecks handles detailed health checks requests
func (s *Server) handleChecks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.config.Timeout)
	defer cancel()

	start := time.Now()
	results := s.manager.CheckAll(ctx)
	overallStatus := s.manager.GetOverallStatus(ctx)
	duration := time.Since(start)

	response := map[string]interface{}{
		"status":    overallStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"duration":  duration.String(),
		"checks":    results,
		"ready":     s.manager.IsReady(),
	}

	w.Header().Set("Content-Type", "application/json")
	
	// Return appropriate status code based on overall health
	switch overallStatus {
	case health.StatusHealthy:
		w.WriteHeader(http.StatusOK)
	case health.StatusDegraded:
		w.WriteHeader(http.StatusOK) // Still OK, but degraded
	case health.StatusUnhealthy:
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	
	json.NewEncoder(w).Encode(response)
}

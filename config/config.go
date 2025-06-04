// Package config provides centralized configuration structures for FastApp.
// All configuration types are organized here for better maintainability.
package config

import (
	"time"
)

// App holds the main configuration for a FastApp application.
// It includes all subsystem configurations in a centralized location.
type App struct {
	// Logger configuration for structured logging
	Logger Logger `json:"logger" yaml:"logger"`

	// AutoMaxProcs automatically configures GOMAXPROCS based on container limits
	AutoMaxProcs AutoMaxProcs `json:"auto_max_procs" yaml:"auto_max_procs"`

	// Observability contains configuration for metrics, health checks, and debugging
	Observability Observability `json:"observability" yaml:"observability"`
}

// Logger contains configuration for the structured logging system.
type Logger struct {
	// AppName is included in all log entries to identify the application
	AppName string `json:"app_name" yaml:"app_name" default:"fastapp"`

	// Level sets the minimum log level (debug, info, warn, error, fatal)
	Level string `json:"level" yaml:"level" default:"info"`

	// DevMode enables development-friendly console output with colors
	DevMode bool `json:"dev_mode" yaml:"dev_mode" default:"false"`

	// MessageKey is the JSON key for the log message
	MessageKey string `json:"message_key" yaml:"message_key" default:"message"`

	// LevelKey is the JSON key for the log level
	LevelKey string `json:"level_key" yaml:"level_key" default:"severity"`

	// TimeKey is the JSON key for the timestamp
	TimeKey string `json:"time_key" yaml:"time_key" default:"timestamp"`
}

// AutoMaxProcs contains configuration for automatic GOMAXPROCS setup.
type AutoMaxProcs struct {
	// Enabled determines if automatic GOMAXPROCS configuration is active
	Enabled bool `json:"enabled" yaml:"enabled" default:"true"`

	// Min sets the minimum number of processors to use
	Min int `json:"min" yaml:"min" default:"1"`
}

// Observability contains configuration for the unified observability server.
// This server provides metrics, health checks, and debugging endpoints on a single port.
type Observability struct {
	// Enabled determines if the observability server should be started
	Enabled bool `json:"enabled" yaml:"enabled" default:"true"`

	// Port specifies the HTTP port for all observability endpoints
	Port int `json:"port" yaml:"port" default:"9090"`

	// Metrics configuration for Prometheus metrics
	Metrics Metrics `json:"metrics" yaml:"metrics"`

	// Health configuration for health check endpoints
	Health Health `json:"health" yaml:"health"`

	// Debug configuration for debugging and profiling endpoints
	Debug Debug `json:"debug" yaml:"debug"`
}

// Metrics contains configuration for Prometheus metrics.
type Metrics struct {
	// Enabled determines if metrics endpoint should be available
	Enabled bool `json:"enabled" yaml:"enabled" default:"true"`

	// Path is the URL path for metrics endpoint
	Path string `json:"path" yaml:"path" default:"/metrics"`
}

// Health contains configuration for health check endpoints.
type Health struct {
	// Enabled determines if health check endpoints should be available
	Enabled bool `json:"enabled" yaml:"enabled" default:"true"`

	// LivePath is the URL path for liveness probe endpoint
	LivePath string `json:"live_path" yaml:"live_path" default:"/health/live"`

	// ReadyPath is the URL path for readiness probe endpoint
	ReadyPath string `json:"ready_path" yaml:"ready_path" default:"/health/ready"`

	// CheckPath is the URL path for detailed health check information
	CheckPath string `json:"check_path" yaml:"check_path" default:"/health/checks"`

	// Timeout is the maximum time to wait for health checks to complete
	Timeout time.Duration `json:"timeout" yaml:"timeout" default:"30s"`

	// CacheTTL is how long to cache health check results
	CacheTTL time.Duration `json:"cache_ttl" yaml:"cache_ttl" default:"5s"`
}

// Debug contains configuration for debugging and profiling endpoints.
type Debug struct {
	// Enabled determines if debug endpoints should be available
	Enabled bool `json:"enabled" yaml:"enabled" default:"true"`

	// PathPrefix is the URL path prefix for debug endpoints
	PathPrefix string `json:"path_prefix" yaml:"path_prefix" default:"/debug"`
}

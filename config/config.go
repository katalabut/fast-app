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
	Logger Logger

	// AutoMaxProcs automatically configures GOMAXPROCS based on container limits
	AutoMaxProcs AutoMaxProcs

	// Observability contains configuration for metrics, health checks, and debugging
	Observability Observability
}

// Logger contains configuration for the structured logging system.
type Logger struct {
	// AppName is included in all log entries to identify the application
	AppName string `default:"fastapp"`

	// Level sets the minimum log level (debug, info, warn, error, fatal)
	Level string `default:"info"`

	// DevMode enables development-friendly console output with colors
	DevMode bool `default:"false"`

	// MessageKey is the JSON key for the log message
	MessageKey string `default:"message"`

	// LevelKey is the JSON key for the log level
	LevelKey string `default:"severity"`

	// TimeKey is the JSON key for the timestamp
	TimeKey string `default:"timestamp"`
}

// AutoMaxProcs contains configuration for automatic GOMAXPROCS setup.
type AutoMaxProcs struct {
	// Enabled determines if automatic GOMAXPROCS configuration is active
	Enabled bool `default:"true"`

	// Min sets the minimum number of processors to use
	Min int `default:"1"`
}

// Observability contains configuration for the unified observability server.
// This server provides metrics, health checks, and debugging endpoints on a single port.
type Observability struct {
	// Enabled determines if the observability server should be started
	Enabled bool `default:"true"`

	// Port specifies the HTTP port for all observability endpoints
	Port int `default:"9090"`

	// Metrics configuration for Prometheus metrics
	Metrics Metrics

	// Health configuration for health check endpoints
	Health Health

	// Debug configuration for debugging and profiling endpoints
	Debug Debug
}

// Metrics contains configuration for Prometheus metrics.
type Metrics struct {
	// Enabled determines if metrics endpoint should be available
	Enabled bool `default:"true"`

	// Path is the URL path for metrics endpoint
	Path string `default:"/metrics"`
}

// Health contains configuration for health check endpoints.
type Health struct {
	// Enabled determines if health check endpoints should be available
	Enabled bool `default:"true"`

	// LivePath is the URL path for liveness probe endpoint
	LivePath string `default:"/health/live"`

	// ReadyPath is the URL path for readiness probe endpoint
	ReadyPath string `default:"/health/ready"`

	// CheckPath is the URL path for detailed health check information
	CheckPath string `default:"/health/checks"`

	// Timeout is the maximum time to wait for health checks to complete
	Timeout time.Duration `default:"30s"`

	// CacheTTL is how long to cache health check results
	CacheTTL time.Duration `default:"5s"`
}

// Debug contains configuration for debugging and profiling endpoints.
type Debug struct {
	// Enabled determines if debug endpoints should be available
	Enabled bool `default:"true"`

	// PathPrefix is the URL path prefix for debug endpoints
	PathPrefix string `default:"/debug"`
}

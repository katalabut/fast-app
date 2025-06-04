package health

import (
	"context"
	"time"
)

// HealthStatus represents the status of a health check
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// HealthResult represents the result of a health check
type HealthResult struct {
	Status   HealthStatus           `json:"status"`
	Message  string                 `json:"message"`
	Details  map[string]interface{} `json:"details,omitempty"`
	Duration time.Duration          `json:"duration"`
}

// HealthChecker is the main interface for health checks
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) HealthResult
}

// HealthProvider interface allows services to provide their own health checks
type HealthProvider interface {
	HealthChecks() []HealthChecker
}

// ReadinessController interface allows services to control their readiness state
type ReadinessController interface {
	SetReady(ready bool)
	IsReady() bool
}

// AggregationStrategy defines how to aggregate multiple health check results
type AggregationStrategy interface {
	Aggregate(results map[string]HealthResult) HealthStatus
}

// ComponentImportance defines the importance level of a component
type ComponentImportance int

const (
	Optional ComponentImportance = iota
	Important
	Critical
)

// HealthCheckOptions contains configuration for health checks
type HealthCheckOptions struct {
	Timeout    time.Duration
	Interval   time.Duration
	Importance ComponentImportance
}

// DefaultHealthCheckOptions returns default options for health checks
func DefaultHealthCheckOptions() HealthCheckOptions {
	return HealthCheckOptions{
		Timeout:    30 * time.Second,
		Interval:   10 * time.Second,
		Importance: Important,
	}
}

// NewHealthyResult creates a healthy result
func NewHealthyResult(message string) HealthResult {
	return HealthResult{
		Status:  StatusHealthy,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NewUnhealthyResult creates an unhealthy result
func NewUnhealthyResult(message string) HealthResult {
	return HealthResult{
		Status:  StatusUnhealthy,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NewDegradedResult creates a degraded result
func NewDegradedResult(message string) HealthResult {
	return HealthResult{
		Status:  StatusDegraded,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetails adds details to a health result
func (hr HealthResult) WithDetails(key string, value interface{}) HealthResult {
	if hr.Details == nil {
		hr.Details = make(map[string]interface{})
	}
	hr.Details[key] = value
	return hr
}

// WithDuration sets the duration for a health result
func (hr HealthResult) WithDuration(duration time.Duration) HealthResult {
	hr.Duration = duration
	return hr
}

// IsHealthy returns true if the status is healthy
func (hr HealthResult) IsHealthy() bool {
	return hr.Status == StatusHealthy
}

// IsDegraded returns true if the status is degraded
func (hr HealthResult) IsDegraded() bool {
	return hr.Status == StatusDegraded
}

// IsUnhealthy returns true if the status is unhealthy
func (hr HealthResult) IsUnhealthy() bool {
	return hr.Status == StatusUnhealthy
}

// AllHealthyStrategy requires all health checks to be healthy
type AllHealthyStrategy struct{}

// Aggregate returns healthy only if all checks are healthy
func (s *AllHealthyStrategy) Aggregate(results map[string]HealthResult) HealthStatus {
	if len(results) == 0 {
		return StatusHealthy
	}

	hasDegraded := false

	for _, result := range results {
		if result.IsUnhealthy() {
			return StatusUnhealthy
		}
		if result.IsDegraded() {
			hasDegraded = true
		}
	}

	if hasDegraded {
		return StatusDegraded
	}

	return StatusHealthy
}

// Convenience functions for creating health checks

// NewCustomCheck creates a new custom health check
func NewCustomCheck(name string, checkFn func(ctx context.Context) HealthResult) HealthChecker {
	return &customCheck{
		name:    name,
		checkFn: checkFn,
	}
}

type customCheck struct {
	name    string
	checkFn func(ctx context.Context) HealthResult
}

func (c *customCheck) Name() string {
	return c.name
}

func (c *customCheck) Check(ctx context.Context) HealthResult {
	return c.checkFn(ctx)
}

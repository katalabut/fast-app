package checks

import (
	"context"
	"time"

	"github.com/katalabut/fast-app/health"
)

// CustomCheckFunc is a function type for custom health checks
type CustomCheckFunc func(ctx context.Context) health.HealthResult

// CustomCheck wraps a custom function as a health checker
type CustomCheck struct {
	name    string
	checkFn CustomCheckFunc
	timeout time.Duration
}

// NewCustomCheck creates a new custom health check
func NewCustomCheck(name string, checkFn CustomCheckFunc) *CustomCheck {
	return &CustomCheck{
		name:    name,
		checkFn: checkFn,
		timeout: 30 * time.Second,
	}
}

// NewCustomCheckWithTimeout creates a new custom health check with timeout
func NewCustomCheckWithTimeout(name string, checkFn CustomCheckFunc, timeout time.Duration) *CustomCheck {
	return &CustomCheck{
		name:    name,
		checkFn: checkFn,
		timeout: timeout,
	}
}

// Name returns the name of the health check
func (c *CustomCheck) Name() string {
	return c.name
}

// Check executes the custom health check function
func (c *CustomCheck) Check(ctx context.Context) health.HealthResult {
	// Create a context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Channel to receive the result
	resultChan := make(chan health.HealthResult, 1)

	// Run the check in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- health.NewUnhealthyResult("panic during health check").
					WithDetails("panic", r)
			}
		}()

		result := c.checkFn(checkCtx)
		resultChan <- result
	}()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result
	case <-checkCtx.Done():
		if checkCtx.Err() == context.DeadlineExceeded {
			return health.NewUnhealthyResult("health check timeout").
				WithDetails("timeout", c.timeout.String())
		}
		return health.NewUnhealthyResult("health check cancelled").
			WithDetails("error", checkCtx.Err().Error())
	}
}

package checks

import (
	"context"
	"testing"
	"time"

	"github.com/katalabut/fast-app/health"
)

func TestCustomCheck(t *testing.T) {
	t.Run("NewCustomCheck", func(t *testing.T) {
		checkFunc := func(ctx context.Context) health.HealthResult {
			return health.NewHealthyResult("test passed")
		}
		
		check := NewCustomCheck("test-check", checkFunc)
		if check.Name() != "test-check" {
			t.Errorf("Expected name 'test-check', got '%s'", check.Name())
		}
	})

	t.Run("NewCustomCheckWithTimeout", func(t *testing.T) {
		checkFunc := func(ctx context.Context) health.HealthResult {
			return health.NewHealthyResult("test passed")
		}
		
		timeout := 10 * time.Second
		check := NewCustomCheckWithTimeout("test-check", checkFunc, timeout)
		
		if check.Name() != "test-check" {
			t.Errorf("Expected name 'test-check', got '%s'", check.Name())
		}
		
		if check.timeout != timeout {
			t.Errorf("Expected timeout %v, got %v", timeout, check.timeout)
		}
	})

	t.Run("SuccessfulCheck", func(t *testing.T) {
		checkFunc := func(ctx context.Context) health.HealthResult {
			return health.NewHealthyResult("all good").
				WithDetails("test_value", 42)
		}
		
		check := NewCustomCheck("test-check", checkFunc)
		result := check.Check(context.Background())
		
		if result.Status != health.StatusHealthy {
			t.Errorf("Expected status %s, got %s", health.StatusHealthy, result.Status)
		}
		
		if result.Message != "all good" {
			t.Errorf("Expected message 'all good', got '%s'", result.Message)
		}
		
		if result.Details["test_value"] != 42 {
			t.Errorf("Expected test_value 42, got %v", result.Details["test_value"])
		}
	})

	t.Run("UnhealthyCheck", func(t *testing.T) {
		checkFunc := func(ctx context.Context) health.HealthResult {
			return health.NewUnhealthyResult("something went wrong").
				WithDetails("error_code", "E001")
		}
		
		check := NewCustomCheck("test-check", checkFunc)
		result := check.Check(context.Background())
		
		if result.Status != health.StatusUnhealthy {
			t.Errorf("Expected status %s, got %s", health.StatusUnhealthy, result.Status)
		}
		
		if result.Message != "something went wrong" {
			t.Errorf("Expected message 'something went wrong', got '%s'", result.Message)
		}
	})

	t.Run("DegradedCheck", func(t *testing.T) {
		checkFunc := func(ctx context.Context) health.HealthResult {
			return health.NewDegradedResult("performance issues").
				WithDetails("response_time", "500ms")
		}
		
		check := NewCustomCheck("test-check", checkFunc)
		result := check.Check(context.Background())
		
		if result.Status != health.StatusDegraded {
			t.Errorf("Expected status %s, got %s", health.StatusDegraded, result.Status)
		}
	})

	t.Run("CheckTimeout", func(t *testing.T) {
		checkFunc := func(ctx context.Context) health.HealthResult {
			// Simulate a slow operation
			time.Sleep(100 * time.Millisecond)
			return health.NewHealthyResult("slow but successful")
		}
		
		check := NewCustomCheckWithTimeout("test-check", checkFunc, 10*time.Millisecond)
		result := check.Check(context.Background())
		
		if result.Status != health.StatusUnhealthy {
			t.Errorf("Expected status %s due to timeout, got %s", health.StatusUnhealthy, result.Status)
		}
		
		if result.Details["timeout"] == nil {
			t.Error("Expected timeout details to be present")
		}
	})

	t.Run("CheckPanic", func(t *testing.T) {
		checkFunc := func(ctx context.Context) health.HealthResult {
			panic("something terrible happened")
		}
		
		check := NewCustomCheck("test-check", checkFunc)
		result := check.Check(context.Background())
		
		if result.Status != health.StatusUnhealthy {
			t.Errorf("Expected status %s due to panic, got %s", health.StatusUnhealthy, result.Status)
		}
		
		if result.Details["panic"] == nil {
			t.Error("Expected panic details to be present")
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		checkFunc := func(ctx context.Context) health.HealthResult {
			select {
			case <-time.After(100 * time.Millisecond):
				return health.NewHealthyResult("completed")
			case <-ctx.Done():
				return health.NewUnhealthyResult("cancelled")
			}
		}
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		check := NewCustomCheck("test-check", checkFunc)
		result := check.Check(ctx)
		
		if result.Status != health.StatusUnhealthy {
			t.Errorf("Expected status %s due to cancellation, got %s", health.StatusUnhealthy, result.Status)
		}
	})

	t.Run("DefaultTimeout", func(t *testing.T) {
		checkFunc := func(ctx context.Context) health.HealthResult {
			return health.NewHealthyResult("test")
		}
		
		check := NewCustomCheck("test-check", checkFunc)
		
		if check.timeout != 30*time.Second {
			t.Errorf("Expected default timeout 30s, got %v", check.timeout)
		}
	})
}

package health

import (
	"context"
	"testing"
	"time"
)

func TestHealthResult(t *testing.T) {
	t.Run("NewHealthyResult", func(t *testing.T) {
		result := NewHealthyResult("test message")
		if result.Status != StatusHealthy {
			t.Errorf("Expected status %s, got %s", StatusHealthy, result.Status)
		}
		if result.Message != "test message" {
			t.Errorf("Expected message 'test message', got '%s'", result.Message)
		}
		if !result.IsHealthy() {
			t.Error("Expected IsHealthy() to return true")
		}
	})

	t.Run("NewUnhealthyResult", func(t *testing.T) {
		result := NewUnhealthyResult("error message")
		if result.Status != StatusUnhealthy {
			t.Errorf("Expected status %s, got %s", StatusUnhealthy, result.Status)
		}
		if !result.IsUnhealthy() {
			t.Error("Expected IsUnhealthy() to return true")
		}
	})

	t.Run("NewDegradedResult", func(t *testing.T) {
		result := NewDegradedResult("degraded message")
		if result.Status != StatusDegraded {
			t.Errorf("Expected status %s, got %s", StatusDegraded, result.Status)
		}
		if !result.IsDegraded() {
			t.Error("Expected IsDegraded() to return true")
		}
	})

	t.Run("WithDetails", func(t *testing.T) {
		result := NewHealthyResult("test").
			WithDetails("key1", "value1").
			WithDetails("key2", 42)
		
		if result.Details["key1"] != "value1" {
			t.Errorf("Expected details key1 to be 'value1', got %v", result.Details["key1"])
		}
		if result.Details["key2"] != 42 {
			t.Errorf("Expected details key2 to be 42, got %v", result.Details["key2"])
		}
	})

	t.Run("WithDuration", func(t *testing.T) {
		duration := 100 * time.Millisecond
		result := NewHealthyResult("test").WithDuration(duration)
		
		if result.Duration != duration {
			t.Errorf("Expected duration %v, got %v", duration, result.Duration)
		}
	})
}

func TestAllHealthyStrategy(t *testing.T) {
	strategy := &AllHealthyStrategy{}

	t.Run("EmptyResults", func(t *testing.T) {
		results := make(map[string]HealthResult)
		status := strategy.Aggregate(results)
		if status != StatusHealthy {
			t.Errorf("Expected %s for empty results, got %s", StatusHealthy, status)
		}
	})

	t.Run("AllHealthy", func(t *testing.T) {
		results := map[string]HealthResult{
			"check1": NewHealthyResult("ok"),
			"check2": NewHealthyResult("ok"),
		}
		status := strategy.Aggregate(results)
		if status != StatusHealthy {
			t.Errorf("Expected %s for all healthy, got %s", StatusHealthy, status)
		}
	})

	t.Run("OneDegraded", func(t *testing.T) {
		results := map[string]HealthResult{
			"check1": NewHealthyResult("ok"),
			"check2": NewDegradedResult("slow"),
		}
		status := strategy.Aggregate(results)
		if status != StatusDegraded {
			t.Errorf("Expected %s for one degraded, got %s", StatusDegraded, status)
		}
	})

	t.Run("OneUnhealthy", func(t *testing.T) {
		results := map[string]HealthResult{
			"check1": NewHealthyResult("ok"),
			"check2": NewUnhealthyResult("failed"),
		}
		status := strategy.Aggregate(results)
		if status != StatusUnhealthy {
			t.Errorf("Expected %s for one unhealthy, got %s", StatusUnhealthy, status)
		}
	})
}

func TestNewCustomCheck(t *testing.T) {
	checkFunc := func(ctx context.Context) HealthResult {
		return NewHealthyResult("custom check passed")
	}

	checker := NewCustomCheck("test-check", checkFunc)
	
	if checker.Name() != "test-check" {
		t.Errorf("Expected name 'test-check', got '%s'", checker.Name())
	}

	result := checker.Check(context.Background())
	if result.Status != StatusHealthy {
		t.Errorf("Expected status %s, got %s", StatusHealthy, result.Status)
	}
	if result.Message != "custom check passed" {
		t.Errorf("Expected message 'custom check passed', got '%s'", result.Message)
	}
}

package health

import (
	"context"
	"testing"
	"time"
)

type mockChecker struct {
	name   string
	result HealthResult
}

func (m *mockChecker) Name() string {
	return m.name
}

func (m *mockChecker) Check(ctx context.Context) HealthResult {
	return m.result
}

func TestManager(t *testing.T) {
	t.Run("NewManager", func(t *testing.T) {
		config := ManagerConfig{
			CacheTTL: 5 * time.Second,
			Strategy: &AllHealthyStrategy{},
		}
		manager := NewManager(config)
		
		if manager == nil {
			t.Fatal("Expected manager to be created")
		}
		if !manager.IsReady() {
			t.Error("Expected manager to be ready by default")
		}
	})

	t.Run("RegisterChecker", func(t *testing.T) {
		manager := NewManager(ManagerConfig{})
		checker := &mockChecker{
			name:   "test-check",
			result: NewHealthyResult("ok"),
		}
		
		manager.RegisterChecker(checker)
		names := manager.GetCheckerNames()
		
		if len(names) != 1 {
			t.Errorf("Expected 1 checker, got %d", len(names))
		}
		if names[0] != "test-check" {
			t.Errorf("Expected checker name 'test-check', got '%s'", names[0])
		}
	})

	t.Run("RegisterCheckers", func(t *testing.T) {
		manager := NewManager(ManagerConfig{})
		checkers := []HealthChecker{
			&mockChecker{name: "check1", result: NewHealthyResult("ok")},
			&mockChecker{name: "check2", result: NewHealthyResult("ok")},
		}
		
		manager.RegisterCheckers(checkers)
		names := manager.GetCheckerNames()
		
		if len(names) != 2 {
			t.Errorf("Expected 2 checkers, got %d", len(names))
		}
	})

	t.Run("UnregisterChecker", func(t *testing.T) {
		manager := NewManager(ManagerConfig{})
		checker := &mockChecker{
			name:   "test-check",
			result: NewHealthyResult("ok"),
		}
		
		manager.RegisterChecker(checker)
		manager.UnregisterChecker("test-check")
		names := manager.GetCheckerNames()
		
		if len(names) != 0 {
			t.Errorf("Expected 0 checkers after unregister, got %d", len(names))
		}
	})

	t.Run("CheckAll", func(t *testing.T) {
		manager := NewManager(ManagerConfig{})
		checkers := []HealthChecker{
			&mockChecker{name: "healthy-check", result: NewHealthyResult("ok")},
			&mockChecker{name: "unhealthy-check", result: NewUnhealthyResult("failed")},
		}
		
		manager.RegisterCheckers(checkers)
		results := manager.CheckAll(context.Background())
		
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		
		if results["healthy-check"].Status != StatusHealthy {
			t.Errorf("Expected healthy-check to be healthy, got %s", results["healthy-check"].Status)
		}
		
		if results["unhealthy-check"].Status != StatusUnhealthy {
			t.Errorf("Expected unhealthy-check to be unhealthy, got %s", results["unhealthy-check"].Status)
		}
	})

	t.Run("GetOverallStatus", func(t *testing.T) {
		manager := NewManager(ManagerConfig{
			Strategy: &AllHealthyStrategy{},
		})
		
		// Test with healthy checks
		healthyChecker := &mockChecker{
			name:   "healthy-check",
			result: NewHealthyResult("ok"),
		}
		manager.RegisterChecker(healthyChecker)
		
		status := manager.GetOverallStatus(context.Background())
		if status != StatusHealthy {
			t.Errorf("Expected overall status to be healthy, got %s", status)
		}
		
		// Add unhealthy check
		unhealthyChecker := &mockChecker{
			name:   "unhealthy-check",
			result: NewUnhealthyResult("failed"),
		}
		manager.RegisterChecker(unhealthyChecker)
		
		status = manager.GetOverallStatus(context.Background())
		if status != StatusUnhealthy {
			t.Errorf("Expected overall status to be unhealthy, got %s", status)
		}
	})

	t.Run("SetReady", func(t *testing.T) {
		manager := NewManager(ManagerConfig{})
		
		if !manager.IsReady() {
			t.Error("Expected manager to be ready initially")
		}
		
		manager.SetReady(false)
		if manager.IsReady() {
			t.Error("Expected manager to be not ready after SetReady(false)")
		}
		
		manager.SetReady(true)
		if !manager.IsReady() {
			t.Error("Expected manager to be ready after SetReady(true)")
		}
	})

	t.Run("ClearCache", func(t *testing.T) {
		manager := NewManager(ManagerConfig{
			CacheTTL: 1 * time.Hour, // Long cache to test clearing
		})
		
		checker := &mockChecker{
			name:   "test-check",
			result: NewHealthyResult("ok"),
		}
		manager.RegisterChecker(checker)
		
		// Run check to populate cache
		manager.CheckAll(context.Background())
		
		// Clear cache
		manager.ClearCache()
		
		// This test mainly ensures ClearCache doesn't panic
		// In a real scenario, we'd need to verify cache is actually cleared
		// but that would require exposing internal cache state
	})
}

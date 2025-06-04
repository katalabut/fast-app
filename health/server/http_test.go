package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/katalabut/fast-app/health"
)

type mockHealthChecker struct {
	name   string
	result health.HealthResult
}

func (m *mockHealthChecker) Name() string {
	return m.name
}

func (m *mockHealthChecker) Check(ctx context.Context) health.HealthResult {
	return m.result
}

func TestServer(t *testing.T) {
	t.Run("NewServer", func(t *testing.T) {
		config := Config{
			Enabled:   true,
			Port:      8080,
			LivePath:  "/health/live",
			ReadyPath: "/health/ready",
			CheckPath: "/health/checks",
			Timeout:   30 * time.Second,
		}
		manager := health.NewManager(health.ManagerConfig{})
		server := NewServer(config, manager)
		
		if server == nil {
			t.Fatal("Expected server to be created")
		}
	})

	t.Run("LivenessEndpoint", func(t *testing.T) {
		config := Config{
			LivePath: "/health/live",
			Timeout:  30 * time.Second,
		}
		manager := health.NewManager(health.ManagerConfig{})
		server := NewServer(config, manager)
		
		req := httptest.NewRequest("GET", "/health/live", nil)
		w := httptest.NewRecorder()
		
		server.handleLiveness(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response["status"] != "alive" {
			t.Errorf("Expected status 'alive', got %v", response["status"])
		}
		
		if response["timestamp"] == nil {
			t.Error("Expected timestamp to be present")
		}
	})

	t.Run("ReadinessEndpointReady", func(t *testing.T) {
		config := Config{
			ReadyPath: "/health/ready",
			Timeout:   30 * time.Second,
		}
		manager := health.NewManager(health.ManagerConfig{})
		manager.SetReady(true)
		server := NewServer(config, manager)
		
		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()
		
		server.handleReadiness(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response["ready"] != true {
			t.Errorf("Expected ready true, got %v", response["ready"])
		}
	})

	t.Run("ReadinessEndpointNotReady", func(t *testing.T) {
		config := Config{
			ReadyPath: "/health/ready",
			Timeout:   30 * time.Second,
		}
		manager := health.NewManager(health.ManagerConfig{})
		manager.SetReady(false)
		server := NewServer(config, manager)
		
		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()
		
		server.handleReadiness(w, req)
		
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
		}
		
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response["ready"] != false {
			t.Errorf("Expected ready false, got %v", response["ready"])
		}
	})

	t.Run("ChecksEndpointHealthy", func(t *testing.T) {
		config := Config{
			CheckPath: "/health/checks",
			Timeout:   30 * time.Second,
		}
		manager := health.NewManager(health.ManagerConfig{})
		
		// Add some mock health checks
		healthyCheck := &mockHealthChecker{
			name:   "healthy-check",
			result: health.NewHealthyResult("all good"),
		}
		manager.RegisterChecker(healthyCheck)
		
		server := NewServer(config, manager)
		
		req := httptest.NewRequest("GET", "/health/checks", nil)
		w := httptest.NewRecorder()
		
		server.handleChecks(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response["status"] != "healthy" {
			t.Errorf("Expected status 'healthy', got %v", response["status"])
		}
		
		checks, ok := response["checks"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected checks to be a map")
		}
		
		if len(checks) != 1 {
			t.Errorf("Expected 1 check, got %d", len(checks))
		}
		
		if checks["healthy-check"] == nil {
			t.Error("Expected healthy-check to be present")
		}
	})

	t.Run("ChecksEndpointUnhealthy", func(t *testing.T) {
		config := Config{
			CheckPath: "/health/checks",
			Timeout:   30 * time.Second,
		}
		manager := health.NewManager(health.ManagerConfig{})
		
		// Add an unhealthy check
		unhealthyCheck := &mockHealthChecker{
			name:   "unhealthy-check",
			result: health.NewUnhealthyResult("something failed"),
		}
		manager.RegisterChecker(unhealthyCheck)
		
		server := NewServer(config, manager)
		
		req := httptest.NewRequest("GET", "/health/checks", nil)
		w := httptest.NewRecorder()
		
		server.handleChecks(w, req)
		
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
		}
		
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response["status"] != "unhealthy" {
			t.Errorf("Expected status 'unhealthy', got %v", response["status"])
		}
	})

	t.Run("ChecksEndpointDegraded", func(t *testing.T) {
		config := Config{
			CheckPath: "/health/checks",
			Timeout:   30 * time.Second,
		}
		manager := health.NewManager(health.ManagerConfig{})
		
		// Add a degraded check
		degradedCheck := &mockHealthChecker{
			name:   "degraded-check",
			result: health.NewDegradedResult("performance issues"),
		}
		manager.RegisterChecker(degradedCheck)
		
		server := NewServer(config, manager)
		
		req := httptest.NewRequest("GET", "/health/checks", nil)
		w := httptest.NewRecorder()
		
		server.handleChecks(w, req)
		
		// Degraded should still return 200 OK
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response["status"] != "degraded" {
			t.Errorf("Expected status 'degraded', got %v", response["status"])
		}
	})

	t.Run("ContentTypeJSON", func(t *testing.T) {
		config := Config{
			LivePath: "/health/live",
			Timeout:  30 * time.Second,
		}
		manager := health.NewManager(health.ManagerConfig{})
		server := NewServer(config, manager)
		
		req := httptest.NewRequest("GET", "/health/live", nil)
		w := httptest.NewRecorder()
		
		server.handleLiveness(w, req)
		
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
	})
}

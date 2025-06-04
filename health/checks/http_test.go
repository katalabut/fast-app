package checks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/katalabut/fast-app/health"
)

func TestHTTPCheck(t *testing.T) {
	t.Run("NewHTTPCheck", func(t *testing.T) {
		check := NewHTTPCheck("test", "http://example.com")
		if check.Name() != "test" {
			t.Errorf("Expected name 'test', got '%s'", check.Name())
		}
	})

	t.Run("SuccessfulCheck", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		check := NewHTTPCheck("test", server.URL)
		result := check.Check(context.Background())

		if result.Status != health.StatusHealthy {
			t.Errorf("Expected status %s, got %s", health.StatusHealthy, result.Status)
		}
		if result.Details["status_code"] != 200 {
			t.Errorf("Expected status code 200, got %v", result.Details["status_code"])
		}
	})

	t.Run("UnhealthyStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		check := NewHTTPCheck("test", server.URL)
		result := check.Check(context.Background())

		if result.Status != health.StatusUnhealthy {
			t.Errorf("Expected status %s, got %s", health.StatusUnhealthy, result.Status)
		}
	})

	t.Run("WithOptions", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"healthy"}`))
		}))
		defer server.Close()

		check := NewHTTPCheckWithOptions("test", server.URL, HTTPOptions{
			Timeout:        5 * time.Second,
			ExpectedStatus: 200,
			ExpectedBody:   "healthy",
			Headers:        map[string]string{"Authorization": "Bearer token"},
		})

		result := check.Check(context.Background())
		if result.Status != health.StatusHealthy {
			t.Errorf("Expected status %s, got %s", health.StatusHealthy, result.Status)
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		check := NewHTTPCheckWithOptions("test", server.URL, HTTPOptions{
			Timeout: 10 * time.Millisecond,
		})

		result := check.Check(context.Background())
		if result.Status != health.StatusUnhealthy {
			t.Errorf("Expected status %s, got %s", health.StatusUnhealthy, result.Status)
		}
	})

	t.Run("ExpectedBodyMismatch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"error"}`))
		}))
		defer server.Close()

		check := NewHTTPCheckWithOptions("test", server.URL, HTTPOptions{
			ExpectedBody: "healthy",
		})

		result := check.Check(context.Background())
		if result.Status != health.StatusUnhealthy {
			t.Errorf("Expected status %s, got %s", health.StatusUnhealthy, result.Status)
		}
	})

	t.Run("SlowResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(60 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		check := NewHTTPCheckWithOptions("test", server.URL, HTTPOptions{
			Timeout: 100 * time.Millisecond,
		})

		result := check.Check(context.Background())
		// Should be degraded due to slow response (> timeout/2)
		if result.Status != health.StatusDegraded {
			t.Errorf("Expected status %s, got %s", health.StatusDegraded, result.Status)
		}
	})

	t.Run("InvalidURL", func(t *testing.T) {
		check := NewHTTPCheck("test", "invalid-url")
		result := check.Check(context.Background())

		if result.Status != health.StatusUnhealthy {
			t.Errorf("Expected status %s, got %s", health.StatusUnhealthy, result.Status)
		}
	})
}

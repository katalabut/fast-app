package checks

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/katalabut/fast-app/health"
)

func TestHTTPBodyLimitAndNoRawBodyDetails(t *testing.T) {
	for _, body := range []string{"dummy-secret", strings.Repeat("x", 70<<10)} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
		result := NewHTTPCheckWithOptions("upstream", s.URL, HTTPOptions{ExpectedBody: "not-present"}).Check(context.Background())
		s.Close()
		if result.Status != health.StatusUnhealthy {
			t.Fatal("invalid upstream reported healthy")
		}
		if strings.Contains(fmt.Sprint(result.Details), "dummy-secret") || result.Details["actual_body"] != nil || result.Details["expected_body"] != nil {
			t.Fatalf("raw body leaked: %+v", result.Details)
		}
		if len(body) > 64<<10 && !strings.Contains(result.Message, "limit") {
			t.Fatalf("oversized body not limited: %+v", result)
		}
	}
}

func TestHTTPClientTimeoutClassified(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer s.Close()
	result := NewHTTPCheckWithOptions("upstream", s.URL, HTTPOptions{Timeout: 10 * time.Millisecond}).Check(context.Background())
	if result.Message != "HTTP request timeout" {
		t.Fatalf("client timeout misclassified: %+v", result)
	}
}

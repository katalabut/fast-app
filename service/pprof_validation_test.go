package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/katalabut/fast-app/config"
)

func TestDormantPprofPatternAllowed(t *testing.T) {
	cfg := config.Observability{Enabled: true, Metrics: config.Metrics{Enabled: true, Path: "/debug/pprof/profile"}, Debug: config.Debug{Enabled: true, PathPrefix: "/debug"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("dormant profile validation: %v", err)
	}
	s := NewObservabilityService(cfg, nil)
	mux, err := s.routes()
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", cfg.Metrics.Path, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("metrics path status=%d", w.Code)
	}
	if _, err = s.WithPprof().routes(); err == nil {
		t.Fatal("active pprof conflict accepted")
	}
}

func TestCustomPprofPatternStillConflicts(t *testing.T) {
	cfg := config.Observability{Enabled: true, Metrics: config.Metrics{Enabled: true, Path: "/custom/pprof/profile"}, Debug: config.Debug{Enabled: true, PathPrefix: "/custom"}}
	if _, err := NewObservabilityService(cfg, nil).routes(); err == nil {
		t.Fatal("custom-prefix pprof conflict accepted")
	}
}

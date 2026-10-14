package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/katalabut/fast-app/config"
	"github.com/katalabut/fast-app/health"
)

func TestPprofPrivateMux(t *testing.T) {
	for _, prefix := range []string{"/debug", "/custom"} {
		t.Run(prefix, func(t *testing.T) {
			s := NewObservabilityService(config.Observability{Debug: config.Debug{Enabled: true, PathPrefix: prefix}}, health.NewManager(health.ManagerConfig{})).WithPprof()
			mux := http.NewServeMux()
			s.registerDebugEndpoints(mux)
			for _, suffix := range []string{"/pprof/", "/pprof/cmdline", "/pprof/goroutine?debug=1"} {
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, httptest.NewRequest("GET", prefix+suffix, nil))
				if w.Code != http.StatusOK {
					t.Errorf("%s%s status=%d", prefix, suffix, w.Code)
				}
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("GET", "/not-a-debug-route", nil))
			if w.Code != http.StatusNotFound {
				t.Error("private mux exposes unrelated routes")
			}
		})
	}
}

func TestDetailsOptInAndSingleSnapshot(t *testing.T) {
	for _, include := range []bool{false, true} {
		var calls atomic.Int32
		m := health.NewManager(health.ManagerConfig{CacheTTL: -1})
		m.RegisterChecker(health.NewCustomCheck("dependency", func(context.Context) health.HealthResult {
			if calls.Add(1) > 1 {
				return health.NewUnhealthyResult("changed")
			}
			return health.NewHealthyResult("ok").WithDetails("diagnostic", "dummy-secret")
		}))
		s := NewObservabilityService(config.Observability{Health: config.Health{Timeout: time.Second}}, m)
		if !include {
			s.WithDetailsRedacted()
		}
		w := httptest.NewRecorder()
		s.handleHealthChecks(w, httptest.NewRequest("GET", "/health/checks", nil))
		if calls.Load() != 1 || w.Code != http.StatusOK {
			t.Fatalf("inconsistent/repeated snapshot: calls=%d status=%d body=%s", calls.Load(), w.Code, w.Body.String())
		}
		if strings.Contains(w.Body.String(), "dummy-secret") != include {
			t.Fatalf("details exposure mismatch: %s", w.Body.String())
		}
	}
}

func TestDisabledPprofAndInvalidRoutes(t *testing.T) {
	s := NewObservabilityService(config.Observability{Enabled: true}, health.NewManager(health.ManagerConfig{}))
	mux, err := s.routes()
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/debug/pprof/", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal("disabled pprof exposed")
	}
	s = NewObservabilityService(config.Observability{Enabled: true, Metrics: config.Metrics{Enabled: true, Path: "invalid"}}, nil)
	if err := s.Run(context.Background()); err == nil {
		t.Fatal("invalid route accepted")
	}
	if s.server.Addr != ":0" {
		t.Fatalf("legacy default listener changed: %s", s.server.Addr)
	}
}

func TestLegacyShutdownBeforeRun(t *testing.T) {
	s := NewDefaultDebugService(DebugServer{Port: 0})
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.server.Addr != ":0" {
		t.Fatalf("legacy listener changed: %s", s.server.Addr)
	}
}

func TestObservabilityShutdownBeforeRun(t *testing.T) {
	s := NewObservabilityService(config.Observability{Enabled: true, Port: 0}, health.NewManager(health.ManagerConfig{}))
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- s.Run(context.Background()) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("server started after completed Shutdown")
		_ = s.Shutdown(context.Background())
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("server did not stop")
		}
	}
}

func TestObservabilityCancellationAndShutdownRace(t *testing.T) {
	for i := 0; i < 30; i++ {
		s := NewObservabilityService(config.Observability{Enabled: true, Port: 0}, health.NewManager(health.ManagerConfig{}))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := s.Run(ctx); err != nil {
				t.Errorf("Run: %v", err)
			}
		}()
		go func() { defer wg.Done(); _ = s.Shutdown(context.Background()) }()
		done := make(chan struct{})
		go func() { wg.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("cancelled server did not stop")
		}
	}
}

func TestPprofRequiresExplicitOptInAtDefaultPrefix(t *testing.T) {
	cfg := config.Observability{Enabled: true, Debug: config.Debug{Enabled: true, PathPrefix: "/debug"}}
	s := NewObservabilityService(cfg, nil)
	mux, err := s.routes()
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/debug/pprof/", nil))
	if w.Code != http.StatusNotFound {
		t.Fatal("new default profiling exposure")
	}
	s.WithListenAddress("127.0.0.1:9090").WithPprof()
	mux, err = s.routes()
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/debug/pprof/", nil))
	if w.Code != http.StatusOK || s.server.Addr != "127.0.0.1:9090" {
		t.Fatal("explicit profile/private address did not work")
	}
}

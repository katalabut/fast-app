package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/katalabut/fast-app/health"
)

func TestShutdownBeforeStart(t *testing.T) {
	s := NewServer(Config{Enabled: true, Port: 0, LivePath: "/live", ReadyPath: "/ready", CheckPath: "/checks", Timeout: time.Second}, health.NewManager(health.ManagerConfig{}))
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.server.Addr != ":0" {
		t.Fatalf("legacy listener changed: %s", s.server.Addr)
	}
}

func TestDetailsAndSingleSnapshot(t *testing.T) {
	for _, include := range []bool{false, true} {
		var calls atomic.Int32
		m := health.NewManager(health.ManagerConfig{CacheTTL: -1})
		m.RegisterChecker(health.NewCustomCheck("dependency", func(context.Context) health.HealthResult {
			if calls.Add(1) > 1 {
				return health.NewUnhealthyResult("changed")
			}
			return health.NewHealthyResult("ok").WithDetails("diagnostic", "dummy-secret")
		}))
		s := NewServer(Config{Timeout: time.Second}, m)
		if !include {
			s.WithDetailsRedacted()
		}
		w := httptest.NewRecorder()
		s.handleChecks(w, httptest.NewRequest("GET", "/checks", nil))
		if calls.Load() != 1 || w.Code != http.StatusOK {
			t.Fatalf("inconsistent snapshot: calls=%d status=%d", calls.Load(), w.Code)
		}
		if strings.Contains(w.Body.String(), "dummy-secret") != include {
			t.Fatal("details policy violated")
		}
	}
}

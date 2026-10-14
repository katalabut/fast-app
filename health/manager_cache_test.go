package health

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestManagerCacheExpires(t *testing.T) {
	var calls atomic.Int32
	m := NewManager(ManagerConfig{CacheTTL: 10 * time.Millisecond})
	m.RegisterChecker(NewCustomCheck("dependency", func(context.Context) HealthResult {
		if calls.Add(1) == 1 {
			return NewHealthyResult("initial")
		}
		return NewUnhealthyResult("changed")
	}))
	m.CheckAll(context.Background())
	m.CheckAll(context.Background())
	if calls.Load() != 1 {
		t.Fatal("cache did not reuse a fresh check")
	}
	time.Sleep(30 * time.Millisecond)
	if result := m.CheckAll(context.Background())["dependency"]; result.Status != StatusUnhealthy || calls.Load() != 2 {
		t.Fatalf("expired result retained: %+v calls=%d", result, calls.Load())
	}
}

func TestManagerReplacementInvalidatesCache(t *testing.T) {
	m := NewManager(ManagerConfig{CacheTTL: time.Hour})
	m.RegisterChecker(NewCustomCheck("dependency", func(context.Context) HealthResult { return NewHealthyResult("old") }))
	m.CheckAll(context.Background())
	m.RegisterChecker(NewCustomCheck("dependency", func(context.Context) HealthResult { return NewUnhealthyResult("new") }))
	if result := m.CheckAll(context.Background())["dependency"]; result.Status != StatusUnhealthy {
		t.Fatalf("replacement uses stale cache: %+v", result)
	}
}

func TestManagerDoesNotCacheStaleInflightResult(t *testing.T) {
	for _, action := range []string{"replace", "clear", "unregister"} {
		t.Run(action, func(t *testing.T) {
			m := NewManager(ManagerConfig{CacheTTL: time.Hour})
			started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
			var calls atomic.Int32
			m.RegisterChecker(NewCustomCheck("dependency", func(context.Context) HealthResult {
				if calls.Add(1) == 1 {
					close(started)
					<-release
					return NewHealthyResult("old")
				}
				return NewUnhealthyResult("fresh")
			}))
			go func() { m.CheckAll(context.Background()); close(done) }()
			<-started
			switch action {
			case "replace":
				m.RegisterChecker(NewCustomCheck("dependency", func(context.Context) HealthResult { return NewUnhealthyResult("replacement") }))
			case "clear":
				m.ClearCache()
			case "unregister":
				m.UnregisterChecker("dependency")
			}
			close(release)
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("inflight check did not finish")
			}
			if action == "unregister" {
				m.mu.RLock()
				count := len(m.cache)
				m.mu.RUnlock()
				if count != 0 {
					t.Fatal("unregistered result repopulated cache")
				}
			} else if result := m.CheckAll(context.Background())["dependency"]; result.Status != StatusUnhealthy {
				t.Fatalf("stale inflight result cached: %+v", result)
			}
		})
	}
}

func TestManagerCacheDetailsSnapshot(t *testing.T) {
	m := NewManager(ManagerConfig{CacheTTL: time.Hour})
	m.RegisterChecker(NewCustomCheck("dependency", func(context.Context) HealthResult { return NewHealthyResult("ok").WithDetails("original", true) }))
	first := m.CheckAll(context.Background())["dependency"]
	first = first.WithDetails("caller", "mutation")
	if _, exists := m.CheckAll(context.Background())["dependency"].Details["caller"]; exists {
		t.Fatalf("caller mutated cached details: %+v", first)
	}
}

func TestCustomCheckPanicIsUnhealthy(t *testing.T) {
	defer func() {
		if x := recover(); x != nil {
			t.Errorf("health panic escaped: %v", x)
		}
	}()
	result := NewCustomCheck("panic", func(context.Context) HealthResult { panic("sensitive panic") }).Check(context.Background())
	if result.Status != StatusUnhealthy {
		t.Fatalf("panic result: %+v", result)
	}
}

func TestCancelledHealthResultNotCached(t *testing.T) {
	m := NewManager(ManagerConfig{CacheTTL: time.Hour})
	m.RegisterChecker(NewCustomCheck("dependency", func(ctx context.Context) HealthResult {
		if ctx.Err() != nil {
			return NewUnhealthyResult("cancelled")
		}
		return NewHealthyResult("ok")
	}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m.CheckAll(ctx)
	if result := m.CheckAll(context.Background())["dependency"]; result.Status != StatusHealthy {
		t.Fatal("cancelled check cached")
	}
}

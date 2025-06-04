package health

import (
	"context"
	"sync"
	"time"

	"github.com/katalabut/fast-app/logger"
)

// Manager coordinates all health checks and manages the overall health state
type Manager struct {
	checkers map[string]HealthChecker
	strategy AggregationStrategy
	cache    map[string]HealthResult
	cacheTTL time.Duration
	mu       sync.RWMutex
	ready    bool
	readyMu  sync.RWMutex
}

// ManagerConfig contains configuration for the health manager
type ManagerConfig struct {
	CacheTTL time.Duration `default:"5s"`
	Strategy AggregationStrategy
}

// NewManager creates a new health manager
func NewManager(config ManagerConfig) *Manager {
	if config.Strategy == nil {
		config.Strategy = &AllHealthyStrategy{}
	}

	if config.CacheTTL == 0 {
		config.CacheTTL = 5 * time.Second
	}

	return &Manager{
		checkers: make(map[string]HealthChecker),
		strategy: config.Strategy,
		cache:    make(map[string]HealthResult),
		cacheTTL: config.CacheTTL,
		ready:    true, // Start as ready by default
	}
}

// RegisterChecker registers a health checker
func (m *Manager) RegisterChecker(checker HealthChecker) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := checker.Name()
	if _, exists := m.checkers[name]; exists {
		logger.Warn(context.Background(), "Health checker with name already exists, overwriting", "name", name)
	}

	m.checkers[name] = checker
	logger.Debug(context.Background(), "Registered health checker", "name", name)
}

// RegisterCheckers registers multiple health checkers
func (m *Manager) RegisterCheckers(checkers []HealthChecker) {
	for _, checker := range checkers {
		m.RegisterChecker(checker)
	}
}

// UnregisterChecker removes a health checker
func (m *Manager) UnregisterChecker(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.checkers, name)
	delete(m.cache, name)
	logger.Debug(context.Background(), "Unregistered health checker", "name", name)
}

// CheckAll runs all registered health checks
func (m *Manager) CheckAll(ctx context.Context) map[string]HealthResult {
	m.mu.RLock()
	checkers := make(map[string]HealthChecker, len(m.checkers))
	for name, checker := range m.checkers {
		checkers[name] = checker
	}
	m.mu.RUnlock()

	results := make(map[string]HealthResult, len(checkers))
	var wg sync.WaitGroup
	var resultsMu sync.Mutex

	for name, checker := range checkers {
		wg.Add(1)
		go func(name string, checker HealthChecker) {
			defer wg.Done()

			start := time.Now()
			result := m.checkWithCache(ctx, name, checker)
			result = result.WithDuration(time.Since(start))

			resultsMu.Lock()
			results[name] = result
			resultsMu.Unlock()
		}(name, checker)
	}

	wg.Wait()
	return results
}

// checkWithCache checks a single health checker with caching
func (m *Manager) checkWithCache(ctx context.Context, name string, checker HealthChecker) HealthResult {
	m.mu.RLock()
	cached, exists := m.cache[name]
	m.mu.RUnlock()

	if exists && time.Since(time.Now().Add(-cached.Duration)) < m.cacheTTL {
		return cached
	}

	result := checker.Check(ctx)

	m.mu.Lock()
	m.cache[name] = result
	m.mu.Unlock()

	return result
}

// GetOverallStatus returns the aggregated health status
func (m *Manager) GetOverallStatus(ctx context.Context) HealthStatus {
	results := m.CheckAll(ctx)
	return m.strategy.Aggregate(results)
}

// IsReady returns the readiness state
func (m *Manager) IsReady() bool {
	m.readyMu.RLock()
	defer m.readyMu.RUnlock()
	return m.ready
}

// SetReady sets the readiness state
func (m *Manager) SetReady(ready bool) {
	m.readyMu.Lock()
	defer m.readyMu.Unlock()

	if m.ready != ready {
		m.ready = ready
		logger.Info(context.Background(), "Application readiness changed", "ready", ready)
	}
}

// GetCheckerNames returns the names of all registered checkers
func (m *Manager) GetCheckerNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.checkers))
	for name := range m.checkers {
		names = append(names, name)
	}
	return names
}

// ClearCache clears the health check cache
func (m *Manager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache = make(map[string]HealthResult)
	logger.Debug(context.Background(), "Health check cache cleared")
}

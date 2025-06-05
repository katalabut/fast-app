package main

import (
	"context"
	"time"

	fastapp "github.com/katalabut/fast-app"
	"github.com/katalabut/fast-app/config"
	"github.com/katalabut/fast-app/configloader"
	"github.com/katalabut/fast-app/health"
	"github.com/katalabut/fast-app/health/checks"
	"github.com/katalabut/fast-app/logger"
)

type AppConfig struct {
	App config.App
}

type SimpleService struct {
	name  string
	ready bool
}

func NewSimpleService(name string) *SimpleService {
	return &SimpleService{
		name: name,
	}
}

func (s *SimpleService) Run(ctx context.Context) error {
	logger.Info(ctx, "Service starting", "service", s.name)

	// Simulate initialization time
	time.Sleep(2 * time.Second)
	s.SetReady(true)
	logger.Info(ctx, "Service is ready", "service", s.name)

	// Simulate periodic work
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info(ctx, "Service doing work", "service", s.name)
		case <-ctx.Done():
			logger.Info(ctx, "Service stopping", "service", s.name)
			return nil
		}
	}
}

func (s *SimpleService) Shutdown(ctx context.Context) error {
	logger.Info(ctx, "Service shutting down", "service", s.name)
	s.SetReady(false)
	return nil
}

// HealthChecks implements health.HealthProvider
func (s *SimpleService) HealthChecks() []health.HealthChecker {
	return []health.HealthChecker{
		health.NewCustomCheck(
			s.name+"-readiness", func(ctx context.Context) health.HealthResult {
				if s.IsReady() {
					return health.NewHealthyResult("Service is ready").
						WithDetails("service_name", s.name).
						WithDetails("uptime", time.Since(time.Now().Add(-10*time.Second)).String())
				}
				return health.NewUnhealthyResult("Service is not ready").
					WithDetails("service_name", s.name)
			},
		),
		health.NewCustomCheck(
			s.name+"-memory", func(ctx context.Context) health.HealthResult {
				// Simulate memory check
				memUsage := 45.6 // MB
				if memUsage > 100 {
					return health.NewDegradedResult("High memory usage").
						WithDetails("memory_mb", memUsage).
						WithDetails("threshold_mb", 100)
				}
				return health.NewHealthyResult("Memory usage normal").
					WithDetails("memory_mb", memUsage)
			},
		),
	}
}

func (s *SimpleService) SetReady(ready bool) {
	s.ready = ready
}

func (s *SimpleService) IsReady() bool {
	return s.ready
}

func main() {
	cfg, err := configloader.New[AppConfig](configloader.WithFile("../config.yaml"))
	if err != nil {
		logger.Fatal(context.Background(), "failed to load config:", err)
	}

	// Create services
	apiService := NewSimpleService("api")
	workerService := NewSimpleService("worker")
	schedulerService := NewSimpleService("scheduler")

	// Global health checks
	globalChecks := []health.HealthChecker{
		checks.NewHTTPCheck("httpbin-200", "https://httpbin.org/status/200"),
		checks.NewHTTPCheck("httpbin-delay", "https://httpbin.org/delay/1"),
		health.NewCustomCheck(
			"system-load", func(ctx context.Context) health.HealthResult {
				// Simulate system load check
				load := 0.75
				if load > 0.9 {
					return health.NewDegradedResult("High system load").
						WithDetails("load", load).
						WithDetails("threshold", 0.9)
				}
				return health.NewHealthyResult("System load normal").
					WithDetails("load", load)
			},
		),
	}

	// Create application
	app := fastapp.New(cfg.App).
		WithHealthChecks(globalChecks...).
		Add(apiService).
		Add(workerService).
		Add(schedulerService)

	// Set application as ready
	app.SetReady(true)

	logger.Info(context.Background(), "🚀 Starting FastApp with Unified Observability")
	logger.Info(context.Background(), "📊 All endpoints available on port 9090:")
	logger.Info(context.Background(), "   • Liveness:  http://localhost:9090/health/live")
	logger.Info(context.Background(), "   • Readiness: http://localhost:9090/health/ready")
	logger.Info(context.Background(), "   • Detailed:  http://localhost:9090/health/checks")
	logger.Info(context.Background(), "   • Metrics:   http://localhost:9090/metrics")
	logger.Info(context.Background(), "   • Profiling: http://localhost:9090/debug/pprof/")

	app.Start()
}

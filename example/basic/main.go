package main

import (
	"context"
	"time"

	fastapp "github.com/katalabut/fast-app"
	"github.com/katalabut/fast-app/configloader"
	"github.com/katalabut/fast-app/health"
	"github.com/katalabut/fast-app/health/checks"
	"github.com/katalabut/fast-app/logger"
	"github.com/katalabut/fast-app/service"
)

type Config struct {
	App         fastapp.Config
	DebugServer service.DebugServer
}

type ApiService struct {
	ready bool
}

func NewApiService() *ApiService {
	return &ApiService{}
}

func (s *ApiService) Run(ctx context.Context) error {
	logger.Info(ctx, "ApiService is running")

	// Simulate initialization
	time.Sleep(2 * time.Second)
	s.SetReady(true)
	logger.Info(ctx, "ApiService is ready")

	// Keep running
	<-ctx.Done()
	return nil
}

func (s *ApiService) Shutdown(ctx context.Context) error {
	logger.Info(ctx, "ApiService is shutting down")
	s.SetReady(false)
	return nil
}

// HealthChecks implements health.HealthProvider
func (s *ApiService) HealthChecks() []health.HealthChecker {
	return []health.HealthChecker{
		health.NewCustomCheck("api-service", func(ctx context.Context) health.HealthResult {
			if s.IsReady() {
				return health.NewHealthyResult("API service is ready")
			}
			return health.NewUnhealthyResult("API service is not ready")
		}),
	}
}

// SetReady implements health.ReadinessController
func (s *ApiService) SetReady(ready bool) {
	s.ready = ready
}

// IsReady implements health.ReadinessController
func (s *ApiService) IsReady() bool {
	return s.ready
}

func main() {
	cfg, err := configloader.New[Config]()
	if err != nil {
		logger.Fatal(context.Background(), "failed to load config:", err)
	}

	apiService := NewApiService()

	// Create some global health checks
	httpCheck := checks.NewHTTPCheck("google", "https://www.google.com")

	app := fastapp.New(cfg.App).
		WithHealthChecks(httpCheck).
		Add(service.NewDefaultDebugService(cfg.DebugServer)).
		Add(apiService)

	// Set application as ready after all services are added
	app.SetReady(true)

	app.Start()
}

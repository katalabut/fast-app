package main

import (
	"context"
	"database/sql"
	"time"

	fastapp "github.com/katalabut/fast-app"
	"github.com/katalabut/fast-app/configloader"
	"github.com/katalabut/fast-app/health"
	"github.com/katalabut/fast-app/health/checks"
	"github.com/katalabut/fast-app/logger"
	"github.com/katalabut/fast-app/service"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type Config struct {
	App         fastapp.Config
	DebugServer service.DebugServer
	Database    DatabaseConfig
}

type DatabaseConfig struct {
	URL string `default:"postgres://user:password@localhost/dbname?sslmode=disable"`
}

type APIService struct {
	db    *sql.DB
	ready bool
}

func NewAPIService(db *sql.DB) *APIService {
	return &APIService{
		db: db,
	}
}

func (s *APIService) Run(ctx context.Context) error {
	logger.Info(ctx, "API Service starting...")
	
	// Simulate initialization
	time.Sleep(1 * time.Second)
	s.SetReady(true)
	logger.Info(ctx, "API Service is ready")
	
	// Keep running until context is cancelled
	<-ctx.Done()
	return nil
}

func (s *APIService) Shutdown(ctx context.Context) error {
	logger.Info(ctx, "API Service shutting down...")
	s.SetReady(false)
	return nil
}

// HealthChecks implements health.HealthProvider
func (s *APIService) HealthChecks() []health.HealthChecker {
	return []health.HealthChecker{
		checks.NewDatabaseCheck("api-database", s.db),
		health.NewCustomCheck("api-readiness", func(ctx context.Context) health.HealthResult {
			if s.IsReady() {
				return health.NewHealthyResult("API service is ready to handle requests")
			}
			return health.NewUnhealthyResult("API service is not ready")
		}),
		health.NewCustomCheck("api-business-logic", s.checkBusinessLogic),
	}
}

func (s *APIService) checkBusinessLogic(ctx context.Context) health.HealthResult {
	// Simulate some business logic check
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables").Scan(&count)
	if err != nil {
		return health.NewUnhealthyResult("Failed to query database").
			WithDetails("error", err.Error())
	}
	
	if count < 1 {
		return health.NewDegradedResult("Database seems empty").
			WithDetails("table_count", count)
	}
	
	return health.NewHealthyResult("Business logic check passed").
		WithDetails("table_count", count)
}

// SetReady implements health.ReadinessController
func (s *APIService) SetReady(ready bool) {
	s.ready = ready
}

// IsReady implements health.ReadinessController
func (s *APIService) IsReady() bool {
	return s.ready
}

type WorkerService struct {
	ready bool
}

func NewWorkerService() *WorkerService {
	return &WorkerService{}
}

func (s *WorkerService) Run(ctx context.Context) error {
	logger.Info(ctx, "Worker Service starting...")
	
	// Simulate longer initialization
	time.Sleep(3 * time.Second)
	s.SetReady(true)
	logger.Info(ctx, "Worker Service is ready")
	
	// Simulate work
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			logger.Info(ctx, "Worker processing batch...")
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *WorkerService) Shutdown(ctx context.Context) error {
	logger.Info(ctx, "Worker Service shutting down...")
	s.SetReady(false)
	return nil
}

// HealthChecks implements health.HealthProvider
func (s *WorkerService) HealthChecks() []health.HealthChecker {
	return []health.HealthChecker{
		health.NewCustomCheck("worker-readiness", func(ctx context.Context) health.HealthResult {
			if s.IsReady() {
				return health.NewHealthyResult("Worker service is processing jobs")
			}
			return health.NewUnhealthyResult("Worker service is not ready")
		}),
	}
}

func (s *WorkerService) SetReady(ready bool) {
	s.ready = ready
}

func (s *WorkerService) IsReady() bool {
	return s.ready
}

func main() {
	cfg, err := configloader.New[Config]()
	if err != nil {
		logger.Fatal(context.Background(), "failed to load config:", err)
	}

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		logger.Fatal(context.Background(), "failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Warn(context.Background(), "Database connection failed, continuing without DB", "error", err)
		db = nil
	}

	// Create services
	var apiService *APIService
	if db != nil {
		apiService = NewAPIService(db)
	}
	workerService := NewWorkerService()

	// Global health checks
	var globalChecks []health.HealthChecker
	globalChecks = append(globalChecks, 
		checks.NewHTTPCheck("google", "https://www.google.com"),
		checks.NewHTTPCheck("httpbin", "https://httpbin.org/status/200"),
	)

	// Create application
	app := fastapp.New(cfg.App).
		WithHealthChecks(globalChecks...)

	// Add debug service
	app.Add(service.NewDefaultDebugService(cfg.DebugServer))

	// Add business services
	if apiService != nil {
		app.Add(apiService)
	}
	app.Add(workerService)

	// Set application as ready
	app.SetReady(true)
	
	logger.Info(context.Background(), "Starting application with health checks enabled")
	logger.Info(context.Background(), "Health endpoints available at:")
	logger.Info(context.Background(), "  - Liveness:  http://localhost:8080/health/live")
	logger.Info(context.Background(), "  - Readiness: http://localhost:8080/health/ready") 
	logger.Info(context.Background(), "  - Detailed:  http://localhost:8080/health/checks")
	
	app.Start()
}

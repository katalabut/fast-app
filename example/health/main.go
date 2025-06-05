package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	fastapp "github.com/katalabut/fast-app"
	"github.com/katalabut/fast-app/config"
	"github.com/katalabut/fast-app/configloader"
	"github.com/katalabut/fast-app/health"
	"github.com/katalabut/fast-app/health/checks"
	"github.com/katalabut/fast-app/logger"
	_ "github.com/lib/pq" // PostgreSQL driver for example
)

type AppConfig struct {
	App      config.App
	Database DatabaseConfig
	External ExternalConfig
}

type DatabaseConfig struct {
	URL string `default:"postgres://user:password@localhost:5432/healthdemo?sslmode=disable"`
}

type ExternalConfig struct {
	APIURL    string `default:"https://httpbin.org/status/200"`
	BackupURL string `default:"https://www.google.com"`
}

// HealthDemoService demonstrates various health check patterns
type HealthDemoService struct {
	name   string
	ready  bool
	status string
	db     *sql.DB
}

func NewHealthDemoService(name string, db *sql.DB) *HealthDemoService {
	return &HealthDemoService{
		name:   name,
		ready:  false,
		status: "initializing",
		db:     db,
	}
}

func (s *HealthDemoService) Run(ctx context.Context) error {
	logger.Info(ctx, "🚀 Starting Health Demo Service", "service", s.name)

	// Simulate initialization process
	s.simulateInitialization(ctx)

	// Start health status simulation
	go s.simulateHealthChanges(ctx)

	// Keep running
	<-ctx.Done()
	logger.Info(ctx, "Health Demo Service shutting down", "service", s.name)
	return nil
}

func (s *HealthDemoService) Shutdown(ctx context.Context) error {
	s.ready = false
	s.status = "shutting_down"
	logger.Info(ctx, "Health Demo Service cleanup completed", "service", s.name)
	return nil
}

// simulateInitialization demonstrates gradual service startup
func (s *HealthDemoService) simulateInitialization(ctx context.Context) {
	logger.Info(ctx, "Initializing service components...")

	// Phase 1: Basic initialization
	s.status = "loading_config"
	time.Sleep(1 * time.Second)
	logger.Info(ctx, "✅ Configuration loaded")

	// Phase 2: Database connection
	s.status = "connecting_database"
	time.Sleep(2 * time.Second)
	logger.Info(ctx, "✅ Database connected")

	// Phase 3: External dependencies
	s.status = "checking_dependencies"
	time.Sleep(1 * time.Second)
	logger.Info(ctx, "✅ Dependencies verified")

	// Phase 4: Ready to serve
	s.status = "ready"
	s.ready = true
	logger.Info(ctx, "🎉 Service is ready to serve traffic")
}

// simulateHealthChanges demonstrates dynamic health status changes
func (s *HealthDemoService) simulateHealthChanges(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Randomly simulate health issues
			if rand.Float32() < 0.2 { // 20% chance of temporary issue
				s.simulateTemporaryIssue(ctx)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *HealthDemoService) simulateTemporaryIssue(ctx context.Context) {
	originalStatus := s.status
	originalReady := s.ready

	// Simulate different types of issues
	issues := []string{"high_latency", "memory_pressure", "connection_pool_exhausted"}
	issue := issues[rand.Intn(len(issues))]

	logger.WarnKV(ctx, "Simulating temporary health issue",
		"issue", issue,
		"duration", "10s")

	s.status = issue
	s.ready = false

	// Recover after 10 seconds
	time.Sleep(10 * time.Second)

	s.status = originalStatus
	s.ready = originalReady

	logger.InfoKV(ctx, "Recovered from health issue",
		"issue", issue)
}

// HealthChecks implements the HealthProvider interface
func (s *HealthDemoService) HealthChecks() []health.HealthChecker {
	return []health.HealthChecker{
		// Custom health check for service readiness
		health.NewCustomCheck("service-readiness", func(ctx context.Context) health.HealthResult {
			if !s.ready {
				return health.NewUnhealthyResult("Service is not ready").
					WithDetails("status", s.status).
					WithDetails("ready", s.ready)
			}

			if s.status != "ready" {
				return health.NewDegradedResult("Service is experiencing issues").
					WithDetails("status", s.status).
					WithDetails("issue_type", s.status)
			}

			return health.NewHealthyResult("Service is ready and healthy").
				WithDetails("status", s.status)
		}),

		// Custom health check for business logic
		health.NewCustomCheck("business-logic", func(ctx context.Context) health.HealthResult {
			// Simulate business logic health check
			start := time.Now()

			// Simulate some business logic validation
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

			duration := time.Since(start)

			// Check if business logic is performing well
			if duration > 50*time.Millisecond {
				return health.NewDegradedResult("Business logic is slow").
					WithDetails("duration_ms", duration.Milliseconds()).
					WithDetails("threshold_ms", 50).
					WithDuration(duration)
			}

			return health.NewHealthyResult("Business logic is performing well").
				WithDetails("duration_ms", duration.Milliseconds()).
				WithDuration(duration)
		}),
	}
}

// SetReady implements the ReadinessController interface
func (s *HealthDemoService) SetReady(ready bool) {
	s.ready = ready
	if ready {
		s.status = "ready"
	} else {
		s.status = "not_ready"
	}
}

// IsReady implements the ReadinessController interface
func (s *HealthDemoService) IsReady() bool {
	return s.ready
}

func main() {
	// Load configuration
	cfg, err := configloader.New[AppConfig](
		configloader.WithFileFromEnv("config.yaml"),
	)
	if err != nil {
		logger.Fatal(context.Background(), "Failed to load configuration", "error", err)
	}

	// Initialize database connection (for demonstration)
	// In real applications, handle connection errors appropriately
	var db *sql.DB
	if cfg.Database.URL != "" {
		db, err = sql.Open("postgres", cfg.Database.URL)
		if err != nil {
			logger.Warn(context.Background(), "Failed to connect to database (demo will continue)", "error", err)
		} else {
			defer db.Close()
		}
	}

	// Create demo service
	demoService := NewHealthDemoService("health-demo", db)

	// Create global health checks
	var globalChecks []health.HealthChecker

	// HTTP health check - external API
	httpCheck := checks.NewHTTPCheck("external-api", cfg.External.APIURL)
	globalChecks = append(globalChecks, httpCheck)

	// HTTP health check with custom options
	backupCheck := checks.NewHTTPCheckWithOptions("backup-service", cfg.External.BackupURL,
		checks.HTTPOptions{
			Timeout:        10 * time.Second,
			ExpectedStatus: http.StatusOK,
			Method:         http.MethodGet,
		})
	globalChecks = append(globalChecks, backupCheck)

	// Database health check (if database is available)
	if db != nil {
		dbCheck := checks.NewDatabaseCheckWithOptions("postgres", db,
			checks.DatabaseOptions{
				PingTimeout: 5 * time.Second,
				Query:       "SELECT 1", // Custom query instead of ping
			})
		globalChecks = append(globalChecks, dbCheck)
	}

	// Custom global health check
	systemCheck := health.NewCustomCheck("system-resources", func(ctx context.Context) health.HealthResult {
		// Simulate system resource check
		start := time.Now()

		// Simulate checking CPU, memory, disk, etc.
		cpuUsage := rand.Float64() * 100
		memoryUsage := rand.Float64() * 100
		diskUsage := rand.Float64() * 100

		duration := time.Since(start)

		result := health.NewHealthyResult("System resources are healthy").
			WithDetails("cpu_usage_percent", fmt.Sprintf("%.1f", cpuUsage)).
			WithDetails("memory_usage_percent", fmt.Sprintf("%.1f", memoryUsage)).
			WithDetails("disk_usage_percent", fmt.Sprintf("%.1f", diskUsage)).
			WithDuration(duration)

		// Check for resource pressure
		if cpuUsage > 90 || memoryUsage > 90 {
			result = health.NewDegradedResult("High resource usage detected").
				WithDetails("cpu_usage_percent", fmt.Sprintf("%.1f", cpuUsage)).
				WithDetails("memory_usage_percent", fmt.Sprintf("%.1f", memoryUsage)).
				WithDetails("disk_usage_percent", fmt.Sprintf("%.1f", diskUsage)).
				WithDuration(duration)
		}

		if diskUsage > 95 {
			result = health.NewUnhealthyResult("Critical disk usage").
				WithDetails("disk_usage_percent", fmt.Sprintf("%.1f", diskUsage)).
				WithDuration(duration)
		}

		return result
	})
	globalChecks = append(globalChecks, systemCheck)

	// Create and configure application
	app := fastapp.New(cfg.App, fastapp.WithVersion("1.0.0"))

	// Add global health checks
	app.WithHealthChecks(globalChecks...)

	// Add service (its health checks will be automatically registered)
	app.Add(demoService)

	// Set application as ready (this affects the readiness probe)
	app.SetReady(true)

	logger.Info(context.Background(), "🎯 Health Checks Demo Application")
	logger.Info(context.Background(), "This demo shows comprehensive health check patterns:")
	logger.Info(context.Background(), "• Service-specific health checks")
	logger.Info(context.Background(), "• Global health checks (HTTP, Database, System)")
	logger.Info(context.Background(), "• Custom health checks with business logic")
	logger.Info(context.Background(), "• Dynamic health status changes")
	logger.Info(context.Background(), "• Readiness control during startup")
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "🏥 Health check endpoints:")
	logger.Info(context.Background(), fmt.Sprintf("   • Liveness:  http://localhost:%d/health/live", cfg.App.Observability.Port))
	logger.Info(context.Background(), fmt.Sprintf("   • Readiness: http://localhost:%d/health/ready", cfg.App.Observability.Port))
	logger.Info(context.Background(), fmt.Sprintf("   • Detailed:  http://localhost:%d/health/checks", cfg.App.Observability.Port))
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "📊 Other endpoints:")
	logger.Info(context.Background(), fmt.Sprintf("   • Metrics:   http://localhost:%d/metrics", cfg.App.Observability.Port))
	logger.Info(context.Background(), fmt.Sprintf("   • Profiling: http://localhost:%d/debug/pprof/", cfg.App.Observability.Port))

	app.Start()
}

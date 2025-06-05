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
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// AppConfig demonstrates complete application configuration
type AppConfig struct {
	// FastApp framework configuration
	App config.App

	// Application-specific configuration
	Database DatabaseConfig
	Redis    RedisConfig
	External ExternalConfig
	Business BusinessConfig
}

type DatabaseConfig struct {
	URL            string        `default:"postgres://user:password@localhost:5432/completedemo?sslmode=disable"`
	MaxConnections int           `default:"25"`
	ConnectTimeout time.Duration `default:"10s"`
	QueryTimeout   time.Duration `default:"30s"`
}

type RedisConfig struct {
	Address string `default:"localhost:6379"`
	Enabled bool   `default:"false"`
}

type ExternalConfig struct {
	APIURL   string        `default:"https://httpbin.org/status/200"`
	Timeout  time.Duration `default:"30s"`
	RetryMax int           `default:"3"`
}

type BusinessConfig struct {
	MaxOrderValue    float64       `default:"10000.0"`
	ProcessingDelay  time.Duration `default:"100ms"`
	EnableNewFeature bool          `default:"false"`
}

// CompleteService demonstrates a full-featured service with all FastApp capabilities
type CompleteService struct {
	name   string
	config *AppConfig
	db     *sql.DB
	ready  bool
	status string

	// Metrics
	requestsTotal    prometheus.Counter
	requestDuration  *prometheus.HistogramVec
	activeOrders     prometheus.Gauge
	orderValue       *prometheus.HistogramVec
	processingErrors *prometheus.CounterVec
}

func NewCompleteService(name string, cfg *AppConfig, db *sql.DB) *CompleteService {
	return &CompleteService{
		name:   name,
		config: cfg,
		db:     db,
		ready:  false,
		status: "initializing",

		// Initialize metrics
		requestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "business_requests_total",
			Help: "Total number of business requests processed",
		}),

		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "business_request_duration_seconds",
				Help:    "Business request processing duration",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"operation", "status"},
		),

		activeOrders: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "active_orders_count",
			Help: "Current number of active orders being processed",
		}),

		orderValue: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "order_value_dollars",
				Help:    "Distribution of order values",
				Buckets: []float64{10, 50, 100, 500, 1000, 5000, 10000},
			},
			[]string{"currency"},
		),

		processingErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "processing_errors_total",
				Help: "Total number of processing errors by type",
			},
			[]string{"error_type", "operation"},
		),
	}
}

func (s *CompleteService) Run(ctx context.Context) error {
	logger.InfoKV(ctx, "🚀 Starting Complete Demo Service",
		"service", s.name,
		"version", "1.0.0")

	// Demonstrate configuration usage
	s.logConfiguration(ctx)

	// Simulate service initialization
	s.initializeService(ctx)

	// Start business operations simulation
	go s.simulateBusinessOperations(ctx)
	go s.simulateOrderProcessing(ctx)

	// Keep running
	<-ctx.Done()
	logger.Info(ctx, "Complete Demo Service shutting down", "service", s.name)
	return nil
}

func (s *CompleteService) Shutdown(ctx context.Context) error {
	s.ready = false
	s.status = "shutting_down"
	logger.Info(ctx, "Complete Demo Service cleanup completed", "service", s.name)
	return nil
}

func (s *CompleteService) logConfiguration(ctx context.Context) {
	logger.Info(ctx, "=== Service Configuration ===")

	// Log FastApp configuration
	logger.InfoKV(ctx, "FastApp Configuration",
		"logger_level", s.config.App.Logger.Level,
		"dev_mode", s.config.App.Logger.DevMode,
		"observability_port", s.config.App.Observability.Port,
		"metrics_enabled", s.config.App.Observability.Metrics.Enabled,
		"health_enabled", s.config.App.Observability.Health.Enabled)

	// Log business configuration
	logger.InfoKV(ctx, "Business Configuration",
		"max_order_value", s.config.Business.MaxOrderValue,
		"processing_delay", s.config.Business.ProcessingDelay,
		"new_feature_enabled", s.config.Business.EnableNewFeature,
		"external_api_timeout", s.config.External.Timeout,
		"database_timeout", s.config.Database.ConnectTimeout)
}

func (s *CompleteService) initializeService(ctx context.Context) {
	logger.Info(ctx, "Initializing service components...")

	// Phase 1: Configuration validation
	s.status = "validating_config"
	time.Sleep(500 * time.Millisecond)
	logger.Info(ctx, "✅ Configuration validated")

	// Phase 2: Database connection
	s.status = "connecting_database"
	time.Sleep(1 * time.Second)
	if s.db != nil {
		logger.Info(ctx, "✅ Database connection established")
	} else {
		logger.Warn(ctx, "⚠️ Database not configured (demo mode)")
	}

	// Phase 3: External dependencies
	s.status = "checking_dependencies"
	time.Sleep(800 * time.Millisecond)
	logger.Info(ctx, "✅ External dependencies verified")

	// Phase 4: Business logic initialization
	s.status = "initializing_business_logic"
	time.Sleep(600 * time.Millisecond)
	logger.Info(ctx, "✅ Business logic initialized")

	// Phase 5: Ready to serve
	s.status = "ready"
	s.ready = true
	logger.InfoKV(ctx, "🎉 Service is ready to serve traffic",
		"initialization_time", "3.9s",
		"status", s.status)
}

func (s *CompleteService) simulateBusinessOperations(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	operations := []string{"user_registration", "profile_update", "password_reset", "data_export"}

	for {
		select {
		case <-ticker.C:
			if !s.ready {
				continue
			}

			operation := operations[rand.Intn(len(operations))]
			s.processBusinessOperation(ctx, operation)

		case <-ctx.Done():
			return
		}
	}
}

func (s *CompleteService) processBusinessOperation(ctx context.Context, operation string) {
	start := time.Now()
	s.requestsTotal.Inc()

	// Add processing delay from configuration
	time.Sleep(s.config.Business.ProcessingDelay)

	// Simulate additional processing time
	processingTime := time.Duration(rand.Intn(200)) * time.Millisecond
	time.Sleep(processingTime)

	duration := time.Since(start)
	status := "success"

	// Simulate occasional errors
	if rand.Float32() < 0.1 { // 10% error rate
		status = "error"
		errorType := "validation_error"
		if rand.Float32() < 0.3 {
			errorType = "timeout_error"
		}
		s.processingErrors.WithLabelValues(errorType, operation).Inc()
		logger.WarnKV(ctx, "Business operation failed",
			"operation", operation,
			"error_type", errorType,
			"duration_ms", duration.Milliseconds())
	} else {
		logger.DebugKV(ctx, "Business operation completed",
			"operation", operation,
			"duration_ms", duration.Milliseconds(),
			"status", status)
	}

	s.requestDuration.WithLabelValues(operation, status).Observe(duration.Seconds())
}

func (s *CompleteService) simulateOrderProcessing(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	activeOrders := 0

	for {
		select {
		case <-ticker.C:
			if !s.ready {
				continue
			}

			// Simulate new orders
			newOrders := rand.Intn(3) + 1
			activeOrders += newOrders

			// Process some orders
			processedOrders := rand.Intn(activeOrders + 1)
			activeOrders -= processedOrders

			if activeOrders < 0 {
				activeOrders = 0
			}

			s.activeOrders.Set(float64(activeOrders))

			// Record order values
			for i := 0; i < processedOrders; i++ {
				orderValue := rand.Float64() * s.config.Business.MaxOrderValue
				s.orderValue.WithLabelValues("USD").Observe(orderValue)
			}

			if processedOrders > 0 {
				logger.DebugKV(ctx, "Order processing update",
					"new_orders", newOrders,
					"processed_orders", processedOrders,
					"active_orders", activeOrders)
			}

		case <-ctx.Done():
			return
		}
	}
}

// HealthChecks implements the HealthProvider interface
func (s *CompleteService) HealthChecks() []health.HealthChecker {
	return []health.HealthChecker{
		// Service readiness check
		health.NewCustomCheck("service-readiness", func(ctx context.Context) health.HealthResult {
			if !s.ready {
				return health.NewUnhealthyResult("Service is not ready").
					WithDetails("status", s.status).
					WithDetails("ready", s.ready)
			}

			return health.NewHealthyResult("Service is ready and operational").
				WithDetails("status", s.status).
				WithDetails("uptime", time.Since(time.Now()).String())
		}),

		// Business logic health check
		health.NewCustomCheck("business-logic", func(ctx context.Context) health.HealthResult {
			start := time.Now()

			// Simulate business logic validation
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

			duration := time.Since(start)

			if duration > 30*time.Millisecond {
				return health.NewDegradedResult("Business logic is slower than expected").
					WithDetails("duration_ms", duration.Milliseconds()).
					WithDetails("threshold_ms", 30).
					WithDuration(duration)
			}

			return health.NewHealthyResult("Business logic is performing well").
				WithDetails("duration_ms", duration.Milliseconds()).
				WithDetails("max_order_value", s.config.Business.MaxOrderValue).
				WithDuration(duration)
		}),

		// Configuration validation check
		health.NewCustomCheck("configuration", func(ctx context.Context) health.HealthResult {
			issues := []string{}

			if s.config.Business.MaxOrderValue <= 0 {
				issues = append(issues, "invalid max order value")
			}
			if s.config.Business.ProcessingDelay < 0 {
				issues = append(issues, "invalid processing delay")
			}
			if s.config.External.Timeout <= 0 {
				issues = append(issues, "invalid external timeout")
			}

			if len(issues) > 0 {
				return health.NewUnhealthyResult("Configuration validation failed").
					WithDetails("issues", issues).
					WithDetails("issue_count", len(issues))
			}

			return health.NewHealthyResult("Configuration is valid").
				WithDetails("validated_at", time.Now().Format(time.RFC3339))
		}),
	}
}

// SetReady implements the ReadinessController interface
func (s *CompleteService) SetReady(ready bool) {
	s.ready = ready
	if ready {
		s.status = "ready"
	} else {
		s.status = "not_ready"
	}
}

// IsReady implements the ReadinessController interface
func (s *CompleteService) IsReady() bool {
	return s.ready
}

func main() {
	// Load configuration with multiple sources
	cfg, err := configloader.New[AppConfig](
		configloader.WithFileFromEnv("config.yaml", "config.json"),
	)
	if err != nil {
		logger.Fatal(context.Background(), "Failed to load configuration", "error", err)
	}

	// Initialize database connection (optional for demo)
	var db *sql.DB
	if cfg.Database.URL != "" {
		db, err = sql.Open("postgres", cfg.Database.URL)
		if err != nil {
			logger.Warn(context.Background(), "Failed to connect to database (demo will continue)", "error", err)
		} else {
			defer db.Close()
		}
	}

	// Create complete service
	completeService := NewCompleteService("complete-demo", cfg, db)

	// Create global health checks
	var globalChecks []health.HealthChecker

	// HTTP health check for external API
	httpCheck := checks.NewHTTPCheckWithOptions("external-api", cfg.External.APIURL,
		checks.HTTPOptions{
			Timeout:        cfg.External.Timeout,
			ExpectedStatus: http.StatusOK,
			Method:         http.MethodGet,
		})
	globalChecks = append(globalChecks, httpCheck)

	// Database health check (if available)
	if db != nil {
		dbCheck := checks.NewDatabaseCheckWithOptions("postgres", db,
			checks.DatabaseOptions{
				PingTimeout: cfg.Database.ConnectTimeout,
				Query:       "SELECT version()",
			})
		globalChecks = append(globalChecks, dbCheck)
	}

	// System resource check
	systemCheck := health.NewCustomCheck("system-resources", func(ctx context.Context) health.HealthResult {
		start := time.Now()

		// Simulate system resource monitoring
		cpuUsage := rand.Float64() * 100
		memoryUsage := rand.Float64() * 100
		diskUsage := rand.Float64() * 100

		duration := time.Since(start)

		result := health.NewHealthyResult("System resources are healthy").
			WithDetails("cpu_usage_percent", fmt.Sprintf("%.1f", cpuUsage)).
			WithDetails("memory_usage_percent", fmt.Sprintf("%.1f", memoryUsage)).
			WithDetails("disk_usage_percent", fmt.Sprintf("%.1f", diskUsage)).
			WithDuration(duration)

		if cpuUsage > 90 || memoryUsage > 90 {
			result = health.NewDegradedResult("High resource usage detected").
				WithDetails("cpu_usage_percent", fmt.Sprintf("%.1f", cpuUsage)).
				WithDetails("memory_usage_percent", fmt.Sprintf("%.1f", memoryUsage)).
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
	app.Add(completeService)

	// Set application as ready
	app.SetReady(true)

	// Display startup information
	logger.Info(context.Background(), "🎯 Complete FastApp Demo Application")
	logger.Info(context.Background(), "This demo showcases ALL FastApp capabilities:")
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "📋 Configuration Management:")
	logger.Info(context.Background(), "   • YAML/JSON file loading")
	logger.Info(context.Background(), "   • Environment variable override")
	logger.Info(context.Background(), "   • Type-safe configuration")
	logger.Info(context.Background(), "   • Default values and validation")
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "📝 Structured Logging:")
	logger.Info(context.Background(), "   • Multiple log levels")
	logger.Info(context.Background(), "   • Context-aware logging")
	logger.Info(context.Background(), "   • Structured key-value pairs")
	logger.Info(context.Background(), "   • Development and production modes")
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "📊 Prometheus Metrics:")
	logger.Info(context.Background(), "   • Counters (requests, errors)")
	logger.Info(context.Background(), "   • Gauges (active orders)")
	logger.Info(context.Background(), "   • Histograms (duration, order values)")
	logger.Info(context.Background(), "   • Multi-dimensional labels")
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "🏥 Health Checks:")
	logger.Info(context.Background(), "   • Service-specific checks")
	logger.Info(context.Background(), "   • Global health checks")
	logger.Info(context.Background(), "   • HTTP dependency monitoring")
	logger.Info(context.Background(), "   • Database connectivity")
	logger.Info(context.Background(), "   • System resource monitoring")
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "🌐 Observability Endpoints:")
	logger.Info(context.Background(), fmt.Sprintf("   • Liveness:  http://localhost:%d/health/live", cfg.App.Observability.Port))
	logger.Info(context.Background(), fmt.Sprintf("   • Readiness: http://localhost:%d/health/ready", cfg.App.Observability.Port))
	logger.Info(context.Background(), fmt.Sprintf("   • Health:    http://localhost:%d/health/checks", cfg.App.Observability.Port))
	logger.Info(context.Background(), fmt.Sprintf("   • Metrics:   http://localhost:%d/metrics", cfg.App.Observability.Port))
	logger.Info(context.Background(), fmt.Sprintf("   • Profiling: http://localhost:%d/debug/pprof/", cfg.App.Observability.Port))
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "💡 Try these commands:")
	logger.Info(context.Background(), fmt.Sprintf("   curl http://localhost:%d/health/checks | jq .", cfg.App.Observability.Port))
	logger.Info(context.Background(), fmt.Sprintf("   curl http://localhost:%d/metrics | grep business_", cfg.App.Observability.Port))

	app.Start()
}

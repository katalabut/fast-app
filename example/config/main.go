package main

import (
	"context"
	"fmt"
	"time"

	fastapp "github.com/katalabut/fast-app"
	"github.com/katalabut/fast-app/config"
	"github.com/katalabut/fast-app/configloader"
	"github.com/katalabut/fast-app/logger"
)

// AppConfig demonstrates the main application configuration structure
type AppConfig struct {
	// FastApp configuration
	App config.App

	// Custom application configuration sections
	Database DatabaseConfig
	Redis    RedisConfig
	API      APIConfig
	Features FeatureFlags
}

// DatabaseConfig shows how to configure database connections
type DatabaseConfig struct {
	URL             string        `default:"postgres://user:password@localhost:5432/mydb?sslmode=disable"`
	MaxConnections  int           `default:"25"`
	MaxIdleTime     time.Duration `default:"15m"`
	ConnectTimeout  time.Duration `default:"10s"`
	QueryTimeout    time.Duration `default:"30s"`
	EnableLogging   bool          `default:"false"`
	MigrationsPath  string        `default:"./migrations"`
}

// RedisConfig demonstrates Redis configuration
type RedisConfig struct {
	Address     string        `default:"localhost:6379"`
	Password    string        `default:""`
	Database    int           `default:"0"`
	PoolSize    int           `default:"10"`
	DialTimeout time.Duration `default:"5s"`
	ReadTimeout time.Duration `default:"3s"`
	Enabled     bool          `default:"true"`
}

// APIConfig shows external API configuration
type APIConfig struct {
	BaseURL       string            `default:"https://api.example.com"`
	Timeout       time.Duration     `default:"30s"`
	RetryAttempts int               `default:"3"`
	RetryDelay    time.Duration     `default:"1s"`
	APIKey        string            `default:""`
	Headers       map[string]string `default:"{}"`
	RateLimit     int               `default:"100"`
}

// FeatureFlags demonstrates feature toggle configuration
type FeatureFlags struct {
	EnableNewUI      bool `default:"false"`
	EnableBetaAPI    bool `default:"false"`
	EnableCaching    bool `default:"true"`
	EnableMetrics    bool `default:"true"`
	MaxFileSize      int  `default:"10485760"` // 10MB
	MaintenanceMode  bool `default:"false"`
}

type ConfigDemoService struct {
	config *AppConfig
}

func NewConfigDemoService(cfg *AppConfig) *ConfigDemoService {
	return &ConfigDemoService{config: cfg}
}

func (s *ConfigDemoService) Run(ctx context.Context) error {
	logger.Info(ctx, "🚀 Starting Configuration Demo Service")

	// Display loaded configuration
	s.displayConfiguration(ctx)

	// Demonstrate configuration validation
	s.validateConfiguration(ctx)

	// Demonstrate configuration usage patterns
	s.demonstrateUsagePatterns(ctx)

	// Keep running
	<-ctx.Done()
	logger.Info(ctx, "Configuration Demo Service shutting down")
	return nil
}

func (s *ConfigDemoService) Shutdown(ctx context.Context) error {
	logger.Info(ctx, "Configuration Demo Service cleanup completed")
	return nil
}

func (s *ConfigDemoService) displayConfiguration(ctx context.Context) {
	logger.Info(ctx, "=== Current Configuration ===")

	// FastApp configuration
	logger.InfoKV(ctx, "FastApp Logger Configuration",
		"app_name", s.config.App.Logger.AppName,
		"level", s.config.App.Logger.Level,
		"dev_mode", s.config.App.Logger.DevMode)

	logger.InfoKV(ctx, "FastApp Observability Configuration",
		"enabled", s.config.App.Observability.Enabled,
		"port", s.config.App.Observability.Port,
		"metrics_enabled", s.config.App.Observability.Metrics.Enabled,
		"health_enabled", s.config.App.Observability.Health.Enabled)

	// Database configuration
	logger.InfoKV(ctx, "Database Configuration",
		"url", maskSensitiveData(s.config.Database.URL),
		"max_connections", s.config.Database.MaxConnections,
		"max_idle_time", s.config.Database.MaxIdleTime,
		"connect_timeout", s.config.Database.ConnectTimeout,
		"logging_enabled", s.config.Database.EnableLogging)

	// Redis configuration
	logger.InfoKV(ctx, "Redis Configuration",
		"address", s.config.Redis.Address,
		"database", s.config.Redis.Database,
		"pool_size", s.config.Redis.PoolSize,
		"enabled", s.config.Redis.Enabled,
		"has_password", s.config.Redis.Password != "")

	// API configuration
	logger.InfoKV(ctx, "External API Configuration",
		"base_url", s.config.API.BaseURL,
		"timeout", s.config.API.Timeout,
		"retry_attempts", s.config.API.RetryAttempts,
		"rate_limit", s.config.API.RateLimit,
		"has_api_key", s.config.API.APIKey != "")

	// Feature flags
	logger.InfoKV(ctx, "Feature Flags",
		"new_ui", s.config.Features.EnableNewUI,
		"beta_api", s.config.Features.EnableBetaAPI,
		"caching", s.config.Features.EnableCaching,
		"metrics", s.config.Features.EnableMetrics,
		"maintenance_mode", s.config.Features.MaintenanceMode,
		"max_file_size_mb", s.config.Features.MaxFileSize/1024/1024)
}

func (s *ConfigDemoService) validateConfiguration(ctx context.Context) {
	logger.Info(ctx, "=== Configuration Validation ===")

	var issues []string

	// Validate database configuration
	if s.config.Database.MaxConnections <= 0 {
		issues = append(issues, "Database max connections must be positive")
	}
	if s.config.Database.ConnectTimeout <= 0 {
		issues = append(issues, "Database connect timeout must be positive")
	}

	// Validate Redis configuration
	if s.config.Redis.Enabled && s.config.Redis.Address == "" {
		issues = append(issues, "Redis address is required when Redis is enabled")
	}
	if s.config.Redis.PoolSize <= 0 {
		issues = append(issues, "Redis pool size must be positive")
	}

	// Validate API configuration
	if s.config.API.BaseURL == "" {
		issues = append(issues, "API base URL is required")
	}
	if s.config.API.RetryAttempts < 0 {
		issues = append(issues, "API retry attempts cannot be negative")
	}

	// Validate feature flags
	if s.config.Features.MaxFileSize <= 0 {
		issues = append(issues, "Max file size must be positive")
	}

	if len(issues) > 0 {
		logger.WarnKV(ctx, "Configuration validation issues found",
			"issues", issues,
			"count", len(issues))
	} else {
		logger.Info(ctx, "✅ Configuration validation passed")
	}
}

func (s *ConfigDemoService) demonstrateUsagePatterns(ctx context.Context) {
	logger.Info(ctx, "=== Configuration Usage Patterns ===")

	// Pattern 1: Feature flags
	if s.config.Features.EnableNewUI {
		logger.Info(ctx, "🎨 New UI is enabled - using modern interface")
	} else {
		logger.Info(ctx, "📱 Using classic UI")
	}

	// Pattern 2: Conditional service initialization
	if s.config.Redis.Enabled {
		logger.InfoKV(ctx, "🔄 Redis caching is enabled",
			"address", s.config.Redis.Address,
			"database", s.config.Redis.Database)
	} else {
		logger.Info(ctx, "💾 Using in-memory caching")
	}

	// Pattern 3: Timeout configuration
	logger.InfoKV(ctx, "⏱️ Service timeouts configured",
		"database_connect", s.config.Database.ConnectTimeout,
		"database_query", s.config.Database.QueryTimeout,
		"api_request", s.config.API.Timeout,
		"redis_dial", s.config.Redis.DialTimeout)

	// Pattern 4: Environment-specific behavior
	if s.config.App.Logger.DevMode {
		logger.Info(ctx, "🔧 Running in development mode")
	} else {
		logger.Info(ctx, "🚀 Running in production mode")
	}

	// Pattern 5: Maintenance mode
	if s.config.Features.MaintenanceMode {
		logger.Warn(ctx, "🚧 Application is in maintenance mode")
	}
}

// maskSensitiveData masks passwords and sensitive information in URLs
func maskSensitiveData(url string) string {
	// Simple masking for demo purposes
	// In real applications, use proper URL parsing and masking
	if len(url) > 20 {
		return url[:10] + "***MASKED***" + url[len(url)-5:]
	}
	return "***MASKED***"
}

func main() {
	logger.Info(context.Background(), "🎯 Configuration Demo Application")
	logger.Info(context.Background(), "This demo shows how to:")
	logger.Info(context.Background(), "• Load configuration from files and environment variables")
	logger.Info(context.Background(), "• Structure complex application configuration")
	logger.Info(context.Background(), "• Validate configuration values")
	logger.Info(context.Background(), "• Use configuration in your application")
	logger.Info(context.Background(), "")

	// Demonstrate different configuration loading methods
	logger.Info(context.Background(), "📁 Loading configuration...")

	// Method 1: Load from file with environment variable override
	cfg, err := configloader.New[AppConfig](
		configloader.WithFileFromEnv("config.yaml", "config.json"),
	)
	if err != nil {
		logger.Fatal(context.Background(), "Failed to load configuration", "error", err)
	}

	logger.Info(context.Background(), "✅ Configuration loaded successfully")

	// Create demo service
	demoService := NewConfigDemoService(cfg)

	// Create and start application
	app := fastapp.New(cfg.App, fastapp.WithVersion("1.0.0"))
	app.Add(demoService)

	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "📊 Observability endpoints:")
	logger.Info(context.Background(), fmt.Sprintf("   • Metrics:   http://localhost:%d/metrics", cfg.App.Observability.Port))
	logger.Info(context.Background(), fmt.Sprintf("   • Health:    http://localhost:%d/health/checks", cfg.App.Observability.Port))
	logger.Info(context.Background(), fmt.Sprintf("   • Profiling: http://localhost:%d/debug/pprof/", cfg.App.Observability.Port))

	app.Start()
}

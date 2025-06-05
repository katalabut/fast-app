package main

import (
	"context"
	"errors"
	"time"

	fastapp "github.com/katalabut/fast-app"
	"github.com/katalabut/fast-app/config"
	"github.com/katalabut/fast-app/configloader"
	"github.com/katalabut/fast-app/logger"
)

type AppConfig struct {
	App config.App
}

type LoggerDemoService struct {
	name string
}

func NewLoggerDemoService(name string) *LoggerDemoService {
	return &LoggerDemoService{name: name}
}

func (s *LoggerDemoService) Run(ctx context.Context) error {
	logger.Info(ctx, "🚀 Starting Logger Demo Service", "service", s.name)

	// Demonstrate different log levels
	s.demonstrateLogLevels(ctx)

	// Demonstrate structured logging
	s.demonstrateStructuredLogging(ctx)

	// Demonstrate context logging
	s.demonstrateContextLogging(ctx)

	// Demonstrate error logging
	s.demonstrateErrorLogging(ctx)

	// Keep running until context is cancelled
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.InfoKV(ctx, "Service heartbeat", 
				"service", s.name,
				"uptime", time.Since(time.Now()).String(),
				"status", "running")
		case <-ctx.Done():
			logger.Info(ctx, "Logger Demo Service shutting down", "service", s.name)
			return nil
		}
	}
}

func (s *LoggerDemoService) Shutdown(ctx context.Context) error {
	logger.Info(ctx, "Logger Demo Service cleanup completed", "service", s.name)
	return nil
}

func (s *LoggerDemoService) demonstrateLogLevels(ctx context.Context) {
	logger.Info(ctx, "=== Demonstrating Different Log Levels ===")

	// Debug level (only visible when log level is debug)
	logger.Debug(ctx, "This is a debug message - useful for development")
	logger.DebugKV(ctx, "Debug with structured data", 
		"user_id", 12345,
		"action", "login_attempt",
		"ip", "192.168.1.100")

	// Info level - general information
	logger.Info(ctx, "This is an info message - general application flow")
	logger.InfoKV(ctx, "User action completed",
		"user_id", 12345,
		"action", "profile_update",
		"duration_ms", 150)

	// Warning level - something unexpected but not critical
	logger.Warn(ctx, "This is a warning - something unusual happened")
	logger.WarnKV(ctx, "High response time detected",
		"endpoint", "/api/users",
		"response_time_ms", 2500,
		"threshold_ms", 1000)

	// Error level - errors that need attention
	logger.Error(ctx, "This is an error message - something went wrong")
	logger.ErrorKV(ctx, "Database connection failed",
		"database", "postgres",
		"host", "localhost:5432",
		"retry_count", 3)
}

func (s *LoggerDemoService) demonstrateStructuredLogging(ctx context.Context) {
	logger.Info(ctx, "=== Demonstrating Structured Logging ===")

	// Using InfoKV for structured logging
	logger.InfoKV(ctx, "Processing user request",
		"request_id", "req-123456",
		"user_id", 98765,
		"endpoint", "/api/orders",
		"method", "POST",
		"user_agent", "Mozilla/5.0...",
		"ip_address", "203.0.113.42")

	// Logging with different data types
	logger.InfoKV(ctx, "Order processing completed",
		"order_id", "ord-789012",
		"customer_id", 98765,
		"total_amount", 299.99,
		"currency", "USD",
		"items_count", 3,
		"processing_time_ms", 1250,
		"success", true,
		"created_at", time.Now())

	// Logging arrays and complex data
	items := []string{"laptop", "mouse", "keyboard"}
	metadata := map[string]interface{}{
		"source": "web",
		"campaign": "summer_sale",
		"discount_applied": true,
	}

	logger.InfoKV(ctx, "Order details",
		"order_id", "ord-789012",
		"items", items,
		"metadata", metadata)
}

func (s *LoggerDemoService) demonstrateContextLogging(ctx context.Context) {
	logger.Info(ctx, "=== Demonstrating Context-Aware Logging ===")

	// Create context with additional fields
	ctxWithUser := logger.WithFields(ctx, 
		"user_id", 12345,
		"session_id", "sess-abcdef",
		"trace_id", "trace-123456")

	// All subsequent logs will include these fields
	logger.Info(ctxWithUser, "User started checkout process")
	logger.Info(ctxWithUser, "Payment method selected")
	logger.Info(ctxWithUser, "Order confirmation sent")

	// Nested context with additional fields
	ctxWithOrder := logger.WithFields(ctxWithUser,
		"order_id", "ord-789012",
		"payment_method", "credit_card")

	logger.Info(ctxWithOrder, "Payment processing started")
	logger.Info(ctxWithOrder, "Payment completed successfully")
}

func (s *LoggerDemoService) demonstrateErrorLogging(ctx context.Context) {
	logger.Info(ctx, "=== Demonstrating Error Logging ===")

	// Simple error logging
	err := errors.New("connection timeout")
	logger.Error(ctx, "Database operation failed", "error", err)

	// Error with additional context
	logger.ErrorKV(ctx, "Failed to process payment",
		"error", err,
		"order_id", "ord-789012",
		"payment_method", "credit_card",
		"amount", 299.99,
		"retry_count", 3)

	// Wrapped error
	wrappedErr := errors.New("payment gateway returned error: " + err.Error())
	logger.ErrorKV(ctx, "Payment gateway error",
		"error", wrappedErr,
		"original_error", err,
		"gateway", "stripe",
		"transaction_id", "txn-456789")

	// Fatal error (use sparingly - this will exit the application)
	// logger.Fatal(ctx, "Critical system failure - shutting down")
	logger.Info(ctx, "Note: Fatal logs would exit the application immediately")
}

func main() {
	// Load configuration
	cfg, err := configloader.New[AppConfig](
		configloader.WithFileFromEnv("config.yaml"),
	)
	if err != nil {
		logger.Fatal(context.Background(), "Failed to load configuration", "error", err)
	}

	// Create demo service
	demoService := NewLoggerDemoService("logger-demo")

	// Create and start application
	app := fastapp.New(cfg.App, fastapp.WithVersion("1.0.0"))
	app.Add(demoService)

	logger.Info(context.Background(), "🎯 Logger Demo Application")
	logger.Info(context.Background(), "This demo shows various logging capabilities:")
	logger.Info(context.Background(), "• Different log levels (debug, info, warn, error)")
	logger.Info(context.Background(), "• Structured logging with key-value pairs")
	logger.Info(context.Background(), "• Context-aware logging")
	logger.Info(context.Background(), "• Error logging patterns")
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "📊 Observability endpoints:")
	logger.Info(context.Background(), "   • Metrics:   http://localhost:9090/metrics")
	logger.Info(context.Background(), "   • Health:    http://localhost:9090/health/checks")
	logger.Info(context.Background(), "   • Profiling: http://localhost:9090/debug/pprof/")

	app.Start()
}

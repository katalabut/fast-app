package main

import (
	"context"
	"math/rand"
	"time"

	fastapp "github.com/katalabut/fast-app"
	"github.com/katalabut/fast-app/config"
	"github.com/katalabut/fast-app/configloader"
	"github.com/katalabut/fast-app/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type AppConfig struct {
	App config.App
}

// MetricsService demonstrates various Prometheus metrics patterns
type MetricsService struct {
	// Counter metrics - values that only increase
	requestsTotal    prometheus.Counter
	errorsTotal      *prometheus.CounterVec
	
	// Gauge metrics - values that can go up and down
	activeConnections prometheus.Gauge
	queueSize        prometheus.Gauge
	
	// Histogram metrics - distribution of values
	requestDuration  *prometheus.HistogramVec
	responseSize     prometheus.Histogram
	
	// Summary metrics - similar to histogram but with quantiles
	processingTime   *prometheus.SummaryVec
}

func NewMetricsService() *MetricsService {
	return &MetricsService{
		// Counter: Total number of HTTP requests
		requestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		}),

		// CounterVec: Errors by type and endpoint
		errorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "errors_total",
				Help: "Total number of errors by type and endpoint",
			},
			[]string{"type", "endpoint"},
		),

		// Gauge: Current number of active connections
		activeConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Current number of active connections",
		}),

		// Gauge: Current queue size
		queueSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "queue_size",
			Help: "Current number of items in processing queue",
		}),

		// HistogramVec: Request duration by method and status
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets, // Default buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
			},
			[]string{"method", "status"},
		),

		// Histogram: Response size distribution
		responseSize: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000}, // Custom buckets
		}),

		// SummaryVec: Processing time with quantiles
		processingTime: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "processing_time_seconds",
				Help:       "Time spent processing requests",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}, // 50th, 90th, 99th percentiles
			},
			[]string{"operation"},
		),
	}
}

func (s *MetricsService) Run(ctx context.Context) error {
	logger.Info(ctx, "🚀 Starting Metrics Demo Service")

	// Start metrics simulation
	go s.simulateHTTPRequests(ctx)
	go s.simulateConnections(ctx)
	go s.simulateQueueProcessing(ctx)
	go s.simulateBusinessOperations(ctx)

	// Keep running
	<-ctx.Done()
	logger.Info(ctx, "Metrics Demo Service shutting down")
	return nil
}

func (s *MetricsService) Shutdown(ctx context.Context) error {
	logger.Info(ctx, "Metrics Demo Service cleanup completed")
	return nil
}

// simulateHTTPRequests demonstrates counter and histogram metrics
func (s *MetricsService) simulateHTTPRequests(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	methods := []string{"GET", "POST", "PUT", "DELETE"}
	statuses := []string{"200", "400", "404", "500"}
	endpoints := []string{"/api/users", "/api/orders", "/api/products"}

	for {
		select {
		case <-ticker.C:
			// Simulate multiple requests
			for i := 0; i < rand.Intn(5)+1; i++ {
				method := methods[rand.Intn(len(methods))]
				status := statuses[rand.Intn(len(statuses))]
				endpoint := endpoints[rand.Intn(len(endpoints))]

				// Increment total requests counter
				s.requestsTotal.Inc()

				// Simulate request duration (50ms to 2s)
				duration := time.Duration(rand.Intn(1950)+50) * time.Millisecond
				s.requestDuration.WithLabelValues(method, status).Observe(duration.Seconds())

				// Simulate response size (100 bytes to 50KB)
				responseSize := float64(rand.Intn(49900) + 100)
				s.responseSize.Observe(responseSize)

				// Record errors for non-2xx status codes
				if status != "200" {
					errorType := "client_error"
					if status == "500" {
						errorType = "server_error"
					}
					s.errorsTotal.WithLabelValues(errorType, endpoint).Inc()
				}

				logger.DebugKV(ctx, "Simulated HTTP request",
					"method", method,
					"status", status,
					"endpoint", endpoint,
					"duration_ms", duration.Milliseconds(),
					"response_size", responseSize)
			}

		case <-ctx.Done():
			return
		}
	}
}

// simulateConnections demonstrates gauge metrics that go up and down
func (s *MetricsService) simulateConnections(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	currentConnections := 0

	for {
		select {
		case <-ticker.C:
			// Randomly change connection count
			change := rand.Intn(21) - 10 // -10 to +10
			currentConnections += change
			
			// Keep connections non-negative
			if currentConnections < 0 {
				currentConnections = 0
			}
			
			// Limit maximum connections
			if currentConnections > 100 {
				currentConnections = 100
			}

			s.activeConnections.Set(float64(currentConnections))

			logger.DebugKV(ctx, "Updated active connections",
				"connections", currentConnections,
				"change", change)

		case <-ctx.Done():
			return
		}
	}
}

// simulateQueueProcessing demonstrates gauge metrics for queue management
func (s *MetricsService) simulateQueueProcessing(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	queueSize := 0

	for {
		select {
		case <-ticker.C:
			// Simulate items being added to queue
			newItems := rand.Intn(5)
			queueSize += newItems

			// Simulate items being processed from queue
			processedItems := rand.Intn(queueSize + 1)
			queueSize -= processedItems

			// Keep queue size non-negative
			if queueSize < 0 {
				queueSize = 0
			}

			s.queueSize.Set(float64(queueSize))

			logger.DebugKV(ctx, "Queue processing update",
				"queue_size", queueSize,
				"new_items", newItems,
				"processed_items", processedItems)

		case <-ctx.Done():
			return
		}
	}
}

// simulateBusinessOperations demonstrates summary metrics with quantiles
func (s *MetricsService) simulateBusinessOperations(ctx context.Context) {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	operations := []string{"user_registration", "order_processing", "payment_processing", "data_export"}

	for {
		select {
		case <-ticker.C:
			operation := operations[rand.Intn(len(operations))]

			// Simulate different processing times for different operations
			var processingTime time.Duration
			switch operation {
			case "user_registration":
				processingTime = time.Duration(rand.Intn(500)+100) * time.Millisecond // 100-600ms
			case "order_processing":
				processingTime = time.Duration(rand.Intn(1000)+200) * time.Millisecond // 200ms-1.2s
			case "payment_processing":
				processingTime = time.Duration(rand.Intn(2000)+500) * time.Millisecond // 500ms-2.5s
			case "data_export":
				processingTime = time.Duration(rand.Intn(5000)+1000) * time.Millisecond // 1-6s
			}

			s.processingTime.WithLabelValues(operation).Observe(processingTime.Seconds())

			logger.DebugKV(ctx, "Business operation completed",
				"operation", operation,
				"processing_time_ms", processingTime.Milliseconds())

		case <-ctx.Done():
			return
		}
	}
}

func main() {
	// Load configuration
	cfg, err := configloader.New[AppConfig](
		configloader.WithFileFromEnv("config.yaml"),
	)
	if err != nil {
		logger.Fatal(context.Background(), "Failed to load configuration", "error", err)
	}

	// Create metrics service
	metricsService := NewMetricsService()

	// Create and start application
	app := fastapp.New(cfg.App, fastapp.WithVersion("1.0.0"))
	app.Add(metricsService)

	logger.Info(context.Background(), "🎯 Metrics Demo Application")
	logger.Info(context.Background(), "This demo shows various Prometheus metrics patterns:")
	logger.Info(context.Background(), "• Counters - values that only increase (requests, errors)")
	logger.Info(context.Background(), "• Gauges - values that go up and down (connections, queue size)")
	logger.Info(context.Background(), "• Histograms - distribution of values (request duration, response size)")
	logger.Info(context.Background(), "• Summaries - quantiles of values (processing time percentiles)")
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "📊 Metrics endpoint:")
	logger.Info(context.Background(), "   • Prometheus: http://localhost:9090/metrics")
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "🔍 Other endpoints:")
	logger.Info(context.Background(), "   • Health:     http://localhost:9090/health/checks")
	logger.Info(context.Background(), "   • Profiling:  http://localhost:9090/debug/pprof/")
	logger.Info(context.Background(), "")
	logger.Info(context.Background(), "💡 Try these Prometheus queries:")
	logger.Info(context.Background(), "   • rate(http_requests_total[5m]) - Request rate")
	logger.Info(context.Background(), "   • histogram_quantile(0.95, http_request_duration_seconds_bucket) - 95th percentile latency")
	logger.Info(context.Background(), "   • active_connections - Current connections")
	logger.Info(context.Background(), "   • processing_time_seconds{quantile=\"0.9\"} - 90th percentile processing time")

	app.Start()
}

# Metrics Example

This example demonstrates comprehensive Prometheus metrics integration with FastApp, showing different metric types and common patterns for application monitoring.

## Features Demonstrated

### 1. Metric Types
- **Counters** - Values that only increase (requests, errors)
- **Gauges** - Values that can go up and down (connections, queue size)
- **Histograms** - Distribution of values with buckets (request duration, response size)
- **Summaries** - Quantiles of values (processing time percentiles)

### 2. Metric Patterns
- **Request metrics** - HTTP request counting and timing
- **Error metrics** - Error tracking by type and endpoint
- **Resource metrics** - Connection and queue monitoring
- **Business metrics** - Application-specific measurements

### 3. Labels and Dimensions
- **Multi-dimensional metrics** - Metrics with labels for filtering
- **Label best practices** - Avoiding high cardinality
- **Dynamic labeling** - Runtime label assignment

## Running the Example

### Basic Usage

```bash
cd example/metrics
go run main.go
```

### With Debug Logging

```bash
cd example/metrics
APP_LOGGER_LEVEL=debug go run main.go
```

### Custom Port

```bash
cd example/metrics
APP_OBSERVABILITY_PORT=8090 go run main.go
```

## Metrics Exposed

### Counters

```
# Total HTTP requests
http_requests_total

# Errors by type and endpoint
errors_total{type="client_error|server_error", endpoint="/api/users|/api/orders|/api/products"}
```

### Gauges

```
# Current active connections
active_connections

# Current queue size
queue_size
```

### Histograms

```
# HTTP request duration with buckets
http_request_duration_seconds_bucket{method="GET|POST|PUT|DELETE", status="200|400|404|500", le="0.005|0.01|..."}
http_request_duration_seconds_sum{method="...", status="..."}
http_request_duration_seconds_count{method="...", status="..."}

# HTTP response size with custom buckets
http_response_size_bytes_bucket{le="100|500|1000|5000|10000|50000|100000"}
http_response_size_bytes_sum
http_response_size_bytes_count
```

### Summaries

```
# Processing time with quantiles
processing_time_seconds{operation="user_registration|order_processing|payment_processing|data_export", quantile="0.5|0.9|0.99"}
processing_time_seconds_sum{operation="..."}
processing_time_seconds_count{operation="..."}
```

## Prometheus Queries

### Request Rate

```promql
# Requests per second
rate(http_requests_total[5m])

# Requests per second by method and status
rate(http_request_duration_seconds_count[5m])
```

### Error Rate

```promql
# Error rate percentage
rate(errors_total[5m]) / rate(http_requests_total[5m]) * 100

# Error rate by type
rate(errors_total[5m]) by (type)
```

### Latency Percentiles

```promql
# 95th percentile request duration
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# 90th percentile processing time
processing_time_seconds{quantile="0.9"}
```

### Resource Utilization

```promql
# Current active connections
active_connections

# Average queue size over time
avg_over_time(queue_size[5m])
```

### Response Size Analysis

```promql
# Average response size
rate(http_response_size_bytes_sum[5m]) / rate(http_response_size_bytes_count[5m])

# 99th percentile response size
histogram_quantile(0.99, rate(http_response_size_bytes_bucket[5m]))
```

## Metric Implementation Patterns

### Counter Pattern

```go
// Define counter
requestsTotal := promauto.NewCounter(prometheus.CounterOpts{
    Name: "http_requests_total",
    Help: "Total number of HTTP requests",
})

// Increment counter
requestsTotal.Inc()
```

### Counter with Labels

```go
// Define counter with labels
errorsTotal := promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "errors_total",
        Help: "Total errors by type and endpoint",
    },
    []string{"type", "endpoint"},
)

// Increment with specific labels
errorsTotal.WithLabelValues("client_error", "/api/users").Inc()
```

### Gauge Pattern

```go
// Define gauge
activeConnections := promauto.NewGauge(prometheus.GaugeOpts{
    Name: "active_connections",
    Help: "Current number of active connections",
})

// Set gauge value
activeConnections.Set(float64(connectionCount))

// Increment/decrement gauge
activeConnections.Inc()
activeConnections.Dec()
activeConnections.Add(5)
activeConnections.Sub(3)
```

### Histogram Pattern

```go
// Define histogram with custom buckets
requestDuration := promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration",
        Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
    },
    []string{"method", "status"},
)

// Observe value
start := time.Now()
// ... do work ...
duration := time.Since(start)
requestDuration.WithLabelValues("GET", "200").Observe(duration.Seconds())
```

### Summary Pattern

```go
// Define summary with quantiles
processingTime := promauto.NewSummaryVec(
    prometheus.SummaryOpts{
        Name:       "processing_time_seconds",
        Help:       "Processing time",
        Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
    },
    []string{"operation"},
)

// Observe value
start := time.Now()
// ... do work ...
duration := time.Since(start)
processingTime.WithLabelValues("user_registration").Observe(duration.Seconds())
```

## Best Practices

### 1. Choose the Right Metric Type

- **Counter**: Use for things that only increase (requests, errors, bytes sent)
- **Gauge**: Use for things that go up and down (memory usage, queue size, temperature)
- **Histogram**: Use for measuring distributions (request duration, response size)
- **Summary**: Use when you need specific quantiles and can't use histograms

### 2. Label Guidelines

```go
// Good - low cardinality labels
errors_total{type="client_error", endpoint="/api/users"}

// Bad - high cardinality labels (avoid user IDs, timestamps, etc.)
errors_total{user_id="12345", timestamp="2024-01-15T10:30:45Z"}
```

### 3. Naming Conventions

```go
// Counters should end with _total
http_requests_total
errors_total

// Gauges should describe what they measure
active_connections
queue_size

// Histograms/Summaries should include units
http_request_duration_seconds
response_size_bytes
```

### 4. Help Text

```go
prometheus.CounterOpts{
    Name: "http_requests_total",
    Help: "Total number of HTTP requests processed", // Clear, descriptive help
}
```

## Observability Endpoints

- **Metrics**: http://localhost:9090/metrics
- **Health Checks**: http://localhost:9090/health/checks
- **Liveness**: http://localhost:9090/health/live
- **Readiness**: http://localhost:9090/health/ready
- **Profiling**: http://localhost:9090/debug/pprof/

## Integration with Monitoring Systems

### Prometheus Configuration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'fastapp-metrics-demo'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 15s
    metrics_path: /metrics
```

### Grafana Dashboard

Create dashboards using the exposed metrics:

1. **Request Rate**: `rate(http_requests_total[5m])`
2. **Error Rate**: `rate(errors_total[5m]) / rate(http_requests_total[5m]) * 100`
3. **Latency**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
4. **Active Connections**: `active_connections`

### Alerting Rules

```yaml
# alerts.yml
groups:
  - name: fastapp-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(errors_total[5m]) / rate(http_requests_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          
      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency detected"
```

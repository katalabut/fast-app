# Complete FastApp Example

This comprehensive example demonstrates ALL FastApp capabilities in a single, production-ready application. It showcases configuration management, structured logging, Prometheus metrics, health checks, and service lifecycle management.

## Features Demonstrated

### 🔧 Configuration Management
- **Multiple sources**: YAML files, JSON files, environment variables
- **Type safety**: Compile-time configuration validation
- **Default values**: Fallback configuration via struct tags
- **Environment-specific**: Different configs for dev/staging/prod
- **Validation**: Runtime configuration validation

### 📝 Structured Logging
- **Multiple levels**: Debug, Info, Warn, Error with appropriate usage
- **Context-aware**: Correlation fields across related log entries
- **Structured data**: Key-value pairs for machine-readable logs
- **Development/Production**: Console vs JSON output modes
- **Performance**: Efficient logging with minimal overhead

### 📊 Prometheus Metrics
- **Counters**: Request counts, error counts (monotonically increasing)
- **Gauges**: Active connections, queue sizes (can go up/down)
- **Histograms**: Request duration, response sizes (distribution with buckets)
- **Labels**: Multi-dimensional metrics for filtering and aggregation
- **Business metrics**: Application-specific measurements

### 🏥 Health Checks
- **Liveness probe**: Process health for container restart decisions
- **Readiness probe**: Traffic readiness for load balancer routing
- **Service checks**: Component-specific health validation
- **Dependency checks**: External service monitoring (HTTP, Database)
- **System checks**: Resource utilization monitoring
- **Custom checks**: Business logic health validation

### 🚀 Service Lifecycle
- **Graceful startup**: Phased initialization with readiness control
- **Graceful shutdown**: Clean resource cleanup on termination
- **Error handling**: Panic recovery and error propagation
- **Signal handling**: SIGTERM/SIGINT handling for container environments

## Running the Example

### Prerequisites

```bash
# Install dependencies (optional - for database demo)
docker run --name postgres-demo \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_USER=user \
  -e POSTGRES_DB=completedemo \
  -p 5432:5432 -d postgres:13
```

### Basic Usage

```bash
cd example/complete
go run main.go
```

### With Custom Configuration

```bash
cd example/complete
cp .env.example .env
# Edit .env as needed
export $(cat .env | xargs)
go run main.go
```

### Different Environments

```bash
# Development mode
APP_LOGGER_LEVEL=debug APP_LOGGER_DEVMODE=true go run main.go

# Production mode
APP_LOGGER_LEVEL=info APP_LOGGER_DEVMODE=false go run main.go

# Custom port
APP_OBSERVABILITY_PORT=8090 go run main.go
```

## Application Architecture

### Service Structure

```go
type CompleteService struct {
    name   string
    config *AppConfig
    db     *sql.DB
    ready  bool
    status string
    
    // Prometheus metrics
    requestsTotal     prometheus.Counter
    requestDuration   *prometheus.HistogramVec
    activeOrders      prometheus.Gauge
    orderValue        *prometheus.HistogramVec
    processingErrors  *prometheus.CounterVec
}
```

### Configuration Structure

```go
type AppConfig struct {
    App      config.App        // FastApp framework config
    Database DatabaseConfig   // Database settings
    Redis    RedisConfig      // Cache settings
    External ExternalConfig   // External API settings
    Business BusinessConfig   // Business logic settings
}
```

### Health Check Implementation

The service implements multiple health check interfaces:

- `HealthProvider`: Provides service-specific health checks
- `ReadinessController`: Controls readiness state during startup

## Observability Endpoints

### Health Endpoints

```bash
# Liveness probe (always 200 if process is alive)
curl http://localhost:9090/health/live

# Readiness probe (200 if ready to serve traffic)
curl http://localhost:9090/health/ready

# Detailed health information
curl http://localhost:9090/health/checks | jq .
```

### Metrics Endpoint

```bash
# Prometheus metrics
curl http://localhost:9090/metrics

# Filter business metrics
curl http://localhost:9090/metrics | grep business_
```

### Debug Endpoints

```bash
# Go profiling
curl http://localhost:9090/debug/pprof/

# Heap profile
curl http://localhost:9090/debug/pprof/heap

# CPU profile (30 seconds)
curl http://localhost:9090/debug/pprof/profile?seconds=30
```

## Metrics Exposed

### Business Metrics

```
# Request counters
business_requests_total

# Request duration histogram
business_request_duration_seconds_bucket{operation="user_registration|profile_update|...", status="success|error", le="..."}
business_request_duration_seconds_sum{operation="...", status="..."}
business_request_duration_seconds_count{operation="...", status="..."}

# Active orders gauge
active_orders_count

# Order value histogram
order_value_dollars_bucket{currency="USD", le="..."}
order_value_dollars_sum{currency="USD"}
order_value_dollars_count{currency="USD"}

# Processing errors counter
processing_errors_total{error_type="validation_error|timeout_error", operation="..."}
```

### System Metrics (built-in)

```
# Go runtime metrics
go_memstats_alloc_bytes
go_memstats_sys_bytes
go_goroutines

# Process metrics
process_cpu_seconds_total
process_resident_memory_bytes
```

## Health Checks Implemented

### Service-Specific Checks

1. **service-readiness**: Service initialization and operational status
2. **business-logic**: Business logic performance validation
3. **configuration**: Configuration validation and consistency

### Global Checks

1. **external-api**: HTTP dependency monitoring
2. **postgres**: Database connectivity and performance
3. **system-resources**: CPU, memory, and disk usage monitoring

## Configuration Examples

### Development Configuration

```yaml
App:
  Logger:
    Level: "debug"
    DevMode: true
  
Business:
  EnableNewFeature: true
  ProcessingDelay: "50ms"

External:
  APIURL: "https://httpbin.org/status/200"
```

### Production Configuration

```yaml
App:
  Logger:
    Level: "info"
    DevMode: false
  Observability:
    Port: 9090

Database:
  MaxConnections: 100
  ConnectTimeout: "5s"

Business:
  MaxOrderValue: 50000.0
  EnableNewFeature: false
```

## Monitoring Integration

### Prometheus Configuration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'fastapp-complete-demo'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 15s
    metrics_path: /metrics
```

### Grafana Dashboard Queries

```promql
# Request rate
rate(business_requests_total[5m])

# Error rate
rate(processing_errors_total[5m]) / rate(business_requests_total[5m]) * 100

# 95th percentile latency
histogram_quantile(0.95, rate(business_request_duration_seconds_bucket[5m]))

# Active orders
active_orders_count

# Average order value
rate(order_value_dollars_sum[5m]) / rate(order_value_dollars_count[5m])
```

### Alerting Rules

```yaml
# alerts.yml
groups:
  - name: fastapp-complete-demo
    rules:
      - alert: HighErrorRate
        expr: rate(processing_errors_total[5m]) / rate(business_requests_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate in complete demo"
          
      - alert: ServiceNotReady
        expr: up{job="fastapp-complete-demo"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Complete demo service is down"
```

## Container Deployment

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o complete-demo ./example/complete

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/complete-demo .
COPY --from=builder /app/example/complete/config.yaml .
EXPOSE 9090
CMD ["./complete-demo"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fastapp-complete-demo
spec:
  replicas: 3
  selector:
    matchLabels:
      app: fastapp-complete-demo
  template:
    metadata:
      labels:
        app: fastapp-complete-demo
    spec:
      containers:
      - name: app
        image: fastapp-complete-demo:latest
        ports:
        - containerPort: 9090
        env:
        - name: APP_LOGGER_LEVEL
          value: "info"
        - name: APP_LOGGER_DEVMODE
          value: "false"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 9090
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
```

## Best Practices Demonstrated

### 1. Configuration Management
- Environment-specific configurations
- Sensitive data handling
- Configuration validation
- Default value management

### 2. Logging Strategy
- Appropriate log levels for different scenarios
- Structured logging for machine processing
- Context propagation for request tracing
- Performance-conscious logging

### 3. Metrics Design
- Business-relevant metrics
- Appropriate metric types for different use cases
- Multi-dimensional labeling
- Performance impact consideration

### 4. Health Check Design
- Different checks for different audiences
- Meaningful health status reporting
- Performance-optimized health checks
- Comprehensive dependency monitoring

### 5. Service Lifecycle
- Graceful startup with readiness control
- Clean shutdown procedures
- Error handling and recovery
- Resource management

This complete example serves as a production-ready template for FastApp applications, demonstrating all framework capabilities in a cohesive, well-structured application.

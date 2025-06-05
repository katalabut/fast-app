# FastApp Examples

This directory contains complete working examples demonstrating various FastApp features and patterns.

## Examples Overview

### [Complete](./complete/) - 🌟 **RECOMMENDED** Production-Ready Application
**START HERE** - Comprehensive example showcasing ALL FastApp capabilities in a cohesive, production-ready application.

**Run:**
```bash
cd complete
go run main.go
```

**Features:**
- Complete configuration management (YAML/JSON + env vars)
- Structured logging with all patterns
- Comprehensive Prometheus metrics
- Full health check implementation
- Service lifecycle management
- Business logic simulation
- Production deployment patterns

### [Logger](./logger/) - Logging Capabilities
Demonstrates comprehensive logging patterns and best practices.

**Run:**
```bash
cd logger
go run main.go
```

**Features:**
- Multiple log levels (debug, info, warn, error, fatal)
- Structured logging with key-value pairs
- Context-aware logging with field propagation
- Development vs production logging modes
- Error logging best practices

### [Config](./config/) - Configuration Management
Shows configuration management patterns including file loading, environment variables, and validation.

**Run:**
```bash
cd config
go run main.go
```

**Features:**
- YAML/JSON configuration files
- Environment variable override
- Type-safe configuration with struct tags
- Configuration validation
- Environment-specific settings
- Feature flags and business configuration

### [Metrics](./metrics/) - Prometheus Metrics
Demonstrates Prometheus metrics integration with various metric types and monitoring patterns.

**Run:**
```bash
cd metrics
go run main.go
```

**Features:**
- Counters (requests, errors)
- Gauges (connections, queue size)
- Histograms (request duration, response size)
- Summaries (processing time percentiles)
- Multi-dimensional labels
- Business metrics patterns

### [Health](./health/) - Health Checks
Comprehensive health check patterns for monitoring application and dependency health.

**Run:**
```bash
cd health
go run main.go
```

**Features:**
- Liveness and readiness probes
- Service-specific health checks
- HTTP dependency monitoring
- Database connectivity checks
- Custom health checks
- Dynamic health status changes
- Gradual startup patterns

### [Basic](./basic/) - Simple Application
A minimal FastApp application demonstrating core features.

**Run:**
```bash
cd basic
go run main.go
```

**Features:**
- Single service with health checks
- HTTP health check for external API
- Custom service health check
- Debug server with metrics

### [Simple](./simple/) - Multiple Services
Multiple services running concurrently with individual health checks.

**Run:**
```bash
cd simple
go run main.go
```

**Features:**
- API, Worker, and Scheduler services
- Memory and system load health checks
- HTTP endpoint health checks
- Service-specific health monitoring

### [Advanced](./advanced/) - Database Integration
Production-ready example with database integration and comprehensive health checks.

**Run:**
```bash
cd advanced
# Note: Requires PostgreSQL database
export DATABASE_URL="postgres://user:password@localhost/dbname?sslmode=disable"
go run main.go
```

**Features:**
- PostgreSQL database integration
- Database health checks with custom queries
- Business logic health validation
- Graceful degradation when database is unavailable

## Common Patterns

### Service Implementation

All examples follow the same service pattern:

```go
type MyService struct {
    ready bool
}

// Required: Service interface
func (s *MyService) Run(ctx context.Context) error {
    // Service initialization
    s.ready = true
    
    // Main service logic
    <-ctx.Done()
    return nil
}

func (s *MyService) Shutdown(ctx context.Context) error {
    s.ready = false
    return nil
}

// Optional: Health checks
func (s *MyService) HealthChecks() []health.HealthChecker {
    return []health.HealthChecker{
        health.NewCustomCheck("my-service", s.checkHealth),
    }
}

// Optional: Readiness control
func (s *MyService) SetReady(ready bool) { s.ready = ready }
func (s *MyService) IsReady() bool { return s.ready }
```

### Configuration Pattern

```go
type Config struct {
    App         fastapp.Config
    DebugServer service.DebugServer
    Database    DatabaseConfig
    // ... your custom config
}

func main() {
    cfg, err := configloader.New[Config]()
    if err != nil {
        logger.Fatal(context.Background(), "Failed to load config", "error", err)
    }
    
    // Use configuration
    app := fastapp.New(cfg.App)
    // ...
}
```

### Health Checks Pattern

```go
// Global health checks
httpCheck := checks.NewHTTPCheck("external-api", "https://api.example.com/health")
dbCheck := checks.NewDatabaseCheck("postgres", db)

app := fastapp.New(cfg.App).
    WithHealthChecks(httpCheck, dbCheck).
    Add(myService)
```

## Testing Examples

Each example can be tested using the observability endpoints:

```bash
# Start the example
go run main.go

# Test liveness (should always return 200)
curl http://localhost:9090/health/live

# Test readiness (returns 200 when ready)
curl http://localhost:9090/health/ready

# Get detailed health information
curl http://localhost:9090/health/checks | jq .

# Check Prometheus metrics
curl http://localhost:9090/metrics

# Check specific business metrics
curl http://localhost:9090/metrics | grep business_

# Access profiling endpoints
curl http://localhost:9090/debug/pprof/
```

## Environment Variables

Examples support configuration via environment variables:

```bash
# Logging
export LOGGER_LEVEL="debug"
export LOGGER_ENCODING="console"

# Health checks
export HEALTH_ENABLED="true"
export HEALTH_PORT="8080"
export HEALTH_TIMEOUT="30s"

# Debug server
export DEBUGSERVER_ENABLED="true"
export DEBUGSERVER_PORT="9090"

# Database (advanced example)
export DATABASE_URL="postgres://localhost/mydb"
```

## Docker Examples

### Basic Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080 9090
CMD ["./main"]
```

### Docker Compose

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"  # Health checks
      - "9090:9090"  # Metrics
    environment:
      - LOGGER_LEVEL=info
      - HEALTH_ENABLED=true
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health/live"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## Kubernetes Examples

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fastapp-example
spec:
  replicas: 3
  selector:
    matchLabels:
      app: fastapp-example
  template:
    metadata:
      labels:
        app: fastapp-example
    spec:
      containers:
      - name: app
        image: fastapp-example:latest
        ports:
        - containerPort: 8080
          name: health
        - containerPort: 9090
          name: metrics
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        env:
        - name: LOGGER_LEVEL
          value: "info"
        - name: HEALTH_ENABLED
          value: "true"
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: fastapp-example
spec:
  selector:
    app: fastapp-example
  ports:
  - name: health
    port: 8080
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
```

## Performance Testing

### Load Testing with hey

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Test health endpoint
hey -n 1000 -c 10 http://localhost:8080/health/ready

# Test detailed health checks
hey -n 100 -c 5 http://localhost:8080/health/checks
```

### Monitoring

```bash
# Watch health status
watch -n 1 'curl -s http://localhost:8080/health/ready | jq .'

# Monitor metrics
curl -s http://localhost:9090/metrics | grep fastapp
```

## Troubleshooting

### Common Issues

1. **Port conflicts**
   ```bash
   # Check what's using the port
   lsof -i :8080
   
   # Use different ports
   export HEALTH_PORT="8081"
   export DEBUGSERVER_PORT="9091"
   ```

2. **Database connection issues (advanced example)**
   ```bash
   # Check database connectivity
   pg_isready -h localhost -p 5432
   
   # Use connection string
   export DATABASE_URL="postgres://user:pass@localhost/db?sslmode=disable"
   ```

3. **Health checks failing**
   ```bash
   # Check detailed health information
   curl http://localhost:8080/health/checks | jq '.checks'
   
   # Enable debug logging
   export LOGGER_LEVEL="debug"
   ```

### Debug Mode

Enable debug logging for troubleshooting:

```bash
export LOGGER_LEVEL="debug"
export LOGGER_ENCODING="console"
go run .
```

## Next Steps

After exploring the examples:

1. **Read the Documentation** - Check out the [docs](../docs) directory
2. **Build Your Own** - Start with the [Getting Started Guide](../docs/getting-started.md)
3. **Contribute** - See the [Contributing Guide](../CONTRIBUTING.md)
4. **Get Help** - Open an issue or start a discussion on GitHub

## Questions?

- 📖 [Documentation](../docs)
- 🐛 [Issues](https://github.com/katalabut/fast-app/issues)
- 💬 [Discussions](https://github.com/katalabut/fast-app/discussions)

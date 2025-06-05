# Health Checks Example

This example demonstrates comprehensive health check patterns in FastApp, including service-specific checks, global checks, and various health check types for monitoring application and dependency health.

## Features Demonstrated

### 1. Health Check Types
- **Liveness Probe** - Checks if the process is alive (for container restart)
- **Readiness Probe** - Checks if the process is ready to serve traffic
- **Custom Health Checks** - Application-specific health validation
- **HTTP Health Checks** - External service dependency monitoring
- **Database Health Checks** - Database connectivity and performance
- **System Health Checks** - Resource utilization monitoring

### 2. Health Check Patterns
- **Service-specific checks** - Health checks provided by individual services
- **Global checks** - Application-wide health checks
- **Dynamic health status** - Health status that changes over time
- **Gradual startup** - Readiness control during service initialization
- **Temporary issues** - Simulation of transient health problems

### 3. Health Check Integration
- **Automatic registration** - Service health checks auto-registered
- **Health aggregation** - Multiple checks combined into overall status
- **Caching** - Health check results cached for performance
- **Timeouts** - Configurable timeouts for health checks

## Running the Example

### Basic Usage

```bash
cd example/health
go run main.go
```

### With Database (PostgreSQL)

```bash
# Start PostgreSQL (using Docker)
docker run --name postgres-health-demo -e POSTGRES_PASSWORD=password -e POSTGRES_USER=user -e POSTGRES_DB=healthdemo -p 5432:5432 -d postgres:13

cd example/health
go run main.go
```

### Custom Configuration

```bash
cd example/health
APP_OBSERVABILITY_PORT=8090 go run main.go
```

## Health Check Endpoints

### Liveness Probe
```bash
curl http://localhost:9090/health/live
```
- Always returns 200 if the process is alive
- Used by container orchestrators for restart decisions

### Readiness Probe
```bash
curl http://localhost:9090/health/ready
```
- Returns 200 if the application is ready to serve traffic
- Used by load balancers for traffic routing decisions

### Detailed Health Checks
```bash
curl http://localhost:9090/health/checks | jq .
```
- Returns detailed information about all health checks
- Includes status, duration, and additional details for each check

## Running the Example

### Basic Usage

```bash
cd example/health
go run main.go
```

### With Database (PostgreSQL)

```bash
# Start PostgreSQL (using Docker)
docker run --name postgres-health-demo -e POSTGRES_PASSWORD=password -e POSTGRES_USER=user -e POSTGRES_DB=healthdemo -p 5432:5432 -d postgres:13

cd example/health
go run main.go
```

### Custom Configuration

```bash
cd example/health
APP_OBSERVABILITY_PORT=8090 go run main.go
```

## Health Check Endpoints

### Liveness Probe
```bash
curl http://localhost:9090/health/live
```
- Always returns 200 if the process is alive
- Used by container orchestrators for restart decisions

### Readiness Probe
```bash
curl http://localhost:9090/health/ready
```
- Returns 200 if the application is ready to serve traffic
- Used by load balancers for traffic routing decisions

### Detailed Health Checks
```bash
curl http://localhost:9090/health/checks | jq .
```
- Returns detailed information about all health checks
- Includes status, duration, and additional details for each check

## Health Check Implementation

### Service Health Checks

Services can provide their own health checks by implementing the `HealthProvider` interface:

```go
func (s *MyService) HealthChecks() []health.HealthChecker {
    return []health.HealthChecker{
        health.NewCustomCheck("service-readiness", func(ctx context.Context) health.HealthResult {
            if !s.ready {
                return health.NewUnhealthyResult("Service is not ready").
                    WithDetails("status", s.status)
            }
            return health.NewHealthyResult("Service is ready")
        }),
    }
}
```

### Global Health Checks

Add application-wide health checks:

```go
// HTTP health check
httpCheck := checks.NewHTTPCheck("external-api", "https://api.example.com/health")

// Database health check
dbCheck := checks.NewDatabaseCheck("postgres", db)

// Custom system check
systemCheck := health.NewCustomCheck("system-resources", func(ctx context.Context) health.HealthResult {
    // Check CPU, memory, disk usage
    return health.NewHealthyResult("System resources OK")
})

app.WithHealthChecks(httpCheck, dbCheck, systemCheck)
```

### Readiness Control

Control service readiness during startup:

```go
type MyService struct {
    ready bool
}

func (s *MyService) SetReady(ready bool) { s.ready = ready }
func (s *MyService) IsReady() bool { return s.ready }

func (s *MyService) Run(ctx context.Context) error {
    // Initialization phase
    s.ready = false
    
    // Do initialization work...
    time.Sleep(5 * time.Second)
    
    // Mark as ready
    s.ready = true
    
    // Continue running...
    <-ctx.Done()
    return nil
}
```

## Built-in Health Check Types

### HTTP Health Check

```go
import "github.com/katalabut/fast-app/health/checks"

// Simple HTTP check
httpCheck := checks.NewHTTPCheck("api", "https://api.example.com/health")

// HTTP check with options
httpCheck := checks.NewHTTPCheckWithOptions("api", "https://api.example.com/health",
    checks.HTTPOptions{
        Timeout:        10 * time.Second,
        ExpectedStatus: 200,
        ExpectedBody:   `{"status":"ok"}`,
        Method:         "GET",
        Headers:        map[string]string{"Authorization": "Bearer token"},
    })
```

### Database Health Check

```go
import "github.com/katalabut/fast-app/health/checks"

// Simple database ping
dbCheck := checks.NewDatabaseCheck("postgres", db)

// Database check with custom query
dbCheck := checks.NewDatabaseCheckWithOptions("postgres", db,
    checks.DatabaseOptions{
        PingTimeout: 5 * time.Second,
        Query:       "SELECT 1",
    })
```

### Custom Health Check

```go
import "github.com/katalabut/fast-app/health"

customCheck := health.NewCustomCheck("business-logic", func(ctx context.Context) health.HealthResult {
    // Your health check logic
    if someCondition {
        return health.NewHealthyResult("All systems operational")
    }
    
    return health.NewDegradedResult("Performance degraded").
        WithDetails("response_time", "500ms").
        WithDetails("threshold", "200ms")
})
```

## Health Status Types

### Healthy
- Service is operating normally
- All dependencies are available
- Performance is within acceptable limits

### Degraded
- Service is operational but with reduced performance
- Some non-critical dependencies may be unavailable
- Performance is below optimal but still acceptable

### Unhealthy
- Service is not operating correctly
- Critical dependencies are unavailable
- Service should not receive traffic

## Health Check Response Format

### Liveness Response
```json
{
  "status": "healthy",
  "message": "Application is alive"
}
```

### Readiness Response
```json
{
  "status": "healthy",
  "message": "Application is ready to serve traffic"
}
```

### Detailed Health Response
```json
{
  "status": "healthy",
  "message": "All health checks passed",
  "checks": {
    "service-readiness": {
      "status": "healthy",
      "message": "Service is ready and healthy",
      "details": {
        "status": "ready"
      },
      "duration": 1234567
    },
    "external-api": {
      "status": "healthy",
      "message": "HTTP check successful",
      "details": {
        "status_code": 200,
        "response_time_ms": 45
      },
      "duration": 45123456
    },
    "postgres": {
      "status": "healthy",
      "message": "Database connection successful",
      "details": {
        "duration": "2ms"
      },
      "duration": 2345678
    }
  }
}
```

## Configuration Options

### Health Check Configuration

```yaml
App:
  Observability:
    Health:
      Enabled: true           # Enable health check endpoints
      LivePath: "/health/live"    # Liveness probe path
      ReadyPath: "/health/ready"  # Readiness probe path
      CheckPath: "/health/checks" # Detailed checks path
      Timeout: "30s"          # Health check timeout
      CacheTTL: "5s"          # Cache duration for results
```

### Environment Variables

```bash
APP_OBSERVABILITY_HEALTH_ENABLED=true
APP_OBSERVABILITY_HEALTH_LIVEPATH=/health/live
APP_OBSERVABILITY_HEALTH_READYPATH=/health/ready
APP_OBSERVABILITY_HEALTH_CHECKPATH=/health/checks
APP_OBSERVABILITY_HEALTH_TIMEOUT=30s
APP_OBSERVABILITY_HEALTH_CACHETTL=5s
```

## Best Practices

### 1. Design for Different Audiences

- **Liveness**: For container orchestrators (Kubernetes, Docker)
- **Readiness**: For load balancers and traffic routing
- **Detailed**: For monitoring systems and debugging

### 2. Include Relevant Details

```go
return health.NewHealthyResult("Database connection successful").
    WithDetails("connection_pool_size", poolSize).
    WithDetails("active_connections", activeConns).
    WithDetails("response_time_ms", duration.Milliseconds())
```

### 3. Use Appropriate Timeouts

```go
// Quick checks for liveness/readiness
checks.DatabaseOptions{PingTimeout: 5 * time.Second}

// Longer timeouts for detailed checks
checks.HTTPOptions{Timeout: 30 * time.Second}
```

### 4. Handle Graceful Degradation

```go
if responseTime > criticalThreshold {
    return health.NewUnhealthyResult("Service is too slow")
} else if responseTime > warningThreshold {
    return health.NewDegradedResult("Service is slower than expected")
}
return health.NewHealthyResult("Service is performing well")
```

### 5. Monitor Health Check Performance

- Health checks should be fast (< 1 second for readiness)
- Cache results when appropriate
- Avoid expensive operations in health checks

## Integration with Container Orchestrators

### Kubernetes

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: my-fastapp
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
```

### Docker Compose

```yaml
version: '3.8'
services:
  app:
    image: my-fastapp
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9090/health/live"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

## Observability Endpoints

- **Health Checks**: http://localhost:9090/health/checks
- **Liveness**: http://localhost:9090/health/live
- **Readiness**: http://localhost:9090/health/ready
- **Metrics**: http://localhost:9090/metrics
- **Profiling**: http://localhost:9090/debug/pprof/

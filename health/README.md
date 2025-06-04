# Health Checks

FastApp provides a comprehensive health check system for monitoring application and dependency health.

## Core Concepts

### Check Types

1. **Liveness Probe** - Checks if the process is alive (for container restart)
2. **Readiness Probe** - Checks if the process is ready to serve traffic (for load balancer)
3. **Health Check** - Checks specific components (DB, Redis, API, etc.)

### HTTP Endpoints

- `GET /health/live` - Liveness probe (always returns 200 if process is alive)
- `GET /health/ready` - Readiness probe (returns 200 if application is ready)
- `GET /health/checks` - Detailed health information for all checks

## Quick Start

```go
package main

import (
    "context"
    
    fastapp "github.com/katalabut/fast-app"
    "github.com/katalabut/fast-app/health"
    "github.com/katalabut/fast-app/health/checks"
)

type MyService struct {
    ready bool
}

// Implement HealthProvider
func (s *MyService) HealthChecks() []health.HealthChecker {
    return []health.HealthChecker{
        health.NewCustomCheck("my-service", func(ctx context.Context) health.HealthResult {
            if s.ready {
                return health.NewHealthyResult("Service is ready")
            }
            return health.NewUnhealthyResult("Service is not ready")
        }),
    }
}

func main() {
    cfg, _ := configloader.New[Config]()
    
    // Global health checks
    httpCheck := checks.NewHTTPCheck("external-api", "https://api.example.com/health")
    
    app := fastapp.New(cfg.App).
        WithHealthChecks(httpCheck).
        Add(&MyService{})
    
    app.SetReady(true)
    app.Start()
}
```

## Built-in Health Checks

### HTTP Check

```go
import "github.com/katalabut/fast-app/health/checks"

// Simple check
httpCheck := checks.NewHTTPCheck("api", "https://api.example.com/health")

// With additional options
httpCheck := checks.NewHTTPCheckWithOptions("api", "https://api.example.com/health", 
    checks.HTTPOptions{
        Timeout:        10 * time.Second,
        ExpectedStatus: 200,
        ExpectedBody:   `{"status":"ok"}`,
        Method:         "GET",
        Headers:        map[string]string{"Authorization": "Bearer token"},
    })
```

### Database Check

```go
import "github.com/katalabut/fast-app/health/checks"

// Simple connection check
dbCheck := checks.NewDatabaseCheck("postgres", db)

// With custom query
dbCheck := checks.NewDatabaseCheckWithOptions("postgres", db,
    checks.DatabaseOptions{
        PingTimeout: 5 * time.Second,
        Query:       "SELECT 1",
    })
```

### Custom Check

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

## Interfaces

### HealthProvider

Services can provide their own health checks:

```go
type HealthProvider interface {
    HealthChecks() []HealthChecker
}

func (s *MyService) HealthChecks() []health.HealthChecker {
    return []health.HealthChecker{
        checks.NewDatabaseCheck("service-db", s.db),
        health.NewCustomCheck("service-logic", s.checkLogic),
    }
}
```

### ReadinessController

Services can control their readiness state:

```go
type ReadinessController interface {
    SetReady(ready bool)
    IsReady() bool
}

func (s *MyService) Run(ctx context.Context) error {
    // Initialization...
    s.SetReady(true) // Signal that we're ready
    
    // Main logic...
    <-ctx.Done()
    return nil
}
```

## Aggregation Strategies

### AllHealthyStrategy (default)

All checks must be healthy for overall healthy status:

```go
strategy := &health.AllHealthyStrategy{}
```

### MajorityHealthyStrategy

Majority of checks must be healthy:

```go
import "github.com/katalabut/fast-app/health/strategies"

strategy := &strategies.MajorityHealthyStrategy{}
```

### WeightedStrategy

Considers component importance:

```go
import "github.com/katalabut/fast-app/health/strategies"

strategy := strategies.NewWeightedStrategy(map[string]health.ComponentImportance{
    "database":     health.Critical,    // must be healthy
    "cache":        health.Important,   // can be degraded
    "external-api": health.Optional,    // can be unhealthy
})
```

## Configuration

```go
type HealthConfig struct {
    Enabled   bool          `default:"true"`
    Port      int           `default:"8080"`
    LivePath  string        `default:"/health/live"`
    ReadyPath string        `default:"/health/ready"`
    CheckPath string        `default:"/health/checks"`
    Timeout   time.Duration `default:"30s"`
    CacheTTL  time.Duration `default:"5s"`
}
```

## Response Examples

### Liveness Probe

```json
{
  "status": "alive",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Readiness Probe

```json
{
  "status": "healthy",
  "ready": true,
  "timestamp": "2024-01-15T10:30:00Z",
  "manager_ready": true,
  "overall_status": "healthy"
}
```

### Detailed Health Checks

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "duration": "45ms",
  "ready": true,
  "checks": {
    "database": {
      "status": "healthy",
      "message": "Connection successful",
      "duration": "12ms"
    },
    "external-api": {
      "status": "degraded",
      "message": "High latency detected",
      "duration": "156ms",
      "details": {
        "latency": "156ms",
        "threshold": "100ms"
      }
    }
  }
}
```

## Kubernetes Integration

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: myapp:latest
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
```

## Testing

### Unit Tests

```go
func TestMyServiceHealthChecks(t *testing.T) {
    service := &MyService{ready: true}
    checks := service.HealthChecks()
    
    if len(checks) != 1 {
        t.Errorf("Expected 1 health check, got %d", len(checks))
    }
    
    result := checks[0].Check(context.Background())
    if result.Status != health.StatusHealthy {
        t.Errorf("Expected healthy status, got %s", result.Status)
    }
}
```

### Integration Tests

```go
func TestHealthEndpoints(t *testing.T) {
    // Start application
    app := fastapp.New(config).Add(myService)
    go app.Start()
    
    // Test liveness endpoint
    resp, err := http.Get("http://localhost:8080/health/live")
    if err != nil {
        t.Fatal(err)
    }
    if resp.StatusCode != 200 {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }
}
```

## Best Practices

1. **Keep checks fast** - Health checks should complete quickly (< 1 second)
2. **Use timeouts** - Always set appropriate timeouts for external dependencies
3. **Fail fast** - Return unhealthy immediately when critical components fail
4. **Use degraded status** - For performance issues that don't require immediate action
5. **Monitor check performance** - Track health check execution times
6. **Test your checks** - Write unit tests for custom health check logic

## Troubleshooting

### Common Issues

1. **Health checks timing out**
   - Increase timeout values
   - Check network connectivity
   - Verify external service availability

2. **False positives**
   - Review check logic
   - Add appropriate error handling
   - Consider using degraded status instead of unhealthy

3. **Performance impact**
   - Enable caching with appropriate TTL
   - Reduce check frequency
   - Optimize check implementation

### Debugging

Enable debug logging to see detailed health check execution:

```go
// Set log level to debug
cfg.Logger.Level = "debug"
```

Check health endpoint directly:

```bash
curl -s http://localhost:8080/health/checks | jq .
```

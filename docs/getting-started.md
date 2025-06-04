# Getting Started with FastApp

This guide will help you get up and running with FastApp quickly.

## Prerequisites

- Go 1.21 or higher
- Basic understanding of Go programming

## Installation

```bash
go get github.com/katalabut/fast-app
```

## Your First FastApp Application

Let's create a simple application step by step.

### 1. Create the main.go file

```go
package main

import (
    "context"
    "time"
    
    fastapp "github.com/katalabut/fast-app"
    "github.com/katalabut/fast-app/configloader"
    "github.com/katalabut/fast-app/logger"
    "github.com/katalabut/fast-app/service"
)

type Config struct {
    App         fastapp.Config
    DebugServer service.DebugServer
}

type HelloService struct{}

func (s *HelloService) Run(ctx context.Context) error {
    logger.Info(ctx, "Hello Service is starting...")
    
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            logger.Info(ctx, "Hello from FastApp!")
        case <-ctx.Done():
            logger.Info(ctx, "Hello Service is shutting down...")
            return nil
        }
    }
}

func (s *HelloService) Shutdown(ctx context.Context) error {
    logger.Info(ctx, "Hello Service cleanup completed")
    return nil
}

func main() {
    cfg, err := configloader.New[Config]()
    if err != nil {
        logger.Fatal(context.Background(), "Failed to load config", "error", err)
    }
    
    fastapp.New(cfg.App).
        Add(service.NewDefaultDebugService(cfg.DebugServer)).
        Add(&HelloService{}).
        Start()
}
```

### 2. Initialize Go module

```bash
go mod init my-fastapp
go mod tidy
```

### 3. Run the application

```bash
go run main.go
```

You should see output similar to:

```
2024-01-15T10:30:00.000Z	INFO	Hello Service is starting...
2024-01-15T10:30:00.000Z	INFO	Starting debug server	{"port": 9090}
2024-01-15T10:30:05.000Z	INFO	Hello from FastApp!
```

### 4. Check the debug endpoints

While your application is running, you can access:

- **Metrics**: http://localhost:9090/metrics
- **Health**: http://localhost:8080/health/live
- **Profiling**: http://localhost:9090/debug/pprof/

## Adding Health Checks

Let's enhance our application with health checks:

```go
package main

import (
    "context"
    "time"
    
    fastapp "github.com/katalabut/fast-app"
    "github.com/katalabut/fast-app/configloader"
    "github.com/katalabut/fast-app/health"
    "github.com/katalabut/fast-app/health/checks"
    "github.com/katalabut/fast-app/logger"
    "github.com/katalabut/fast-app/service"
)

type Config struct {
    App         fastapp.Config
    DebugServer service.DebugServer
}

type HelloService struct {
    ready bool
}

func (s *HelloService) Run(ctx context.Context) error {
    logger.Info(ctx, "Hello Service is starting...")
    
    // Simulate initialization
    time.Sleep(2 * time.Second)
    s.ready = true
    logger.Info(ctx, "Hello Service is ready!")
    
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            logger.Info(ctx, "Hello from FastApp!")
        case <-ctx.Done():
            logger.Info(ctx, "Hello Service is shutting down...")
            s.ready = false
            return nil
        }
    }
}

func (s *HelloService) Shutdown(ctx context.Context) error {
    s.ready = false
    logger.Info(ctx, "Hello Service cleanup completed")
    return nil
}

// Implement HealthProvider interface
func (s *HelloService) HealthChecks() []health.HealthChecker {
    return []health.HealthChecker{
        health.NewCustomCheck("hello-service", func(ctx context.Context) health.HealthResult {
            if s.ready {
                return health.NewHealthyResult("Hello service is ready")
            }
            return health.NewUnhealthyResult("Hello service is not ready")
        }),
    }
}

func main() {
    cfg, err := configloader.New[Config]()
    if err != nil {
        logger.Fatal(context.Background(), "Failed to load config", "error", err)
    }
    
    // Add global health checks
    httpCheck := checks.NewHTTPCheck("google", "https://www.google.com")
    
    app := fastapp.New(cfg.App).
        WithHealthChecks(httpCheck).
        Add(service.NewDefaultDebugService(cfg.DebugServer)).
        Add(&HelloService{})
    
    // Set application as ready
    app.SetReady(true)
    
    logger.Info(context.Background(), "Starting FastApp application")
    logger.Info(context.Background(), "Health endpoints available:")
    logger.Info(context.Background(), "  - Liveness:  http://localhost:8080/health/live")
    logger.Info(context.Background(), "  - Readiness: http://localhost:8080/health/ready")
    logger.Info(context.Background(), "  - Detailed:  http://localhost:8080/health/checks")
    
    app.Start()
}
```

Now you can check the health endpoints:

```bash
# Check if the application is alive
curl http://localhost:8080/health/live

# Check if the application is ready
curl http://localhost:8080/health/ready

# Get detailed health information
curl http://localhost:8080/health/checks | jq .
```

## Configuration

FastApp uses struct-based configuration with automatic environment variable binding:

```go
type Config struct {
    App         fastapp.Config
    DebugServer service.DebugServer
    Database    DatabaseConfig
    Redis       RedisConfig
}

type DatabaseConfig struct {
    URL         string `default:"postgres://localhost/mydb"`
    MaxConns    int    `default:"10"`
    MaxIdleTime string `default:"5m"`
}

type RedisConfig struct {
    URL      string `default:"redis://localhost:6379"`
    Password string `default:""`
    DB       int    `default:"0"`
}
```

You can override configuration values using environment variables:

```bash
export DATABASE_URL="postgres://user:pass@localhost/mydb"
export DATABASE_MAXCONNS="20"
export REDIS_URL="redis://localhost:6380"
```

## Next Steps

- [Health Checks Guide](./health-checks.md) - Learn about comprehensive health monitoring
- [Configuration Guide](./configuration.md) - Deep dive into configuration management
- [Service Development](./services.md) - Best practices for building services
- [Deployment](./deployment.md) - Production deployment guidelines

## Examples

Check out the complete examples in the [examples](../example) directory:

- **[Basic](../example/basic)** - Simple application
- **[Simple](../example/simple)** - Multiple services with health checks
- **[Advanced](../example/advanced)** - Database integration

## Common Patterns

### Service with Dependencies

```go
type APIService struct {
    db    *sql.DB
    cache *redis.Client
    ready bool
}

func NewAPIService(db *sql.DB, cache *redis.Client) *APIService {
    return &APIService{
        db:    db,
        cache: cache,
    }
}

func (s *APIService) HealthChecks() []health.HealthChecker {
    return []health.HealthChecker{
        checks.NewDatabaseCheck("api-db", s.db),
        health.NewCustomCheck("api-cache", s.checkCache),
    }
}

func (s *APIService) checkCache(ctx context.Context) health.HealthResult {
    err := s.cache.Ping(ctx).Err()
    if err != nil {
        return health.NewUnhealthyResult("Cache connection failed").
            WithDetails("error", err.Error())
    }
    return health.NewHealthyResult("Cache is healthy")
}
```

### Graceful Shutdown

```go
func (s *APIService) Shutdown(ctx context.Context) error {
    logger.Info(ctx, "API Service shutting down...")
    
    // Mark as not ready first
    s.ready = false
    
    // Close connections
    if s.db != nil {
        s.db.Close()
    }
    if s.cache != nil {
        s.cache.Close()
    }
    
    logger.Info(ctx, "API Service shutdown completed")
    return nil
}
```

## Troubleshooting

### Application won't start

1. Check Go version: `go version`
2. Verify dependencies: `go mod tidy`
3. Check for port conflicts
4. Review configuration values

### Health checks failing

1. Check endpoint URLs
2. Verify network connectivity
3. Review timeout settings
4. Check service dependencies

### Performance issues

1. Enable profiling endpoints
2. Check resource usage
3. Review health check frequency
4. Monitor application metrics

For more help, check our [troubleshooting guide](./troubleshooting.md) or open an issue on GitHub.

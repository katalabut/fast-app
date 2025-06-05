# FastApp

<div align="center">

![FastApp Logo](./assets/logo.jpeg)

**A lightweight, production-ready application framework for Go**

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue.svg)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/katalabut/fast-app)](https://goreportcard.com/report/github.com/katalabut/fast-app)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Coverage](https://img.shields.io/badge/Coverage-85%25-green.svg)](https://github.com/katalabut/fast-app)

[Features](#features) •
[Installation](#installation) •
[Quick Start](#quick-start) •
[Documentation](#documentation) •
[Examples](#examples) •
[Contributing](#contributing)

</div>

---

## Overview

FastApp is a lightweight, opinionated framework for building production-ready Go applications with minimal boilerplate. It provides essential features like graceful shutdown, health checks, configuration management, and observability out of the box.

## Features

### 🚀 **Core Features**
- **Simple API** - Minimal boilerplate, maximum productivity
- **Graceful Shutdown** - Proper service lifecycle management with timeouts
- **Configuration Management** - Struct-based configuration with environment variable support
- **Structured Logging** - Built-in zap integration with context support

### 🏥 **Health & Monitoring**
- **Health Checks** - Built-in liveness and readiness probes
- **Kubernetes Ready** - Standard `/health/live` and `/health/ready` endpoints
- **Auto-Discovery** - Automatic health check collection from services
- **Multiple Strategies** - Flexible aggregation strategies (all-healthy, majority, weighted)

### 📊 **Unified Observability**
- **Single Port** - All observability endpoints on one port (9090 by default)
- **Metrics** - Prometheus metrics at `/metrics`
- **Health Checks** - Kubernetes-compatible health endpoints
- **Debug & Profiling** - Go pprof endpoints at `/debug/pprof/*`
- **Panic Recovery** - Automatic panic handling with logging
- **Auto MaxProcs** - Automatic GOMAXPROCS configuration

### ⚙️ **Developer Experience**
- **Type Safety** - Leverages Go generics for type-safe configuration
- **Hot Reload** - Development-friendly configuration reloading
- **Extensible** - Plugin-friendly architecture
- **Well Tested** - Comprehensive test coverage

## Installation

```bash
go get github.com/katalabut/fast-app
```

**Requirements:**
- Go 1.21 or higher
- No external dependencies for core functionality

## Quick Start

### Basic Application

```go
package main

import (
    "context"

    fastapp "github.com/katalabut/fast-app"
    "github.com/katalabut/fast-app/config"
    "github.com/katalabut/fast-app/configloader"
)

type AppConfig struct {
    App config.App
}

type MyService struct{}

func (s *MyService) Run(ctx context.Context) error {
    // Your service logic here
    <-ctx.Done()
    return nil
}

func (s *MyService) Shutdown(ctx context.Context) error {
    // Cleanup logic here
    return nil
}

func main() {
    cfg, _ := configloader.New[AppConfig]()

    fastapp.New(cfg.App).
        Add(&MyService{}).
        Start()
}
```

### With Health Checks

```go
package main

import (
    "context"

    fastapp "github.com/katalabut/fast-app"
    "github.com/katalabut/fast-app/config"
    "github.com/katalabut/fast-app/configloader"
    "github.com/katalabut/fast-app/health"
    "github.com/katalabut/fast-app/health/checks"
)

type APIService struct {
    ready bool
}

func (s *APIService) Run(ctx context.Context) error {
    // Initialize service
    s.ready = true
    <-ctx.Done()
    return nil
}

func (s *APIService) Shutdown(ctx context.Context) error {
    s.ready = false
    return nil
}

// Implement HealthProvider interface
func (s *APIService) HealthChecks() []health.HealthChecker {
    return []health.HealthChecker{
        health.NewCustomCheck("api-readiness", func(ctx context.Context) health.HealthResult {
            if s.ready {
                return health.NewHealthyResult("API service is ready")
            }
            return health.NewUnhealthyResult("API service is not ready")
        }),
    }
}

func main() {
    cfg, _ := configloader.New[Config]()
    
    // Add global health checks
    httpCheck := checks.NewHTTPCheck("external-api", "https://api.example.com/health")
    
    app := fastapp.New(cfg.App).
        WithHealthChecks(httpCheck).
        Add(&APIService{})
    
    app.SetReady(true)
    app.Start()
}
```

## Health Checks

FastApp provides a comprehensive health check system for monitoring application and dependency health.

### HTTP Endpoints

All endpoints are available on the same port (9090 by default):

- `GET /health/live` - Liveness probe (always returns 200 if process is alive)
- `GET /health/ready` - Readiness probe (returns 200 if application is ready to serve traffic)
- `GET /health/checks` - Detailed health information for all registered checks
- `GET /metrics` - Prometheus metrics endpoint
- `GET /debug/pprof/*` - Go profiling endpoints (heap, goroutine, cpu, etc.)

### Built-in Health Checks

```go
import "github.com/katalabut/fast-app/health/checks"

// HTTP endpoint check
httpCheck := checks.NewHTTPCheck("api", "https://api.example.com/health")

// Database check
dbCheck := checks.NewDatabaseCheck("postgres", db)

// Custom check
customCheck := health.NewCustomCheck("business-logic", func(ctx context.Context) health.HealthResult {
    // Your health check logic
    return health.NewHealthyResult("All systems operational")
})
```

### Service Health Checks

Services can provide their own health checks by implementing the `HealthProvider` interface:

```go
func (s *MyService) HealthChecks() []health.HealthChecker {
    return []health.HealthChecker{
        health.NewCustomCheck("my-service-check", s.checkHealth),
    }
}
```

## Configuration

FastApp uses struct-based configuration with automatic environment variable binding:

```go
type AppConfig struct {
    App      config.App
    Database DatabaseConfig
}

type DatabaseConfig struct {
    URL      string `default:"postgres://localhost/mydb"`
    MaxConns int    `default:"10"`
}
```

### Logger Configuration

The logger is automatically configured when you create the FastApp instance with `fastapp.New()`. This means logging will use your configuration immediately, not just when `Start()` is called:

```go
func main() {
    cfg, _ := configloader.New[AppConfig]()

    // Logger is configured here with your settings
    app := fastapp.New(cfg.App)

    // This log will use your configured logger (AppName, DevMode, etc.)
    logger.Info(context.Background(), "Application initialized")

    app.Add(&MyService{}).Start()
}
```

## Examples

Check out the [examples](./example) directory for complete working examples:

- **[Basic](./example/basic)** - Simple application with health checks
- **[Simple](./example/simple)** - Multiple services with comprehensive health monitoring
- **[Advanced](./example/advanced)** - Database integration and complex health checks

## Documentation

### Core Interfaces

```go
// Service interface that all services must implement
type Service interface {
    Run(ctx context.Context) error
    Shutdown(ctx context.Context) error
}

// Optional: Provide health checks
type HealthProvider interface {
    HealthChecks() []HealthChecker
}

// Optional: Control service readiness
type ReadinessController interface {
    SetReady(ready bool)
    IsReady() bool
}
```

### Health Check Strategies

- **AllHealthyStrategy** (default) - All checks must be healthy
- **MajorityHealthyStrategy** - Majority of checks must be healthy
- **WeightedStrategy** - Considers component importance (Critical, Important, Optional)

### Kubernetes Integration

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: myapp:latest
    ports:
    - containerPort: 8080  # Your application port
      name: http
    - containerPort: 9090  # Observability port
      name: observability
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

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development

```bash
# Clone the repository
git clone https://github.com/katalabut/fast-app.git
cd fast-app

# Run tests
go test ./...

# Run examples
cd example/simple
go run .
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [x] Health Checks & Readiness/Liveness Probes
- [ ] Named Services & Selective Running
- [ ] Dependency Injection Container
- [ ] Enhanced Metrics & Observability
- [ ] Configuration Hot Reload

## Support

- 📖 [Documentation](./docs)
- 🐛 [Issue Tracker](https://github.com/katalabut/fast-app/issues)
- 💬 [Discussions](https://github.com/katalabut/fast-app/discussions)

---

<div align="center">
Made with ❤️ by the FastApp team
</div>

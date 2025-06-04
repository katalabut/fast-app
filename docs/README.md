# FastApp Documentation

Welcome to the FastApp documentation! This guide will help you understand and use FastApp effectively.

## Table of Contents

### Getting Started
- [**Getting Started**](./getting-started.md) - Your first FastApp application
- [**Installation**](./installation.md) - Installation and setup guide
- [**Quick Start**](./quick-start.md) - 5-minute tutorial

### Core Concepts
- [**Architecture**](./architecture.md) - System design and components
- [**Services**](./services.md) - Building and managing services
- [**Configuration**](./configuration.md) - Configuration management
- [**Health Checks**](./health-checks.md) - Monitoring and health probes

### Advanced Topics
- [**Deployment**](./deployment.md) - Production deployment guide
- [**Monitoring**](./monitoring.md) - Observability and metrics
- [**Performance**](./performance.md) - Performance optimization
- [**Security**](./security.md) - Security best practices

### Reference
- [**API Reference**](./api-reference.md) - Complete API documentation
- [**Configuration Reference**](./configuration-reference.md) - All configuration options
- [**Examples**](./examples.md) - Code examples and patterns
- [**Troubleshooting**](./troubleshooting.md) - Common issues and solutions

## Quick Links

### 🚀 **New to FastApp?**
Start with the [Getting Started Guide](./getting-started.md) to build your first application in minutes.

### 🏥 **Need Health Checks?**
Check out the [Health Checks Guide](./health-checks.md) for comprehensive monitoring setup.

### 🔧 **Configuration Help?**
See the [Configuration Guide](./configuration.md) for environment-specific setups.

### 🚢 **Ready to Deploy?**
Follow the [Deployment Guide](./deployment.md) for production-ready deployments.

## Examples

### Basic Application
```go
package main

import (
    "context"
    fastapp "github.com/katalabut/fast-app"
    "github.com/katalabut/fast-app/configloader"
)

type MyService struct{}

func (s *MyService) Run(ctx context.Context) error {
    <-ctx.Done()
    return nil
}

func (s *MyService) Shutdown(ctx context.Context) error {
    return nil
}

func main() {
    cfg, _ := configloader.New[fastapp.Config]()
    fastapp.New(cfg).Add(&MyService{}).Start()
}
```

### With Health Checks
```go
func (s *MyService) HealthChecks() []health.HealthChecker {
    return []health.HealthChecker{
        health.NewCustomCheck("my-service", func(ctx context.Context) health.HealthResult {
            return health.NewHealthyResult("Service is healthy")
        }),
    }
}
```

## Community

### Getting Help
- 📖 [Documentation](https://github.com/katalabut/fast-app/docs)
- 🐛 [Issue Tracker](https://github.com/katalabut/fast-app/issues)
- 💬 [Discussions](https://github.com/katalabut/fast-app/discussions)
- 📧 [Email Support](mailto:support@fastapp.dev)

### Contributing
- 🤝 [Contributing Guide](../CONTRIBUTING.md)
- 🔄 [Pull Requests](https://github.com/katalabut/fast-app/pulls)
- 📋 [Roadmap](../ROADMAP.md)
- 🎯 [Good First Issues](https://github.com/katalabut/fast-app/labels/good%20first%20issue)

## Frequently Asked Questions

### General

**Q: What makes FastApp different from other Go frameworks?**
A: FastApp focuses on production-ready applications with minimal boilerplate. It provides essential features like health checks, graceful shutdown, and observability out of the box.

**Q: Is FastApp suitable for microservices?**
A: Yes! FastApp is designed with microservices in mind, providing health checks, service discovery, and container-friendly features.

**Q: Can I use FastApp with existing Go applications?**
A: Absolutely! FastApp is designed to be incrementally adoptable. You can start by wrapping existing services.

### Technical

**Q: How do I add custom health checks?**
A: Implement the `HealthProvider` interface in your service or add global checks using `WithHealthChecks()`.

**Q: Can I customize the health check endpoints?**
A: Yes! Configure the paths and ports in the `HealthConfig` section of your configuration.

**Q: How do I handle database connections?**
A: Use the built-in database health checks and pass your database connection to services through dependency injection.

**Q: Is there support for graceful shutdown?**
A: Yes! FastApp handles graceful shutdown automatically with configurable timeouts.

### Deployment

**Q: How do I deploy FastApp to Kubernetes?**
A: FastApp provides standard health check endpoints that work with Kubernetes liveness and readiness probes. See the [Deployment Guide](./deployment.md).

**Q: Can I run multiple services in one application?**
A: Yes! Add multiple services using `app.Add()`. Each service runs in its own goroutine.

**Q: How do I configure different environments?**
A: Use environment variables to override configuration values. FastApp automatically binds struct fields to environment variables.

## Version Compatibility

| FastApp Version | Go Version | Status |
|----------------|------------|---------|
| v1.x.x         | >= 1.21    | ✅ Active |
| v0.x.x         | >= 1.19    | 🔄 Beta |

## License

FastApp is released under the [MIT License](../LICENSE).

## Changelog

See [CHANGELOG.md](../CHANGELOG.md) for version history and breaking changes.

---

**Need help?** Don't hesitate to [open an issue](https://github.com/katalabut/fast-app/issues) or start a [discussion](https://github.com/katalabut/fast-app/discussions)!

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Named services and selective running
- Dependency injection container
- Configuration hot reload
- Enhanced metrics and observability

## [0.2.0] - 2024-01-15

### Added
- **Health Checks System** - Comprehensive health monitoring
  - Liveness and readiness probes
  - HTTP endpoints (`/health/live`, `/health/ready`, `/health/checks`)
  - Built-in health checks (HTTP, Database, Custom)
  - Automatic health check collection from services
  - Multiple aggregation strategies (AllHealthy, Majority, Weighted)
  - Health check caching with configurable TTL
  - Kubernetes-compatible health endpoints
- **Health Check Interfaces**
  - `HealthProvider` interface for services
  - `ReadinessController` interface for service readiness management
  - `HealthChecker` interface for custom health checks
- **Built-in Health Checks**
  - `HTTPCheck` - HTTP endpoint health verification
  - `DatabaseCheck` - Database connection health verification
  - `CustomCheck` - User-defined health check logic
- **Health Server**
  - Dedicated HTTP server for health endpoints
  - Configurable ports and paths
  - JSON response format with detailed information
  - Proper HTTP status codes for different health states
- **Configuration**
  - `HealthConfig` for health system configuration
  - Environment variable binding for health settings
  - Configurable timeouts and cache TTL
- **Examples**
  - Basic health check example
  - Simple multi-service example
  - Advanced database integration example
- **Documentation**
  - Comprehensive health checks documentation
  - Architecture documentation
  - Getting started guide
  - API reference

### Changed
- **Package Renaming**
  - Renamed main package from `app` to `fastapp` for better external usage
  - Renamed `config` package to `configloader` for clarity
- **Application Interface**
  - Added `WithHealthChecks()` method for global health checks
  - Added `SetReady()` and `IsReady()` methods for application readiness
  - Enhanced service registration to automatically collect health checks

### Fixed
- Improved error handling in service lifecycle management
- Better panic recovery with detailed logging
- Fixed graceful shutdown timeout handling

### Security
- Health endpoints don't expose sensitive information
- Configurable health server port for isolation

## [0.1.0] - 2024-01-01

### Added
- **Core Framework**
  - Basic application container (`fastapp.App`)
  - Service interface (`Service`) with `Run()` and `Shutdown()` methods
  - Graceful shutdown with configurable timeouts
  - Service lifecycle management
- **Configuration System**
  - Struct-based configuration with `configloader`
  - Environment variable binding
  - Default value support through struct tags
  - Type-safe configuration loading
- **Logging**
  - Structured logging with zap integration
  - Context-aware logging
  - Configurable log levels
- **Debug Server**
  - Built-in debug server with metrics endpoint
  - Prometheus metrics support
  - pprof profiling endpoints
  - Health check endpoint (basic)
- **Observability**
  - Automatic GOMAXPROCS configuration
  - Panic recovery with logging
  - Application metrics collection
- **Examples**
  - Basic application example
  - Service with configuration example
- **Documentation**
  - README with quick start guide
  - Basic API documentation
  - Contributing guidelines

### Technical Details
- Go 1.21+ support
- Minimal external dependencies
- Production-ready defaults
- Container-friendly design

## [0.0.1] - 2023-12-15

### Added
- Initial project setup
- Basic project structure
- Go module initialization
- MIT License
- Initial README

---

## Release Notes

### v0.2.0 - Health Checks Release

This release introduces a comprehensive health check system that makes FastApp truly production-ready. The health check system provides:

- **Kubernetes Integration**: Standard `/health/live` and `/health/ready` endpoints
- **Automatic Discovery**: Services can provide their own health checks
- **Flexible Strategies**: Multiple ways to aggregate health check results
- **Built-in Checks**: Common health checks for HTTP endpoints and databases
- **Performance**: Caching and timeout management for efficient health monitoring

**Breaking Changes:**
- Package `app` renamed to `fastapp`
- Package `config` renamed to `configloader`
- Import paths need to be updated

**Migration Guide:**
```go
// Old
import "github.com/katalabut/fast-app/app"
import "github.com/katalabut/fast-app/config"

// New
import fastapp "github.com/katalabut/fast-app"
import "github.com/katalabut/fast-app/configloader"
```

### v0.1.0 - Initial Release

The first stable release of FastApp provides a solid foundation for building Go applications with:

- Simple service management
- Configuration handling
- Graceful shutdown
- Basic observability

This release establishes the core patterns and interfaces that will be extended in future versions.

---

## Upgrade Guides

### Upgrading from v0.1.x to v0.2.x

1. **Update import paths:**
   ```bash
   # Use your preferred method to update imports
   find . -name "*.go" -exec sed -i 's|github.com/katalabut/fast-app/app|github.com/katalabut/fast-app|g' {} \;
   find . -name "*.go" -exec sed -i 's|github.com/katalabut/fast-app/config|github.com/katalabut/fast-app/configloader|g' {} \;
   ```

2. **Update package references:**
   ```go
   // Old
   app.New(cfg)
   config.New[Config]()
   
   // New
   fastapp.New(cfg)
   configloader.New[Config]()
   ```

3. **Add health checks (optional but recommended):**
   ```go
   // Add to your services
   func (s *MyService) HealthChecks() []health.HealthChecker {
       return []health.HealthChecker{
           health.NewCustomCheck("my-service", s.checkHealth),
       }
   }
   ```

4. **Update configuration (optional):**
   ```go
   type Config struct {
       App    fastapp.Config
       Health fastapp.HealthConfig  // Add health configuration
       // ... your other config
   }
   ```

## Support

For questions about releases or upgrade issues:
- 📖 [Documentation](./docs)
- 🐛 [Issue Tracker](https://github.com/katalabut/fast-app/issues)
- 💬 [Discussions](https://github.com/katalabut/fast-app/discussions)

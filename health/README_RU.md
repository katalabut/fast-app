# Health Checks

FastApp provides a comprehensive health check system for monitoring application and dependency health.

## Основные концепции

### Типы проверок

1. **Liveness Probe** - проверяет, жив ли процесс (для рестарта контейнера)
2. **Readiness Probe** - проверяет, готов ли процесс принимать трафик (для load balancer)
3. **Health Check** - проверка конкретного компонента (DB, Redis, API, etc.)

### HTTP Endpoints

- `GET /health/live` - Liveness probe (всегда 200 если процесс жив)
- `GET /health/ready` - Readiness probe (200 если приложение готово)
- `GET /health/checks` - Детальная информация по всем health checks

## Быстрый старт

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

// Реализуем HealthProvider
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
    
    // Глобальные health checks
    httpCheck := checks.NewHTTPCheck("external-api", "https://api.example.com/health")
    
    app := fastapp.New(cfg.App).
        WithHealthChecks(httpCheck).
        Add(&MyService{})
    
    app.SetReady(true)
    app.Start()
}
```

## Встроенные Health Checks

### HTTP Check

```go
import "github.com/katalabut/fast-app/health/checks"

// Простая проверка
httpCheck := checks.NewHTTPCheck("api", "https://api.example.com/health")

// С дополнительными опциями
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

// Простая проверка подключения
dbCheck := checks.NewDatabaseCheck("postgres", db)

// С кастомным запросом
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
    // Ваша логика проверки
    if someCondition {
        return health.NewHealthyResult("All systems operational")
    }
    return health.NewDegradedResult("Performance degraded").
        WithDetails("response_time", "500ms").
        WithDetails("threshold", "200ms")
})
```

## Интерфейсы

### HealthProvider

Сервисы могут предоставлять свои health checks:

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

Сервисы могут управлять своим состоянием готовности:

```go
type ReadinessController interface {
    SetReady(ready bool)
    IsReady() bool
}

func (s *MyService) Run(ctx context.Context) error {
    // Инициализация...
    s.SetReady(true) // Сигнализируем что готовы
    
    // Основная логика...
    <-ctx.Done()
    return nil
}
```

## Стратегии агрегации

### AllHealthyStrategy (по умолчанию)

Все проверки должны быть healthy для общего статуса healthy:

```go
strategy := &health.AllHealthyStrategy{}
```

### MajorityHealthyStrategy

Большинство проверок должно быть healthy:

```go
import "github.com/katalabut/fast-app/health/strategies"

strategy := &strategies.MajorityHealthyStrategy{}
```

### WeightedStrategy

Учитывает важность компонентов:

```go
import "github.com/katalabut/fast-app/health/strategies"

strategy := strategies.NewWeightedStrategy(map[string]health.ComponentImportance{
    "database":     health.Critical,    // обязательно healthy
    "cache":        health.Important,   // может быть degraded
    "external-api": health.Optional,    // может быть unhealthy
})
```

## Конфигурация

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

## Примеры ответов

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

## Интеграция с Kubernetes

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

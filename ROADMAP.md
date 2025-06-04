# FastApp Roadmap - План развития

## Философия
FastApp - это легковесная библиотека для быстрого создания production-ready приложений на Go. Цель: предоставить все необходимое для запуска приложения, но не превращаться в тяжелый фреймворк.

## Основные принципы
- ✅ Простота использования
- ✅ Минимальные зависимости  
- ✅ Graceful shutdown из коробки
- ✅ Конфигурация через структуры
- ❌ НЕ HTTP фреймворк (пользователь сам выбирает gin/echo/fiber/etc)
- ❌ НЕ ORM или database layer

---

## 🎯 Приоритетные фичи

### 1. Health Checks & Readiness/Liveness Probes
**Статус:** � В разработке
**Приоритет:** Высокий

#### Архитектура Health Checks

##### Основные концепции
1. **Liveness Probe** - "жив ли процесс?" (для рестарта контейнера)
2. **Readiness Probe** - "готов ли принимать трафик?" (для load balancer)
3. **Health Check** - проверка конкретного компонента (DB, Redis, API)

##### Интерфейсы

```go
// Основной интерфейс для health check
type HealthChecker interface {
    Name() string
    Check(ctx context.Context) HealthResult
}

// Результат проверки
type HealthResult struct {
    Status   HealthStatus
    Message  string
    Details  map[string]interface{}
    Duration time.Duration
}

type HealthStatus string
const (
    StatusHealthy   HealthStatus = "healthy"
    StatusUnhealthy HealthStatus = "unhealthy"
    StatusDegraded  HealthStatus = "degraded"
)

// Сервисы могут предоставлять свои health checks
type HealthProvider interface {
    HealthChecks() []HealthChecker
}

// Сервисы могут управлять своим состоянием
type ReadinessController interface {
    SetReady(ready bool)
    IsReady() bool
}
```

##### Встроенные Health Checks

```go
// Database health check
health.NewDatabaseCheck("postgres", db, health.DatabaseOptions{
    PingTimeout: 5*time.Second,
    Query: "SELECT 1", // optional custom query
})

// Redis health check
health.NewRedisCheck("cache", redis, health.RedisOptions{
    PingTimeout: 3*time.Second,
})

// HTTP endpoint health check
health.NewHTTPCheck("external-api", "https://api.example.com/health",
    health.HTTPOptions{
        Timeout: 10*time.Second,
        ExpectedStatus: 200,
        ExpectedBody: `{"status":"ok"}`, // optional
    })

// Custom health check
health.NewCustomCheck("business-logic", func(ctx context.Context) health.HealthResult {
    // custom logic
    return health.HealthResult{
        Status: health.StatusHealthy,
        Message: "All good",
    }
})
```

##### Использование в сервисах

```go
type APIService struct {
    db    *sql.DB
    redis *redis.Client
    ready bool
}

// Сервис предоставляет свои health checks
func (s *APIService) HealthChecks() []health.HealthChecker {
    return []health.HealthChecker{
        health.NewDatabaseCheck("api-db", s.db),
        health.NewRedisCheck("api-cache", s.redis),
        health.NewCustomCheck("api-business", s.checkBusinessLogic),
    }
}

// Сервис может управлять своей готовностью
func (s *APIService) SetReady(ready bool) {
    s.ready = ready
}

func (s *APIService) IsReady() bool {
    return s.ready
}

func (s *APIService) Run(ctx context.Context) error {
    // Инициализация...
    s.SetReady(true) // Сигнализируем что готовы

    // Основная логика...
    return nil
}
```

##### Конфигурация в приложении

```go
type HealthConfig struct {
    Enabled     bool          `default:"true"`
    LivePath    string        `default:"/health/live"`
    ReadyPath   string        `default:"/health/ready"`
    Timeout     time.Duration `default:"30s"`
    Port        int           `default:"8080"` // отдельный порт для health checks
}

app := fastapp.New(cfg).
    WithHealthChecks(cfg.Health,
        // Глобальные health checks
        health.NewHTTPCheck("external-service", "https://api.example.com/health"),
    ).
    Add(apiService).  // автоматически соберет health checks из сервиса
    Add(workerService).
    Start()
```

##### HTTP Endpoints

```
GET /health/live   -> Liveness probe (всегда 200 если процесс жив)
GET /health/ready  -> Readiness probe (200 если все сервисы ready)
GET /health/checks -> Детальная информация по всем checks

Response format:
{
  "status": "healthy|degraded|unhealthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "duration": "45ms",
  "checks": {
    "api-db": {
      "status": "healthy",
      "message": "Connection successful",
      "duration": "12ms"
    },
    "api-cache": {
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

##### Стратегии агрегации

```go
// Различные стратегии для определения общего статуса
type AggregationStrategy interface {
    Aggregate(results map[string]HealthResult) HealthStatus
}

// Все должны быть healthy
health.AllHealthyStrategy{}

// Большинство должно быть healthy
health.MajorityHealthyStrategy{}

// Критичные компоненты должны быть healthy
health.WeightedStrategy{
    "database": health.Critical,    // обязательно healthy
    "cache":    health.Important,   // может быть degraded
    "metrics":  health.Optional,    // может быть unhealthy
}
```

### 2. Named Services & Selective Running
**Статус:** 🔴 Не реализовано  
**Приоритет:** Высокий

#### Описание
Возможность именовать сервисы и запускать только выбранные через CLI флаги.

#### Функциональность
- Именование сервисов при добавлении
- CLI флаги для выборочного запуска
- Dependency resolution между сервисами
- Валидация зависимостей при старте

#### Пример использования
```go
app := fastapp.New(cfg).
    AddNamed("api", apiService).
    AddNamed("worker", workerService).
    AddNamed("scheduler", schedulerService, 
        fastapp.DependsOn("api")). // scheduler зависит от api
    Start()

// Запуск из командной строки:
// ./app --services=api,worker
// ./app --services=scheduler  // автоматически запустит api тоже
// ./app --exclude=worker      // запустит все кроме worker
```

### 3. Dependency Injection Container
**Статус:** 🔴 Не реализовано  
**Приоритет:** Средний

#### Описание
Простой DI контейнер для управления зависимостями между сервисами.

#### Функциональность
- Регистрация зависимостей (singleton, transient)
- Автоматическое внедрение в сервисы
- Lifecycle management (инициализация, cleanup)
- Type-safe API через generics

#### Пример использования
```go
// Регистрация зависимостей
container := fastapp.NewContainer().
    RegisterSingleton(func() *sql.DB { return db }).
    RegisterSingleton(func() *redis.Client { return rdb }).
    RegisterTransient(func(db *sql.DB) *UserRepository { 
        return &UserRepository{db: db} 
    })

// Сервис с автоматическим внедрением зависимостей
type APIService struct {
    DB       *sql.DB          `inject:""`
    Redis    *redis.Client    `inject:""`
    UserRepo *UserRepository  `inject:""`
}

app := fastapp.New(cfg).
    WithContainer(container).
    Add(&APIService{}).
    Start()
```

---

## 🚀 Дополнительные фичи

### 4. Metrics & Observability
**Статус:** 🟡 Частично (есть prometheus endpoint)  
**Приоритет:** Средний

#### Что добавить
- Встроенные метрики приложения (uptime, goroutines, memory)
- Автоматические метрики для сервисов (start/stop events)
- Интеграция с OpenTelemetry для tracing
- Structured logging с correlation ID

### 5. Configuration Validation & Hot Reload
**Статус:** 🔴 Не реализовано  
**Приоритет:** Низкий

#### Функциональность
- Валидация конфигурации с подробными ошибками
- Hot reload конфигурации без перезапуска
- Configuration profiles (dev, staging, prod)
- Secrets management integration

### 6. Service Discovery & Communication
**Статус:** 🔴 Не реализовано  
**Приоритет:** Низкий

#### Функциональность
- Простой service registry для межсервисного общения
- Event bus для асинхронной коммуникации между сервисами
- Circuit breaker pattern для внешних зависимостей

### 7. Development Tools
**Статус:** 🔴 Не реализовано  
**Приоритет:** Низкий

#### Функциональность
- Live reload в development режиме
- Автогенерация документации по сервисам
- CLI для управления приложением
- Профилирование и debugging endpoints

---

## 📋 Детальный план реализации

### Фаза 1: Health Checks (2-3 недели)

#### Этап 1.1: Базовые интерфейсы и структуры (2-3 дня)
1. Создать пакет `health/` с основными интерфейсами
2. Реализовать `HealthChecker`, `HealthResult`, `HealthStatus`
3. Создать базовый `HealthManager` для сбора и выполнения проверок
4. Добавить поддержку timeout и context cancellation

#### Этап 1.2: Встроенные health checks (3-4 дня)
1. `DatabaseCheck` - проверка подключения к БД (SQL ping + custom query)
2. `RedisCheck` - проверка Redis (ping + optional get/set test)
3. `HTTPCheck` - проверка HTTP endpoints (status code + optional body match)
4. `CustomCheck` - wrapper для пользовательских функций
5. Добавить конфигурируемые timeouts и retry logic

#### Этап 1.3: Интеграция с сервисами (2-3 дня)
1. Расширить интерфейс `Service` для поддержки `HealthProvider`
2. Добавить `ReadinessController` для управления готовностью сервисов
3. Реализовать автоматический сбор health checks из сервисов
4. Добавить lifecycle management (старт/стоп health checks)

#### Этап 1.4: HTTP endpoints и агрегация (3-4 дня)
1. Создать HTTP handler для health endpoints
2. Реализовать стратегии агрегации результатов
3. Добавить JSON serialization с детальной информацией
4. Интегрировать с существующим debug server или создать отдельный

#### Этап 1.5: Тестирование и документация (2-3 дня)
1. Unit тесты для всех компонентов
2. Integration тесты с реальными зависимостями
3. Примеры использования
4. Обновить README с документацией по health checks

#### Структура файлов:
```
health/
├── health.go           # Основные интерфейсы и типы
├── manager.go          # HealthManager для координации
├── checks/
│   ├── database.go     # DatabaseCheck
│   ├── redis.go        # RedisCheck
│   ├── http.go         # HTTPCheck
│   └── custom.go       # CustomCheck
├── strategies/
│   ├── all_healthy.go  # AllHealthyStrategy
│   ├── majority.go     # MajorityHealthyStrategy
│   └── weighted.go     # WeightedStrategy
├── server/
│   └── http.go         # HTTP endpoints handler
└── examples/
    └── basic/          # Примеры использования
```

#### Ключевые решения:
1. **Отдельный порт для health checks** - изоляция от основного трафика
2. **Graceful degradation** - приложение продолжает работать при проблемах с некритичными компонентами
3. **Кэширование результатов** - избежать частых проверок тяжелых компонентов
4. **Structured logging** - детальные логи для debugging
5. **Prometheus metrics** - экспорт метрик health checks

### Фаза 2: Named Services (1-2 недели)  
1. Расширить структуру Service для поддержки имен
2. Добавить CLI parsing для service selection
3. Реализовать dependency resolution
4. Обновить примеры и документацию

### Фаза 3: Dependency Injection (3-4 недели)
1. Создать пакет `container` с DI функциональностью
2. Реализовать reflection-based injection
3. Добавить lifecycle management
4. Интегрировать с основным приложением

---

## 🎯 Метрики успеха

- Время от идеи до working prototype: < 30 минут
- Размер типичного main.go: < 50 строк
- Время старта приложения: < 1 секунда
- Memory overhead: < 10MB для базового приложения
- Количество внешних зависимостей: < 15

---

## 🤔 Вопросы для обсуждения

1. Нужна ли поддержка gRPC сервисов из коробки?
2. Стоит ли добавить встроенную поддержку rate limiting?
3. Нужен ли встроенный scheduler для cron jobs?
4. Как лучше организовать plugin систему для расширений?
5. Стоит ли добавить поддержку WebSocket connections?

---

## ✅ Статус реализации

### Health Checks & Readiness/Liveness Probes - РЕАЛИЗОВАНО

**Что реализовано:**
- ✅ Базовые интерфейсы и типы (`HealthChecker`, `HealthResult`, `HealthStatus`)
- ✅ Health Manager для координации проверок
- ✅ HTTP сервер с endpoints `/health/live`, `/health/ready`, `/health/checks`
- ✅ Встроенные health checks (HTTP, Database, Custom)
- ✅ Стратегии агрегации (AllHealthy, Majority, Weighted)
- ✅ Автоматический сбор health checks из сервисов
- ✅ Интеграция с основным приложением
- ✅ Кэширование результатов проверок
- ✅ Управление готовностью сервисов
- ✅ Unit тесты
- ✅ Примеры использования
- ✅ Документация

**Примеры:**
- `example/basic/` - базовый пример с health checks
- `example/simple/` - расширенный пример с несколькими сервисами
- `example/advanced/` - пример с database health checks

**Тесты:**
```bash
go test ./health/...
```

**Использование:**
```go
app := fastapp.New(cfg).
    WithHealthChecks(httpCheck, dbCheck).
    Add(myService).
    Start()
```

---

*Этот документ будет обновляться по мере развития проекта*

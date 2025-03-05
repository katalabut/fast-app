# Fast App

Fast App - это легковесная библиотека на Go для быстрого создания приложений с поддержкой graceful shutdown и управления жизненным циклом сервисов.

## Особенности

- 🚀 Простой и понятный API
- 🔄 Graceful shutdown с таймаутами
- 📊 Встроенный debug-сервер
- 📝 Интегрированное логирование (с использованием zap)
- ⚙️ Конфигурация через структуры
- 🔧 Автоматическая настройка GOMAXPROCS
- 🛡️ Обработка паник с логированием

## Установка

```bash
go get github.com/katalabut/fast-app
```

## Быстрый старт

```go
package main

import (
    app "github.com/katalabut/fast-app"
    "github.com/katalabut/fast-app/config"
)

type Config struct {
    App         app.Config
    DebugServer app.DebugServer
}

type MyService struct {}

func (s *MyService) Run(ctx context.Context) error {
    // Ваш код сервиса
    return nil
}

func (s *MyService) Shutdown(ctx context.Context) error {
    // Код для graceful shutdown
    return nil
}

func main() {
    cfg, _ := config.New[Config]()
    
    app.New(cfg.App).
        Add(app.NewDefaultDebugService(cfg.DebugServer)).
        Add(&MyService{}).
        Start()
}
```

## Конфигурация

Библиотека поддерживает следующие опции конфигурации:

```go
type Config struct {
    Logger       logger.Config
    AutoMaxProcs struct {
        Enabled bool
        Min     int
    }
}
```

## Интерфейсы

Каждый сервис должен реализовывать интерфейс `Service`:

```go
type Service interface {
    Run(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

## Лицензия

MIT 
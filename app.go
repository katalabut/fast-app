// Package fastapp provides a lightweight, production-ready application framework for Go.
// It offers essential features like graceful shutdown, health checks, configuration management,
// and observability out of the box with minimal boilerplate.
package fastapp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/katalabut/fast-app/health"
	"github.com/katalabut/fast-app/logger"
	"github.com/katalabut/fast-app/service"
	"github.com/pkg/errors"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	exitCodeOk             = 0
	exitCodeApplicationErr = 1
	exitCodeWatchdog       = 1

	defaultShutdownTimeout = time.Second * 5
	watchdogTimeout        = defaultShutdownTimeout + time.Second*5
)

// App represents the main application instance that manages services,
// health checks, and graceful shutdown.
type App struct {
	config Config
	logger *zap.SugaredLogger

	opts                 options
	runners              []Runner
	healthManager        *health.Manager
	observabilityService *service.ObservabilityService
}

// Runner wraps a service for execution within the application.
type Runner struct {
	service Service
}

// Service defines the interface that all services must implement.
// Services are the core building blocks of a FastApp application.
type Service interface {
	// Run starts the service and blocks until the context is cancelled.
	// It should return nil on graceful shutdown or an error if something goes wrong.
	Run(ctx context.Context) error

	// Shutdown gracefully stops the service within the given context timeout.
	// It should clean up resources and return nil on successful shutdown.
	Shutdown(ctx context.Context) error
}

// New creates a new FastApp application instance with the given configuration and options.
// It initializes the logger with the provided configuration, sets up the health management
// system and prepares the observability server endpoints.
//
// The logger is configured immediately when this function is called, so any subsequent
// logging will use the configuration from the provided config.
//
// Example:
//
//	cfg := fastapp.Config{...}
//	app := fastapp.New(cfg, fastapp.WithVersion("1.0.0"))
//	// Logger is now configured and ready to use
func New(config Config, opts ...Option) *App {
	op := options{
		version:         "",
		stopAllOnErr:    true,
		shutdownTimeout: defaultShutdownTimeout,

		ctx: context.Background(),
		mux: http.NewServeMux(),
	}

	for _, o := range opts {
		o.apply(&op)
	}

	// Initialize logger with configuration
	loggerConfig := logger.Config{
		AppName:    config.Logger.AppName,
		Level:      config.Logger.Level,
		DevMode:    config.Logger.DevMode,
		MessageKey: config.Logger.MessageKey,
		LevelKey:   config.Logger.LevelKey,
		TimeKey:    config.Logger.TimeKey,
	}

	lg, err := logger.InitLogger(loggerConfig, op.version)
	if err != nil {
		panic(errors.Wrap(err, "failed to init logger"))
	}

	// Initialize health manager
	healthManager := health.NewManager(health.ManagerConfig{
		CacheTTL: config.Observability.Health.CacheTTL,
		Strategy: &health.AllHealthyStrategy{},
	})

	// Initialize observability service
	observabilityService := service.NewObservabilityService(config.Observability, healthManager)

	return &App{
		config:               config,
		logger:               lg,
		opts:                 op,
		healthManager:        healthManager,
		observabilityService: observabilityService,
	}
}

// Start begins the application lifecycle, starting all registered services
// and setting up signal handling for graceful shutdown. This method blocks
// until the application is terminated by a signal or an error occurs.
//
// The method handles:
//   - GOMAXPROCS configuration (if enabled)
//   - Health server startup
//   - Service startup with panic recovery
//   - Graceful shutdown coordination
//   - Watchdog timer for forced shutdown
func (a *App) Start() {
	var (
		config = a.config
		lg     = a.logger
	)

	ctx, cancel := signal.NotifyContext(a.opts.ctx, os.Interrupt)
	defer cancel()

	defer func() { _ = lg.Sync() }()

	lg.Info("Starting")

	{
		// Automatically setting GOMAXPROCS.
		if config.AutoMaxProcs.Enabled && config.AutoMaxProcs.Min > 0 {
			if _, err := maxprocs.Set(
				maxprocs.Logger(lg.Infof),
				maxprocs.Min(config.AutoMaxProcs.Min),
			); err != nil {
				lg.Warn("Failed to set GOMAXPROCS", zap.Error(err))
			}
		}
	}

	g, ctx := errgroup.WithContext(ctx)

	// Start observability server (includes health checks, metrics, and debug endpoints)
	if a.config.Observability.Enabled {
		g.Go(a.GracefulShutdown(ctx, a.observabilityService.Shutdown))
		g.Go(func() error {
			return a.observabilityService.Run(ctx)
		})
	}

	for _, run := range a.runners {
		run := run

		g.Go(a.GracefulShutdown(ctx, run.service.Shutdown))
		g.Go(
			func() (rerr error) {
				defer func() {
					// Recovering panic to log it and return error.
					if ec := recover(); ec != nil {
						lg.Errorw(
							"Panic",
							zap.String("panic", fmt.Sprintf("%v", ec)),
							zap.StackSkip("stack", 1),
						)
						rerr = fmt.Errorf("shutting down (panic): %v", ec)

						// Also shutting down all services on error.
						if a.opts.stopAllOnErr {
							cancel()
						}
					}
				}()

				if err := run.service.Run(ctx); err != nil {
					if errors.Is(err, ctx.Err()) {
						// Parent context got cancelled, error is expected.
						lg.Debug("Graceful shutdown")
						return nil
					}
					return err
				}

				return nil
			},
		)
	}

	// if a.config.DebugServer.Enabled {
	// 	lg.Info("Starting debug server", zap.Int("port", a.config.DebugServer.Port))
	// 	g.Go(
	// 		func() error {
	// 			return defaultDebugServer(ctx, a.config.DebugServer.Port)
	// 		},
	// 	)
	// }

	go func() {
		// Guaranteed way to kill application.
		// Helps if f is stuck, e.g. deadlock during shutdown.
		<-ctx.Done()

		// Context is canceled, giving application time to shut down gracefully.

		lg.Info("Waiting for application shutdown")
		time.Sleep(watchdogTimeout)

		// Application is not shutting down gracefully, kill it.
		// This code should not be executed if f is already returned.

		lg.Warn("Graceful shutdown watchdog triggered: forcing shutdown")
		os.Exit(exitCodeWatchdog)
	}()

	if err := g.Wait(); err != nil {
		lg.Errorw("Failed", zap.Error(err))
		os.Exit(exitCodeApplicationErr)
	}

	lg.Info("Application stopped")
	os.Exit(exitCodeOk)
}

// GracefulShutdown creates a shutdown function that waits for the context to be cancelled
// and then executes the provided shutdown functions with a timeout.
// This is used internally to coordinate graceful shutdown of services.
func (a *App) GracefulShutdown(ctx context.Context, f ...func(context.Context) error) func() error {
	return func() error {
		// Wait until g ctx canceled, then try to shut down server.
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.opts.shutdownTimeout)
		defer cancel()

		for _, ff := range f {
			if err := ff(shutdownCtx); err != nil {
				return errors.Wrap(err, "failed to shut down service")
			}
		}

		return nil
	}
}

// Add registers a service with the application. The service will be started
// when Start() is called and will be gracefully shut down on application termination.
//
// If the service implements health.HealthProvider, its health checks will be
// automatically registered with the health management system.
//
// Example:
//
//	app.Add(&MyService{})
//	app.Add(service.NewDefaultDebugService(cfg.DebugServer))
func (a *App) Add(svc Service) *App {
	a.runners = append(
		a.runners, Runner{
			service: svc,
		},
	)

	// Check if service provides health checks
	if healthProvider, ok := svc.(health.HealthProvider); ok {
		healthChecks := healthProvider.HealthChecks()
		a.healthManager.RegisterCheckers(healthChecks)
		logger.Debug(context.Background(), "Registered health checks from service",
			"service_type", fmt.Sprintf("%T", svc),
			"checks_count", len(healthChecks))
	}

	return a
}

// WithHealthChecks adds global health checks to the application
func (a *App) WithHealthChecks(checkers ...health.HealthChecker) *App {
	a.healthManager.RegisterCheckers(checkers)
	logger.Debug(context.Background(), "Registered global health checks",
		"checks_count", len(checkers))
	return a
}

// SetReady sets the application readiness state
func (a *App) SetReady(ready bool) {
	a.healthManager.SetReady(ready)
}

// IsReady returns the application readiness state
func (a *App) IsReady() bool {
	return a.healthManager.IsReady()
}

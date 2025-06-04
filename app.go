package fastapp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/katalabut/fast-app/health"
	"github.com/katalabut/fast-app/health/server"
	"github.com/katalabut/fast-app/logger"
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

type (
	App struct {
		config Config

		opts          options
		runners       []Runner
		healthManager *health.Manager
		healthServer  *server.Server
	}
	Runner struct {
		service Service
	}

	Service interface {
		Run(ctx context.Context) error
		Shutdown(ctx context.Context) error
	}
)

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

	// Initialize health manager
	healthManager := health.NewManager(health.ManagerConfig{
		CacheTTL: config.Health.CacheTTL,
		Strategy: &health.AllHealthyStrategy{},
	})

	// Initialize health server
	healthServer := server.NewServer(server.Config{
		Enabled:   config.Health.Enabled,
		Port:      config.Health.Port,
		LivePath:  config.Health.LivePath,
		ReadyPath: config.Health.ReadyPath,
		CheckPath: config.Health.CheckPath,
		Timeout:   config.Health.Timeout,
	}, healthManager)

	return &App{
		config:        config,
		opts:          op,
		healthManager: healthManager,
		healthServer:  healthServer,
	}
}

func (a *App) Start() {
	var (
		config = a.config
	)

	ctx, cancel := signal.NotifyContext(a.opts.ctx, os.Interrupt)
	defer cancel()

	lg, err := logger.InitLogger(config.Logger, a.opts.version)
	if err != nil {
		panic(errors.Wrap(err, "failed to init logger"))
	}

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

	// Start health server
	if a.config.Health.Enabled {
		g.Go(a.GracefulShutdown(ctx, a.healthServer.Shutdown))
		g.Go(func() error {
			return a.healthServer.Start(ctx)
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

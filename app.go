package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

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

		opts    options
		runners []Runner
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

	return &App{
		config: config,
		opts:   op,
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
	return a
}

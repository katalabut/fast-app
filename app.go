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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	exitCodeOk             = 0
	exitCodeApplicationErr = 1
	exitCodeWatchdog       = 1
)

const (
	defaultShutdownTimeout = time.Second * 5
	watchdogTimeout        = defaultShutdownTimeout + time.Second*5
)

type App struct {
	config Config

	opts options
}

func New(config Config, opts ...Option) *App {
	op := options{
		version:         "",
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

func (a *App) Run(f func(context.Context) error) {
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
	g.Go(
		func() (rerr error) {
			defer lg.Info("Shutting down")
			defer func() {
				// Recovering panic to log it and return error.
				if ec := recover(); ec != nil {
					lg.Error(
						"Panic",
						zap.String("panic", fmt.Sprintf("%v", ec)),
						zap.StackSkip("stack", 1),
					)
					rerr = fmt.Errorf("shutting down (panic): %v", ec)
				}
			}()
			if err := f(ctx); err != nil {
				if errors.Is(err, ctx.Err()) {
					// Parent context got cancelled, error is expected.
					lg.Debug("Graceful shutdown")
					return nil
				}
				return err
			}

			// Also shutting down metrics server to stop error group.
			cancel()

			return nil
		},
	)

	if a.config.DebugServer.Enabled {
		lg.Info("Starting debug server", zap.Int("port", a.config.DebugServer.Port))
		g.Go(
			func() error {
				return defaultDebugServer(ctx, a.config.DebugServer.Port)
			},
		)
	}

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
		lg.Error("Failed", zap.Error(err))
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

func defaultDebugServer(ctx context.Context, port int) (err error) {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	server := http.Server{
		Addr:        fmt.Sprintf(":%d", port),
		ReadTimeout: 5 * time.Second,
		IdleTimeout: 120 * time.Second,
		Handler:     mux,
	}

	go func() {
		<-ctx.Done()
		err = server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "failed to start debug server")
	}

	return
}

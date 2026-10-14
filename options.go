package fastapp

import (
	"context"
	"net/http"
	"time"
)

type options struct {
	observabilityAddress string
	pprofEnabled         bool
	redactHealthDetails  bool
	version              string
	stopAllOnErr         bool
	shutdownTimeout      time.Duration

	ctx context.Context
	mux *http.ServeMux
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

// Option is a functional option for the application.
type Option interface {
	apply(o *options)
}

// WithContext sets the base context for the application. Background context is used by default.
func WithContext(ctx context.Context) Option {
	return optionFunc(
		func(o *options) {
			o.ctx = ctx
		},
	)
}

// WithVersion sets the application version for logging and metrics.
// The version is included in log output and can be used for observability.
func WithVersion(version string) Option {
	return optionFunc(
		func(o *options) {
			o.version = version
		},
	)
}

// WithDisableStopAllOnErr lets other services continue after a Run/Shutdown error or Run panic.
// The failed service is shut down independently. Failures are logged immediately and
// still cause exit code 1 when the remaining services stop; no restart is performed.
// By default, any service failure cancels all services.
func WithDisableStopAllOnErr() Option {
	return optionFunc(
		func(o *options) {
			o.stopAllOnErr = false
		},
	)
}

// WithShutdownTimeout sets the timeout passed to service Shutdown methods.
// The watchdog triggers after this timeout plus 5 seconds, saturating at the
// largest time.Duration. It then allows at most 100 milliseconds for its final
// log flush before forcibly terminating the process.
// Default timeout is 5 seconds.
func WithShutdownTimeout(timeout time.Duration) Option {
	return optionFunc(
		func(o *options) {
			o.shutdownTimeout = timeout
		},
	)
}

// WithObservabilityAddress opts into a specific listener address; the default remains :port.
func WithObservabilityAddress(address string) Option {
	return optionFunc(func(o *options) { o.observabilityAddress = address })
}

// WithPprof opts into default-prefix pprof endpoints if Debug.Enabled is true.
// Prefer a private listener via WithObservabilityAddress before enabling diagnostics.
func WithPprof() Option { return optionFunc(func(o *options) { o.pprofEnabled = true }) }

// WithHealthDetailsRedacted omits checker Details from built-in HTTP responses.
// Checker Message and Status remain public and must not contain secrets.
func WithHealthDetailsRedacted() Option {
	return optionFunc(func(o *options) { o.redactHealthDetails = true })
}

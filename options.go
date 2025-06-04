package fastapp

import (
	"context"
	"net/http"
	"time"
)

type options struct {
	version         string
	stopAllOnErr    bool
	shutdownTimeout time.Duration

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

// WithDisableStopAllOnErr disables the default behavior of stopping all services
// when one service encounters an error. By default, if any service fails,
// all services are gracefully shut down.
func WithDisableStopAllOnErr() Option {
	return optionFunc(
		func(o *options) {
			o.stopAllOnErr = false
		},
	)
}

// WithShutdownTimeout sets the maximum time to wait for services to shut down gracefully.
// If services don't shut down within this timeout, the application will be forcefully terminated.
// Default timeout is 5 seconds.
func WithShutdownTimeout(timeout time.Duration) Option {
	return optionFunc(
		func(o *options) {
			o.shutdownTimeout = timeout
		},
	)
}

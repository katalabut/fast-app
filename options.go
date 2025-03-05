package app

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

func WithVersion(version string) Option {
	return optionFunc(
		func(o *options) {
			o.version = version
		},
	)
}

func WithDisableStopAllOnErr() Option {
	return optionFunc(
		func(o *options) {
			o.stopAllOnErr = false
		},
	)
}

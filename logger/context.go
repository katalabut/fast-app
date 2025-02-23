package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type contextKey struct{}
type contextTags struct{}

var loggerContextKey = contextKey{}
var loggerContextTags = contextTags{}

// ToContext returns new context with specified sugared logger inside.
func ToContext(ctx context.Context, l *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerContextKey, l)
}

// ContextWithKV returns new context with specified logger with field
func ContextWithKV(ctx context.Context, kvs ...interface{}) context.Context {
	l := FromContext(ctx).Desugar()
	result := make([]zap.Field, 0, len(kvs)/2)
	for i := 0; i < len(kvs); i += 2 {
		if i == len(kvs)-1 {
			// Пришло нечетное кол-во kvs
			break
		}
		key, ok := kvs[i].(string)
		if !ok {
			//Ключ поля не является строкой
			l.Warn(fmt.Sprintf("invalid KVs key %v", key))
			continue
		}
		result = append(result, zap.Any(key, kvs[i+1]))
	}
	l.With(result...)

	return ToContext(ctx, l.With(result...).Sugar())
}

func ContextWithTags(ctx context.Context, tags ...string) context.Context {
	if v, ok := ctx.Value(loggerContextTags).([]string); ok {
		tags = append(v, tags...)
	}

	return context.WithValue(ctx, loggerContextTags, tags)
}

// FromContext returns logger from context if set. Otherwise returns global `global` logger.
// In both cases returned logger is populated with `trace_id` & `span_id`.
func FromContext(ctx context.Context) *zap.SugaredLogger {
	l := global

	if logger, ok := ctx.Value(loggerContextKey).(*zap.SugaredLogger); ok {
		l = logger
	}

	//if span, ok := tracer.SpanFromContext(ctx); ok {
	//	if span.Context() != nil {
	//		l = withTraceID(l, span.Context().TraceID())
	//		l = withSpanID(l, span.Context().SpanID())
	//	}
	//}

	return l
}

func withTraceID(l *zap.SugaredLogger, traceId uint64) *zap.SugaredLogger {
	log := l.With(zap.Uint64("trace_id", traceId))
	return log
}

func withSpanID(l *zap.SugaredLogger, spanId uint64) *zap.SugaredLogger {
	log := l.With(zap.Uint64("span_id", spanId))
	return log
}

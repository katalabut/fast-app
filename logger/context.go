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
	l := contextLogger(ctx).Desugar()
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

	return ToContext(ctx, l.With(result...).Sugar())
}

func ContextWithTags(ctx context.Context, tags ...string) context.Context {
	previous, _ := ctx.Value(loggerContextTags).([]string)
	combined := append([]string(nil), previous...)
	combined = append(combined, tags...)
	return context.WithValue(ctx, loggerContextTags, combined)
}

// FromContext returns a context logger, or the synchronized global logger, enriched with context tags.
// Trace/span fields must be provided explicitly through ContextWithKV.
func FromContext(ctx context.Context) *zap.SugaredLogger {
	l := contextLogger(ctx)
	if tags, ok := ctx.Value(loggerContextTags).([]string); ok && len(tags) > 0 {
		l = l.With("tags", tags)
	}
	return l
}

func contextLogger(ctx context.Context) *zap.SugaredLogger {
	if l, ok := ctx.Value(loggerContextKey).(*zap.SugaredLogger); ok && l != nil {
		return l
	}
	return Logger()
}

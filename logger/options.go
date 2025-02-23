package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type coreWithLevel struct {
	zapcore.Core
	level zapcore.Level
}

func (c *coreWithLevel) Enabled(l zapcore.Level) bool {
	return c.level.Enabled(l)
}

func (c *coreWithLevel) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *coreWithLevel) With(fields []zapcore.Field) zapcore.Core {
	return &coreWithLevel{
		c.Core.With(fields),
		c.level,
	}
}

// WithLevel returns `zap.Option` that can be used to create a new logger
// from an existing one with a new logging level
//
// Usage:
//
//	logger.Logger().Desugar().WithOptions(logger.WithLevel(level)).Sugar()
func WithLevel(lvl zapcore.Level) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &coreWithLevel{core, lvl}
	})
}

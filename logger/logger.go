package logger

import (
	"context"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	AppName string

	Level   string `default:"info"`
	DevMode bool   `default:"false"`

	MessageKey string `default:"message"`
	LevelKey   string `default:"severity"`
	TimeKey    string `default:"timestamp"`
}

var (
	// global logger instance.
	global      *zap.SugaredLogger
	globalGuard sync.RWMutex

	level      = zap.NewAtomicLevelAt(zap.InfoLevel)
	defaultCfg = Config{
		Level:      "info",
		MessageKey: "message",
		LevelKey:   "severity",
		TimeKey:    "timestamp",
		AppName:    "app",
		DevMode:    false,
	}

	globalVersion string
)

func init() {
	SetLogger(New(level, defaultCfg))
}

func InitLogger(cfg Config, version string) (*zap.SugaredLogger, error) {
	globalVersion = version

	lvl, err := zapLevelFromString(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to unmurshal log level: %s; err: %v", cfg.Level, err)
	}

	logger := New(lvl, cfg)
	SetLogger(logger)
	return logger, nil
}

func zapLevelFromString(newLogLevel string) (zap.AtomicLevel, error) {
	lvl := zap.NewAtomicLevel()
	err := lvl.UnmarshalText([]byte(newLogLevel))
	return lvl, err
}

// New creates new *zap.SugaredLogger with standard EncoderConfig
func New(lvl zapcore.LevelEnabler, cfg Config, options ...zap.Option) *zap.SugaredLogger {
	if lvl == nil {
		lvl = level
	}
	sink := zapcore.AddSync(os.Stdout)
	options = append(options, zap.ErrorOutput(sink))

	config := zapcore.EncoderConfig{
		TimeKey:        cfg.TimeKey,
		LevelKey:       cfg.LevelKey,
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     cfg.MessageKey,
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	var encoder zapcore.Encoder
	if cfg.DevMode {
		config.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(config)
	} else {
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoder = zapcore.NewJSONEncoder(config)
	}

	return zap.New(zapcore.NewCore(encoder, sink, lvl), options...).With(getZapFields(cfg)...).Sugar()
}

func getZapFields(config Config) []zapcore.Field {
	var fields []zapcore.Field

	if globalVersion != "" {
		fields = append(fields, zap.String("version", globalVersion))
	}

	if config.AppName != "" {
		fields = append(fields, zap.String("application_name", config.AppName))
	}

	return fields
}

// Logger returns current global logger.
func Logger() *zap.SugaredLogger {
	globalGuard.RLock()
	defer globalGuard.RUnlock()
	return global
}

// SetLogger sets global used logger. This function is not thread-safe.
func SetLogger(l *zap.SugaredLogger) {
	globalGuard.Lock()
	defer globalGuard.Unlock()
	global = l
}

func Debug(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Debug(args...)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Debugf(format, args...)
}

func DebugKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Debugw(message, kvs...)
}

func Info(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Info(args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Infof(format, args...)
}

func InfoKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Infow(message, kvs...)
}

func Warn(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Warn(args...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Warnf(format, args...)
}

func WarnKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Warnw(message, kvs...)
}

func Error(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Error(args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Errorf(format, args...)
}

func ErrorKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Errorw(message, kvs...)
}

func Fatal(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Fatal(args...)
}

func Fatalf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Fatalf(format, args...)
}

func FatalKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Fatalw(message, kvs...)
}

package logger

import (
	"context"
	"log/slog"
	"time"
)

type SlogAdapter struct {
	logger *slog.Logger
}

func newSlogAdapter(opts ...Option) *SlogAdapter {
	cfg := defaultConfigs()
	for _, opt := range opts {
		opt(cfg)
	}

	handler := slog.NewJSONHandler(cfg.GetWriter(), &slog.HandlerOptions{
		Level: toSlogLevel(cfg.Level),
	})

	return &SlogAdapter{logger: slog.New(handler)}
}

func (a *SlogAdapter) Debug(msg string, args ...any) { a.logger.Debug(msg, args...) }
func (a *SlogAdapter) Info(msg string, args ...any)  { a.logger.Info(msg, args...) }
func (a *SlogAdapter) Warn(msg string, args ...any)  { a.logger.Warn(msg, args...) }
func (a *SlogAdapter) Error(msg string, args ...any) { a.logger.Error(msg, args...) }

func (a *SlogAdapter) Debugw(msg string, keysAndValues ...any) { a.logger.Debug(msg, keysAndValues...) }
func (a *SlogAdapter) Infow(msg string, keysAndValues ...any)  { a.logger.Info(msg, keysAndValues...) }
func (a *SlogAdapter) Warnw(msg string, keysAndValues ...any)  { a.logger.Warn(msg, keysAndValues...) }
func (a *SlogAdapter) Errorw(msg string, keysAndValues ...any) { a.logger.Error(msg, keysAndValues...) }

func (a *SlogAdapter) Ctx(ctx context.Context) Logger {
	requestID := GetRequestID(ctx)
	if requestID == "" {
		return a
	}

	return &SlogAdapter{logger: a.logger.With("request_id", requestID)}
}

func (a *SlogAdapter) With(args ...any) Logger {
	return &SlogAdapter{logger: a.logger.With(args...)}
}

func (a *SlogAdapter) WithGroup(name string) Logger {
	return &SlogAdapter{logger: a.logger.WithGroup(name)}
}

func (a *SlogAdapter) Log(level Level, msg string, attrs ...Attr) {
	slogLevel := toSlogLevel(level)
	if !a.logger.Enabled(context.Background(), slogLevel) {
		return
	}
	a.logger.Log(context.Background(), slogLevel, msg, toSlogAttrs(attrs)...)
}

func (a *SlogAdapter) LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	a.Ctx(ctx).Log(level, msg, attrs...)
}

func (a *SlogAdapter) LogRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	a.Ctx(ctx).Info("request",
		"method", method,
		"path", path,
		"status", status,
		"duration", duration,
		"status_class", status/100,
	)
}

func toSlogLevel(l Level) slog.Level {
	switch l {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func toSlogAttrs(attrs []Attr) []any {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = slog.Any(attr.Key, attr.Value)
	}
	return args
}

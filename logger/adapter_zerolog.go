package logger

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type ZerologAdapter struct {
	l zerolog.Logger
}

func NewZerologAdapter(appName, env string, opts ...Option) *ZerologAdapter {
	cfg := defaultConfigs()
	for _, opt := range opts {
		opt(cfg)
	}

	zl := zerolog.New(cfg.GetWriter()).With().
		Timestamp().
		Str("service", appName).
		Str("env", env).
		Logger()

	return &ZerologAdapter{l: zl}
}

func (a *ZerologAdapter) Debug(msg string, args ...any) { a.l.Debug().Fields(args).Msg(msg) }
func (a *ZerologAdapter) Info(msg string, args ...any)  { a.l.Info().Fields(args).Msg(msg) }
func (a *ZerologAdapter) Warn(msg string, args ...any)  { a.l.Warn().Fields(args).Msg(msg) }
func (a *ZerologAdapter) Error(msg string, args ...any) { a.l.Error().Fields(args).Msg(msg) }

func (a *ZerologAdapter) Debugw(msg string, kvs ...any) { a.l.Debug().Fields(kvs).Msg(msg) }
func (a *ZerologAdapter) Infow(msg string, kvs ...any)  { a.l.Info().Fields(kvs).Msg(msg) }
func (a *ZerologAdapter) Warnw(msg string, kvs ...any)  { a.l.Warn().Fields(kvs).Msg(msg) }
func (a *ZerologAdapter) Errorw(msg string, kvs ...any) { a.l.Error().Fields(kvs).Msg(msg) }

func (a *ZerologAdapter) Ctx(ctx context.Context) Logger {
	requestID := GetRequestID(ctx)
	if requestID == "" {
		return a
	}
	return &ZerologAdapter{l: a.l.With().Str("request_id", requestID).Logger()}
}

func (a *ZerologAdapter) With(args ...any) Logger {
	return &ZerologAdapter{l: a.l.With().Fields(args).Logger()}
}

func (a *ZerologAdapter) WithGroup(name string) Logger {
	return &ZerologAdapter{l: a.l.With().Dict(name, zerolog.Dict()).Logger()}
}

func (a *ZerologAdapter) Log(level Level, msg string, attrs ...Attr) {
	zlLevel := toZerologLevel(level)
	if zlLevel == zerolog.Disabled {
		return
	}

	event := a.l.WithLevel(zlLevel)
	for _, attr := range attrs {
		event.Any(attr.Key, attr.Value)
	}
	event.Msg(msg)
}

func (a *ZerologAdapter) LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	a.Ctx(ctx).Log(level, msg, attrs...)
}

func (a *ZerologAdapter) LogRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	a.Ctx(ctx).Info("http request",
		"method", method,
		"path", path,
		"status", status,
		"duration", duration,
	)
}

func toZerologLevel(l Level) zerolog.Level {
	switch l {
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

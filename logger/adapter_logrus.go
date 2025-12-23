package logger

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type LogrusAdapter struct {
	entry *logrus.Entry
}

func NewLogrusAdapter(appName, env string, opts ...Option) *LogrusAdapter {
	cfg := defaultConfigs()
	for _, opt := range opts {
		opt(cfg)
	}

	l := logrus.New()

	l.SetOutput(cfg.GetWriter())

	entry := l.WithFields(logrus.Fields{
		"service": appName,
		"env":     env,
	})

	return &LogrusAdapter{entry: entry}
}

func (a *LogrusAdapter) Debug(msg string, args ...any) { a.entry.Debug(args...) }
func (a *LogrusAdapter) Info(msg string, args ...any)  { a.entry.Info(args...) }
func (a *LogrusAdapter) Warn(msg string, args ...any)  { a.entry.Warn(args...) }
func (a *LogrusAdapter) Error(msg string, args ...any) { a.entry.Error(args...) }

func (a *LogrusAdapter) Debugw(msg string, kvs ...any) { a.With(kvs...).Debug(msg) }
func (a *LogrusAdapter) Infow(msg string, kvs ...any)  { a.With(kvs...).Info(msg) }
func (a *LogrusAdapter) Warnw(msg string, kvs ...any)  { a.With(kvs...).Warn(msg) }
func (a *LogrusAdapter) Errorw(msg string, kvs ...any) { a.With(kvs...).Error(msg) }

func (a *LogrusAdapter) Ctx(ctx context.Context) Logger {
	requestID := GetRequestID(ctx)
	if requestID == "" {
		return a
	}
	return &LogrusAdapter{
		entry: a.entry.WithField("request_id", requestID),
	}
}

func (a *LogrusAdapter) With(args ...any) Logger {
	if len(args) == 0 {
		return a
	}

	fields := make(logrus.Fields)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		if i+1 < len(args) {
			fields[key] = args[i+1]
		}
	}

	return &LogrusAdapter{
		entry: a.entry.WithFields(fields),
	}
}

func (a *LogrusAdapter) WithGroup(name string) Logger {
	return a
}

func (a *LogrusAdapter) Log(level Level, msg string, attrs ...Attr) {
	fields := make(logrus.Fields)
	for _, attr := range attrs {
		fields[attr.Key] = attr.Value
	}

	a.entry.WithFields(fields).Log(toLogrusLevel(level), msg)
}

func (a *LogrusAdapter) LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	a.Ctx(ctx).Log(level, msg, attrs...)
}

func (a *LogrusAdapter) LogRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	a.Ctx(ctx).With(
		"method", method,
		"path", path,
		"status", status,
		"duration", duration,
	).Info("http request")
}

func toLogrusLevel(l Level) logrus.Level {
	switch l {
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

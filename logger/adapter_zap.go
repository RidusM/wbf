package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	_argsPairs = 2
)

type ZapLogger struct {
	logger *zap.Logger
	level  zapcore.Level
}

func NewZapLogger(appName, env string, opts ...Option) (*ZapLogger, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	}

	cfg := defaultConfigs()
	for _, opt := range opts {
		opt(cfg)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(cfg.GetWriter()),
		toZapLevel(cfg.Level),
	)

	logger := &ZapLogger{
		logger: zap.New(core,
			zap.Fields(
				zap.String("service", appName),
				zap.String("env", env),
			),
			zap.AddCaller(),
			zap.AddStacktrace(zap.ErrorLevel),
		),
	}
	return logger, nil
}

func (l *ZapLogger) Zap() *zap.Logger {
	return l.logger
}

type ZapAdapter struct {
	zapLogger *ZapLogger
}

func NewZapAdapter(appName, env string) (*ZapAdapter, error) {
	zl, err := NewZapLogger(appName, env)
	if err != nil {
		return nil, err
	}
	return &ZapAdapter{zapLogger: zl}, nil
}

func (a *ZapAdapter) Debug(msg string, args ...any) {
	a.zapLogger.Zap().Sugar().Debugw(msg, args...)
}

func (a *ZapAdapter) Info(msg string, args ...any) {
	a.zapLogger.Zap().Sugar().Infow(msg, args...)
}

func (a *ZapAdapter) Warn(msg string, args ...any) {
	a.zapLogger.Zap().Sugar().Warnw(msg, args...)
}

func (a *ZapAdapter) Error(msg string, args ...any) {
	a.zapLogger.Zap().Sugar().Errorw(msg, args...)
}

func (a *ZapAdapter) Debugw(msg string, args ...any) {
	a.zapLogger.Zap().Sugar().Debugw(msg, args...)
}

func (a *ZapAdapter) Infow(msg string, args ...any) {
	a.zapLogger.Zap().Sugar().Infow(msg, args...)
}

func (a *ZapAdapter) Warnw(msg string, args ...any) {
	a.zapLogger.Zap().Sugar().Warnw(msg, args...)
}

func (a *ZapAdapter) Errorw(msg string, args ...any) {
	a.zapLogger.Zap().Sugar().Errorw(msg, args...)
}

func (a *ZapAdapter) Ctx(ctx context.Context) Logger {
	requestID := GetRequestID(ctx)

	if requestID == "" {
		return a
	}

	return &ZapAdapter{
		zapLogger: &ZapLogger{
			logger: a.zapLogger.logger.With(zap.String("request_id", requestID)),
		},
	}
}

func (a *ZapAdapter) With(args ...any) Logger {
	newAdapter := &ZapAdapter{zapLogger: &ZapLogger{}}
	newAdapter.zapLogger.logger = a.zapLogger.Zap().With(toZapFields(args)...)
	return newAdapter
}

func (a *ZapAdapter) WithGroup(name string) Logger {
	newAdapter := &ZapAdapter{zapLogger: &ZapLogger{}}
	newAdapter.zapLogger.logger = a.zapLogger.Zap().With(zap.Namespace(name))
	return newAdapter
}

func (a *ZapAdapter) Log(level Level, msg string, attrs ...Attr) {
	zapLevel := toZapLevel(level)

	if ce := a.zapLogger.Zap().Check(zapLevel, msg); ce != nil {
		ce.Write(toZapFieldsFromAttrs(attrs)...)
	}
}

func (a *ZapAdapter) LogAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	l := a.Ctx(ctx)

	l.Log(level, msg, attrs...)
}

func (a *ZapAdapter) LogRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	a.Ctx(ctx).Info("request",
		"method", method,
		"path", path,
		"status", status,
		"duration", duration,
	)
}

func toZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func toZapFields(args []any) []zap.Field {
	if len(args)%2 != 0 {
		args = append(args, "<missing>")
	}
	fields := make([]zap.Field, 0, len(args)/_argsPairs)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = "UNKNOWN"
		}
		fields = append(fields, zap.Any(key, args[i+1]))
	}
	return fields
}

func toZapFieldsFromAttrs(attrs []Attr) []zap.Field {
	fields := make([]zap.Field, 0, len(attrs))
	for _, a := range attrs {
		fields = append(fields, zap.Any(a.Key, a.Value))
	}
	return fields
}

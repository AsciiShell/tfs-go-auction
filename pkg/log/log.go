package log

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...interface{})
	Info(msg string)
	Infof(format string, args ...interface{})
	Warn(msg string)
	Warnf(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	Fatal(msg string)
	Fatalf(format string, args ...interface{})

	With(fields ...Field) Logger
	WithError(err error) Logger
}

func New() Logger {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			// Keys can be anything except the empty string.
			TimeKey:  "T",
			LevelKey: "L",
			NameKey:  "N",
			//CallerKey:      "C",
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	zapLogger, err := cfg.Build()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Can't create logger: %s", err)
		os.Exit(1)
	}

	return &structuredLogger{zapLogger: zapLogger}
}

type Field = zap.Field

type structuredLogger struct {
	zapLogger *zap.Logger
}

func (l *structuredLogger) Debug(msg string) {
	l.zapLogger.Debug(msg)
}
func (l *structuredLogger) Debugf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Debugf(format, args...)
}

func (l *structuredLogger) Info(msg string) {
	l.zapLogger.Info(msg)
}

func (l *structuredLogger) Infof(format string, args ...interface{}) {
	l.zapLogger.Sugar().Infof(format, args...)
}

func (l *structuredLogger) Warn(msg string) {
	l.zapLogger.Warn(msg)
}

func (l *structuredLogger) Warnf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Warnf(format, args...)
}

func (l *structuredLogger) Error(msg string) {
	l.zapLogger.Error(msg)
}

func (l *structuredLogger) Errorf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Errorf(format, args...)
}

func (l *structuredLogger) Fatal(msg string) {
	l.zapLogger.Fatal(msg)
}

func (l *structuredLogger) Fatalf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Fatalf(format, args...)
}

func (l *structuredLogger) With(fields ...Field) Logger {
	l.zapLogger = l.zapLogger.With(fields...)
	return l
}

func (l *structuredLogger) WithError(err error) Logger {
	l.zapLogger = l.zapLogger.With(zap.Error(err))
	return l
}

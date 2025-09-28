package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const (
	LoggerKey    contextKey = "logger"
	RequestIDKey contextKey = "requestID"
	RequestID               = "requestID"
	ServiceName             = "service"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Fatal(ctx context.Context, msg string, fields ...zap.Field)

	With(fields ...zap.Field) Logger
	Sync() error
}

func GetLoggerFromContext(ctx context.Context) Logger {
	return ctx.Value(LoggerKey).(Logger)
}

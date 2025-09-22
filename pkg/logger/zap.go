package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type logger struct {
	serviceName string
	logger      *zap.Logger
}

func (l *logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Debug(msg, l.withContext(ctx, fields...)...)
}

func (l *logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Info(msg, l.withContext(ctx, fields...)...)
}

func (l *logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Warn(msg, l.withContext(ctx, fields...)...)
}

func (l *logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Error(msg, l.withContext(ctx, fields...)...)
}

func (l *logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, l.withContext(ctx, fields...)...)
}

func (l *logger) withContext(ctx context.Context, fields ...zap.Field) []zap.Field {
	fields = append(fields, zap.String(ServiceName, l.serviceName))
	if v := ctx.Value(RequestIDKey); v != nil {
		fields = append(fields, zap.String(RequestID, v.(string)))
	}
	return fields
}

func (l *logger) With(fields ...zap.Field) Logger {
	return &logger{
		serviceName: l.serviceName,
		logger:      l.logger.With(fields...),
	}
}

func NewZapLogger(serviceName string, env string) (Logger, error) {
	if env == "dev" || env == "prod" {
		if err := os.MkdirAll("logs", 0755); err != nil {
			return nil, fmt.Errorf("failed to create logs directory: %w", err)
		}
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	config := zap.Config{
		DisableCaller:     false,
		DisableStacktrace: false,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	var options []zap.Option

	switch env {
	case "local":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.Development = true
		config.DisableStacktrace = true
		config.Encoding = "console"

		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

		options = []zap.Option{zap.AddCaller(), zap.AddCallerSkip(1)}

		config.EncoderConfig = encoderConfig

		zapLogger, err := config.Build(options...)
		if err != nil {
			return nil, err
		}

		return &logger{serviceName: serviceName, logger: zapLogger}, nil

	case "dev":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.Development = true

		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")

		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zap.DebugLevel,
		)

		rotator := &lumberjack.Logger{
			Filename:   "./logs/dev.log",
			MaxSize:    50,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}

		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(rotator),
			zap.DebugLevel,
		)

		tee := zapcore.NewTee(consoleCore, fileCore)

		options = []zap.Option{zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.ErrorLevel)}
		baseLogger := zap.New(tee, options...)
		return &logger{serviceName: serviceName, logger: baseLogger}, nil

	case "prod":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		config.Development = false
		config.Encoding = "json"

		consoleCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zap.InfoLevel,
		)

		rotator := &lumberjack.Logger{
			Filename:   "./logs/app.log",
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     60,
			Compress:   true,
		}
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(rotator),
			zap.InfoLevel,
		)

		tee := zapcore.NewTee(consoleCore, fileCore)

		options = []zap.Option{zap.AddStacktrace(zap.ErrorLevel)}
		baseLogger := zap.New(tee, options...)
		return &logger{serviceName: serviceName, logger: baseLogger}, nil

	default:
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.Development = true
		config.DisableStacktrace = true
		config.Encoding = "console"

		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

		options = []zap.Option{zap.AddCaller(), zap.AddCallerSkip(1)}

		config.EncoderConfig = encoderConfig

		zapLogger, err := config.Build(options...)
		if err != nil {
			return nil, err
		}

		return &logger{serviceName: serviceName, logger: zapLogger}, nil
	}
}

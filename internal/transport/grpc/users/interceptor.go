package usersgrpc

import (
	"context"
	"time"

	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	requestIDKey = logger.RequestIDKey
	loggerKey    = logger.LoggerKey
)

func RecoveryInterceptor(ctx context.Context, log logger.Logger) grpc.UnaryServerInterceptor {
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			log.Error(ctx, "Recovered from panic", zap.Error(err))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	return recovery.UnaryServerInterceptor(recoveryOpts...)
}

func RequestIDInterceptor(baseLog logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		var reqID string
		if ok {
			if vals := md.Get(string(requestIDKey)); len(vals) > 0 {
				reqID = vals[0]
			}
		}
		if reqID == "" {
			reqID = uuid.NewString()
		}

		log := baseLog.With(zap.String(logger.RequestID, reqID))

		ctx = context.WithValue(ctx, loggerKey, log)
		ctx = context.WithValue(ctx, requestIDKey, reqID)

		return handler(ctx, req)
	}
}

func LoggingInterceptor(baseLog logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()

		log, ok := logger.GetLoggerFromContext(ctx)
		if !ok {
			log = baseLog
		}

		log.Info(ctx, "gRPC request: started", zap.String("method", info.FullMethod))

		resp, err = handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			log.Warn(ctx, "gRPC request: finished with error",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		} else {
			log.Info(ctx, "gRPC request: finished",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
			)
		}

		return resp, err
	}
}

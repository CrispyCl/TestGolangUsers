package interceptors

import (
	"context"
	"time"

	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

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

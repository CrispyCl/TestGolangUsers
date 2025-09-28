package interceptors

import (
	"context"

	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	requestIDKey = logger.RequestIDKey
	loggerKey    = logger.LoggerKey
)

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

		return handler(ctx, req)
	}
}

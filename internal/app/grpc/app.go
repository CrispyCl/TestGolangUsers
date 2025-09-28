package grpcapp

import (
	"context"
	"fmt"
	"net"

	"github.com/CrispyCl/TestGolangUsers/internal/domain/models"
	"github.com/CrispyCl/TestGolangUsers/internal/transport/grpc/interceptors"
	usersgrpc "github.com/CrispyCl/TestGolangUsers/internal/transport/grpc/users"
	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type UserService interface {
	Create(ctx context.Context, email, password string, role models.UserRole) (int64, error)
	CheckPassword(ctx context.Context, email, password string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	UpdateLastSeen(ctx context.Context, id int64) error
}

type App struct {
	log        logger.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(ctx context.Context, log logger.Logger, usersService UserService, port int) *App {
	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptors.RequestIDInterceptor(log),
		interceptors.LoggingInterceptor(log),
		interceptors.RecoveryInterceptor(ctx, log),
	))

	usersgrpc.Register(gRPCServer, usersService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) Start(ctx context.Context) error {
	const op = "grpcapp.Start"

	log := a.log.With(zap.String("op", op))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info(ctx, "gRPC: server started", zap.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) {
	const op = "grpcapp.Stop"
	log := a.log.With(zap.String("op", op))

	done := make(chan struct{})
	go func() {
		a.gRPCServer.GracefulStop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		log.Warn(ctx, "Graceful stop timed out, forcing stop")
		a.gRPCServer.Stop()
	case <-done:
		log.Info(ctx, "gRPC: server stopped gracefully")
	}
}

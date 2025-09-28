package app

import (
	"context"
	"fmt"
	"time"

	grpcapp "github.com/CrispyCl/TestGolangUsers/internal/app/grpc"
	"github.com/CrispyCl/TestGolangUsers/internal/config"
	"github.com/CrispyCl/TestGolangUsers/internal/repository/pg"
	"github.com/CrispyCl/TestGolangUsers/internal/service/users"
	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
	"github.com/CrispyCl/TestGolangUsers/pkg/storage/postgres"
	"golang.org/x/sync/errgroup"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	const op = "app.New"

	log, ok := logger.GetLoggerFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("%s: could not get logger from context", op)
	}

	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("%s: postgres: connection failed: %w", op, err)
	}

	userRepo := pg.NewUserRepository(db)
	userServ := users.NewUserService(log, userRepo)

	grpcApp := grpcapp.New(ctx, log, userServ, cfg.GRPCServerPort)

	return &App{GRPCServer: grpcApp}, nil
}

func (a *App) Start(ctx context.Context) error {
	const op = "app.Start"

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return a.GRPCServer.Start(ctx)
	})

	return eg.Wait()
}

func (a *App) Stop(ctx context.Context) error {
	const op = "app.Stop"
	const gracefulStopTimeout = 5 * time.Second

	eg := errgroup.Group{}

	eg.Go(func() error {
		ctxStop, cancel := context.WithTimeout(ctx, gracefulStopTimeout)
		defer cancel()

		a.GRPCServer.Stop(ctxStop)
		return nil
	})

	return eg.Wait()
}

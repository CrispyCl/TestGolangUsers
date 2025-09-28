package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/CrispyCl/TestGolangUsers/internal/app"
	"github.com/CrispyCl/TestGolangUsers/internal/config"
	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
	"go.uber.org/zap"
)

const (
	serviceName = "users"
	loggerKey   = logger.LoggerKey
)

func main() {
	cfg := config.MustLoad()

	mainLogger, err := logger.NewZapLogger(serviceName, cfg.Env)
	if err != nil {
		panic(fmt.Errorf("failed to load logger: %w", err))
	}
	defer mainLogger.Sync()

	ctx := context.WithValue(context.Background(), loggerKey, mainLogger)

	app, err := app.New(ctx, cfg)
	if err != nil {
		mainLogger.Fatal(ctx, err.Error())
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGTERM, syscall.SIGINT)

	errCh := make(chan error, 1)

	go func() {
		mainLogger.Info(ctx, "server.start")
		errCh <- app.Start(ctx)
	}()

	select {
	case sig := <-graceCh:
		mainLogger.Info(ctx, "received signal, stopping server", zap.String("signal", sig.String()))
	case err := <-errCh:
		mainLogger.Error(ctx, "server stopped unexpectedly", zap.Error(err))
	}

	if err := app.Stop(ctx); err != nil {
		mainLogger.Error(ctx, "error while stopping the server", zap.Error(err))
	}
	mainLogger.Info(ctx, "server.stop")
}

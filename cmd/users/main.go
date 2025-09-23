package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/CrispyCl/TestGolangUsers/internal/config"
	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
)

const (
	serviceName = "users"
	loggerKey   = logger.LoggerKey
)

func main() {
	cfg := config.MustLoad()

	log, err := logger.NewZapLogger(serviceName, cfg.Env)
	if err != nil {
		panic(fmt.Errorf("failed to load logger: %+v", err))
	}

	ctx := context.WithValue(context.Background(), loggerKey, log)

	log.Info(ctx, "Run the application")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	log.Info(ctx, "Gracefully stopped")
}

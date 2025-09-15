package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/CrispyCl/TestGolangUsers/internal/config"
	"github.com/CrispyCl/TestGolangUsers/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("Run the application")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	log.Info("Gracefully stopped")
}

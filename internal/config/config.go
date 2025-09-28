package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/CrispyCl/TestGolangUsers/pkg/storage/postgres"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres postgres.Config

	Env            string `env:"ENV" env-default:"local"`
	GRPCServerPort int    `env:"GRPC_SERVER_PORT"`
}

func MustLoad() Config {
	configPath := fetchConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Errorf("failed to read config: %+v", err))
	}

	return cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	if res == "" {
		res = "config/.env"
	}

	return res
}

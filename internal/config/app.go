package config

import (
	"context"
	"flag"
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/mrhyman/shortner/internal/model"
)

const (
	DefaultServerAddress = "localhost:8080"
	DefaultBaseURL       = "http://localhost:8080"
)

type AppConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func Load() AppConfig {
	var cfg AppConfig

	serverFlag := flag.String("a", DefaultServerAddress, "HTTP server address, e.g. localhost:8888")
	baseFlag := flag.String("b", DefaultBaseURL, "Base URL for short links, e.g. http://localhost:8080")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		slog.ErrorContext(context.Background(), model.ErrEnvParsing.Error(), slog.String("err", err.Error()))
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = *serverFlag
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = *baseFlag
	}

	return cfg
}
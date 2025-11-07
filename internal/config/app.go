package config

import (
	"context"
	"flag"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/mrhyman/shortner/internal/logger"
)

const (
	DefaultServerAddress   = "localhost:8080"
	DefaultBaseURL         = "http://localhost:8080"
	DefaultFileStoragePath = "/.storage/db.json"
)

type AppConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	StoragePath   string `env:"FILE_STORAGE_PATH"`
}

func Load(ctx context.Context) AppConfig {
	var cfg AppConfig

	log := logger.FromContext(ctx)

	serverFlag := flag.String("a", DefaultServerAddress, "HTTP server address, e.g. localhost:8888")
	baseFlag := flag.String("b", DefaultBaseURL, "Base URL for short links, e.g. http://localhost:8080")
	fileFlag := flag.String("f", DefaultFileStoragePath, "Base storage path, e.g. /.storage/db.json")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.With("err", err.Error()).Fatal()
	}

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = *serverFlag
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = *baseFlag
	}

	if cfg.StoragePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.With("err", err.Error()).Fatal()
		}

		cfg.StoragePath = filepath.Join(cwd, *fileFlag)
	}

	return cfg
}

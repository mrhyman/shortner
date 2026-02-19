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
	DefaultFileStoragePath = ""
	DefaultDBDSN           = ""
	DefaultHashKey         = "qwerty12345"
)

type StorageMode int

const (
	StorageDB StorageMode = iota
	StorageFile
	StorageMemory
)

type AppConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	StoragePath   string `env:"FILE_STORAGE_PATH"`
	DBDSN         string `env:"DATABASE_DSN"`
	HashKey       string `env:"HASH_KEY"`
	AuditFile     string `env:"AUDIT_FILE"`
	AuditURL      string `env:"AUDIT_URL"`
	EnableHTTPS   bool   `env:"ENABLE_HTTPS"`
	CertFile      string
	KeyFile       string
	StorageMode   StorageMode
}

func Load(ctx context.Context) AppConfig {
	var cfg AppConfig

	log := logger.Get()

	serverFlag := flag.String("a", DefaultServerAddress, "HTTP server address, e.g. localhost:8888")
	baseFlag := flag.String("b", DefaultBaseURL, "Base URL for short links, e.g. http://localhost:8080")
	fileFlag := flag.String("f", DefaultFileStoragePath, "Base storage path, e.g. /.storage/db.json")
	dbFlag := flag.String("d", DefaultDBDSN, "Database connection string. postgres://postgres:postgres@localhost:5432/postgres")
	hashKey := flag.String("hk", DefaultHashKey, "Auth hash key. e.g. qwerty12345")
	auditFileFlag := flag.String("audit-file", "", "Audit file path. e.g. /var/log/audit.log")
	auditURLFlag := flag.String("audit-url", "", "External audit URL. e.g. http://somehost:8080/audit")
	enableHTTPSFlag := flag.Bool("s", false, "Enable HTTPS server")
	certFileFlag := flag.String("cert-file", "certs/server.crt", "Path to certificate file")
	keyFileFlag := flag.String("key-file", "certs/server.key", "Path to private key file")
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

	if cfg.DBDSN == "" {
		cfg.DBDSN = *dbFlag
	}

	if cfg.HashKey == "" {
		cfg.HashKey = *hashKey
	}

	if cfg.AuditFile == "" {
		cfg.AuditFile = *auditFileFlag
	}

	if cfg.AuditURL == "" {
		cfg.AuditURL = *auditURLFlag
	}

	if !cfg.EnableHTTPS {
		cfg.EnableHTTPS = *enableHTTPSFlag
	}

	if cfg.EnableHTTPS {
		cfg.CertFile = *certFileFlag
		cfg.KeyFile = *keyFileFlag
	}

	return setStorageMode(&cfg)
}

func setStorageMode(cfg *AppConfig) AppConfig {
	_, envDBSet := os.LookupEnv("DATABASE_DSN")
	_, envFileSet := os.LookupEnv("FILE_STORAGE_PATH")

	dbFlag := flag.Lookup("d")
	fileFlag := flag.Lookup("f")

	dbFlagSet := dbFlag != nil && dbFlag.Value.String() != "" && dbFlag.Value.String() != DefaultDBDSN
	fileFlagSet := fileFlag != nil && fileFlag.Value.String() != "" && fileFlag.Value.String() != DefaultFileStoragePath

	switch {
	case cfg.DBDSN != "" && (envDBSet || dbFlagSet):
		cfg.StorageMode = StorageDB

	case envFileSet || fileFlagSet:
		cfg.StorageMode = StorageFile

	default:
		cfg.StorageMode = StorageMemory
	}
	return *cfg
}

package config

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/mrhyman/shortner/internal/logger"
)

const (
	DefaultServerAddress   = "localhost:8080"
	DefaultBaseURL         = "http://localhost:8080"
	DefaultFileStoragePath = ""
	DefaultDBDSN           = ""
	DefaultHashKey         = "qwerty12345"
	ShutdownTimeout        = 10 * time.Second
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
type JSONConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

func Load(ctx context.Context) AppConfig {
	var cfg AppConfig

	log := logger.Get()

	configFlag := flag.String("c", "", "Path to config file")
	configAltFlag := flag.String("config", "", "Path to config file (alternative)")
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

	serverFlagValue := *serverFlag
	baseFlagValue := *baseFlag
	fileFlagValue := *fileFlag
	dbFlagValue := *dbFlag
	hashKeyFlagValue := *hashKey
	auditFileFlagValue := *auditFileFlag
	auditURLFlagValue := *auditURLFlag
	enableHTTPSFlagValue := *enableHTTPSFlag

	cfg.ServerAddress = serverFlagValue
	cfg.BaseURL = baseFlagValue

	if fileFlagValue != "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.With("err", err.Error()).Fatal()
		}
		cfg.StoragePath = filepath.Join(cwd, fileFlagValue)
	}

	cfg.DBDSN = dbFlagValue
	cfg.HashKey = hashKeyFlagValue
	cfg.AuditFile = auditFileFlagValue
	cfg.AuditURL = auditURLFlagValue
	cfg.EnableHTTPS = enableHTTPSFlagValue

	if cfg.EnableHTTPS {
		cfg.CertFile = *certFileFlag
		cfg.KeyFile = *keyFileFlag
	}

	configPath := *configFlag
	if configPath == "" {
		configPath = *configAltFlag
	}
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	var jsonCfg JSONConfig
	jsonConfigLoaded := false
	if configPath != "" {
		if file, err := os.Open(configPath); err == nil {
			defer file.Close()
			if err := json.NewDecoder(file).Decode(&jsonCfg); err == nil {
				jsonConfigLoaded = true
			} else {
				log.With("err", err.Error()).Warn("Failed to parse config file")
			}
		} else {
			log.With("err", err.Error()).Warn("Failed to open config file")
		}
	}

	if jsonConfigLoaded {
		if serverFlagValue == DefaultServerAddress && jsonCfg.ServerAddress != "" {
			cfg.ServerAddress = jsonCfg.ServerAddress
		}
		if baseFlagValue == DefaultBaseURL && jsonCfg.BaseURL != "" {
			cfg.BaseURL = jsonCfg.BaseURL
		}
		if fileFlagValue == DefaultFileStoragePath && jsonCfg.FileStoragePath != "" {
			cfg.StoragePath = jsonCfg.FileStoragePath
		}
		if dbFlagValue == DefaultDBDSN && jsonCfg.DatabaseDSN != "" {
			cfg.DBDSN = jsonCfg.DatabaseDSN
		}
		if !enableHTTPSFlagValue && jsonCfg.EnableHTTPS {
			cfg.EnableHTTPS = jsonCfg.EnableHTTPS
		}
	}

	if err := env.Parse(&cfg); err != nil {
		log.With("err", err.Error()).Fatal()
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

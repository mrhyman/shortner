package config

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mrhyman/shortner/internal/logger"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	DefaultServerAddress   = "localhost:8080"
	DefaultBaseURL         = "http://localhost:8080"
	DefaultFileStoragePath = ""
	DefaultDBDSN           = ""
	DefaultTrustedSubnet   = "127.0.0.0/8"
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
	ServerAddress string
	BaseURL       string
	StoragePath   string
	DBDSN         string
	HashKey       string
	AuditFile     string
	AuditURL      string
	EnableHTTPS   bool
	CertFile      string
	KeyFile       string
	StorageMode   StorageMode
	TrustedSubnet string
}

var (
	once sync.Once
	mu   sync.Mutex
)

// сброс флагов для тестов
func ResetForTest() {
	mu.Lock()
	defer mu.Unlock()
	once = sync.Once{}
}

func createFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	fs.StringP("config", "c", "", "Path to config file")
	fs.StringP("server-address", "a", DefaultServerAddress, "HTTP server address, e.g. localhost:8888")
	fs.StringP("base-url", "b", DefaultBaseURL, "Base URL for short links, e.g. http://localhost:8080")
	fs.StringP("file-storage-path", "f", DefaultFileStoragePath, "Base storage path, e.g. /.storage/db.json")
	fs.StringP("database-dsn", "d", DefaultDBDSN, "Database connection string. postgres://postgres:postgres@localhost:5432/postgres")
	fs.StringP("trusted-subnet", "t", DefaultTrustedSubnet, "Trusted subnet (CIDR)")
	fs.String("hash-key", DefaultHashKey, "Auth hash key. e.g. qwerty12345")
	fs.String("audit-file", "", "Audit file path. e.g. /var/log/audit.log")
	fs.String("audit-url", "", "External audit URL. e.g. http://somehost:8080/audit")
	fs.BoolP("enable-https", "s", false, "Enable HTTPS server")
	fs.String("cert-file", "certs/server.crt", "Path to certificate file")
	fs.String("key-file", "certs/server.key", "Path to private key file")

	return fs
}

func getFlagDefault(flagName string) string {
	switch flagName {
	case "server-address":
		return DefaultServerAddress
	case "base-url":
		return DefaultBaseURL
	case "file-storage-path":
		return DefaultFileStoragePath
	case "database-dsn":
		return DefaultDBDSN
	case "hash-key":
		return DefaultHashKey
	case "enable-https":
		return "false"
	case "cert-file":
		return "certs/server.crt"
	case "key-file":
		return "certs/server.key"
	default:
		return ""
	}
}

func Load(ctx context.Context) AppConfig {
	log := logger.Get()

	mu.Lock()
	fs := createFlagSet()

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.With("err", err.Error()).Warn("Failed to parse flags")
	}

	once.Do(func() {
		pflag.CommandLine = fs
	})
	mu.Unlock()

	v := viper.New()

	// дефолтные значения (самый низкий приоритет)
	v.SetDefault("server-address", DefaultServerAddress)
	v.SetDefault("base-url", DefaultBaseURL)
	v.SetDefault("file-storage-path", DefaultFileStoragePath)
	v.SetDefault("database-dsn", DefaultDBDSN)
	v.SetDefault("hash-key", DefaultHashKey)
	v.SetDefault("enable-https", false)
	v.SetDefault("cert-file", "certs/server.crt")
	v.SetDefault("key-file", "certs/server.key")
	v.SetDefault("trusted-subnet", DefaultTrustedSubnet)

	// конфигурационный файл (приоритет выше дефолтов)
	configPath := ""
	if flag := fs.Lookup("config"); flag != nil {
		configPath = flag.Value.String()
	}
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			log.With("err", err.Error()).Warn("Failed to read config file")
		} else {
			log.With("file", configPath).Info("Config file loaded")

			if v.IsSet("server_address") {
				v.Set("server-address", v.GetString("server_address"))
			}
			if v.IsSet("base_url") {
				v.Set("base-url", v.GetString("base_url"))
			}
			if v.IsSet("file_storage_path") {
				v.Set("file-storage-path", v.GetString("file_storage_path"))
			}
			if v.IsSet("database_dsn") {
				v.Set("database-dsn", v.GetString("database_dsn"))
			}
			if v.IsSet("enable_https") {
				v.Set("enable-https", v.GetBool("enable_https"))
			}
			if v.IsSet("trusted-subnet") {
				v.Set("trusted-subnet", v.GetString("trusted_subnet"))
			}
		}
	}

	// флаги (приоритет выше конфига)
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Name == "config" {
			return
		}

		currentValue := f.Value.String()
		defaultValue := getFlagDefault(f.Name)

		if currentValue != defaultValue {
			if f.Value.Type() == "bool" {
				v.Set(f.Name, currentValue == "true")
			} else {
				v.Set(f.Name, currentValue)
			}
		}
	})

	// переменные окружения (наивысший приоритет)
	envMappings := map[string]string{
		"SERVER_ADDRESS":    "server-address",
		"BASE_URL":          "base-url",
		"FILE_STORAGE_PATH": "file-storage-path",
		"DATABASE_DSN":      "database-dsn",
		"HASH_KEY":          "hash-key",
		"AUDIT_FILE":        "audit-file",
		"AUDIT_URL":         "audit-url",
		"ENABLE_HTTPS":      "enable-https",
	}

	for envKey, viperKey := range envMappings {
		if envValue := os.Getenv(envKey); envValue != "" {
			if viperKey == "enable-https" {
				v.Set(viperKey, envValue == "true" || envValue == "1")
			} else {
				v.Set(viperKey, envValue)
			}
		}
	}

	cfg := AppConfig{
		ServerAddress: v.GetString("server-address"),
		BaseURL:       v.GetString("base-url"),
		StoragePath:   v.GetString("file-storage-path"),
		DBDSN:         v.GetString("database-dsn"),
		HashKey:       v.GetString("hash-key"),
		AuditFile:     v.GetString("audit-file"),
		AuditURL:      v.GetString("audit-url"),
		EnableHTTPS:   v.GetBool("enable-https"),
		CertFile:      v.GetString("cert-file"),
		KeyFile:       v.GetString("key-file"),
		TrustedSubnet: v.GetString("trusted-subnet"),
	}

	if cfg.StoragePath != "" && !filepath.IsAbs(cfg.StoragePath) {
		cwd, err := os.Getwd()
		if err != nil {
			log.With("err", err.Error()).Fatal("Failed to get working directory")
		}
		cfg.StoragePath = filepath.Join(cwd, cfg.StoragePath)
	}

	return setStorageMode(&cfg)
}

func setStorageMode(cfg *AppConfig) AppConfig {
	switch {
	case cfg.DBDSN != "":
		cfg.StorageMode = StorageDB
	case cfg.StoragePath != "":
		cfg.StorageMode = StorageFile
	default:
		cfg.StorageMode = StorageMemory
	}
	return *cfg
}

package config_test

import (
	"flag"
	"os"
	"testing"

	"github.com/mrhyman/shortner/internal/config"
)

func resetEnvAndFlags() {
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("CONFIG")
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestSetup_ConfigPriorities(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		args        []string
		wantAddress string
		wantBaseURL string
	}{
		{
			name:        "defaults when no env or flags",
			envVars:     nil,
			args:        []string{"cmd"},
			wantAddress: config.DefaultServerAddress,
			wantBaseURL: config.DefaultBaseURL,
		},
		{
			name:        "from flags only",
			args:        []string{"cmd", "-a", "127.0.0.1:9999", "-b", "http://127.0.0.1:9999"},
			wantAddress: "127.0.0.1:9999",
			wantBaseURL: "http://127.0.0.1:9999",
		},
		{
			name: "from env only",
			envVars: map[string]string{
				"SERVER_ADDRESS": "env:7777",
				"BASE_URL":       "http://env:9090",
			},
			args:        []string{"cmd"},
			wantAddress: "env:7777",
			wantBaseURL: "http://env:9090",
		},
		{
			name: "env overrides flags",
			envVars: map[string]string{
				"SERVER_ADDRESS": "env:7777",
			},
			args:        []string{"cmd", "-a", "flag:8888"},
			wantAddress: "env:7777",
			wantBaseURL: config.DefaultBaseURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			resetEnvAndFlags()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// act
			os.Args = tt.args
			cfg := config.Load(t.Context())

			// assert
			if cfg.ServerAddress != tt.wantAddress {
				t.Errorf("ServerAddress: got %s, want %s", cfg.ServerAddress, tt.wantAddress)
			}
			if cfg.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL: got %s, want %s", cfg.BaseURL, tt.wantBaseURL)
			}
		})
	}
}

func TestSetup_JSONConfig(t *testing.T) {
	configContent := `{
		"server_address": "json:8080",
		"base_url": "http://json:8080",
		"file_storage_path": "/tmp/test.db",
		"database_dsn": "test_db_dsn",
		"enable_https": true
	}`

	tmpFile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	tests := []struct {
		name            string
		envVars         map[string]string
		args            []string
		configFile      string
		wantAddress     string
		wantBaseURL     string
		wantStorage     string
		wantDBDSN       string
		wantEnableHTTPS bool
	}{
		{
			name:            "from json config only",
			args:            []string{"cmd"},
			configFile:      tmpFile.Name(),
			wantAddress:     "json:8080",
			wantBaseURL:     "http://json:8080",
			wantStorage:     "/tmp/test.db",
			wantDBDSN:       "test_db_dsn",
			wantEnableHTTPS: true,
		},
		{
			name:            "flags override json config",
			args:            []string{"cmd", "-a", "flag:9999", "-b", "http://flag:9999"},
			configFile:      tmpFile.Name(),
			wantAddress:     "flag:9999",
			wantBaseURL:     "http://flag:9999",
			wantStorage:     "/tmp/test.db",
			wantDBDSN:       "test_db_dsn",
			wantEnableHTTPS: true,
		},
		{
			name: "env overrides json config",
			envVars: map[string]string{
				"SERVER_ADDRESS": "env:7777",
			},
			args:            []string{"cmd"},
			configFile:      tmpFile.Name(),
			wantAddress:     "env:7777",
			wantBaseURL:     "http://json:8080",
			wantStorage:     "/tmp/test.db",
			wantDBDSN:       "test_db_dsn",
			wantEnableHTTPS: true,
		},
		{
			name: "env and flags override json config",
			envVars: map[string]string{
				"SERVER_ADDRESS": "env:7777",
			},
			args:            []string{"cmd", "-b", "http://flag:9999"},
			configFile:      tmpFile.Name(),
			wantAddress:     "env:7777",
			wantBaseURL:     "http://flag:9999",
			wantStorage:     "/tmp/test.db",
			wantDBDSN:       "test_db_dsn",
			wantEnableHTTPS: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			resetEnvAndFlags()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			if tt.configFile != "" {
				os.Setenv("CONFIG", tt.configFile)
			}

			// act
			os.Args = tt.args
			cfg := config.Load(t.Context())

			// assert
			if cfg.ServerAddress != tt.wantAddress {
				t.Errorf("ServerAddress: got %s, want %s", cfg.ServerAddress, tt.wantAddress)
			}
			if cfg.BaseURL != tt.wantBaseURL {
				t.Errorf("BaseURL: got %s, want %s", cfg.BaseURL, tt.wantBaseURL)
			}
			if cfg.StoragePath != tt.wantStorage {
				t.Errorf("StoragePath: got %s, want %s", cfg.StoragePath, tt.wantStorage)
			}
			if cfg.DBDSN != tt.wantDBDSN {
				t.Errorf("DBDSN: got %s, want %s", cfg.DBDSN, tt.wantDBDSN)
			}
			if cfg.EnableHTTPS != tt.wantEnableHTTPS {
				t.Errorf("EnableHTTPS: got %v, want %v", cfg.EnableHTTPS, tt.wantEnableHTTPS)
			}
		})
	}
}

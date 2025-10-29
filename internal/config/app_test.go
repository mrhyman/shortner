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
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestSetup_ConfigPriorities(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		args           []string
		wantAddress    string
		wantBaseURL    string
	}{
		{
			name:        "defaults when no env or flags",
			envVars:     nil,
			args:        []string{"cmd"},
			wantAddress: config.DefaultServerAddress,
			wantBaseURL: config.DefaultBaseURL,
		},
		{
			name: "from flags only",
			args: []string{"cmd", "-a", "127.0.0.1:9999", "-b", "http://127.0.0.1:9999"},
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

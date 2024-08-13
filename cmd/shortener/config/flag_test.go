package config

import (
	"os"
	"testing"
)

// TestParseFlags проверяет, правильно ли парсятся флаги
func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Args
	}{
		{
			name: "Parse flags correctly",
			args: []string{
				"-a=:8081",
				"-b=http://localhost:8081",
				"-f=/tmp/storage.txt",
				"-d=postgres://user:pass@localhost/dbname",
			},
			expected: Args{
				ServerAddress:   ":8081",
				BaseURL:         "http://localhost:8081",
				FileStoragePath: "/tmp/storage.txt",
				DatabaseDNS:     "postgres://user:pass@localhost/dbname",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Port: "8080"}
			args := NewArgs(config)

			args.ParseFlags(tt.args)

			if args.ServerAddress != tt.expected.ServerAddress {
				t.Errorf("expected ServerAddress to be '%s', got '%s'", tt.expected.ServerAddress, args.ServerAddress)
			}
			if args.BaseURL != tt.expected.BaseURL {
				t.Errorf("expected BaseURL to be '%s', got '%s'", tt.expected.BaseURL, args.BaseURL)
			}
			if args.FileStoragePath != tt.expected.FileStoragePath {
				t.Errorf("expected FileStoragePath to be '%s', got '%s'", tt.expected.FileStoragePath, args.FileStoragePath)
			}
			if args.DatabaseDNS != tt.expected.DatabaseDNS {
				t.Errorf("expected DatabaseDNS to be '%s', got '%s'", tt.expected.DatabaseDNS, args.DatabaseDNS)
			}
		})
	}
}

// TestOverrideWithEnvVars проверяет, переопределяют ли переменные среды флаги
func TestOverrideWithEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Args
	}{
		{
			name: "Override flags with environment variables",
			envVars: map[string]string{
				"SHORTENER_SERVER_ADDRESS": ":9090",
				"SHORTENER_BASE_URL":       "http://localhost:9090",
				"FILE_STORAGE_PATH":        "/data/storage.txt",
				"DATABASE_DSN":             "postgres://user:pass@localhost/newdb",
			},
			expected: Args{
				ServerAddress:   ":9090",
				BaseURL:         "http://localhost:9090",
				FileStoragePath: "/data/storage.txt",
				DatabaseDNS:     "postgres://user:pass@localhost/newdb",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Port: "8080"}
			args := NewArgs(config)

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer os.Clearenv()

			args.ParseFlags([]string{})

			if args.ServerAddress != tt.expected.ServerAddress {
				t.Errorf("expected ServerAddress to be '%s', got '%s'", tt.expected.ServerAddress, args.ServerAddress)
			}
			if args.BaseURL != tt.expected.BaseURL {
				t.Errorf("expected BaseURL to be '%s', got '%s'", tt.expected.BaseURL, args.BaseURL)
			}
			if args.FileStoragePath != tt.expected.FileStoragePath {
				t.Errorf("expected FileStoragePath to be '%s', got '%s'", tt.expected.FileStoragePath, args.FileStoragePath)
			}
			if args.DatabaseDNS != tt.expected.DatabaseDNS {
				t.Errorf("expected DatabaseDNS to be '%s', got '%s'", tt.expected.DatabaseDNS, args.DatabaseDNS)
			}
		})
	}
}

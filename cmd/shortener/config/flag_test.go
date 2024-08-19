package config

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestParseFlags проверяет, правильно ли парсятся флаги
func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Args
	}{
		{
			name: "parse_flags_correctly",
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
			args := NewArgs()
			args.ParseFlags(tt.args)

			t.Logf("Expected: %+v", tt.expected)
			t.Logf("Got: %+v", *args)

			if diff := cmp.Diff(tt.expected, *args); diff != "" {
				t.Fatalf("unexpected config parameters (-want +got):\n%s", diff)
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
			name: "override_flags_with_environment_variables",
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
			args := NewArgs()

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer os.Clearenv()

			args.ParseFlags([]string{})

			if diff := cmp.Diff(tt.expected, *args); diff != "" {
				t.Fatalf("unexpected config parameters (-want +got):\n%s", diff)
			}
		})
	}
}

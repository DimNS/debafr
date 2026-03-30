package application

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		setEnvFunc func()
		want       *Configuration
		wantErr    error
	}{
		{
			name: "Should load configuration from env with default values",
			setEnvFunc: func() {
				// put env vars
			},
			want: &Configuration{
				DevMode: func() bool {
					fromEnv := os.Getenv("DEBAFR_DEV_MODE")
					if fromEnv == "" {
						return false
					}

					return fromEnv == "true"
				}(),
				Toml: TomlConfig{
					App: AppConfig{
						Name: "capuchin",
					},
					Files: FilesConfig{
						ComposeBlue:  "compose.blue.yaml",
						ComposeGreen: "compose.green.yaml",
						NginxConf:    "nginx.conf",
					},
					BinPaths: BinPathsConfig{
						Docker: "/usr/bin/docker",
						Curl:   "/usr/bin/curl",
						Nginx:  "/usr/sbin/nginx",
					},
					Timeouts: TimeoutsConfig{
						Default: 30 * time.Second,
					},
					Healthcheck: HealthcheckConfig{
						MaxRetries: 10,
						RetryDelay: 3 * time.Second,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should load configuration from env",
			setEnvFunc: func() {
				t.Setenv("DEBAFR_DEV_MODE", "true")
			},
			want: &Configuration{
				DevMode: true,
				Toml: TomlConfig{
					App: AppConfig{
						Name: "capuchin",
					},
					Files: FilesConfig{
						ComposeBlue:  "compose.blue.yaml",
						ComposeGreen: "compose.green.yaml",
						NginxConf:    "nginx.conf",
					},
					BinPaths: BinPathsConfig{
						Docker: "/usr/bin/docker",
						Curl:   "/usr/bin/curl",
						Nginx:  "/usr/sbin/nginx",
					},
					Timeouts: TimeoutsConfig{
						Default: 30 * time.Second,
					},
					Healthcheck: HealthcheckConfig{
						MaxRetries: 10,
						RetryDelay: 3 * time.Second,
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setEnvFunc()
			got, err := LoadConfiguration()
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

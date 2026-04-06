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
		filename   string
		want       *Configuration
		wantErr    error
	}{
		{
			name: "Should load configuration from config file with default values",
			setEnvFunc: func() {
				// put env vars
			},
			filename: "testdata/config_mini.toml",
			want: &Configuration{
				DevMode: func() bool {
					fromEnv := os.Getenv("DEBAFR_DEV_MODE")
					if fromEnv == "" {
						return false
					}

					return fromEnv == "true"
				}(),
				Toml: TomlConfig{
					App: AppConfig{ //nolint:gosec // ignore
						ProjectName:     "myapp",
						ProxyPassPrefix: "proxy_pass http://127.0.0.1:",
						LocationPorts: []LocationPort{
							{
								Location:  "/api",
								BluePort:  "3001",
								GreenPort: "3011",
							},
							{
								Location:  "/ws",
								BluePort:  "3003",
								GreenPort: "3013",
							},
							{
								Location:  "/",
								BluePort:  "3002",
								GreenPort: "3012",
							},
						},
					},
					DockerLogin: DockerLoginConfig{},
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
						Default: Duration(30 * time.Second),
					},
					Healthcheck: HealthcheckConfig{
						MaxRetries: 10,
						RetryDelay: Duration(3 * time.Second),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Should load configuration from config file",
			setEnvFunc: func() {
				t.Setenv("DEBAFR_DEV_MODE", "true")
			},
			filename: "testdata/config_full.toml",
			want: &Configuration{
				DevMode: true,
				Toml: TomlConfig{
					App: AppConfig{ //nolint:gosec // ignore
						ProjectName:     "myapp",
						ProxyPassPrefix: "proxy_pass http://127.0.0.1:",
						LocationPorts: []LocationPort{
							{
								Location:  "/api",
								BluePort:  "3001",
								GreenPort: "3011",
							},
							{
								Location:  "/ws",
								BluePort:  "3003",
								GreenPort: "3013",
							},
							{
								Location:  "/",
								BluePort:  "3002",
								GreenPort: "3012",
							},
						},
					},
					DockerLogin: DockerLoginConfig{
						Enabled:  true,
						Registry: "ghcr.io",
						Username: "username2",
						Password: "password2",
					},
					Files: FilesConfig{
						ComposeBlue:  "myapp_compose.blue.yaml",
						ComposeGreen: "myapp_compose.green.yaml",
						NginxConf:    "myapp_nginx.conf",
					},
					BinPaths: BinPathsConfig{
						Docker: "docker",
						Curl:   "curl",
						Nginx:  "nginx",
					},
					Timeouts: TimeoutsConfig{
						Default: Duration(60 * time.Second),
					},
					Healthcheck: HealthcheckConfig{
						MaxRetries: 5,
						RetryDelay: Duration(5 * time.Second),
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setEnvFunc()
			got, err := LoadConfiguration(tt.filename)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

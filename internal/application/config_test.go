package application

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
					App: AppConfig{
						ProjectName: "myapp",
						Images: []string{
							"ghcr.io/myapp/api",
							"ghcr.io/myapp/ui",
						},
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
						Default: Duration(5 * time.Minute),
					},
					Healthcheck: HealthcheckConfig{
						MaxRetries: 30,
						RetryDelay: Duration(5 * time.Second),
					},
					Nginx: NginxConfig{
						Enabled: boolPtr(true),
						Test:    []string{"/usr/sbin/nginx", "-t"},
						Reload:  []string{"/usr/sbin/nginx", "-s", "reload"},
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
					App: AppConfig{
						ProjectName: "myapp",
						Images: []string{
							"ghcr.io/myapp/api",
							"ghcr.io/myapp/ui",
						},
						Replacer: []Replacer{
							{
								Path:  "myapp.json",
								Blue:  "myapp_blue_api:8080",
								Green: "myapp_green_api:8080",
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
						Default: Duration(120 * time.Second),
					},
					Healthcheck: HealthcheckConfig{
						MaxRetries: 5,
						RetryDelay: Duration(10 * time.Second),
					},
					Nginx: NginxConfig{
						Enabled: boolPtr(true),
						Test:    []string{"nginx", "-t"},
						Reload:  []string{"nginx", "-s", "reload"},
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

func TestLoadConfigurationWithoutNginx(t *testing.T) {
	t.Parallel()

	got, err := LoadConfiguration("testdata/config_no_nginx.toml")
	require.NoError(t, err)
	require.NotNil(t, got.Toml.Nginx.Enabled)
	assert.False(t, *got.Toml.Nginx.Enabled)
	assert.Empty(t, got.Toml.Nginx.Test)
	assert.Empty(t, got.Toml.Nginx.Reload)
}

func boolPtr(value bool) *bool {
	return &value
}

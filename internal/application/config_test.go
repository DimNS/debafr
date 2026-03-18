package application

import (
	"os"
	"testing"

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
				t.Setenv("DEBAFR_PROJECT_NAME", "capuchin")
			},
			want: &Configuration{
				DevMode: func() bool {
					// Это надо из-за .envrc и запуска тестов через make test

					fromEnv := os.Getenv("DEBAFR_DEV_MODE")
					if fromEnv == "" {
						return false
					}

					return fromEnv == "true"
				}(),
				ProjectName: "capuchin",
				Filename: Filename{
					ComposeBlue:  "compose.blue.yaml",
					ComposeGreen: "compose.green.yaml",
					NginxConf:    "nginx.conf",
				},
			},
			wantErr: nil,
		},
		{
			name: "Should load configuration from env",
			setEnvFunc: func() {
				t.Setenv("DEBAFR_DEV_MODE", "true")

				t.Setenv("DEBAFR_PROJECT_NAME", "capuchin")

				t.Setenv("DEBAFR_FILENAME_COMPOSE_BLUE", "compose-blue")
				t.Setenv("DEBAFR_FILENAME_COMPOSE_GREEN", "compose-green")
				t.Setenv("DEBAFR_FILENAME_NGINX_CONF", "nginx-conf")
			},
			want: &Configuration{
				DevMode:     true,
				ProjectName: "capuchin",
				Filename: Filename{
					ComposeBlue:  "compose-blue",
					ComposeGreen: "compose-green",
					NginxConf:    "nginx-conf",
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

package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"debafr/internal/domain"
)

// Порт 3002 стоит сразу в двух location'ах, как в .dev/nginx.conf: замена
// глобальная, поэтому переключиться должны оба, а не только первый.
const (
	nginxConfBlue = `server {
    listen 8080;

    location /api {
        proxy_pass http://127.0.0.1:3001; # blue
        #proxy_pass http://127.0.0.1:3011; # green
    }

    location /ws {
        proxy_pass http://127.0.0.1:3002; # blue
        #proxy_pass http://127.0.0.1:3012; # green
    }

    location / {
        proxy_pass http://127.0.0.1:3002; # blue
        #proxy_pass http://127.0.0.1:3012; # green
    }
}
`

	// На чистой установке активных строк нет: приложение только
	// раскомментирует следующую стратегию, и если что-то уже активно, в
	// одном location'е окажутся два proxy_pass, а такой конфиг nginx -t не
	// пропускает.
	nginxConfIdle = `server {
    listen 8080;

    location /api {
        #proxy_pass http://127.0.0.1:3001; # blue
        #proxy_pass http://127.0.0.1:3011; # green
    }

    location /ws {
        #proxy_pass http://127.0.0.1:3002; # blue
        #proxy_pass http://127.0.0.1:3012; # green
    }

    location / {
        #proxy_pass http://127.0.0.1:3002; # blue
        #proxy_pass http://127.0.0.1:3012; # green
    }
}
`

	nginxConfGreen = `server {
    listen 8080;

    location /api {
        #proxy_pass http://127.0.0.1:3001; # blue
        proxy_pass http://127.0.0.1:3011; # green
    }

    location /ws {
        #proxy_pass http://127.0.0.1:3002; # blue
        proxy_pass http://127.0.0.1:3012; # green
    }

    location / {
        #proxy_pass http://127.0.0.1:3002; # blue
        proxy_pass http://127.0.0.1:3012; # green
    }
}
`
)

func TestSwitchNginxSwapsEveryOccurrence(t *testing.T) {
	t.Parallel()

	ok := func() ([]byte, error) { return []byte("ok"), nil }

	tests := []struct {
		name  string
		given string
		ports []CurrNextPort
		want  string
	}{
		{
			name:  "blue -> green",
			given: nginxConfBlue,
			ports: []CurrNextPort{
				{Location: "/api", CurrentPort: "3001", NextPort: "3011"},
				{Location: "/ws", CurrentPort: "3002", NextPort: "3012"},
				{Location: "/", CurrentPort: "3002", NextPort: "3012"},
			},
			want: nginxConfGreen,
		},
		{
			name:  "green -> blue",
			given: nginxConfGreen,
			ports: []CurrNextPort{
				{Location: "/api", CurrentPort: "3011", NextPort: "3001"},
				{Location: "/ws", CurrentPort: "3012", NextPort: "3002"},
				{Location: "/", CurrentPort: "3012", NextPort: "3002"},
			},
			want: nginxConfBlue,
		},
		{
			name:  "install: включается следующая стратегия",
			given: nginxConfIdle,
			ports: []CurrNextPort{
				{Location: "/api", CurrentPort: EmptyValue, NextPort: "3001"},
				{Location: "/ws", CurrentPort: EmptyValue, NextPort: "3002"},
				{Location: "/", CurrentPort: EmptyValue, NextPort: "3002"},
			},
			want: nginxConfBlue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "nginx.conf")
			require.NoError(t, os.WriteFile(path, []byte(tt.given), fileMode))

			// Не литерал в поле proxyPass: gosec G101 читает "pass" в
			// имени поля как пароль.
			prefix := "proxy_pass http://127.0.0.1:"

			res := switchNginx(switchConfig{
				proxyPass: prefix,
				filePath:  path,
				ports:     tt.ports,
				test:      ok,
				reload:    ok,
			})
			require.Equal(t, domain.ExecResultStatusSuccess, res.Status, res.Output)

			got, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

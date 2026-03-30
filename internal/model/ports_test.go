package model

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_findPorts(t *testing.T) {
	tests := []struct {
		name     string
		pathFile string
		ports    []LocationPort
		wantErr  error
	}{
		{
			name:     "Should find ports - blue strategy",
			pathFile: "nginx.blue.conf",
			ports: []LocationPort{
				{
					Location:    "/api",
					CurrentPort: "3001",
					NextPort:    "3011",
				},
				{
					Location:    "/ws",
					CurrentPort: "3003",
					NextPort:    "3013",
				},
				{
					Location:    "/",
					CurrentPort: "3002",
					NextPort:    "3012",
				},
			},
			wantErr: nil,
		},
		{
			name:     "Should find ports - green strategy",
			pathFile: "nginx.green.conf",
			ports: []LocationPort{
				{
					Location:    "/api",
					CurrentPort: "3011",
					NextPort:    "3001",
				},
				{
					Location:    "/ws",
					CurrentPort: "3013",
					NextPort:    "3003",
				},
				{
					Location:    "/",
					CurrentPort: "3012",
					NextPort:    "3002",
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := os.OpenRoot("testdata/ports")
			if err != nil {
				t.Errorf("Failed to open root: %v", err)
			}
			defer root.Close()

			content, err := root.ReadFile(tt.pathFile)
			if err != nil {
				t.Errorf("Failed to read nginx config: %v", err)
			}

			ports, err := parseNginxLocationPorts(string(content))
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.ports, ports)
		})
	}
}

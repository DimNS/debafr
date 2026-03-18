package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_findPorts(t *testing.T) {
	type args struct {
		pathFile    string
		projectName string
	}
	tests := []struct {
		name         string
		args         args
		wantBackend  string
		wantFrontend string
		wantErr      error
	}{
		{
			name: "Should find ports - blue strategy",
			args: args{
				pathFile:    "testdata/ports/compose.blue.yaml",
				projectName: "capuchin",
			},
			wantBackend:  "3001",
			wantFrontend: "3002",
			wantErr:      nil,
		},
		{
			name: "Should find ports - green strategy",
			args: args{
				pathFile:    "testdata/ports/compose.green.yaml",
				projectName: "capuchin",
			},
			wantBackend:  "3011",
			wantFrontend: "3012",
			wantErr:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBackend, gotFrontend, err := findPorts(tt.args.pathFile, tt.args.projectName)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantBackend, gotBackend)
			assert.Equal(t, tt.wantFrontend, gotFrontend)
		})
	}
}

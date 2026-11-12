package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"debafr/internal/domain"
)

func TestReplaceInFile(t *testing.T) {
	t.Parallel()

	const (
		blue  = "app_blue:80"
		green = "app_green:80"
	)

	tests := []struct {
		name  string
		next  domain.Strategy
		given string
		want  string
		count int
	}{
		{
			name:  "blue -> green replaces every occurrence",
			next:  domain.StrategyGreen,
			given: `{"targets":["app_blue:80","app_blue:80"]}`,
			want:  `{"targets":["app_green:80","app_green:80"]}`,
			count: 2,
		},
		{
			name:  "green -> blue",
			next:  domain.StrategyBlue,
			given: `{"targets":["app_green:80"]}`,
			want:  `{"targets":["app_blue:80"]}`,
			count: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "targets.json")
			require.NoError(t, os.WriteFile(path, []byte(tt.given), fileMode))

			count, err := replaceInFile(
				domain.AppConfigReplacer{Path: path, Blue: blue, Green: green},
				tt.next,
			)
			require.NoError(t, err)
			assert.Equal(t, tt.count, count)

			got, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}

	t.Run("Should succeed when the next strategy value is already in place", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "targets.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"targets":["app_blue:80"]}`), fileMode))

		count, err := replaceInFile(
			domain.AppConfigReplacer{Path: path, Blue: blue, Green: green},
			domain.StrategyBlue,
		)
		require.NoError(t, err)
		assert.Zero(t, count)
	})

	t.Run("Should fail when the value of the current strategy is missing", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "targets.json")
		require.NoError(t, os.WriteFile(path, []byte("nothing here"), fileMode))

		_, err := replaceInFile(
			domain.AppConfigReplacer{Path: path, Blue: blue, Green: green},
			domain.StrategyGreen,
		)
		require.Error(t, err)
	})
}

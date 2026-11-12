package model

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"debafr/internal/domain"
)

func TestRunSwitching(t *testing.T) {
	t.Parallel()

	t.Run("Should skip everything when nginx is disabled", func(t *testing.T) {
		t.Parallel()

		res := runSwitching(context.Background(), domain.AppConfigNginx{
			Enabled: false,
			Test:    []string{"sh", "-c", "echo should-not-run; exit 1"},
		})

		require.Equal(t, domain.ExecResultStatusSuccess, res.Status)
		assert.Contains(t, res.Output, "disabled")
		assert.NotContains(t, res.Output, "should-not-run")
	})

	t.Run("Should run test and reload in order", func(t *testing.T) {
		t.Parallel()

		res := runSwitching(context.Background(), domain.AppConfigNginx{
			Enabled: true,
			Test:    []string{"echo", "test-ok"},
			Reload:  []string{"echo", "reload-ok"},
		})

		require.Equal(t, domain.ExecResultStatusSuccess, res.Status)
		assert.Less(t,
			strings.Index(res.Output, "test-ok"),
			strings.Index(res.Output, "reload-ok"),
			"the config is tested before it is reloaded",
		)
	})

	t.Run("Should skip empty command lists", func(t *testing.T) {
		t.Parallel()

		res := runSwitching(context.Background(), domain.AppConfigNginx{Enabled: true})

		require.Equal(t, domain.ExecResultStatusSuccess, res.Status)
		assert.Contains(t, res.Output, "No nginx commands configured")
	})

	t.Run("Should not reload when the test fails", func(t *testing.T) {
		t.Parallel()

		res := runSwitching(context.Background(), domain.AppConfigNginx{
			Enabled: true,
			Test:    []string{"false"},
			Reload:  []string{"echo", "reload-ok"},
		})

		require.Equal(t, domain.ExecResultStatusError, res.Status)
		require.Error(t, res.Err)
		assert.NotContains(t, res.Output, "reload-ok")
	})
}

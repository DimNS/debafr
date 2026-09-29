package model

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"debafr/internal/domain"
	"debafr/internal/provider/docker"
)

func Test_pullRatio(t *testing.T) {
	t.Parallel()

	t.Run("Should be empty while the size of the first layer is unknown", func(t *testing.T) {
		t.Parallel()

		assert.Zero(t, pullRatio(docker.PullProgress{Current: 10}), "nothing to compare with")
	})

	t.Run("Should count the downloaded bytes", func(t *testing.T) {
		t.Parallel()

		assert.InDelta(t, 0.25, pullRatio(docker.PullProgress{Current: 25, Total: 100}), 0.001)
	})

	t.Run("Should stay full when more bytes arrive than the layers announced", func(t *testing.T) {
		t.Parallel()

		assert.InEpsilon(t, 1, pullRatio(docker.PullProgress{Current: 120, Total: 100}), 0.001, "the bar must not overflow")
	})
}

func Test_barView(t *testing.T) {
	t.Parallel()

	theme := domain.NewTheme()

	for _, tt := range []struct {
		name  string
		ratio float64
		width int
	}{
		{name: "half", ratio: 0.5, width: 10},
		{name: "empty", ratio: 0, width: 10},
		{name: "full", ratio: 1, width: 10},
		{name: "narrow panel", ratio: 0.5, width: -4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			view := barView(theme, tt.ratio, tt.width)

			assert.Equal(t, max(tt.width, 0)+percentColumn, lipgloss.Width(view), "the bar has to fit into the panel")
		})
	}
}

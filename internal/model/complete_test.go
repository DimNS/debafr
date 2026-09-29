package model

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"debafr/internal/domain"
)

// completeDIC is fakeDIC plus the summary the final screen reads.
type completeDIC struct{ fakeDIC }

func (completeDIC) GetSummary() *Summary {
	return NewSummary(SummaryConfig{Theme: domain.NewTheme()})
}

func TestCompleteBanner(t *testing.T) {
	t.Parallel()

	// the width of the right panel, as Exec sees it
	const panelWidth = 120 - 50 - compensationWidth

	t.Run("Should centre the success line and show the total time", func(t *testing.T) {
		t.Parallel()

		c := &Complete{dic: completeDIC{}, theme: domain.NewTheme()}

		lines := strings.Split(c.banner(), "\n")
		require.Len(t, lines, 2)
		for _, line := range lines {
			assert.Equal(t, panelWidth, lipgloss.Width(line), "the banner fills the panel")
		}
		assert.Contains(t, lines[0], "Application deployed successfully")
		assert.Contains(t, lines[1], "Deployed in 0:00")
	})

	t.Run("Should not claim a success when the containers were not listed", func(t *testing.T) {
		t.Parallel()

		c := &Complete{dic: completeDIC{}, theme: domain.NewTheme(), output: "Error: nope", failed: true}

		assert.Equal(t, "Error: nope", c.View())
	})
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "0:00", formatDuration(0))
	assert.Equal(t, "0:09", formatDuration(9*time.Second))
	assert.Equal(t, "1:05", formatDuration(65*time.Second))
	assert.Equal(t, "9:59", formatDuration(599*time.Second))
}

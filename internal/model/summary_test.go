package model

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"

	"debafr/internal/domain"
)

func TestSplitString(t *testing.T) {
	t.Run("Should not cut utf-8 runes in the middle", func(t *testing.T) {
		// 10 runes, width 4 -> "ыбл" is impossible, a byte based cut would
		// produce invalid utf-8 here
		dir := "/srv/ыблы/проекты/debafr"

		got := splitString(dir, 10)

		for line := range strings.SplitSeq(got, "\n") {
			assert.True(t, utf8.ValidString(line), "line %q is not valid utf-8", line)
			assert.LessOrEqual(t, utf8.RuneCountInString(line), 10)
		}
		assert.Equal(t, dir, strings.ReplaceAll(got, "\n", ""))
	})

	t.Run("Should keep a short string as is", func(t *testing.T) {
		assert.Equal(t, "/srv/юзер", splitString("/srv/юзер", 10))
	})

	t.Run("Should return an empty string for an empty input", func(t *testing.T) {
		assert.Empty(t, splitString("", 10))
	})
}

func TestSteps(t *testing.T) {
	t.Parallel()

	withVM := domain.AppConfig{VictoriaMetrics: domain.AppConfigVictoriaMetrics{Enabled: true}}

	t.Run("Should count the pipeline of both modes", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 5, totalSteps(domain.AppConfig{}, domain.ModeInstall))
		assert.Equal(t, 7, totalSteps(domain.AppConfig{}, domain.ModeUpdate))
		assert.Equal(t, 8, totalSteps(withVM, domain.ModeUpdate))
	})

	t.Run("Should not count the steps before the mode is chosen", func(t *testing.T) {
		t.Parallel()

		s := &Summary{}
		s.NextStep()

		step, total := s.GetStep()
		assert.Zero(t, total)
		assert.Zero(t, step)
	})

	t.Run("Should count the steps and stop at the total", func(t *testing.T) {
		t.Parallel()

		s := &Summary{}
		s.SetTotalSteps(2)
		s.NextStep()
		s.NextStep()
		s.NextStep()

		step, total := s.GetStep()
		assert.Equal(t, 2, total)
		assert.Equal(t, 2, step)
	})
}

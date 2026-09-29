package model

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"debafr/internal/domain"
)

// fakeDIC only implements what Exec touches, everything else panics.
type fakeDIC struct{ DIC }

func (fakeDIC) GetTheme() *domain.Theme { return domain.NewTheme() }

func (fakeDIC) GetPhysicalWidth() int { return 120 }

func (fakeDIC) GetPhysicalHeight() int { return 40 }

func (fakeDIC) GetSummaryWidth() int { return 50 }

// stubModel stands for the next step of the pipeline.
type stubModel struct{}

func (*stubModel) Init() tea.Cmd { return nil }

func (s *stubModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return s, nil }

func (*stubModel) View() string { return "" }

// nextStep returns the model the command moves to, nil when it does not move.
func nextStep(cmd tea.Cmd) tea.Model {
	if cmd == nil {
		return nil
	}

	if msg, ok := cmd().(NextCmdMsg); ok {
		return msg.NextCmd
	}

	return nil
}

func TestExecAutoAdvance(t *testing.T) {
	t.Parallel()

	run := func(t *testing.T, autoAdvance bool, result domain.ExecResult) *Exec {
		t.Helper()

		c := NewExec(fakeDIC{}, domain.ExecConfig{
			Name:        "Step",
			StartFunc:   func() domain.ExecResult { return result },
			SuccessFunc: func() {},
			ErrorFunc:   func() {},
			AutoAdvance: autoAdvance,
			NextCmd:     &stubModel{},
		})
		c.exec()

		return c
	}

	success := domain.ExecResult{Status: domain.ExecResultStatusSuccess, Output: "ok"}
	failure := domain.ExecResult{Status: domain.ExecResultStatusError, Err: assert.AnError}

	t.Run("Should move to the next step right after a success", func(t *testing.T) {
		t.Parallel()

		c := run(t, true, success)

		_, cmd := c.Update(StatusDone{true})

		require.Same(t, c.execCfg.NextCmd, nextStep(cmd), "success must advance without enter")
	})

	t.Run("Should stop on a failure", func(t *testing.T) {
		t.Parallel()

		c := run(t, true, failure)

		_, cmd := c.Update(StatusDone{true})

		assert.Nil(t, nextStep(cmd), "failure must wait for the user")
	})

	t.Run("Should wait for enter when auto advance is off", func(t *testing.T) {
		t.Parallel()

		c := run(t, false, success)

		_, cmd := c.Update(StatusDone{true})
		require.Nil(t, nextStep(cmd), "nothing should happen until enter")

		_, cmd = c.Update(tea.KeyMsg{Type: tea.KeyEnter})
		require.Same(t, c.execCfg.NextCmd, nextStep(cmd), "enter must advance in the manual mode")
	})
}

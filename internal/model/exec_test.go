package model

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"debafr/internal/domain"
)

// fakeDIC only implements what Exec touches, everything else panics. It is a
// pointer, so a test can resize the terminal behind the running step.
type fakeDIC struct {
	DIC

	width, height int
}

func (f *fakeDIC) GetTheme() *domain.Theme { return domain.NewTheme() }

func (f *fakeDIC) GetPhysicalWidth() int { return f.width }

func (f *fakeDIC) GetPhysicalHeight() int { return f.height }

func (*fakeDIC) GetSummaryWidth() int { return 50 }

// the terminal of the tests, wide enough to always show the hint of the footer.
func newFakeDIC() *fakeDIC { return &fakeDIC{width: 120, height: 40} }

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

		c := NewExec(newFakeDIC(), domain.ExecConfig{
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

func TestExecFooter(t *testing.T) {
	t.Parallel()

	newExec := func(dic *fakeDIC, output string) *Exec {
		t.Helper()

		c := NewExec(dic, domain.ExecConfig{
			Name: "Step",
			StartFunc: func() domain.ExecResult {
				return domain.ExecResult{Status: domain.ExecResultStatusSuccess, Output: output}
			},
			SuccessFunc: func() {},
			ErrorFunc:   func() {},
		})
		c.exec()
		c.Update(StatusDone{true})

		return c
	}

	short := newExec(newFakeDIC(), "ok")
	long := newExec(newFakeDIC(), strings.Repeat("line\n", 200))

	t.Run("Should stay out of the way when the output fits", func(t *testing.T) {
		t.Parallel()

		assert.False(t, short.scrollable())
		assert.NotContains(t, short.View(), "Scroll:", "the footer is only noise here")
		assert.Equal(t, short.dic.GetPhysicalHeight()-compensationHeight, short.pager.Height,
			"the whole panel belongs to the output")
	})

	t.Run("Should take two lines from the output and give the last one back", func(t *testing.T) {
		t.Parallel()

		require.True(t, long.scrollable())
		assert.Contains(t, long.View(), "Scroll:")
		// the border of the panel is what compensationHeight pays for, the two
		// lines of the footer are taken from the output and nothing more
		assert.Equal(t, long.dic.GetPhysicalHeight()-compensationHeight, long.pager.Height+footerHeight,
			"pager plus footer have to fill the panel exactly")
	})

	t.Run("Should fit into the panel of any width", func(t *testing.T) {
		t.Parallel()

		// 120 - 50 - compensationWidth is the width of the panel, the hint is
		// dropped before the scroll info itself gets cut
		assert.Equal(t, long.pager.Width, lipgloss.Width(long.footer()))

		narrow := newExec(&fakeDIC{width: 80, height: 40}, strings.Repeat("line\n", 200))
		footer := narrow.footer()

		assert.Equal(t, narrow.pager.Width, lipgloss.Width(footer))
		assert.Contains(t, footer, "Scroll:")
		assert.NotContains(t, footer, "Press enter", "the hint has to yield to the scroll info")
	})
}

func TestExecResize(t *testing.T) {
	t.Parallel()

	// the layout has to follow the terminal, not the size it was built with
	dic := newFakeDIC()
	c := NewExec(dic, domain.ExecConfig{
		Name:        "Step",
		StartFunc:   func() domain.ExecResult { return domain.ExecResult{} },
		SuccessFunc: func() {},
		ErrorFunc:   func() {},
	})
	before := c.pager.Height

	dic.width, dic.height = 80, 20
	c.Update(tea.WindowSizeMsg{Width: 80, Height: 20})

	assert.NotEqual(t, before, c.pager.Height, "the resize has to reach the pager")
	assert.Equal(t, 80-dic.GetSummaryWidth()-compensationWidth, c.pager.Width)
	assert.LessOrEqual(t, c.pager.Height+compensationHeight, dic.GetPhysicalHeight())
}

// A terminal can report no size at all before it is laid out, and a negative
// pager height makes bubbles slice past the end of the content.
func TestExecResizeWithoutSize(t *testing.T) {
	t.Parallel()

	dic := &fakeDIC{width: 0, height: 0}
	c := NewExec(dic, domain.ExecConfig{
		Name: "Step",
		StartFunc: func() domain.ExecResult {
			return domain.ExecResult{Status: domain.ExecResultStatusError, Err: assert.AnError, Output: "line\n"}
		},
		SuccessFunc: func() {},
		ErrorFunc:   func() {},
	})

	c.exec()

	assert.NotPanics(t, func() { c.Update(StatusDone{true}) })
	assert.Positive(t, c.pager.Height, "the pager needs a height to slice the content by")
}

package model

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"debafr/internal/domain"
)

const (
	// horizontal space eaten by the panel border and its padding.
	compensationWidth = 6
	// vertical space eaten by the panel border, the footer is on top of that.
	compensationHeight = 2

	footerHeight = 2

	// how long the finished step stays on the screen before the next one
	// starts, so a fast step does not just flash.
	autoAdvancePause = 300 * time.Millisecond

	secondsInMinute = 60
)

type Exec struct {
	dic     DIC
	theme   *domain.Theme
	spinner spinner.Model
	pager   viewport.Model

	status *bool
	result domain.ExecResult

	startedAt time.Time

	execCfg domain.ExecConfig
}

func NewExec(dic DIC, execCfg domain.ExecConfig) *Exec {
	c := &Exec{
		dic:     dic,
		theme:   dic.GetTheme(),
		spinner: domain.NewSpinner(dic.GetTheme().StyleGreen),
		pager:   viewport.New(0, 0),

		execCfg: execCfg,
	}
	c.resize()

	return c
}

func (c *Exec) Init() tea.Cmd {
	c.startedAt = time.Now()
	c.dic.GetSummary().NextStep()

	return tea.Batch(c.exec, c.spinner.Tick)
}

func (c *Exec) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" && c.status != nil && *c.status {
			return c, c.next(0)
		}

	case StatusDone:
		switch c.result.Status {
		case domain.ExecResultStatusError:
			c.execCfg.ErrorFunc()
			c.status = func() *bool { b := false; return &b }()
		case domain.ExecResultStatusSuccess:
			c.execCfg.SuccessFunc()
			c.status = func() *bool { b := true; return &b }()
		default:
		}

		c.showOutput()
		c.resize()

		// a successful step does not need the user to read it, a failed one does.
		if c.execCfg.AutoAdvance && c.result.Status == domain.ExecResultStatusSuccess {
			return c, c.next(autoAdvancePause)
		}

	default:
	}

	if _, ok := msg.(tea.WindowSizeMsg); ok {
		c.resize()
	}

	var (
		cmdPager, cmdSpinner tea.Cmd
		cmds                 []tea.Cmd
	)

	c.pager, cmdPager = c.pager.Update(msg)
	c.spinner, cmdSpinner = c.spinner.Update(msg)
	cmds = []tea.Cmd{cmdPager, cmdSpinner}

	return c, tea.Batch(cmds...)
}

func (c *Exec) View() string {
	if c.status == nil {
		c.pager.SetContent(fmt.Sprintf(
			"%s %s  %s  %s",
			c.execCfg.Name, c.spinner.View(), c.stepsView(), c.elapsedView(),
		))

		return c.pager.View()
	}

	if !c.scrollable() {
		return c.pager.View()
	}

	return fmt.Sprintf("%s\n%s", c.pager.View(), c.footer())
}

// stepsView is the honest progress metric: the steps are not weighted by time,
// so counting them says more than a percentage would.
func (c *Exec) stepsView() string {
	step, total := c.dic.GetSummary().GetStep()
	if total == 0 {
		return ""
	}

	return fmt.Sprintf("▸ %d/%d", step, total)
}

// elapsedView is recomputed on every spinner frame, so the step needs no timer
// state of its own.
func (c *Exec) elapsedView() string {
	d := time.Since(c.startedAt).Truncate(time.Second)

	return fmt.Sprintf(
		"%d:%02d", int(d.Minutes()), int(d.Seconds())%secondsInMinute,
	)
}

// scrollable reports whether the content is taller than the panel, the only case
// where the footer earns its two lines.
func (c *Exec) scrollable() bool {
	return c.pager.TotalLineCount() > c.dic.GetPhysicalHeight()-compensationHeight
}

// resize refits the pager into the current terminal size, the footer is
// accounted for only when it is actually rendered.
func (c *Exec) resize() {
	height := c.dic.GetPhysicalHeight() - compensationHeight
	if c.scrollable() {
		height -= footerHeight
	}

	c.pager.Width = c.dic.GetPhysicalWidth() - c.dic.GetSummaryWidth() - compensationWidth
	c.pager.Height = height
}

// footer is kept to exactly two lines, so a narrow terminal drops the hint
// instead of pushing the last line of the output out of the screen.
func (c *Exec) footer() string {
	text := fmt.Sprintf("Scroll: %3.f%%", c.pager.ScrollPercent()*100) //nolint:mnd // ignore
	if hint := c.theme.TextPressEnterToContinue; lipgloss.Width(text+hint)+2 <= c.pager.Width {
		text += "  " + hint
	}

	return lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(c.theme.ColorGray).
		Foreground(c.theme.ColorYellow).
		Width(c.pager.Width).
		Render(text)
}

func (c *Exec) exec() tea.Msg {
	c.result = c.execCfg.StartFunc()

	return StatusDone{true}
}

// next moves to the next step of the pipeline after the given pause, zero
// meaning without waiting.
func (c *Exec) next(pause time.Duration) tea.Cmd {
	return tea.Tick(pause, func(time.Time) tea.Msg {
		return NextCmdMsg{
			NextCmd: c.execCfg.NextCmd,
		}
	})
}

// showOutput puts the result of the finished step into the pager.
func (c *Exec) showOutput() {
	if c.result.Status != domain.ExecResultStatusError {
		c.pager.SetContent(c.result.Output)
		c.pager.GotoBottom()

		return
	}

	var output string
	if c.result.Output != "" {
		output = "\n\nOutput:\n" + c.result.Output
	}

	c.pager.SetContent(
		fmt.Sprintf(
			"%s...\n\n%s%s",
			c.execCfg.Name,
			c.theme.StyleRed.Render("Error: "+c.result.Err.Error()),
			output,
		),
	)
	c.pager.GotoBottom()
}

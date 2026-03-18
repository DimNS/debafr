package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"debafr/internal/domain"
)

type Exec struct {
	theme   *domain.Theme
	spinner spinner.Model
	pager   viewport.Model

	status *bool
	result domain.ExecResult

	execCfg domain.ExecConfig
}

func NewExec(dic DIC, execCfg domain.ExecConfig) *Exec {
	compensationWidth := 6
	compensationHeight := 4

	pager := viewport.New(
		dic.GetPhysicalWidth()-dic.GetSummaryWidth()-compensationWidth,
		dic.GetPhysicalHeight()-compensationHeight,
	)

	return &Exec{
		theme:   dic.GetTheme(),
		spinner: domain.NewSpinner(dic.GetTheme().StyleGreen),
		pager:   pager,

		execCfg: execCfg,
	}
}

func (c *Exec) Init() tea.Cmd {
	return tea.Batch(c.exec, c.spinner.Tick)
}

func (c *Exec) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" && c.status != nil && *c.status {
			return c, func() tea.Msg {
				return NextCmdMsg{
					NextCmd: c.execCfg.NextCmd,
				}
			}
		}

	case StatusDone:
		switch c.result.Status {
		case domain.ExecResultStatusError:
			c.execCfg.ErrorFunc()
			c.status = func() *bool { b := false; return &b }()
		case domain.ExecResultStatusSuccess:
			c.execCfg.SuccessFunc()
			c.status = func() *bool { b := true; return &b }()
		}

		if c.result.Status == domain.ExecResultStatusError {
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
		} else {
			c.pager.SetContent(c.result.Output + "\n" + c.theme.TextPressEnterToContinue)
			c.pager.GotoBottom()
		}
	}

	c.pager, cmd = c.pager.Update(msg)
	cmds = append(cmds, cmd)

	c.spinner, cmd = c.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *Exec) View() string {
	footer := lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(c.theme.ColorGray).
		Foreground(c.theme.ColorYellow).
		Render(fmt.Sprintf("Scroll: %3.f%%", c.pager.ScrollPercent()*100)) //nolint:mnd // ignore

	if c.status == nil {
		c.pager.SetContent(fmt.Sprintf("%s %s", c.execCfg.Name, c.spinner.View()))

		return fmt.Sprintf("%s\n%s", c.pager.View(), footer)
	}

	return fmt.Sprintf("%s\n%s", c.pager.View(), footer)
}

func (c *Exec) exec() tea.Msg {
	c.result = c.execCfg.StartFunc()

	return StatusDone{true}
}

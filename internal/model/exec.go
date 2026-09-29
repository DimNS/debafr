package model

import (
	"fmt"

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
	footerHeight       = 2
)

type Exec struct {
	dic     DIC
	theme   *domain.Theme
	spinner spinner.Model
	pager   viewport.Model

	status *bool
	result domain.ExecResult

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
	return tea.Batch(c.exec, c.spinner.Tick)
}

func (c *Exec) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		default:
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
			c.pager.SetContent(c.result.Output)
			c.pager.GotoBottom()
		}

		c.resize()

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
		c.pager.SetContent(fmt.Sprintf("%s %s", c.execCfg.Name, c.spinner.View()))

		return c.pager.View()
	}

	if !c.scrollable() {
		return c.pager.View()
	}

	return fmt.Sprintf("%s\n%s", c.pager.View(), c.footer())
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

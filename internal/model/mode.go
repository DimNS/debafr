package model

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"debafr/internal/domain"
)

type Mode struct {
	dic DIC

	mode string

	lg   *lipgloss.Renderer
	form *huh.Form
}

func NewMode(dic DIC) *Mode {
	c := &Mode{
		dic: dic,
		lg:  lipgloss.DefaultRenderer(),
	}

	c.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose mode").
				Options(huh.NewOptions(
					domain.ModeUpdate.String(),
					domain.ModeInstall.String(),
				)...).
				Value(&c.mode),
		).WithShowHelp(false),
	)

	return c
}

func (c *Mode) Init() tea.Cmd {
	return c.form.Init()
}

func (c *Mode) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	form, cmd := c.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		c.form = f
		cmds = append(cmds, cmd)
	}

	if c.form.State == huh.StateCompleted {
		c.dic.GetSummary().UpdateMode(domain.Mode(c.mode))

		cmds = append(cmds, func() tea.Msg {
			return NextCmdMsg{
				NextCmd: NewExecRequirements(c.dic),
			}
		})
	}

	return c, tea.Batch(cmds...)
}

func (c *Mode) View() string {
	v := strings.TrimSuffix(c.form.View(), "\n\n")

	return c.lg.NewStyle().Render(v)
}

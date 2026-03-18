package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-version"
)

type NextDeploy struct {
	dic DIC

	nextVersion string

	lg   *lipgloss.Renderer
	form *huh.Form
}

func NewNextDeploy(dic DIC) *NextDeploy {
	c := &NextDeploy{
		dic: dic,

		nextVersion: dic.GetSummary().GetNextVersion(),

		lg: lipgloss.DefaultRenderer(),
	}

	c.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Input next version to deploy").
				Validate(isVersion).
				Value(&c.nextVersion),
		).WithShowHelp(false),
	)

	return c
}

func (c *NextDeploy) Init() tea.Cmd {
	return c.form.Init()
}

func (c *NextDeploy) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	form, cmd := c.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		c.form = f
		cmds = append(cmds, cmd)
	}

	if c.form.State == huh.StateCompleted {
		c.dic.GetSummary().UpdateNextVersion(c.nextVersion)

		cmds = append(cmds, func() tea.Msg {
			return NextCmdMsg{
				NextCmd: NewExecLaunchingDeploy(c.dic),
			}
		})
	}

	return c, tea.Batch(cmds...)
}

func (c *NextDeploy) View() string {
	v := strings.TrimSuffix(c.form.View(), "\n\n")

	return c.lg.NewStyle().Render(v)
}

func isVersion(s string) error {
	_, err := version.NewSemver(s)
	if err != nil {
		return fmt.Errorf("invalid version: %v", err)
	}

	return nil
}

package model

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"debafr/internal/domain"
)

type Dir struct {
	dic     DIC
	spinner spinner.Model

	dir string
	err error
}

func NewDir(dic DIC) *Dir {
	return &Dir{
		dic:     dic,
		spinner: domain.NewSpinner(dic.GetTheme().StyleGreen),
	}
}

func (c *Dir) Init() tea.Cmd {
	return tea.Batch(c.getDir, c.spinner.Tick)
}

func (c *Dir) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusDone:
		c.dic.GetSummary().UpdateDir(c.dir)

		return c, func() tea.Msg {
			return NextCmdMsg{
				NextCmd: NewMode(c.dic),
			}
		}
	case StatusError:
		c.err = msg
		return c, nil
	default:
		var cmd tea.Cmd
		c.spinner, cmd = c.spinner.Update(msg)
		return c, cmd
	}
}

func (c *Dir) View() string {
	prefix := "Getting current directory"

	if c.err != nil {
		return prefix + "... " + c.dic.GetTheme().StyleRed.Render(
			fmt.Sprintf("something went wrong: %s", c.err),
		)
	}

	return fmt.Sprintf("%s %s", prefix, c.spinner.View())
}

func (c *Dir) getDir() tea.Msg {
	dir, err := os.Getwd()
	if err != nil {
		return StatusError{
			fmt.Errorf("failed to get current directory: %v", err),
		}
	}
	if dir == "" {
		return StatusError{
			errors.New("failed to get current directory"),
		}
	}

	if c.dic.GetDevMode() {
		c.dir = dir + "/.dev"
	} else {
		c.dir = dir
	}

	return StatusDone{true}
}

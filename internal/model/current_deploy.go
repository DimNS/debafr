package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"debafr/internal/domain"
)

type CurrentDeploy struct {
	dic     DIC
	summary *Summary
	spinner spinner.Model

	version  string
	strategy domain.Strategy
	err      error
}

func NewCurrentDeploy(dic DIC) *CurrentDeploy {
	return &CurrentDeploy{
		dic:     dic,
		summary: dic.GetSummary(),
		spinner: domain.NewSpinner(dic.GetTheme().StyleGreen),
	}
}

func (c *CurrentDeploy) Init() tea.Cmd {
	return tea.Batch(c.getCurrentDeploy, c.spinner.Tick)
}

func (c *CurrentDeploy) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusDone:
		c.summary.UpdateCurrentVersion(c.version)
		c.summary.UpdateCurrentStrategy(c.strategy)

		nextStrategy := domain.StrategyBlue
		if c.strategy == domain.StrategyBlue {
			nextStrategy = domain.StrategyGreen
		}
		c.summary.UpdateNextStrategy(nextStrategy)

		return c, func() tea.Msg {
			return NextCmdMsg{
				NextCmd: NewExecPorts(c.dic),
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

func (c *CurrentDeploy) View() string {
	prefix := "Getting current deploy"

	if c.err != nil {
		return prefix + "... " + c.dic.GetTheme().StyleRed.Render(
			fmt.Sprintf("something went wrong: %s", c.err),
		)
	}

	return fmt.Sprintf("%s %s", prefix, c.spinner.View())
}

func (c *CurrentDeploy) getCurrentDeploy() tea.Msg {
	if c.summary.GetMode() == domain.ModeUpdate {
		v, strategy, err := c.dic.GetDockerService().GetCurrentDeploy(c.summary.GetProjectName())
		if err != nil {
			return StatusError{
				fmt.Errorf("failed to get current deploy: %v", err),
			}
		}

		c.version = v
		c.strategy = strategy

		return StatusDone{true}
	}

	c.version = EmptyValue
	c.strategy = EmptyValue

	return StatusDone{true}
}

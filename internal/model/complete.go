package model

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"debafr/internal/domain"
)

type Complete struct {
	theme *domain.Theme

	output string
}

func NewComplete(theme *domain.Theme) *Complete {
	return &Complete{
		theme: theme,
	}
}

func (c *Complete) Init() tea.Cmd {
	out, err := exec.Command(PathDocker, "ps").CombinedOutput()
	if err != nil {
		outString := string(out)

		var output string
		if outString != "" {
			output = "\n\nOutput:\n" + outString
		}

		c.output = fmt.Sprintf(
			"%s%s",
			c.theme.StyleRed.Render("Error: "+err.Error()),
			output,
		)
	}

	c.output = string(out)

	return nil
}

func (c *Complete) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return c, nil
}

func (c *Complete) View() string {
	return c.output + "\n🎉 Application updated successfully"
}

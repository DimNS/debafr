package model

import (
	"context"
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"debafr/internal/domain"
)

type Complete struct {
	theme *domain.Theme
	cfg   domain.AppConfig

	output string
}

func NewComplete(dic DIC) *Complete {
	return &Complete{
		theme: dic.GetTheme(),
		cfg:   dic.GetAppConfig(),
	}
}

func (c *Complete) Init() tea.Cmd {
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.Timeouts.Default)
	defer cancel()

	out, err := exec.CommandContext(ctx, c.cfg.BinPaths.Docker, "ps").CombinedOutput()
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
	return c.output + "\n🎉 Application deployed successfully"
}

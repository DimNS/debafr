package model

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"debafr/internal/domain"
	"debafr/internal/provider/docker"
)

type Complete struct {
	theme         *domain.Theme
	cfg           domain.AppConfig
	dockerService *docker.Docker

	output string
}

const shortIDLength = 12

func NewComplete(dic DIC) *Complete {
	return &Complete{
		theme:         dic.GetTheme(),
		cfg:           dic.GetAppConfig(),
		dockerService: dic.GetDockerService(),
	}
}

func (c *Complete) Init() tea.Cmd {
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.Timeouts.Default)
	defer cancel()

	containers, err := c.dockerService.ListRunning(ctx)
	if err != nil {
		c.output = c.theme.StyleRed.Render("Error: " + err.Error())

		return nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%-14s %-24s %-12s %s\n", "CONTAINER ID", "IMAGE", "STATUS", "NAMES")
	for _, ctr := range containers {
		id := ctr.ID
		if len(id) > shortIDLength {
			id = id[:shortIDLength]
		}

		fmt.Fprintf(
			&sb,
			"%-14s %-24s %-12s %s\n",
			id,
			ctr.Image,
			ctr.Status,
			strings.TrimPrefix(strings.Join(ctr.Names, ", "), "/"),
		)
	}
	c.output = sb.String()

	return nil
}

func (c *Complete) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return c, nil
}

func (c *Complete) View() string {
	return c.output + "\n🎉 Application deployed successfully"
}

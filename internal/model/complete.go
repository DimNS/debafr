package model

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"debafr/internal/domain"
	"debafr/internal/provider/docker"
)

type Complete struct {
	dic           DIC
	theme         *domain.Theme
	cfg           domain.AppConfig
	dockerService *docker.Docker

	output string
	failed bool
}

const shortIDLength = 12

func NewComplete(dic DIC) *Complete {
	return &Complete{
		dic:           dic,
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
		c.failed = true

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
	if c.failed {
		return c.output
	}

	return c.output + "\n" + c.banner()
}

// banner is the last thing the user reads, so it takes the whole width of the
// panel and tells how long the deploy took.
func (c *Complete) banner() string {
	return c.theme.StyleGreen.
		Width(c.dic.GetPhysicalWidth() - c.dic.GetSummaryWidth() - compensationWidth).
		Align(lipgloss.Left).
		Render("🎉 Application deployed successfully\nDeployed in " + formatDuration(c.dic.GetSummary().GetElapsed()))
}

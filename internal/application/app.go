package application

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"debafr/internal/domain"
	"debafr/internal/model"
	"debafr/internal/provider/docker"
)

const (
	summaryWidth = 50

	compensationWidth  = 4
	compensationHeight = 2

	defaultTimeout = 10 * time.Second
)

type App struct {
	summary *model.Summary

	physicalWidth  int
	physicalHeight int

	leftStyle  lipgloss.Style
	rightStyle lipgloss.Style

	currentCmd tea.Model
	keys       domain.KeyMap
}

func New(appVersion string) (*App, error) {
	conf, err := LoadConfiguration(".debafr.toml")
	if err != nil {
		return nil, fmt.Errorf("load configuration: %v", err)
	}

	physicalWidth, physicalHeight, err := term.GetSize(int(os.Stdout.Fd())) //nolint:gosec // ignore
	if err != nil {
		return nil, fmt.Errorf("get terminal size: %v", err)
	}

	dockerService, err := docker.New(conf.DevMode)
	if err != nil {
		return nil, fmt.Errorf("new docker: %v", err)
	}

	theme := domain.NewTheme()

	summary := model.NewSummary(model.SummaryConfig{
		AppVersion: appVersion,
		DevMode:    conf.DevMode,

		ProjectName: conf.Toml.App.ProjectName,

		Width: summaryWidth,

		Theme: theme,

		FilenameComposeBlue:  conf.Toml.Files.ComposeBlue,
		FilenameComposeGreen: conf.Toml.Files.ComposeGreen,
		FilenameNginxConf:    conf.Toml.Files.NginxConf,
	})

	dic := NewDIC(DICConfig{
		DevMode: conf.DevMode,

		SummaryWidth: summaryWidth,

		PhysicalWidth:  physicalWidth,
		PhysicalHeight: physicalHeight,

		Theme: theme,

		Summary: summary,

		DockerService: dockerService,

		AppConfig: conf.Toml.GetDomainConfig(),
	})

	return &App{
		summary: summary,

		physicalWidth:  physicalWidth,
		physicalHeight: physicalHeight,

		leftStyle: lipgloss.NewStyle().
			Width(summaryWidth).
			Height(physicalHeight-compensationHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorGreen).
			Padding(0, 1),
		rightStyle: lipgloss.NewStyle().
			Width(physicalWidth-summaryWidth-compensationWidth).
			Height(physicalHeight-compensationHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorGreen).
			Padding(0, 1),

		currentCmd: model.NewDir(dic),
		keys:       domain.NewKeyMap(),
	}, nil
}

func (a *App) Init() tea.Cmd {
	return a.currentCmd.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var cmd tea.Cmd

	a.currentCmd, cmd = a.currentCmd.Update(msg)

	switch msg := msg.(type) {
	case model.NextCmdMsg:
		a.currentCmd = msg.NextCmd
		return a, a.currentCmd.Init()

	case tea.KeyMsg:
		if key.Matches(msg, a.keys.Quit) { //nolint:nestif // ignore
			if a.summary.GetDevMode() {
				output, err := exec.CommandContext(ctx, "/usr/bin/docker", "rm", "-f", "debafr_app").CombinedOutput()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(output))

				output, err = exec.CommandContext(ctx, "/usr/bin/docker", "rmi", "-f", "nginx:alpine").CombinedOutput()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(output))
			}

			return a, tea.Quit
		}

	default:
	}

	return a, cmd
}

func (a *App) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		a.leftStyle.Render(a.summary.View()),
		a.rightStyle.Render(a.currentCmd.View()),
	)
}

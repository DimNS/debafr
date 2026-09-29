package application

import (
	"context"
	"fmt"
	"os"
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

	defaultUpdateTimeout = 5 * time.Minute
)

type App struct {
	summary *model.Summary

	dockerService *docker.Docker

	dic *DIC

	leftStyle  lipgloss.Style
	rightStyle lipgloss.Style

	currentCmd tea.Model
	keys       domain.KeyMap
}

func New(appVersion string) (*App, error) {
	conf, err := LoadConfiguration("debafr.toml")
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

	app := &App{
		summary: summary,

		dockerService: dockerService,

		dic: dic,

		currentCmd: model.NewDir(dic),
		keys:       domain.NewKeyMap(),
	}
	app.rebuildStyles(physicalWidth, physicalHeight)

	return app, nil
}

func (a *App) Init() tea.Cmd {
	return a.currentCmd.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// the size has to be known before anything is rendered, so update the
	// styles here, before the message is handed over to the current model
	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		a.dic.SetPhysicalSize(sizeMsg.Width, sizeMsg.Height)
		a.rebuildStyles(sizeMsg.Width, sizeMsg.Height)
	}

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
				if err := a.dockerService.RemoveContainer(ctx, model.TestContainerName); err != nil {
					fmt.Println(err)
				}
				if err := a.dockerService.RemoveImage(ctx, "nginx:alpine"); err != nil {
					fmt.Println(err)
				}
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

// rebuildStyles refits both panels into the given terminal size.
func (a *App) rebuildStyles(physicalWidth, physicalHeight int) {
	theme := a.dic.GetTheme()

	a.leftStyle = lipgloss.NewStyle().
		Width(summaryWidth).
		Height(physicalHeight-compensationHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorGreen).
		Padding(0, 1)
	a.rightStyle = lipgloss.NewStyle().
		Width(physicalWidth-summaryWidth-compensationWidth).
		Height(physicalHeight-compensationHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorGreen).
		Padding(0, 1)
}

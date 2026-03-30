package model

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"debafr/internal/domain"
)

type SummaryConfig struct {
	AppVersion string
	DevMode    bool

	ProjectName string

	Width int

	Theme *domain.Theme

	FilenameComposeBlue  string
	FilenameComposeGreen string
	FilenameNginxConf    string
}

type Summary struct {
	appVersion string
	devMode    bool
	theme      *domain.Theme

	dirMaxWidth int

	projectName string
	mode        domain.Mode

	requirementsCurlVersion          string
	requirementsDockerVersion        string
	requirementsDockerComposeVersion string
	requirementsNginxVersion         string

	filenameComposeBlue  string
	filenameComposeGreen string
	filenameNginxConf    string

	currentDir string

	currentVersion  string
	currentStrategy domain.Strategy

	nextVersion  string
	nextStrategy domain.Strategy

	ports []LocationPort

	deployLaunching   *bool
	deployHealthcheck *bool

	switchingNginx *bool

	shutdownStopping *bool

	styles styles
}

type LocationPort struct {
	Location    string
	CurrentPort string
	NextPort    string
}

type styles struct {
	category lipgloss.Style
	title    lipgloss.Style
	text     lipgloss.Style
}

func NewSummary(cfg SummaryConfig) *Summary {
	marginCompensation := 3

	return &Summary{
		appVersion: cfg.AppVersion,
		devMode:    cfg.DevMode,
		theme:      cfg.Theme,

		dirMaxWidth: cfg.Width - marginCompensation,

		projectName: cfg.ProjectName,
		mode:        "???",

		requirementsCurlVersion:          "???",
		requirementsDockerVersion:        "???",
		requirementsDockerComposeVersion: "???",
		requirementsNginxVersion:         "???",

		filenameComposeBlue:  cfg.FilenameComposeBlue,
		filenameComposeGreen: cfg.FilenameComposeGreen,
		filenameNginxConf:    cfg.FilenameNginxConf,

		currentDir: "???",

		currentVersion:  "???",
		currentStrategy: "???",

		nextVersion:  "???",
		nextStrategy: "???",

		ports: nil,

		styles: styles{
			category: lipgloss.NewStyle().
				Foreground(cfg.Theme.ColorWhite).
				Bold(true).
				Transform(strings.ToUpper).
				MarginTop(1),

			title: lipgloss.NewStyle().
				Foreground(cfg.Theme.ColorGreen).
				Bold(true),

			text: lipgloss.NewStyle().
				Foreground(cfg.Theme.ColorYellow).
				Bold(true),
		},
	}
}

func (s *Summary) View() string {
	title := lipgloss.NewStyle().
		Foreground(s.theme.ColorOrange).
		Bold(true).
		Transform(strings.ToUpper).
		Render("🛠️ Debafr")
	version := lipgloss.NewStyle().
		Foreground(s.theme.ColorWhite).
		Render(s.appVersion)
	header := lipgloss.NewStyle().
		MarginBottom(1).
		Render(title + " v" + version)

	devModeStr := s.styles.text.Render("off")
	if s.devMode {
		devModeStr = lipgloss.NewStyle().
			Foreground(s.theme.ColorRed).
			Bold(true).
			Render("on")
	}

	var deploy string
	if s.mode == domain.ModeInstall || s.mode == domain.ModeUpdate {
		deploy = fmt.Sprintf(
			"%s\n%s\n%s",
			s.styles.category.Render("Deploy ("+s.nextVersion+")"),
			s.styles.title.Render("Launching:   ")+s.boolToIcon(s.deployLaunching),
			s.styles.title.Render("Healthcheck: ")+s.boolToIcon(s.deployHealthcheck),
		)
	}

	var switchStrategy string
	var shutdown string
	if s.mode == domain.ModeUpdate {
		switchStrategy = fmt.Sprintf(
			"%s\n%s",
			s.styles.category.Render("Switch strategy"),
			s.styles.title.Render("Switching - Nginx: ")+s.boolToIcon(s.switchingNginx),
		)

		shutdown = fmt.Sprintf(
			"%s\n%s",
			s.styles.category.Render("Shutdown ("+s.currentVersion+")"),
			s.styles.title.Render("Stopping the old version: ")+s.boolToIcon(s.shutdownStopping),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		s.styles.title.Render("DevMode: ")+devModeStr,
		s.styles.title.Render("Project: ")+s.styles.text.Render(s.projectName),
		s.styles.title.Render("Mode:    ")+s.styles.text.Render(s.mode.String()),

		s.styles.category.Render("Requirements"),
		s.styles.title.Render("curl:           ")+s.styles.text.Render(s.requirementsCurlVersion),
		s.styles.title.Render("docker:         ")+s.styles.text.Render(s.requirementsDockerVersion),
		s.styles.title.Render("docker compose: ")+s.styles.text.Render(s.requirementsDockerComposeVersion),
		s.styles.title.Render("nginx:          ")+s.styles.text.Render(s.requirementsNginxVersion),

		s.styles.category.Render("Directory"),
		s.styles.text.Render(splitString(s.currentDir, s.dirMaxWidth)),

		s.styles.category.Render("Files"),
		s.styles.title.Render("compose.blue.yaml:    ")+s.styles.text.Render(s.filenameComposeBlue),
		s.styles.title.Render("compose.green.yaml:   ")+s.styles.text.Render(s.filenameComposeGreen),
		s.styles.title.Render("nginx.conf (symlink): ")+s.styles.text.Render(s.filenameNginxConf),

		s.styles.category.Render("Deploy strategy"),
		s.styles.title.Render("Version:  ")+s.styles.text.Render(s.currentVersion)+" >> "+s.styles.text.Render(s.nextVersion),
		s.styles.title.Render("Strategy: ")+s.styles.text.Render(s.currentStrategy.String())+" >> "+s.styles.text.Render(s.nextStrategy.String()),

		s.portsView(),

		deploy,
		switchStrategy,
		shutdown,
	)
}

func (s *Summary) GetDevMode() bool {
	return s.devMode
}

func (s *Summary) GetProjectName() string {
	return s.projectName
}

func (s *Summary) GetMode() domain.Mode {
	return s.mode
}

func (s *Summary) GetRoot() (*os.Root, error) {
	root, err := os.OpenRoot(s.currentDir)
	if err != nil {
		return nil, fmt.Errorf("open root: %v", err)
	}

	return root, nil
}

func (s *Summary) GetFilenameComposeBlue() string {
	return s.filenameComposeBlue
}

func (s *Summary) GetFilenameComposeGreen() string {
	return s.filenameComposeGreen
}

func (s *Summary) GetFilenameNginxConf() string {
	return s.filenameNginxConf
}

func (s *Summary) GetCurrentVersion() string {
	return s.currentVersion
}

func (s *Summary) GetCurrentStrategy() domain.Strategy {
	return s.currentStrategy
}

func (s *Summary) GetNextVersion() string {
	return s.nextVersion
}

func (s *Summary) GetNextStrategy() domain.Strategy {
	return s.nextStrategy
}

func (s *Summary) GetPorts() []LocationPort {
	return s.ports
}

func (s *Summary) UpdateDir(value string) {
	s.currentDir = value
}

func (s *Summary) UpdateMode(value domain.Mode) {
	s.mode = value
}

func (s *Summary) UpdateRequirementsCurlVersion(value string) {
	s.requirementsCurlVersion = value
}

func (s *Summary) UpdateRequirementsDockerVersion(value string) {
	s.requirementsDockerVersion = value
}

func (s *Summary) UpdateRequirementsDockerComposeVersion(value string) {
	s.requirementsDockerComposeVersion = value
}

func (s *Summary) UpdateRequirementsNginxVersion(value string) {
	s.requirementsNginxVersion = value
}

func (s *Summary) UpdateCurrentVersion(value string) {
	s.currentVersion = value
}

func (s *Summary) UpdateCurrentStrategy(value domain.Strategy) {
	s.currentStrategy = value
}

func (s *Summary) UpdateNextVersion(value string) {
	s.nextVersion = value
}

func (s *Summary) UpdateNextStrategy(value domain.Strategy) {
	s.nextStrategy = value
}

func (s *Summary) UpdatePorts(ports []LocationPort) {
	s.ports = ports
}

func (s *Summary) UpdateDeployLaunching(value bool) {
	s.deployLaunching = &value
}

func (s *Summary) UpdateDeployHealthcheck(value bool) {
	s.deployHealthcheck = &value
}

func (s *Summary) UpdateSwitchingNginx(value bool) {
	s.switchingNginx = &value
}

func (s *Summary) UpdateShutdownStopping(value bool) {
	s.shutdownStopping = &value
}

func (s *Summary) boolToIcon(b *bool) string {
	if b == nil {
		return s.styles.text.Render("???")
	}

	if *b {
		return s.theme.StyleGreen.Render("✅")
	}

	return s.theme.StyleRed.Render("❌")
}

func (s *Summary) portsView() string {
	if len(s.ports) == 0 {
		return ""
	}

	lines := []string{
		s.styles.category.Render("Ports"),
	}

	for _, p := range s.ports {
		loc := s.styles.title.Render(fmt.Sprintf("%s: ", p.Location))
		cp := s.styles.text.Render(p.CurrentPort)
		np := s.styles.text.Render(p.NextPort)
		lines = append(lines,
			loc+cp+" >> "+np,
		)
	}

	return strings.Join(lines, "\n")
}

func splitString(s string, width int) string {
	var result []string

	for len(s) > width {
		result = append(result, s[:width])
		s = s[width:]
	}

	if len(s) > 0 {
		result = append(result, s)
	}

	return strings.Join(result, "\n")
}

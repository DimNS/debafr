package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"debafr/internal/domain"
)

const (
	// pendingValue is the placeholder of everything the steps have not filled
	// in yet, the spinner frame takes its place on the screen.
	pendingValue = "⏳"

	// labelWidth fits the longest label of the panel, so all the values line up
	// in one column no matter which section they belong to.
	labelWidth = 26
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

	currNextPorts []CurrNextPort

	deployLaunching   *bool
	deployHealthcheck *bool

	switchingNginx *bool

	shutdownStopping *bool

	step       int
	totalSteps int

	// pendingFrame is the current frame of the spinner, shown instead of the
	// placeholder until the value is known.
	pendingFrame string

	styles styles
}

type CurrNextPort struct {
	Location    string
	CurrentPort string
	NextPort    string
}

type styles struct {
	category lipgloss.Style
	label    lipgloss.Style
	value    lipgloss.Style
}

func NewSummary(cfg SummaryConfig) *Summary {
	marginCompensation := 3

	return &Summary{
		appVersion: cfg.AppVersion,
		devMode:    cfg.DevMode,
		theme:      cfg.Theme,

		dirMaxWidth: cfg.Width - marginCompensation,

		projectName: cfg.ProjectName,
		mode:        pendingValue,

		requirementsCurlVersion:          pendingValue,
		requirementsDockerVersion:        pendingValue,
		requirementsDockerComposeVersion: pendingValue,
		requirementsNginxVersion:         pendingValue,

		filenameComposeBlue:  cfg.FilenameComposeBlue,
		filenameComposeGreen: cfg.FilenameComposeGreen,
		filenameNginxConf:    cfg.FilenameNginxConf,

		currentDir: pendingValue,

		currentVersion:  pendingValue,
		currentStrategy: pendingValue,

		nextVersion:  pendingValue,
		nextStrategy: pendingValue,

		currNextPorts: nil,

		pendingFrame: pendingValue,

		styles: styles{
			category: lipgloss.NewStyle().
				Foreground(cfg.Theme.ColorWhite).
				Bold(true).
				Transform(strings.ToUpper).
				MarginTop(1),

			label: cfg.Theme.StyleDim,

			value: cfg.Theme.StyleValue,
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

	devModeStr := s.value("off")
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
			s.styles.category.Render("Deploy ("+s.pending(s.nextVersion)+")"),
			s.stepLine("Launching:", s.deployLaunching),
			s.stepLine("Healthcheck:", s.deployHealthcheck),
		)
	}

	var switchStrategy string
	var shutdown string
	if s.mode == domain.ModeUpdate {
		switchStrategy = fmt.Sprintf(
			"%s\n%s",
			s.styles.category.Render("Switch strategy"),
			s.stepLine("Switching - Nginx:", s.switchingNginx),
		)

		shutdown = fmt.Sprintf(
			"%s\n%s",
			s.styles.category.Render("Shutdown ("+s.pending(s.currentVersion)+")"),
			s.stepLine("Stopping the old version:", s.shutdownStopping),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		s.label("DevMode:")+devModeStr,
		s.label("Project:")+s.value(s.projectName),
		s.label("Mode:")+s.value(s.mode.String()),

		s.styles.category.Render("Requirements"),
		s.label("curl:")+s.value(s.requirementsCurlVersion),
		s.label("docker:")+s.value(s.requirementsDockerVersion),
		s.label("docker compose:")+s.value(s.requirementsDockerComposeVersion),
		s.label("nginx:")+s.value(s.requirementsNginxVersion),

		s.styles.category.Render("Directory"),
		s.value(splitString(s.currentDir, s.dirMaxWidth)),

		s.styles.category.Render("Files"),
		s.label("compose.blue.yaml:")+s.value(s.filenameComposeBlue),
		s.label("compose.green.yaml:")+s.value(s.filenameComposeGreen),
		s.label("nginx.conf (symlink):")+s.value(s.filenameNginxConf),

		s.styles.category.Render("Deploy strategy"),
		s.label("Version:")+s.value(s.currentVersion)+" >> "+s.value(s.nextVersion),
		s.label("Strategy:")+s.value(s.currentStrategy.String())+" >> "+s.value(s.nextStrategy.String()),

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

func (s *Summary) GetDir() string {
	return s.currentDir
}

func (s *Summary) GetFilenameNginxConf() string {
	return s.filenameNginxConf
}

func (s *Summary) GetCurrentVersion() string {
	return s.currentVersion
}

func (s *Summary) GetNextVersion() string {
	return s.nextVersion
}

func (s *Summary) GetNextStrategy() domain.Strategy {
	return s.nextStrategy
}

func (s *Summary) GetPorts() []CurrNextPort {
	return s.currNextPorts
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

func (s *Summary) UpdatePorts(ports []CurrNextPort) {
	s.currNextPorts = ports
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

// SetTotalSteps fixes the size of the pipeline, it is known as soon as the mode
// is chosen.
func (s *Summary) SetTotalSteps(total int) {
	s.totalSteps = total
}

// NextStep counts the step the pipeline has just entered. Before the mode is
// chosen the total is unknown and nothing is counted, so the walk from the
// current directory to the requirements check stays out of the counter.
func (s *Summary) NextStep() {
	if s.totalSteps > 0 && s.step < s.totalSteps {
		s.step++
	}
}

// GetStep returns the step in progress and the size of the pipeline.
func (s *Summary) GetStep() (step, total int) {
	return s.step, s.totalSteps
}

// SetPendingFrame stores the current frame of the spinner, the summary has no
// animation of its own and borrows the frames of the running step.
func (s *Summary) SetPendingFrame(frame string) {
	s.pendingFrame = frame
}

// totalSteps counts the steps of the pipeline. Both modes share the deploy
// itself, update additionally switches the strategy and stops the old deploy,
// which is also the only way into the VictoriaMetrics step.
func totalSteps(cfg domain.AppConfig, mode domain.Mode) int {
	const (
		sharedSteps = 5 // requirements, ports, images, deploy, healthcheck
		updateSteps = 2 // switching strategy, stopping the current deploy
	)

	steps := sharedSteps
	if mode == domain.ModeUpdate {
		steps += updateSteps
		if cfg.VictoriaMetrics.Enabled {
			steps++
		}
	}

	return steps
}

// label renders a label of the panel, dimmed and padded, so the values keep
// their column.
func (s *Summary) label(text string) string {
	return s.styles.label.Width(labelWidth).Render(text)
}

// value renders a value of the panel, the ones not known yet show the frame of
// the running step instead of the placeholder.
func (s *Summary) value(text string) string {
	return s.styles.value.Render(s.pending(text))
}

// pending swaps the placeholder for the current frame of the spinner.
func (s *Summary) pending(text string) string {
	if text != pendingValue {
		return text
	}

	return s.pendingFrame
}

// stepLine renders a step of the pipeline. An untouched step is dimmed, so it
// cannot be mistaken for the one running right now.
func (s *Summary) stepLine(text string, done *bool) string {
	line := s.label(text) + s.boolToIcon(done)
	if done == nil {
		return s.theme.StyleDim.Render(line)
	}

	return line
}

func (s *Summary) boolToIcon(b *bool) string {
	if b == nil {
		return s.pending(pendingValue)
	}

	if *b {
		return s.theme.StyleGreen.Render("✅")
	}

	return s.theme.StyleRed.Render("❌")
}

func (s *Summary) portsView() string {
	if len(s.currNextPorts) == 0 {
		return ""
	}

	lines := []string{
		s.styles.category.Render("Ports"),
	}

	// the column follows the longest location, so the ports of every app stay
	// aligned with each other.
	locationWidth := 0
	for _, p := range s.currNextPorts {
		locationWidth = max(locationWidth, lipgloss.Width(p.Location)+len(": "))
	}

	for _, p := range s.currNextPorts {
		loc := s.styles.label.Width(locationWidth).Render(p.Location + ":")
		cp := s.value(p.CurrentPort)
		np := s.value(p.NextPort)

		lines = append(lines,
			loc+cp+" >> "+np,
		)
	}

	return strings.Join(lines, "\n")
}

func splitString(s string, width int) string {
	var result []string

	runes := []rune(s)
	for len(runes) > width {
		result = append(result, string(runes[:width]))
		runes = runes[width:]
	}

	if len(runes) > 0 {
		result = append(result, string(runes))
	}

	return strings.Join(result, "\n")
}

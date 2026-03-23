package model

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/goccy/go-yaml"

	"debafr/internal/domain"
)

type Ports struct {
	dic     DIC
	summary *Summary
	theme   *domain.Theme
	spinner spinner.Model

	ports portsInfo
	err   error
}

type portsInfo struct {
	blue  ports
	green ports
}

type ports struct {
	backend  string
	frontend string
}

func NewPorts(dic DIC) *Ports {
	return &Ports{
		dic:     dic,
		summary: dic.GetSummary(),
		theme:   dic.GetTheme(),
		spinner: domain.NewSpinner(dic.GetTheme().StyleGreen),
	}
}

func (c *Ports) Init() tea.Cmd {
	return tea.Batch(c.processing, c.spinner.Tick)
}

func (c *Ports) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusDone:
		if c.summary.GetMode() == domain.ModeInstall {
			c.summary.UpdateCurrentPorts("---", "---")
			c.summary.UpdateNextPorts(c.ports.blue.backend, c.ports.blue.frontend)

			return c, func() tea.Msg {
				return NextCmdMsg{
					NextCmd: NewNextDeploy(c.dic),
				}
			}
		}

		if c.summary.GetCurrentStrategy() == domain.StrategyBlue {
			c.summary.UpdateCurrentPorts(c.ports.blue.backend, c.ports.blue.frontend)
			c.summary.UpdateNextPorts(c.ports.green.backend, c.ports.green.frontend)

			return c, func() tea.Msg {
				return NextCmdMsg{
					NextCmd: NewNextDeploy(c.dic),
				}
			}
		}

		c.summary.UpdateCurrentPorts(c.ports.green.backend, c.ports.green.frontend)
		c.summary.UpdateNextPorts(c.ports.blue.backend, c.ports.blue.frontend)

		return c, func() tea.Msg {
			return NextCmdMsg{
				NextCmd: NewNextDeploy(c.dic),
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

func (c *Ports) View() string {
	prefix := "Searching ports from docker compose files"

	if c.err != nil {
		return prefix + "... " + c.theme.StyleRed.Render(
			fmt.Sprintf("something went wrong: %s", c.err),
		)
	}

	return fmt.Sprintf("%s %s", prefix, c.spinner.View())
}

func (c *Ports) processing() tea.Msg {
	root, err := c.summary.GetRoot()
	if err != nil {
		return domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    err,
			Output: "failed to open root",
		}
	}
	defer root.Close()

	projectName := c.summary.GetProjectName()

	c.ports.blue.backend, c.ports.blue.frontend, err = findPorts(root, c.summary.GetFilenameComposeBlue(), projectName)
	if err != nil {
		return StatusError{
			fmt.Errorf("failed to find blue ports: %v", err),
		}
	}

	c.ports.green.backend, c.ports.green.frontend, err = findPorts(root, c.summary.GetFilenameComposeGreen(), projectName)
	if err != nil {
		return StatusError{
			fmt.Errorf("failed to find green ports: %v", err),
		}
	}

	return StatusDone{true}
}

func findPorts(root *os.Root, pathFile string, projectName string) (backend string, frontend string, err error) {
	data, err := root.ReadFile(pathFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %v", err)
	}

	var compose DockerCompose
	err = yaml.Unmarshal(data, &compose)
	if err != nil {
		return "", "", fmt.Errorf("yaml unmarshal: %v", err)
	}

	for _, service := range compose.Services {
		if len(service.Labels) == 0 {
			continue
		}
		if len(service.Ports) == 0 {
			continue
		}

		if slices.Contains(service.Labels, "app.project.name="+projectName) &&
			slices.Contains(service.Labels, "app.service.type="+domain.ContainerAppServiceTypeBackend.String()) {
			backend = strings.Split(service.Ports[0], ":")[0]
			continue
		}

		if slices.Contains(service.Labels, "app.project.name="+projectName) &&
			slices.Contains(service.Labels, "app.service.type="+domain.ContainerAppServiceTypeFrontend.String()) {
			frontend = strings.Split(service.Ports[0], ":")[0]
			continue
		}
	}

	return backend, frontend, nil
}

type DockerCompose struct {
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Labels []string `yaml:"labels"`
	Ports  []string `yaml:"ports"`
}

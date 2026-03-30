package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"debafr/internal/domain"
)

var (
	reLocationBlock = regexp.MustCompile(`(?m)^\s*location\s+(\S+)\s*\{`)
	reProxyPass     = regexp.MustCompile(`(?m)^(\s*)(#?\s*)proxy_pass\s+http://127\.0\.0\.1:(\d+)\s*;\s*(#\s*(blue|green))?`)
)

type Ports struct {
	dic     DIC
	summary *Summary
	theme   *domain.Theme
	spinner spinner.Model

	ports []LocationPort

	err error
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
			for i := range c.ports {
				c.ports[i].CurrentPort = "---"
			}

			c.summary.UpdatePorts(c.ports)

			return c, func() tea.Msg {
				return NextCmdMsg{
					NextCmd: NewNextDeploy(c.dic),
				}
			}
		}

		c.summary.UpdatePorts(c.ports)

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
	prefix := "Searching ports from nginx config"

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

	content, err := root.ReadFile(c.summary.GetFilenameNginxConf())
	if err != nil {
		return StatusError{
			fmt.Errorf("failed to read nginx config: %v", err),
		}
	}

	c.ports, err = parseNginxLocationPorts(string(content))
	if err != nil {
		return StatusError{
			fmt.Errorf("failed to parse nginx config: %v", err),
		}
	}

	if len(c.ports) == 0 {
		return StatusError{
			fmt.Errorf("no location ports found in nginx config"),
		}
	}

	return StatusDone{true}
}

func parseNginxLocationPorts(content string) ([]LocationPort, error) {
	var result []LocationPort

	locMatches := reLocationBlock.FindAllStringSubmatchIndex(content, -1)

	for i, locMatch := range locMatches {
		locPath := content[locMatch[2]:locMatch[3]]

		blockStart := locMatch[1]

		var blockEnd int
		if i+1 < len(locMatches) {
			blockEnd = locMatches[i+1][0]
		} else {
			blockEnd = len(content)
		}

		block := content[blockStart:blockEnd]

		depth := 0
		cutAt := -1
		for pos, ch := range block {
			if ch == '{' {
				depth++
			} else if ch == '}' {
				depth--
				if depth == 0 {
					cutAt = blockStart + pos + 1
					break
				}
			}
		}
		if cutAt != -1 {
			block = content[blockStart:cutAt]
		}

		proxyMatches := reProxyPass.FindAllStringSubmatch(block, -1)

		var activePort, commentedPort string

		for _, m := range proxyMatches {
			isCommented := strings.Contains(m[0], "#proxy_pass") || strings.HasPrefix(strings.TrimSpace(m[0]), "#")
			port := m[3]

			if isCommented {
				commentedPort = port
			} else {
				activePort = port
			}
		}

		if activePort != "" && commentedPort != "" {
			result = append(result, LocationPort{
				Location:    locPath,
				CurrentPort: activePort,
				NextPort:    commentedPort,
			})
		}
	}

	return result, nil
}

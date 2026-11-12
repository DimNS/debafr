package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/docker/api/types/container"

	"debafr/internal/domain"
)

func NewExecHealthcheck(dic DIC) *Exec {
	summary := dic.GetSummary()
	dockerService := dic.GetDockerService()
	cfg := dic.GetAppConfig()

	return NewExec(dic, domain.ExecConfig{
		Name: "Healthcheck",

		StartFunc: func() domain.ExecResult {
			projectName := summary.GetProjectName()
			nextVersion := summary.GetNextVersion()
			containers, err := dockerService.GetContainers(projectName, nextVersion)
			if err != nil {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: "failed to get containers",
				}
			}

			return healthcheck(dockerService, projectName, containers, cfg.Healthcheck.MaxRetries, cfg.Healthcheck.RetryDelay)
		},

		SuccessFunc: func() {
			summary.UpdateDeployHealthcheck(true)
		},

		ErrorFunc: func() {
			summary.UpdateDeployHealthcheck(false)
		},

		AutoAdvance: true,

		NextCmd: func() tea.Model {
			if summary.GetMode() == domain.ModeUpdate {
				if len(cfg.Replacer) > 0 {
					return NewExecReplacer(dic)
				}

				return NewExecSwitchingStrategy(dic)
			}

			return NewComplete(dic)
		}(),
	})
}

// containerStateGetter is the part of the docker service the healthcheck needs.
type containerStateGetter interface {
	GetState(containerID string) (domain.ContainerState, error)
}

func healthcheck(
	dockerService containerStateGetter,
	projectName string,
	containers []container.Summary,
	maxRetries int,
	retryDelay time.Duration,
) domain.ExecResult {
	var lastResult domain.ExecResult

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ready := true
		var output strings.Builder

		for _, ctr := range containers {
			state, err := dockerService.GetState(ctr.ID)
			if err != nil {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: fmt.Sprintf("failed to get %s container state", containerService(ctr, projectName)),
				}
			}

			if !containerReady(state) {
				ready = false
			}

			fmt.Fprintf(&output, "### %s State: %s (%s)\n", containerService(ctr, projectName), state.Status, state.Health)
		}

		if ready {
			return domain.ExecResult{
				Status: domain.ExecResultStatusSuccess,
				Output: strings.TrimSuffix(output.String(), "\n"),
			}
		}

		lastResult = domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    errors.New("not ready"),
			Output: fmt.Sprintf("%s(attempt %d/%d)", output.String(), attempt, maxRetries),
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return lastResult
}

// containerReady reports whether the container is up. A container without a
// healthcheck can only be checked by its running state.
func containerReady(state domain.ContainerState) bool {
	if state.Status != domain.ContainerStateStatusRunning {
		return false
	}

	return state.Health == domain.ContainerStateHealthHealthy ||
		state.Health == domain.ContainerStateHealthNoHealthcheck
}

func containerService(ctr container.Summary, projectName string) domain.ContainerAppServiceType {
	service, _ := domain.ContainerServiceType(ctr.Labels, projectName)

	return service
}

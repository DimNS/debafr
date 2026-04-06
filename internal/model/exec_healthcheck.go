package model

import (
	"errors"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/docker/api/types/container"

	"debafr/internal/domain"
	"debafr/internal/provider/docker"
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
			frontend, backend, err := dockerService.GetContainers(projectName, nextVersion)
			if err != nil {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: "failed to get containers",
				}
			}

			return healthcheck(dockerService, frontend, backend, cfg.Healthcheck.MaxRetries, cfg.Healthcheck.RetryDelay)
		},

		SuccessFunc: func() {
			summary.UpdateDeployHealthcheck(true)
		},

		ErrorFunc: func() {
			summary.UpdateDeployHealthcheck(false)
		},

		NextCmd: func() tea.Model {
			if summary.GetMode() == domain.ModeUpdate {
				return NewExecSwitchingStrategy(dic)
			}

			return NewComplete(dic)
		}(),
	})
}

func healthcheck(
	dockerService *docker.Docker,
	frontend, backend *container.Summary,
	maxRetries int,
	retryDelay time.Duration,
) domain.ExecResult {
	var lastResult domain.ExecResult

	for attempt := 1; attempt <= maxRetries; attempt++ {
		frontendState, err := dockerService.GetState(frontend.ID)
		if err != nil {
			return domain.ExecResult{
				Status: domain.ExecResultStatusError,
				Err:    err,
				Output: "failed to get frontend container state",
			}
		}

		backendState, err := dockerService.GetState(backend.ID)
		if err != nil {
			return domain.ExecResult{
				Status: domain.ExecResultStatusError,
				Err:    err,
				Output: "failed to get backend container state",
			}
		}

		if frontendState.Status == domain.ContainerStateStatusRunning &&
			frontendState.Health == domain.ContainerStateHealthHealthy &&
			backendState.Status == domain.ContainerStateStatusRunning &&
			backendState.Health == domain.ContainerStateHealthHealthy {
			return domain.ExecResult{
				Status: domain.ExecResultStatusSuccess,
				Err:    nil,
				Output: fmt.Sprintf(
					"### Frontend State: %s (%s)\n### Backend State: %s (%s)",
					frontendState.Status,
					frontendState.Health,
					backendState.Status,
					backendState.Health,
				),
			}
		}

		lastResult = domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    errors.New("not ready"),
			Output: fmt.Sprintf(
				"### Frontend State: %s (%s)\n### Backend State: %s (%s)\n(attempt %d/%d)",
				frontendState.Status,
				frontendState.Health,
				backendState.Status,
				backendState.Health,
				attempt,
				maxRetries,
			),
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return lastResult
}

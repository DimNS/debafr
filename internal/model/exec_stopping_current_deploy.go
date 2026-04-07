package model

import (
	"debafr/internal/domain"

	tea "github.com/charmbracelet/bubbletea"
)

func NewExecStoppingCurrentDeploy(dic DIC) *Exec {
	summary := dic.GetSummary()
	dockerService := dic.GetDockerService()
	cfg := dic.GetAppConfig()

	return NewExec(dic, domain.ExecConfig{
		Name: "Stopping current deploy",

		StartFunc: func() domain.ExecResult {
			projectName := summary.GetProjectName()
			currentVersion := summary.GetCurrentVersion()
			frontend, backend, err := dockerService.GetContainers(projectName, currentVersion)
			if err != nil {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: "failed to get containers",
				}
			}

			if err := dockerService.ContainerStop(frontend.ID); err != nil {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: "failed to stop frontend container",
				}
			}

			if err := dockerService.ContainerStop(backend.ID); err != nil {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: "failed to stop backend container",
				}
			}

			return domain.ExecResult{
				Status: domain.ExecResultStatusSuccess,
				Output: "Current deploy stopped successfully",
			}
		},

		SuccessFunc: func() {
			summary.UpdateShutdownStopping(true)
		},

		ErrorFunc: func() {
			summary.UpdateShutdownStopping(false)
		},

		NextCmd: func() tea.Model {
			if cfg.VictoriaMetrics.Enabled {
				return NewExecWriteVictoriaMetricsTargets(dic)
			}

			return NewComplete(dic)
		}(),
	})
}

package model

import (
	"fmt"

	"debafr/internal/domain"
)

func NewExecStoppingCurrentDeploy(dic DIC) *Exec {
	summary := dic.GetSummary()
	dockerService := dic.GetDockerService()

	return NewExec(dic, domain.ExecConfig{
		Name: "Stopping current deploy",

		StartFunc: func() domain.ExecResult {
			projectName := summary.GetProjectName()
			currentVersion := summary.GetCurrentVersion()
			containers, err := dockerService.GetContainers(projectName, currentVersion)
			if err != nil {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: "failed to get containers",
				}
			}

			for _, ctr := range containers {
				service, _ := domain.ContainerServiceType(ctr.Labels, projectName)

				if err := dockerService.ContainerStop(ctr.ID); err != nil {
					return domain.ExecResult{
						Status: domain.ExecResultStatusError,
						Err:    err,
						Output: fmt.Sprintf("failed to stop %s container", service),
					}
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

		AutoAdvance: true,

		NextCmd: NewComplete(dic),
	})
}

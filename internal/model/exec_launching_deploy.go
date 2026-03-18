package model

import (
	"fmt"
	"os"
	"os/exec"

	"debafr/internal/domain"
)

func NewExecLaunchingDeploy(dic DIC) *Exec {
	summary := dic.GetSummary()

	return NewExec(dic, domain.ExecConfig{
		Name: "Deploying",

		StartFunc: func() domain.ExecResult {
			command := exec.Command( //#nosec G204 -- This is a false positive
				PathDocker,
				"compose",
				"-f",
				fmt.Sprintf("./compose.%s.yaml", summary.GetNextStrategy()),
				"up",
				"-d",
			)
			command.Env = append(os.Environ(), "APP_VERSION="+summary.GetNextVersion())

			if dic.GetDevMode() {
				command = exec.Command(
					PathDocker,
					"run",
					"-d",
					"--name",
					TestContainerName,
					"-p",
					"8585:80",
					"nginx:alpine",
				)
			}

			output, err := command.CombinedOutput()
			if err != nil {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: string(output),
				}
			}

			return domain.ExecResult{
				Status: domain.ExecResultStatusSuccess,
				Output: string(output),
			}
		},

		SuccessFunc: func() {
			summary.UpdateDeployLaunching(true)
		},

		ErrorFunc: func() {
			summary.UpdateDeployLaunching(false)
		},

		NextCmd: NewExecHealthcheck(dic),
	})
}

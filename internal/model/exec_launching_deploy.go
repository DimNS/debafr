package model

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"debafr/internal/domain"
)

func NewExecLaunchingDeploy(dic DIC) *Exec {
	summary := dic.GetSummary()
	cfg := dic.GetAppConfig()
	dockerService := dic.GetDockerService()

	return NewExec(dic, domain.ExecConfig{
		Name: "Deploying",

		StartFunc: func() domain.ExecResult {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeouts.Default)
			defer cancel()

			if dic.GetDevMode() {
				id, err := dockerService.RunContainer(ctx, TestContainerName, "nginx:alpine", "8585", "80")
				if err != nil {
					return domain.ExecResult{
						Status: domain.ExecResultStatusError,
						Err:    err,
					}
				}

				return domain.ExecResult{
					Status: domain.ExecResultStatusSuccess,
					Output: "Container started: " + id[:12],
				}
			}

			var f string
			switch summary.GetNextStrategy() {
			case domain.StrategyBlue:
				f = cfg.Files.ComposeBlue
			case domain.StrategyGreen:
				f = cfg.Files.ComposeGreen
			default:
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    fmt.Errorf("unknown strategy: %s", summary.GetNextStrategy()),
				}
			}

			// Compose: docker SDK не умеет compose, остаётся на CLI
			command := exec.CommandContext(
				ctx,
				cfg.BinPaths.Docker,
				"compose",
				"-f",
				f,
				"up",
				"-d",
			)
			command.Env = append(os.Environ(), "APP_VERSION="+summary.GetNextVersion())

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

		AutoAdvance: true,

		NextCmd: NewExecHealthcheck(dic),
	})
}

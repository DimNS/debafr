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

	return NewExec(dic, domain.ExecConfig{
		Name: "Deploying",

		StartFunc: func() domain.ExecResult {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeouts.Default)
			defer cancel()

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

			if dic.GetDevMode() {
				command = exec.CommandContext(
					ctx,
					cfg.BinPaths.Docker,
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

package model

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"debafr/internal/domain"
)

func NewExecPullImage(dic DIC) *Exec {
	summary := dic.GetSummary()
	cfg := dic.GetAppConfig()
	dockerService := dic.GetDockerService()

	return NewExec(dic, domain.ExecConfig{
		Name: "Pulling images",

		StartFunc: func() domain.ExecResult {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeouts.Default)
			defer cancel()

			if cfg.DockerLogin.Enabled {
				loginCmd := exec.CommandContext(
					ctx,
					cfg.BinPaths.Docker,
					"login",
					cfg.DockerLogin.Registry,
					"-u", cfg.DockerLogin.Username,
					"-p", cfg.DockerLogin.Password,
				)
				if err := loginCmd.Run(); err != nil {
					return domain.ExecResult{
						Status: domain.ExecResultStatusError,
						Err:    fmt.Errorf("docker login failed: %w", err),
					}
				}
			}

			v := summary.GetNextVersion()

			output := make([]string, 0, len(cfg.Images)+1)
			output = append(output, "Images pulled successfully:")
			for _, img := range cfg.Images {
				image := img + ":" + v

				if err := dockerService.ImagePull(image); err != nil {
					return domain.ExecResult{
						Status: domain.ExecResultStatusError,
						Err:    fmt.Errorf("docker pull failed: %w", err),
					}
				}

				output = append(output, image)
			}

			return domain.ExecResult{
				Status: domain.ExecResultStatusSuccess,
				Output: strings.Join(output, "\n"),
			}
		},

		SuccessFunc: func() {
			// do nothing
		},

		ErrorFunc: func() {
			// do nothing
		},

		NextCmd: NewExecLaunchingDeploy(dic),
	})
}

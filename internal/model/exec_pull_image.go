package model

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"debafr/internal/domain"

	"github.com/docker/docker/api/types/image"
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

			pullOpts := image.PullOptions{}
			if cfg.DockerLogin.Enabled {
				authConfig := map[string]string{
					"username": cfg.DockerLogin.Username,
					"password": cfg.DockerLogin.Password,
				}
				jsonBytes, err := json.Marshal(authConfig)
				if err != nil {
					return domain.ExecResult{
						Status: domain.ExecResultStatusError,
						Err:    fmt.Errorf("json marshal failed: %v", err),
					}
				}

				pullOpts.RegistryAuth = base64.URLEncoding.EncodeToString(jsonBytes)
			}

			v := summary.GetNextVersion()

			output := make([]string, 0, len(cfg.Images)+1)
			output = append(output, "Images pulled successfully:")
			for _, img := range cfg.Images {
				image := img + ":" + v

				if err := dockerService.ImagePull(ctx, image, pullOpts); err != nil {
					return domain.ExecResult{
						Status: domain.ExecResultStatusError,
						Err:    fmt.Errorf("docker pull failed: %v", err),
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

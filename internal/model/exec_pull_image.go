package model

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/docker/docker/api/types/image"

	"debafr/internal/domain"
	"debafr/internal/provider/docker"
)

// percentColumn is the width of the percentage next to the bar, the bar itself
// has to fit into what is left of the panel.
const percentColumn = 5

func NewExecPullImage(dic DIC) *Exec {
	summary := dic.GetSummary()
	cfg := dic.GetAppConfig()
	dockerService := dic.GetDockerService()

	// the docker goroutine swaps the snapshot whole and the view reads it as
	// it is, that is the only state the two of them share.
	var downloaded atomic.Pointer[docker.PullProgress]

	// the bytes of the images pulled before the current one, the bar covers
	// every image of the step and not only the one running right now.
	var done, total int64

	var last docker.PullProgress

	report := func(p docker.PullProgress) {
		last = p
		downloaded.Store(&docker.PullProgress{
			Current: done + p.Current,
			Total:   total + p.Total,
		})
	}

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

				if err := dockerService.ImagePull(ctx, image, pullOpts, report); err != nil {
					return domain.ExecResult{
						Status: domain.ExecResultStatusError,
						Err:    fmt.Errorf("docker pull failed: %v", err),
					}
				}

				done += last.Current
				total += last.Total

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

		AutoAdvance: true,

		LiveFunc: func() string {
			var ratio float64
			if p := downloaded.Load(); p != nil {
				ratio = pullRatio(*p)
			}

			width := dic.GetPhysicalWidth() - dic.GetSummaryWidth() - compensationWidth - percentColumn

			return barView(dic.GetTheme(), ratio, width)
		},

		NextCmd: NewExecLaunchingDeploy(dic),
	})
}

// pullRatio is the share of the download that is done, 0 while the size of the
// first layer is still unknown.
func pullRatio(p docker.PullProgress) float64 {
	if p.Total <= 0 {
		return 0
	}
	if p.Current >= p.Total {
		return 1
	}

	return float64(p.Current) / float64(p.Total)
}

// barView renders the share of the download that is done.
func barView(theme *domain.Theme, ratio float64, width int) string {
	width = max(width, 0)
	filled := min(max(int(ratio*float64(width)), 0), width)

	return theme.StyleGreen.Render(strings.Repeat("█", filled)) +
		theme.StyleDim.Render(strings.Repeat("░", width-filled)) +
		" " + fmt.Sprintf("%3.0f%%", ratio*100) //nolint:mnd // percent
}

package model

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"debafr/internal/domain"
)

func NewExecSwitchingStrategy(dic DIC) *Exec {
	summary := dic.GetSummary()
	cfg := dic.GetAppConfig()

	return NewExec(dic, domain.ExecConfig{
		Name: "Switching strategy",

		StartFunc: func() domain.ExecResult {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeouts.Default)
			defer cancel()

			return runSwitching(ctx, cfg.Nginx)
		},

		SuccessFunc: func() {
			summary.UpdateSwitchingNginx(true)
		},

		ErrorFunc: func() {
			summary.UpdateSwitchingNginx(false)
		},

		AutoAdvance: true,

		NextCmd: NewExecStoppingCurrentDeploy(dic),
	})
}

// runSwitching runs the configured nginx commands in order. A command with an
// empty argument list is skipped, so nginx is optional.
func runSwitching(ctx context.Context, cfg domain.AppConfigNginx) domain.ExecResult {
	if !cfg.Enabled {
		return domain.ExecResult{
			Status: domain.ExecResultStatusSuccess,
			Output: "Nginx is disabled, switching skipped",
		}
	}

	var output strings.Builder

	commands := []struct {
		name string
		args []string
	}{
		{"Test config", cfg.Test},
		{"Reload", cfg.Reload},
	}

	for _, command := range commands {
		if len(command.args) == 0 {
			continue
		}

		out, err := exec.CommandContext(ctx, command.args[0], command.args[1:]...).CombinedOutput()
		if len(out) > 0 {
			fmt.Fprintf(&output, "### %s:\n%s\n", command.name, out)
		}
		if err != nil {
			return domain.ExecResult{
				Status: domain.ExecResultStatusError,
				Err:    fmt.Errorf("%s: %w", strings.ToLower(command.name), err),
				Output: output.String(),
			}
		}
	}

	if output.Len() == 0 {
		output.WriteString("No nginx commands configured")
	}

	return domain.ExecResult{
		Status: domain.ExecResultStatusSuccess,
		Output: output.String(),
	}
}

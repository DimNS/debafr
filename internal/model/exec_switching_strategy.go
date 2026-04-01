package model

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"debafr/internal/domain"
)

const fileMode = 0644

type switchConfig struct {
	proxyPass string
	filePath  string
	ports     []CurrNextPort
	cmdTest   *exec.Cmd
	cmdReload *exec.Cmd
}

func NewExecSwitchingStrategy(dic DIC) *Exec {
	summary := dic.GetSummary()
	cfg := dic.GetAppConfig()

	return NewExec(dic, domain.ExecConfig{
		Name: "Switching strategy",

		StartFunc: func() domain.ExecResult {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeouts.Default)
			defer cancel()

			cmdTest := exec.CommandContext(ctx, cfg.BinPaths.Nginx, "-t")
			cmdReload := exec.CommandContext(ctx, cfg.BinPaths.Nginx, "-s", "reload")
			if dic.GetDevMode() {
				cmdTest = exec.CommandContext(ctx, cfg.BinPaths.Docker, "exec", TestContainerName, cfg.BinPaths.Nginx, "-t")
				cmdReload = exec.CommandContext(ctx, cfg.BinPaths.Docker, "exec", TestContainerName, cfg.BinPaths.Nginx, "-s", "reload")
			}
			resNginx := switchNginx(switchConfig{
				proxyPass: cfg.ProxyPassPrefix,
				filePath:  path.Join(summary.GetDir(), summary.GetFilenameNginxConf()),
				ports:     summary.GetPorts(),
				cmdTest:   cmdTest,
				cmdReload: cmdReload,
			})
			if resNginx.Status == domain.ExecResultStatusError {
				return resNginx
			}

			return domain.ExecResult{
				Status: domain.ExecResultStatusSuccess,
				Output: fmt.Sprintf( //nolint:perfsprint // ignore
					"### Switching - Nginx:\n%s",
					resNginx.Output,
				),
			}
		},

		SuccessFunc: func() {
			summary.UpdateSwitchingNginx(true)
		},

		ErrorFunc: func() {
			summary.UpdateSwitchingNginx(false)
		},

		NextCmd: NewExecStoppingCurrentDeploy(dic),
	})
}

func switchNginx(cfg switchConfig) domain.ExecResult {
	content, err := os.ReadFile(cfg.filePath)
	if err != nil {
		return domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    fmt.Errorf("switchNginx: failed to read file: %v", err),
		}
	}

	fileContent := string(content)

	for _, p := range cfg.ports {
		if p.CurrentPort != EmptyValue {
			curr := cfg.proxyPass + p.CurrentPort
			if !strings.Contains(fileContent, "#"+curr) {
				fileContent = strings.Replace(fileContent, curr, "#"+curr, 1)
			}
		}

		next := cfg.proxyPass + p.NextPort
		if strings.Contains(fileContent, "#"+next) {
			fileContent = strings.Replace(fileContent, "#"+next, next, 1)
		}
	}

	err = os.WriteFile(cfg.filePath, []byte(fileContent), fileMode) //#nosec G306,G703 -- This is a false positive
	if err != nil {
		return domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    fmt.Errorf("switchNginx: failed to write file: %v", err),
		}
	}

	outputTest, err := cfg.cmdTest.CombinedOutput()
	if err != nil {
		return domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    err,
			Output: string(outputTest),
		}
	}

	outputReload, err := cfg.cmdReload.CombinedOutput()
	if err != nil {
		return domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    err,
			Output: string(outputReload),
		}
	}

	return domain.ExecResult{
		Status: domain.ExecResultStatusSuccess,
		Output: fmt.Sprintf(
			"### Test config:\n%s\n### Reload Nginx:\n%s",
			string(outputTest),
			string(outputReload),
		),
	}
}

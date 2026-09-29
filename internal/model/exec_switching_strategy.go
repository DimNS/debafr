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
	test      func() ([]byte, error)
	reload    func() ([]byte, error)
}

func NewExecSwitchingStrategy(dic DIC) *Exec {
	summary := dic.GetSummary()
	cfg := dic.GetAppConfig()

	return NewExec(dic, domain.ExecConfig{
		Name: "Switching strategy",

		StartFunc: func() domain.ExecResult {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeouts.Default)
			defer cancel()

			cmdTest := func() ([]byte, error) {
				return exec.CommandContext(ctx, cfg.BinPaths.Nginx, "-t").CombinedOutput()
			}
			cmdReload := func() ([]byte, error) {
				return exec.CommandContext(ctx, cfg.BinPaths.Nginx, "-s", "reload").CombinedOutput()
			}
			if dic.GetDevMode() {
				dockerService := dic.GetDockerService()
				cmdTest = func() ([]byte, error) {
					return dockerService.Exec(ctx, TestContainerName, []string{"nginx", "-t"})
				}
				cmdReload = func() ([]byte, error) {
					return dockerService.Exec(ctx, TestContainerName, []string{"nginx", "-s", "reload"})
				}
			}
			resNginx := switchNginx(switchConfig{
				proxyPass: cfg.ProxyPassPrefix,
				filePath:  path.Join(summary.GetDir(), summary.GetFilenameNginxConf()),
				ports:     summary.GetPorts(),
				test:      cmdTest,
				reload:    cmdReload,
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
				fileContent = strings.ReplaceAll(fileContent, curr, "#"+curr)
			}
		}

		next := cfg.proxyPass + p.NextPort
		if strings.Contains(fileContent, "#"+next) {
			fileContent = strings.ReplaceAll(fileContent, "#"+next, next)
		}
	}

	err = os.WriteFile(cfg.filePath, []byte(fileContent), fileMode) //#nosec G306,G703 -- This is a false positive
	if err != nil {
		return domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    fmt.Errorf("switchNginx: failed to write file: %v", err),
		}
	}

	outputTest, err := cfg.test()
	if err != nil {
		return domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    err,
			Output: string(outputTest),
		}
	}

	outputReload, err := cfg.reload()
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

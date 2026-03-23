package model

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"debafr/internal/domain"
)

const (
	fileMode = 0644
)

type switchConfig struct {
	root             *os.Root
	filePath         string
	currPortBackend  string
	currPortFrontend string
	nextPortBackend  string
	nextPortFrontend string
	cmdTest          *exec.Cmd
	cmdReload        *exec.Cmd
}

func NewExecSwitchingStrategy(dic DIC) *Exec {
	summary := dic.GetSummary()

	return NewExec(dic, domain.ExecConfig{
		Name: "Switching strategy",

		StartFunc: func() domain.ExecResult {
			ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
			defer cancel()

			root, err := summary.GetRoot()
			if err != nil {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: "failed to open root",
				}
			}
			defer root.Close()

			devMode := dic.GetDevMode()
			currPortBackend, currPortFrontend := summary.GetCurrentPorts()
			nextPortBackend, nextPortFrontend := summary.GetNextPorts()

			cmdTest := exec.CommandContext(ctx, PathNginx, "-t")
			cmdReload := exec.CommandContext(ctx, PathNginx, "-s", "reload")
			if devMode {
				cmdTest = exec.CommandContext(ctx, PathDocker, "exec", TestContainerName, PathNginx, "-t")
				cmdReload = exec.CommandContext(ctx, PathDocker, "exec", TestContainerName, PathNginx, "-s", "reload")
			}
			resNginx := switchNginx(switchConfig{
				root:             root,
				filePath:         summary.GetFilenameNginxConf(),
				currPortBackend:  currPortBackend,
				currPortFrontend: currPortFrontend,
				nextPortBackend:  nextPortBackend,
				nextPortFrontend: nextPortFrontend,
				cmdTest:          cmdTest,
				cmdReload:        cmdReload,
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
	content, err := cfg.root.ReadFile(cfg.filePath)
	if err != nil {
		return domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    fmt.Errorf("switchNginx: failed to read file: %v", err),
		}
	}

	fileContent := string(content)

	const proxyPass = "proxy_pass http://127.0.0.1:" //#nosec G101 -- This is a false positive

	currBackendString := proxyPass + cfg.currPortBackend
	nextBackendString := proxyPass + cfg.nextPortBackend

	currFrontendString := proxyPass + cfg.currPortFrontend
	nextFrontendString := proxyPass + cfg.nextPortFrontend

	if !strings.Contains(fileContent, "#"+currBackendString) {
		fileContent = strings.Replace(fileContent, currBackendString, "#"+currBackendString, 1)
	}

	if strings.Contains(fileContent, "#"+nextBackendString) {
		fileContent = strings.Replace(fileContent, "#"+nextBackendString, nextBackendString, 1)
	}

	if !strings.Contains(fileContent, "#"+currFrontendString) {
		fileContent = strings.Replace(fileContent, currFrontendString, "#"+currFrontendString, 1)
	}

	if strings.Contains(fileContent, "#"+nextFrontendString) {
		fileContent = strings.Replace(fileContent, "#"+nextFrontendString, nextFrontendString, 1)
	}

	err = cfg.root.WriteFile(cfg.filePath, []byte(fileContent), fileMode) //#nosec G306 -- This is a false positive
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

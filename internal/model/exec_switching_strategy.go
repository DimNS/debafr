package model

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"debafr/internal/domain"
)

const (
	fileMode = 0644
)

func NewExecSwitchingStrategy(dic DIC) *Exec {
	summary := dic.GetSummary()

	return NewExec(dic, domain.ExecConfig{
		Name: "Switching strategy",

		StartFunc: func() domain.ExecResult {
			dir := summary.GetDir()
			devMode := dic.GetDevMode()
			currPortBackend, currPortFrontend := summary.GetCurrentPorts()
			nextPortBackend, nextPortFrontend := summary.GetNextPorts()

			pathNginx := path.Join(dir, summary.GetFilenameNginxConf())

			cmdTest := exec.Command(PathNginx, "-t")
			cmdReload := exec.Command(PathNginx, "-s", "reload")
			if devMode {
				cmdTest = exec.Command(PathDocker, "exec", TestContainerName, PathNginx, "-t")
				cmdReload = exec.Command(PathDocker, "exec", TestContainerName, PathNginx, "-s", "reload")
			}
			resNginx := switchNginx(pathNginx, currPortBackend, currPortFrontend, nextPortBackend, nextPortFrontend, cmdTest, cmdReload)
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

func switchNginx(
	filePath string,
	currPortBackend string,
	currPortFrontend string,
	nextPortBackend string,
	nextPortFrontend string,
	cmdTest *exec.Cmd,
	cmdReload *exec.Cmd,
) domain.ExecResult {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    fmt.Errorf("switchNginx: failed to read file: %v", err),
		}
	}

	fileContent := string(content)

	const proxyPass = "proxy_pass http://127.0.0.1:" //#nosec G101 -- This is a false positive

	currBackendString := proxyPass + currPortBackend
	nextBackendString := proxyPass + nextPortBackend

	currFrontendString := proxyPass + currPortFrontend
	nextFrontendString := proxyPass + nextPortFrontend

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

	err = os.WriteFile(filePath, []byte(fileContent), fileMode) //#nosec G306 -- This is a false positive
	if err != nil {
		return domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    fmt.Errorf("switchNginx: failed to write file: %v", err),
		}
	}

	outputTest, err := cmdTest.CombinedOutput()
	if err != nil {
		return domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    err,
			Output: string(outputTest),
		}
	}

	outputReload, err := cmdReload.CombinedOutput()
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

package model

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"

	"debafr/internal/domain"
)

func NewExecRequirements(dic DIC) *Exec {
	summary := dic.GetSummary()

	return NewExec(dic, domain.ExecConfig{
		Name: "Requirements",

		StartFunc: func() domain.ExecResult {
			var (
				err *domain.ExecResult

				curlVersion          string
				dockerVersion        string
				dockerComposeVersion string
				nginxVersion         string

				curlRegex          = regexp.MustCompile(`curl (?P<version>\d+(\.\d+)*(\.\w+)*)`)
				dockerRegex        = regexp.MustCompile(`Docker version (?P<version>\d+\.\d+\.\d+)`)
				dockerComposeRegex = regexp.MustCompile(`Docker Compose version v(?P<version>\d+(\.\d+)*(\.\w+)*)`)
				nginxRegex         = regexp.MustCompile(`nginx/(?P<version>[\d\.]+)`)
			)

			curlVersion, err = getVersion(exec.Command(PathCurl, "--version"), curlRegex)
			if err != nil {
				return *err
			}
			summary.UpdateRequirementsCurlVersion(curlVersion)

			dockerVersion, err = getVersion(exec.Command(PathDocker, "-v"), dockerRegex)
			if err != nil {
				return *err
			}
			summary.UpdateRequirementsDockerVersion(dockerVersion)

			dockerComposeVersion, err = getVersion(exec.Command(PathDocker, "compose", "version"), dockerComposeRegex)
			if err != nil {
				return *err
			}
			summary.UpdateRequirementsDockerComposeVersion(dockerComposeVersion)

			nginxVersion, err = getVersion(exec.Command(PathNginx, "-v"), nginxRegex)
			if err != nil {
				return *err
			}
			summary.UpdateRequirementsNginxVersion(nginxVersion)

			return domain.ExecResult{
				Status: domain.ExecResultStatusSuccess,
				Output: fmt.Sprintf(
					"curl: %s\ndocker: %s\ndocker compose: %s\nnginx: %s\n",
					curlVersion,
					dockerVersion,
					dockerComposeVersion,
					nginxVersion,
				),
			}
		},

		SuccessFunc: func() {
			// Нечего делать
		},
		ErrorFunc: func() {
			// Нечего делать
		},

		NextCmd: NewCurrentDeploy(dic),
	})
}

func getVersion(cmd *exec.Cmd, re *regexp.Regexp) (version string, result *domain.ExecResult) {
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", &domain.ExecResult{
			Status: domain.ExecResultStatusError,
			Err:    err,
			Output: string(output),
		}
	}

	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		v := matches[re.SubexpIndex("version")]
		if v == "" {
			return "", &domain.ExecResult{
				Status: domain.ExecResultStatusError,
				Err:    errors.New("version not found"),
				Output: string(output),
			}
		}

		return "v" + v, nil
	}

	return "", &domain.ExecResult{
		Status: domain.ExecResultStatusError,
		Err:    errors.New("version not found"),
		Output: string(output),
	}
}

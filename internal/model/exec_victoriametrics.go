package model

import (
	"encoding/json"
	"os"

	"debafr/internal/domain"
)

type TargetGroup struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func NewExecWriteVictoriaMetricsTargets(dic DIC) *Exec {
	summary := dic.GetSummary()
	cfg := dic.GetAppConfig()

	return NewExec(dic, domain.ExecConfig{
		Name: "Write VictoriaMetrics Targets",

		StartFunc: func() domain.ExecResult {
			var activeTarget string
			if summary.GetNextStrategy() == domain.StrategyBlue {
				activeTarget = cfg.VictoriaMetrics.TargetBlue
			} else {
				activeTarget = cfg.VictoriaMetrics.TargetGreen
			}

			targetGroups := []TargetGroup{
				{
					Targets: []string{activeTarget},
					Labels: map[string]string{
						"project_name": cfg.ProjectName,
					},
				},
			}

			jsonData, err := json.MarshalIndent(targetGroups, "", "  ")
			if err != nil {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: "failed to marshal json",
				}
			}

			tmpFile := cfg.VictoriaMetrics.TargetsOutputFilePath + ".tmp"
			if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil { //nolint:gosec,mnd // It's ok
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: "failed to write temp file",
				}
			}

			if err := os.Rename(tmpFile, cfg.VictoriaMetrics.TargetsOutputFilePath); err != nil {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    err,
					Output: "failed to rename file",
				}
			}

			return domain.ExecResult{
				Status: domain.ExecResultStatusSuccess,
				Output: "VictoriaMetrics targets written successfully",
			}
		},

		SuccessFunc: func() {},

		ErrorFunc: func() {},

		NextCmd: NewComplete(dic),
	})
}

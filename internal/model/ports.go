package model

import (
	"errors"
	"fmt"
	"strings"

	"debafr/internal/domain"
)

func NewExecPorts(dic DIC) *Exec {
	summary := dic.GetSummary()
	cfg := dic.GetAppConfig()

	return NewExec(dic, domain.ExecConfig{
		Name: "Parse ports",

		StartFunc: func() domain.ExecResult {
			currNextPorts := parseCurrNextPorts(
				cfg.LocationPorts,
				summary.GetNextStrategy(),
				summary.GetMode(),
			)
			if len(currNextPorts) == 0 {
				return domain.ExecResult{
					Status: domain.ExecResultStatusError,
					Err:    errors.New("parse ports error"),
					Output: "ports not found",
				}
			}

			summary.UpdatePorts(currNextPorts)

			lines := make([]string, 0, len(currNextPorts)+1)
			lines = append(lines, "Found ports")
			for _, p := range currNextPorts {
				lines = append(lines, fmt.Sprintf("  %s: %s >> %s", p.Location, p.CurrentPort, p.NextPort))
			}

			return domain.ExecResult{
				Status: domain.ExecResultStatusSuccess,
				Output: strings.Join(lines, "\n"),
			}
		},

		SuccessFunc: func() {
			// Нечего делать
		},
		ErrorFunc: func() {
			// Нечего делать
		},

		NextCmd: NewNextDeploy(dic),
	})
}

func parseCurrNextPorts(
	locPorts []domain.AppConfigLocationPort,
	nextStrategy domain.Strategy,
	mode domain.Mode,
) []CurrNextPort {
	if len(locPorts) == 0 {
		return nil
	}

	currNextPorts := make([]CurrNextPort, 0, len(locPorts))

	for _, item := range locPorts {
		var cp, np string
		if nextStrategy == domain.StrategyBlue {
			cp = item.GreenPort
			np = item.BluePort
		} else {
			cp = item.BluePort
			np = item.GreenPort
		}

		if mode == domain.ModeInstall {
			cp = EmptyValue
		}

		currNextPorts = append(currNextPorts, CurrNextPort{
			Location:    item.Location,
			CurrentPort: cp,
			NextPort:    np,
		})
	}

	return currNextPorts
}

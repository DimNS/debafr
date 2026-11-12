package model

import (
	"fmt"
	"os"
	"strings"

	"debafr/internal/domain"
)

func NewExecReplacer(dic DIC) *Exec {
	summary := dic.GetSummary()
	cfg := dic.GetAppConfig()

	return NewExec(dic, domain.ExecConfig{
		Name: "Replacer",

		StartFunc: func() domain.ExecResult {
			next := summary.GetNextStrategy()

			output := make([]string, 0, len(cfg.Replacer))
			for _, r := range cfg.Replacer {
				count, err := replaceInFile(r, next)
				if err != nil {
					return domain.ExecResult{
						Status: domain.ExecResultStatusError,
						Err:    fmt.Errorf("%s: %w", r.Path, err),
						Output: "failed to replace in file",
					}
				}

				output = append(output, fmt.Sprintf("%s: %d replacements", r.Path, count))
			}

			return domain.ExecResult{
				Status: domain.ExecResultStatusSuccess,
				Output: strings.Join(output, "\n"),
			}
		},

		SuccessFunc: func() {},

		ErrorFunc: func() {},

		AutoAdvance: true,

		NextCmd: NewComplete(dic),
	})
}

// replaceInFile swaps the value of the current strategy for the next one.
func replaceInFile(r domain.AppConfigReplacer, next domain.Strategy) (int, error) {
	search, replace := r.Green, r.Blue
	if next == domain.StrategyGreen {
		search, replace = r.Blue, r.Green
	}

	content, err := os.ReadFile(r.Path)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	count := strings.Count(string(content), search)
	if count == 0 {
		// Повторный update после сбоя: файл уже в состоянии следующей стратегии.
		if strings.Contains(string(content), replace) {
			return 0, nil
		}

		return 0, fmt.Errorf("neither %q nor %q found", search, replace)
	}

	replaced := strings.ReplaceAll(string(content), search, replace)
	// #nosec G703 -- путь берётся из конфига проекта, как и у nginx.conf.
	if err := os.WriteFile(r.Path, []byte(replaced), fileMode); err != nil {
		return 0, fmt.Errorf("write file: %w", err)
	}

	return count, nil
}

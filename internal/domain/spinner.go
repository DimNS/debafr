package domain

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

func NewSpinner(style lipgloss.Style) spinner.Model {
	return spinner.New(
		spinner.WithSpinner(spinner.Points),
		spinner.WithStyle(style),
	)
}

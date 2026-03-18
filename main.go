package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"debafr/internal/application"
)

var appVersion = "0.0.0" //nolint:gochecknoglobals // все в порядке

func main() {
	app, err := application.New(appVersion)
	if err != nil {
		log.Fatalf("New: %v\n", err)
	}

	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Run: %v\n", err)
	}
}

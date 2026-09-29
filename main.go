package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"debafr/internal/application"
)

var appVersion = "0.0.0" //nolint:gochecknoglobals // все в порядке

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Printf("debafr %s\n", appVersion)

			return
		case "update":
			if err := application.Update(appVersion); err != nil {
				log.Fatalf("Update: %v\n", err)
			}

			return
		}
	}

	app, err := application.New(appVersion)
	if err != nil {
		log.Fatalf("New: %v\n", err)
	}

	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Run: %v\n", err)
	}
}

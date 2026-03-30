package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

const TestContainerName = "debafr_app"

type NextCmdMsg struct {
	NextCmd tea.Model
}

type StatusDone struct {
	bool
}

type StatusError struct {
	error
}

func (e StatusError) Error() string {
	return e.error.Error()
}

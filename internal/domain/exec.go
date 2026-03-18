package domain

import tea "github.com/charmbracelet/bubbletea"

type ExecConfig struct {
	Name string

	StartFunc func() ExecResult

	SuccessFunc func()
	ErrorFunc   func()

	NextCmd tea.Model
}

type ExecResult struct {
	Status ExecResultStatus
	Output string
	Err    error
}

type ExecResultStatus string

const (
	ExecResultStatusSuccess ExecResultStatus = "SUCCESS"
	ExecResultStatusError   ExecResultStatus = "ERROR"
)

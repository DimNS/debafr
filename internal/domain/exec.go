package domain

import tea "github.com/charmbracelet/bubbletea"

type ExecConfig struct {
	Name string

	StartFunc func() ExecResult

	SuccessFunc func()
	ErrorFunc   func()

	// AutoAdvance moves to NextCmd right after a successful run, without
	// waiting for enter. Set it to false to keep the manual mode.
	AutoAdvance bool

	// LiveFunc renders the lines a long running step shows under its name,
	// a progress bar for instance. It is read on every frame of the spinner, so
	// it has to be safe to call from another goroutine than the step itself.
	LiveFunc func() string

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

package model

import tea "github.com/charmbracelet/bubbletea"

const (
	TestContainerName = "debafr_app"

	PathDocker = "/usr/bin/docker"
	PathCurl   = "/usr/bin/curl"
	PathNginx  = "/usr/sbin/nginx"
)

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

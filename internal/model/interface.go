package model

import (
	"debafr/internal/domain"
	"debafr/internal/provider/docker"
)

type DIC interface {
	GetDevMode() bool

	GetSummaryWidth() int

	GetPhysicalWidth() int
	GetPhysicalHeight() int

	GetTheme() *domain.Theme

	GetSummary() *Summary

	GetDockerService() *docker.Docker

	GetAppConfig() domain.AppConfig
}

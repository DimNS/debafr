package application

import (
	"debafr/internal/domain"
	"debafr/internal/model"
	"debafr/internal/provider/docker"
)

type DICConfig struct {
	DevMode bool

	SummaryWidth int

	PhysicalWidth  int
	PhysicalHeight int

	Theme *domain.Theme

	Summary *model.Summary

	DockerService *docker.Docker
}

type DIC struct {
	devMode bool

	summaryWidth int

	physicalWidth  int
	physicalHeight int

	theme *domain.Theme

	summary *model.Summary

	dockerService *docker.Docker
}

func NewDIC(cfg DICConfig) *DIC {
	return &DIC{
		devMode: cfg.DevMode,

		summaryWidth: cfg.SummaryWidth,

		physicalWidth:  cfg.PhysicalWidth,
		physicalHeight: cfg.PhysicalHeight,

		theme: cfg.Theme,

		summary: cfg.Summary,

		dockerService: cfg.DockerService,
	}
}

func (d *DIC) GetDevMode() bool {
	return d.devMode
}

func (d *DIC) GetSummaryWidth() int {
	return d.summaryWidth
}

func (d *DIC) GetPhysicalWidth() int {
	return d.physicalWidth
}

func (d *DIC) GetPhysicalHeight() int {
	return d.physicalHeight
}

func (d *DIC) GetTheme() *domain.Theme {
	return d.theme
}

func (d *DIC) GetSummary() *model.Summary {
	return d.summary
}

func (d *DIC) GetDockerService() *docker.Docker {
	return d.dockerService
}

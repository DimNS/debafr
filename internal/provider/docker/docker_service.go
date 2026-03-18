package docker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"debafr/internal/domain"
)

const (
	defaultTimeout = 10 * time.Second
)

// Docker представляет работу с Docker.
type Docker struct {
	devMode bool

	cli *client.Client
}

// New возвращает новый экземпляр Docker.
func New(devMode bool) (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("client.NewClientWithOpts: %v", err)
	}

	return &Docker{
		devMode: devMode,

		cli: cli,
	}, nil
}

func (d *Docker) GetCurrentDeploy(needProjectName string) (currVersion string, currStrategy domain.Strategy, err error) { //nolint:gocognit,gocyclo,cyclop // It's ok
	if d.devMode {
		return "v0.8.0", "blue", nil
	}

	containers, err := d.cli.ContainerList(context.Background(), container.ListOptions{
		All: true, // Include stopped containers too
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to list containers: %v", err)
	}

	versions := make(map[string]struct{})
	strategies := make(map[domain.Strategy]struct{})

	for _, ctr := range containers {
		if len(ctr.Labels) == 0 {
			continue
		}

		projectName, projectNameOK := ctr.Labels["app.project.name"]
		version, versionOK := ctr.Labels["app.version"]
		deployStrategy, deployStrategyOK := ctr.Labels["app.deployment.strategy"]

		if !projectNameOK || !deployStrategyOK || !versionOK {
			continue
		}
		if projectName != needProjectName {
			continue
		}

		versions[version] = struct{}{}

		switch deployStrategy {
		case domain.StrategyBlue.String():
			strategies[domain.StrategyBlue] = struct{}{}
		case domain.StrategyGreen.String():
			strategies[domain.StrategyGreen] = struct{}{}
		}
	}

	if len(versions) == 0 {
		return "", "", errors.New("version not found")
	}
	if len(versions) > 1 {
		return "", "", errors.New("multiple versions found")
	}

	if len(strategies) == 0 {
		return "", "", errors.New("strategy not found")
	}
	if len(strategies) > 1 {
		return "", "", errors.New("multiple strategies found")
	}

	for k := range versions {
		currVersion = "v" + k
		break
	}

	for k := range strategies {
		currStrategy = k
		break
	}

	return currVersion, currStrategy, nil
}

func (d *Docker) GetContainers(projectName string, version string) (frontend, backend *container.Summary, err error) {
	if d.devMode {
		return &container.Summary{
				Image: "ghcr.io/capuchinapp/cloud/ui:v0.8.0",
				Names: []string{"/project_blue_v0.8.0_ui"},
			}, &container.Summary{
				Image: "ghcr.io/capuchinapp/cloud/api:v0.8.0",
				Names: []string{"/project_blue_v0.8.0_api"},
			}, nil
	}

	containers, err := d.cli.ContainerList(context.Background(), container.ListOptions{
		All: true, // Include stopped containers too
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list containers: %v", err)
	}

	for _, ctr := range containers {
		if domain.ContainerExistLabel(ctr.Labels, projectName, domain.ContainerAppServiceTypeFrontend) &&
			domain.ContainerExistVersion(ctr.Names, version) {
			frontend = &ctr
		} else if domain.ContainerExistLabel(ctr.Labels, projectName, domain.ContainerAppServiceTypeBackend) &&
			domain.ContainerExistVersion(ctr.Names, version) {
			backend = &ctr
		}
	}

	if frontend == nil {
		return nil, nil, errors.New("frontend container not found")
	}
	if backend == nil {
		return nil, nil, errors.New("backend container not found")
	}

	return frontend, backend, nil
}

func (d *Docker) GetState(containerID string) (domain.ContainerState, error) {
	if d.devMode {
		return domain.ContainerState{
			Status: domain.ContainerStateStatusRunning,
			Health: domain.ContainerStateHealthHealthy,
		}, nil
	}

	inspect, err := d.cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return domain.ContainerState{}, fmt.Errorf("failed to inspect container: %v", err)
	}
	if inspect.State == nil {
		return domain.ContainerState{}, errors.New("container state is nil")
	}
	if inspect.State.Health == nil {
		return domain.ContainerState{}, errors.New("container health is nil")
	}

	return domain.ContainerState{
		Status: statusToDomain(inspect.State.Status),
		Health: healthToDomain(inspect.State.Health.Status),
	}, nil
}

func (d *Docker) ContainerStop(containerID string) error {
	if d.devMode {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	err := d.cli.ContainerStop(ctx, containerID, container.StopOptions{})
	if err != nil {
		return fmt.Errorf("failed to stop container: %v", err)
	}

	return nil
}

func statusToDomain(status string) domain.ContainerStateStatus {
	switch status {
	case "created":
		return domain.ContainerStateStatusCreated
	case "running":
		return domain.ContainerStateStatusRunning
	case "paused":
		return domain.ContainerStateStatusPaused
	case "restarting":
		return domain.ContainerStateStatusRestarting
	case "removing":
		return domain.ContainerStateStatusRemoving
	case "exited":
		return domain.ContainerStateStatusExited
	case "dead":
		return domain.ContainerStateStatusDead
	default:
		return domain.ContainerStateStatusExited
	}
}

func healthToDomain(status string) domain.ContainerStateHealth {
	switch status {
	case container.Starting:
		return domain.ContainerStateHealthStarting
	case container.Healthy:
		return domain.ContainerStateHealthHealthy
	case container.Unhealthy:
		return domain.ContainerStateHealthUnhealthy
	default:
		return domain.ContainerStateHealthNoHealthcheck
	}
}

package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"

	"debafr/internal/domain"
)

const (
	defaultTimeout = 10 * time.Second
)

// ErrDeployNotFound means no containers of the project were found.
var ErrDeployNotFound = errors.New("version not found")

// Docker представляет работу с Docker.
type Docker struct {
	devMode bool

	cli *client.Client
}

// New возвращает новый экземпляр Docker.
func New(devMode bool) (*Docker, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
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

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	containers, err := d.cli.ContainerList(ctx, container.ListOptions{
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
		default:
			continue
		}
	}

	if len(versions) == 0 {
		return "", "", ErrDeployNotFound
	}
	if len(versions) > 1 {
		vs := make([]string, 0, len(versions))
		for k := range versions {
			vs = append(vs, k)
		}
		return "", "", fmt.Errorf("multiple versions found: %s", strings.Join(vs, ", "))
	}

	if len(strategies) == 0 {
		return "", "", errors.New("strategy not found")
	}
	if len(strategies) > 1 {
		ss := make([]string, 0, len(strategies))
		for k := range strategies {
			ss = append(ss, k.String())
		}
		return "", "", fmt.Errorf("multiple strategies found: %s", strings.Join(ss, ", "))
	}

	for k := range versions {
		currVersion = k
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

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	containers, err := d.cli.ContainerList(ctx, container.ListOptions{
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

// GetContainersStatus returns a state snapshot of the frontend and backend containers of the deploy.
func (d *Docker) GetContainersStatus(projectName, version string) ([]domain.ContainerStatus, error) {
	frontend, backend, err := d.GetContainers(projectName, version)
	if err != nil {
		return nil, fmt.Errorf("get containers: %v", err)
	}

	summaries := []struct {
		service domain.ContainerAppServiceType
		ctr     *container.Summary
	}{
		{domain.ContainerAppServiceTypeFrontend, frontend},
		{domain.ContainerAppServiceTypeBackend, backend},
	}

	statuses := make([]domain.ContainerStatus, 0, len(summaries))
	for _, s := range summaries {
		state, err := d.GetState(s.ctr.ID)
		if err != nil {
			return nil, fmt.Errorf("get state of %s: %v", s.service, err)
		}

		statuses = append(statuses, domain.ContainerStatus{
			Service: s.service,
			Name:    strings.TrimPrefix(strings.Join(s.ctr.Names, ","), "/"),
			Image:   s.ctr.Image,
			State:   state,
		})
	}

	return statuses, nil
}

func (d *Docker) GetState(containerID string) (domain.ContainerState, error) {
	if d.devMode {
		return domain.ContainerState{
			Status: domain.ContainerStateStatusRunning,
			Health: domain.ContainerStateHealthHealthy,
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	inspect, err := d.cli.ContainerInspect(ctx, containerID)
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

func (d *Docker) RunContainer(ctx context.Context, name, img, hostPort, containerPort string) (string, error) {
	port := nat.Port(containerPort)

	resp, err := d.cli.ContainerCreate(ctx, &container.Config{
		Image:        img,
		ExposedPorts: nat.PortSet{port: struct{}{}},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{port: []nat.PortBinding{{HostPort: hostPort}}},
	}, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	if err := d.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}

	return resp.ID, nil
}

func (d *Docker) Exec(ctx context.Context, containerID string, cmd []string) ([]byte, error) {
	execResp, err := d.cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %v", err)
	}

	attach, err := d.cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach exec: %v", err)
	}
	defer attach.Close()

	var buf bytes.Buffer
	if _, err := stdcopy.StdCopy(&buf, &buf, attach.Reader); err != nil {
		return nil, fmt.Errorf("failed to read exec output: %v", err)
	}
	attach.Close()

	inspect, err := d.cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect exec: %v", err)
	}
	if inspect.ExitCode != 0 {
		return buf.Bytes(), fmt.Errorf("exec exited with code %d: %s", inspect.ExitCode, buf.String())
	}

	return buf.Bytes(), nil
}

func (d *Docker) RemoveContainer(ctx context.Context, containerID string) error {
	if err := d.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("failed to remove container: %v", err)
	}

	return nil
}

func (d *Docker) RemoveImage(ctx context.Context, imageID string) error {
	if _, err := d.cli.ImageRemove(ctx, imageID, image.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("failed to remove image: %v", err)
	}

	return nil
}

func (d *Docker) ListRunning(ctx context.Context) ([]container.Summary, error) {
	containers, err := d.cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	return containers, nil
}

func (d *Docker) ImagePull(ctx context.Context, img string, pullOpts image.PullOptions) error {
	if d.devMode {
		return nil
	}

	_, err := d.cli.ImagePull(ctx, img, pullOpts)
	if err != nil {
		return fmt.Errorf("failed to pull image: %v", err)
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

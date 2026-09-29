package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"

	"debafr/internal/domain"
)

const (
	defaultTimeout = 10 * time.Second
)

// ErrDeployNotFound means no containers of the project were found.
var ErrDeployNotFound = errors.New("version not found")

// Docker представляет работу с Docker.
type Docker struct {
	cli *client.Client
}

// New возвращает новый экземпляр Docker.
func New() (*Docker, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("client.NewClientWithOpts: %v", err)
	}

	return &Docker{cli: cli}, nil
}

func (d *Docker) GetCurrentDeploy(needProjectName string) (currVersion string, currStrategy domain.Strategy, err error) { //nolint:gocognit,gocyclo,cyclop // It's ok
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// Только запущенные: остановленный прошлый деплой остаётся в docker
	// навсегда, и вместе с работающим он давал бы две стратегии в ответе.
	containers, err := d.cli.ContainerList(ctx, container.ListOptions{
		All: false,
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
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	containers, err := d.cli.ContainerList(ctx, container.ListOptions{
		All: true, // Include stopped containers too
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list containers: %v", err)
	}

	// Остановленный контейнер прошлого деплоя с той же версией не должен
	// перебивать работающий. Если запущенных нет, берём любые.
	if running := runningOnly(containers); len(running) > 0 {
		containers = running
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

// runningOnly оставляет только запущенные контейнеры.
func runningOnly(containers []container.Summary) []container.Summary {
	running := make([]container.Summary, 0, len(containers))
	for _, ctr := range containers {
		if ctr.State == "running" {
			running = append(running, ctr)
		}
	}
	return running
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
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	err := d.cli.ContainerStop(ctx, containerID, container.StopOptions{})
	if err != nil {
		return fmt.Errorf("failed to stop container: %v", err)
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

// PullProgress is how much of an image is downloaded at the moment. Docker
// reports the size of a layer only once the layer starts arriving, so Total
// grows while the pull is running.
type PullProgress struct {
	Current int64
	Total   int64
}

// ImagePull pulls the image and reports how far it has got through progress,
// which may be nil.
//
// ponytail: the sizes of the layers are known only as they come in, so the
// share of the pull jumps every time a new layer shows up. The bar shows what
// has actually arrived, it does not smooth it out on its own.
func (d *Docker) ImagePull(ctx context.Context, img string, pullOpts image.PullOptions, progress func(PullProgress)) error {
	body, err := d.cli.ImagePull(ctx, img, pullOpts)
	if err != nil {
		return fmt.Errorf("failed to pull image: %v", err)
	}
	defer body.Close()

	stats := newPullStats()

	dec := json.NewDecoder(body)
	for {
		var msg jsonmessage.JSONMessage
		if err := dec.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("failed to read pull progress: %v", err)
		}

		if msg.Error != nil {
			return fmt.Errorf("failed to pull image: %v", msg.Error)
		}
		if msg.Progress == nil {
			continue
		}

		if progress != nil {
			progress(stats.add(msg.ID, msg.Progress))
		}
	}
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

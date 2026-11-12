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

// GetContainers returns all app containers of the deploy, whatever services
// the project consists of.
func (d *Docker) GetContainers(projectName string, version string) ([]container.Summary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	containers, err := d.cli.ContainerList(ctx, container.ListOptions{
		All: true, // Include stopped containers too
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	matched := deployContainers(containers, projectName, version)
	if len(matched) == 0 {
		return nil, errors.New("containers not found")
	}

	return matched, nil
}

func deployContainers(containers []container.Summary, projectName, version string) []container.Summary {
	matched := make([]container.Summary, 0, len(containers))

	for _, ctr := range containers {
		if _, ok := domain.ContainerServiceType(ctr.Labels, projectName); !ok {
			continue
		}
		if !domain.ContainerExistVersion(ctr.Labels, ctr.Names, version) {
			continue
		}

		matched = append(matched, ctr)
	}

	// Остановленный контейнер прошлого деплоя с той же версией не должен
	// перебивать работающий. Если запущенных нет, берём любые: healthcheck
	// должен показать их состояние, а не «containers not found».
	if running := runningOnly(matched); len(running) > 0 {
		matched = running
	}

	return matched
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

// GetContainersStatus returns a state snapshot of all app containers of the deploy.
func (d *Docker) GetContainersStatus(projectName, version string) ([]domain.ContainerStatus, error) {
	containers, err := d.GetContainers(projectName, version)
	if err != nil {
		return nil, fmt.Errorf("get containers: %v", err)
	}

	statuses := make([]domain.ContainerStatus, 0, len(containers))
	for _, ctr := range containers {
		service, _ := domain.ContainerServiceType(ctr.Labels, projectName)

		state, err := d.GetState(ctr.ID)
		if err != nil {
			return nil, fmt.Errorf("get state of %s: %v", service, err)
		}

		statuses = append(statuses, domain.ContainerStatus{
			Service: service,
			Name:    strings.TrimPrefix(strings.Join(ctr.Names, ","), "/"),
			Image:   ctr.Image,
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

	health := domain.ContainerStateHealthNoHealthcheck
	if inspect.State.Health != nil {
		health = healthToDomain(inspect.State.Health.Status)
	}

	return domain.ContainerState{
		Status: statusToDomain(inspect.State.Status),
		Health: health,
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

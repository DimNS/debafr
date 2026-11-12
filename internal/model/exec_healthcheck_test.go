package model

import (
	"errors"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"debafr/internal/domain"
)

type fakeStateGetter struct {
	states map[string]domain.ContainerState
	err    error
}

func (f fakeStateGetter) GetState(containerID string) (domain.ContainerState, error) {
	if f.err != nil {
		return domain.ContainerState{}, f.err
	}

	return f.states[containerID], nil
}

func TestHealthcheck(t *testing.T) {
	t.Parallel()

	ctr := func(id, service string) container.Summary {
		return container.Summary{
			ID: id,
			Labels: map[string]string{
				"app.project.name": "myapp",
				"app.service.type": service,
			},
		}
	}

	running := func(health domain.ContainerStateHealth) domain.ContainerState {
		return domain.ContainerState{Status: domain.ContainerStateStatusRunning, Health: health}
	}

	t.Run("Should pass for one container", func(t *testing.T) {
		t.Parallel()

		res := healthcheck(
			fakeStateGetter{states: map[string]domain.ContainerState{
				"api": running(domain.ContainerStateHealthHealthy),
			}},
			"myapp",
			[]container.Summary{ctr("api", "backend")},
			1,
			0,
		)

		require.Equal(t, domain.ExecResultStatusSuccess, res.Status)
		assert.Contains(t, res.Output, "backend")
	})

	t.Run("Should pass for three containers", func(t *testing.T) {
		t.Parallel()

		res := healthcheck(
			fakeStateGetter{states: map[string]domain.ContainerState{
				"api":    running(domain.ContainerStateHealthHealthy),
				"ui":     running(domain.ContainerStateHealthHealthy),
				"worker": running(domain.ContainerStateHealthHealthy),
			}},
			"myapp",
			[]container.Summary{ctr("api", "backend"), ctr("ui", "frontend"), ctr("worker", "worker")},
			1,
			0,
		)

		require.Equal(t, domain.ExecResultStatusSuccess, res.Status)
		assert.Contains(t, res.Output, "backend")
		assert.Contains(t, res.Output, "frontend")
		assert.Contains(t, res.Output, "worker")
	})

	t.Run("Should pass for a running container without a healthcheck", func(t *testing.T) {
		t.Parallel()

		res := healthcheck(
			fakeStateGetter{states: map[string]domain.ContainerState{
				"api": running(domain.ContainerStateHealthNoHealthcheck),
			}},
			"myapp",
			[]container.Summary{ctr("api", "backend")},
			1,
			0,
		)

		require.Equal(t, domain.ExecResultStatusSuccess, res.Status)
	})

	t.Run("Should fail when any container is not ready", func(t *testing.T) {
		t.Parallel()

		res := healthcheck(
			fakeStateGetter{states: map[string]domain.ContainerState{
				"api": running(domain.ContainerStateHealthHealthy),
				"ui":  running(domain.ContainerStateHealthStarting),
			}},
			"myapp",
			[]container.Summary{ctr("api", "backend"), ctr("ui", "frontend")},
			2,
			0,
		)

		require.Equal(t, domain.ExecResultStatusError, res.Status)
		assert.Contains(t, res.Output, "frontend")
		assert.Contains(t, res.Output, "attempt 2/2")
	})

	t.Run("Should fail on state error", func(t *testing.T) {
		t.Parallel()

		res := healthcheck(
			fakeStateGetter{err: errors.New("docker is gone")},
			"myapp",
			[]container.Summary{ctr("api", "backend")},
			1,
			0,
		)

		require.Equal(t, domain.ExecResultStatusError, res.Status)
		assert.ErrorContains(t, res.Err, "docker is gone")
		assert.Contains(t, res.Output, "backend")
	})
}

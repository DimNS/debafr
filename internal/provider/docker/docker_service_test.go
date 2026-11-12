package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"

	"debafr/internal/domain"
)

func Test_statusToDomain(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   domain.ContainerStateStatus
	}{
		{
			name:   "created",
			status: "created",
			want:   domain.ContainerStateStatusCreated,
		},
		{
			name:   "running",
			status: "running",
			want:   domain.ContainerStateStatusRunning,
		},
		{
			name:   "paused",
			status: "paused",
			want:   domain.ContainerStateStatusPaused,
		},
		{
			name:   "restarting",
			status: "restarting",
			want:   domain.ContainerStateStatusRestarting,
		},
		{
			name:   "removing",
			status: "removing",
			want:   domain.ContainerStateStatusRemoving,
		},
		{
			name:   "exited",
			status: "exited",
			want:   domain.ContainerStateStatusExited,
		},
		{
			name:   "dead",
			status: "dead",
			want:   domain.ContainerStateStatusDead,
		},
		{
			name:   "unknown",
			status: "unknown",
			want:   domain.ContainerStateStatusExited,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusToDomain(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_healthToDomain(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   domain.ContainerStateHealth
	}{
		{
			name:   "starting",
			status: "starting",
			want:   domain.ContainerStateHealthStarting,
		},
		{
			name:   "healthy",
			status: "healthy",
			want:   domain.ContainerStateHealthHealthy,
		},
		{
			name:   "unhealthy",
			status: "unhealthy",
			want:   domain.ContainerStateHealthUnhealthy,
		},
		{
			name:   "no healthcheck",
			status: "none",
			want:   domain.ContainerStateHealthNoHealthcheck,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := healthToDomain(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_runningOnly(t *testing.T) {
	tests := []struct {
		name  string
		given []container.Summary
		want  int
	}{
		{"остались только запущенные", []container.Summary{
			{State: "running"}, {State: "exited"},
		}, 1},
		{"запущенных нет", []container.Summary{
			{State: "exited"}, {State: "created"},
		}, 0},
		{"все запущены", []container.Summary{
			{State: "running"}, {State: "running"},
		}, 2},
		{"пусто", []container.Summary{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runningOnly(tt.given)

			assert.Len(t, got, tt.want)
			for _, ctr := range got {
				assert.Equal(t, "running", ctr.State)
			}
		})
	}
}

func Test_deployContainers(t *testing.T) {
	t.Parallel()

	ctr := func(id, project, service, version string) container.Summary {
		return container.Summary{
			ID:    id,
			Names: []string{"/" + project + "_blue_" + service + "_" + version},
			Labels: map[string]string{
				"app.project.name": project,
				"app.service.type": service,
			},
		}
	}

	tests := []struct {
		name    string
		given   []container.Summary
		wantIDs []string
	}{
		{
			name: "Should find a single container",
			given: []container.Summary{
				ctr("api", "myapp", "backend", "v1"),
			},
			wantIDs: []string{"api"},
		},
		{
			name: "Should find three containers of any service type",
			given: []container.Summary{
				ctr("api", "myapp", "backend", "v1"),
				ctr("ui", "myapp", "frontend", "v1"),
				ctr("worker", "myapp", "worker", "v1"),
			},
			wantIDs: []string{"api", "ui", "worker"},
		},
		{
			name: "Should ignore other projects, other versions and infrastructure",
			given: []container.Summary{
				ctr("api", "myapp", "backend", "v1"),
				ctr("old", "myapp", "backend", "v0"),
				ctr("other", "other", "backend", "v1"),
				ctr("db", "myapp", "", "v1"),
			},
			wantIDs: []string{"api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := deployContainers(tt.given, "myapp", "v1")

			ids := make([]string, 0, len(got))
			for _, ctr := range got {
				ids = append(ids, ctr.ID)
			}

			assert.Equal(t, tt.wantIDs, ids)
		})
	}
}

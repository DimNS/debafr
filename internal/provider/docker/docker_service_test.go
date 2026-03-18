package docker

import (
	"testing"

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

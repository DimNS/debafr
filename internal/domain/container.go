package domain

import (
	"strings"
)

type ContainerState struct {
	Status ContainerStateStatus
	Health ContainerStateHealth
}

type ContainerStateStatus string

const (
	ContainerStateStatusCreated    ContainerStateStatus = "created"
	ContainerStateStatusRunning    ContainerStateStatus = "running"
	ContainerStateStatusPaused     ContainerStateStatus = "paused"
	ContainerStateStatusRestarting ContainerStateStatus = "restarting"
	ContainerStateStatusRemoving   ContainerStateStatus = "removing"
	ContainerStateStatusExited     ContainerStateStatus = "exited"
	ContainerStateStatusDead       ContainerStateStatus = "dead"
)

func (s ContainerStateStatus) String() string {
	return string(s)
}

type ContainerStateHealth string

const (
	ContainerStateHealthNoHealthcheck ContainerStateHealth = "none"      // Indicates there is no healthcheck
	ContainerStateHealthStarting      ContainerStateHealth = "starting"  // Starting indicates that the container is not yet ready
	ContainerStateHealthHealthy       ContainerStateHealth = "healthy"   // Healthy indicates that the container is running correctly
	ContainerStateHealthUnhealthy     ContainerStateHealth = "unhealthy" // Unhealthy indicates that the container has a problem
)

func (s ContainerStateHealth) String() string {
	return string(s)
}

// ContainerStatus is a snapshot of one project container.
type ContainerStatus struct {
	Service ContainerAppServiceType
	Name    string
	Image   string
	State   ContainerState
}

type ContainerAppServiceType string

const (
	ContainerAppServiceTypeBackend  ContainerAppServiceType = "backend"
	ContainerAppServiceTypeFrontend ContainerAppServiceType = "frontend"
)

func (v ContainerAppServiceType) String() string {
	return string(v)
}

// ContainerServiceType returns the app service type of a container. Containers
// without the label are infrastructure and are not part of the deploy.
func ContainerServiceType(labels map[string]string, needProjectName string) (ContainerAppServiceType, bool) {
	if len(labels) == 0 || labels["app.project.name"] != needProjectName {
		return "", false
	}

	serviceType := labels["app.service.type"]
	if serviceType == "" {
		return "", false
	}

	return ContainerAppServiceType(serviceType), true
}

func ContainerExistVersion(names []string, version string) bool {
	if len(names) == 0 {
		return false
	}

	for _, name := range names {
		if strings.Contains(name, version) {
			return true
		}
	}

	return false
}

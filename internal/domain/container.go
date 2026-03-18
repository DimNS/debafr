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

type ContainerAppServiceType string

const (
	ContainerAppServiceTypeBackend  ContainerAppServiceType = "backend"
	ContainerAppServiceTypeFrontend ContainerAppServiceType = "frontend"
)

func (v ContainerAppServiceType) String() string {
	return string(v)
}

func ContainerExistLabel(labels map[string]string, needProjectName string, needServiceType ContainerAppServiceType) bool {
	if len(labels) == 0 {
		return false
	}

	name, nameOK := labels["app.project.name"]
	serviceType, serviceTypeOK := labels["app.service.type"]

	return nameOK && name == needProjectName && serviceTypeOK && serviceType == needServiceType.String()
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

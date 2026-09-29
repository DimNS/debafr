package application

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"debafr/internal/domain"
	"debafr/internal/provider/docker"
)

const tabPadding = 2

// Status prints the current deploy of the project: version, strategy, ports and container states.
func Status() error {
	conf, err := LoadConfiguration("debafr.toml")
	if err != nil {
		return fmt.Errorf("load configuration: %v", err)
	}

	projectName := conf.Toml.App.ProjectName

	dockerService, err := docker.New()
	if err != nil {
		return fmt.Errorf("new docker: %v", err)
	}

	version, strategy, err := dockerService.GetCurrentDeploy(projectName)
	if errors.Is(err, docker.ErrDeployNotFound) {
		fmt.Printf("%s: no deploy found\n", projectName)

		return nil
	}
	if err != nil {
		return fmt.Errorf("get current deploy: %v", err)
	}

	fmt.Printf("%s %s (%s)\n\n", projectName, version, strategy)

	fmt.Println("Ports")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, tabPadding, ' ', 0)
	for _, p := range activePorts(conf.Toml.App.LocationPorts, strategy) {
		fmt.Fprintf(w, "  %s\t%s\n", p.Location, p.Port)
	}
	_ = w.Flush()

	statuses, err := dockerService.GetContainersStatus(projectName, version)
	if err != nil {
		return fmt.Errorf("get containers status: %v", err)
	}

	fmt.Println()
	fmt.Println("Containers")

	w = tabwriter.NewWriter(os.Stdout, 0, 0, tabPadding, ' ', 0)
	for _, s := range statuses {
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s/%s\n", s.Service, s.Name, s.Image, s.State.Status, s.State.Health)
	}
	_ = w.Flush()

	return nil
}

type activePort struct {
	Location string
	Port     string
}

// activePorts returns the ports serving traffic in the given strategy.
func activePorts(locPorts []LocationPort, strategy domain.Strategy) []activePort {
	ports := make([]activePort, 0, len(locPorts))

	for _, p := range locPorts {
		port := p.GreenPort
		if strategy == domain.StrategyBlue {
			port = p.BluePort
		}

		ports = append(ports, activePort{Location: p.Location, Port: port})
	}

	return ports
}

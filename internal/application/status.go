package application

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"debafr/internal/provider/docker"
)

const tabPadding = 2

// Status prints the current deploy of the project: version, strategy and container states.
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

	statuses, err := dockerService.GetContainersStatus(projectName, version)
	if err != nil {
		return fmt.Errorf("get containers status: %v", err)
	}

	fmt.Println("Containers")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, tabPadding, ' ', 0)
	for _, s := range statuses {
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s/%s\n", s.Service, s.Name, s.Image, s.State.Status, s.State.Health)
	}
	_ = w.Flush()

	return nil
}

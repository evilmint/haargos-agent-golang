package dockergatherer

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

type DockerAPIContainer struct {
	ID     string `json:"Id"`
	Names  []string
	Image  string
	State  string
	Status string
}

type DockerAPIContainerDetails struct {
	Name  string
	State struct {
		Running    bool
		Restarting bool
		StartedAt  string `json:"StartedAt"`
		FinishedAt string `json:"FinishedAt"`
	}
	// Include other fields as necessary...
}

type DockerGatherer struct {
	log        *logrus.Logger
	httpClient *http.Client
	socketPath string
}

func NewDockerGatherer(socketPath string) *DockerGatherer {
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
		Timeout: 30 * time.Second,
	}

	return &DockerGatherer{
		log:        logrus.New(),
		httpClient: httpClient,
		socketPath: socketPath,
	}
}

func (dg *DockerGatherer) GatherDocker() types.Docker {
	req, err := http.NewRequest("GET", "http://localhost/containers/json", nil)
	if err != nil {
		dg.log.Error("Error encountered while connecting to Docker socket.")
		return types.Docker{Containers: []types.DockerContainer{}}
	}

	resp, err := dg.httpClient.Do(req)
	if err != nil {
		dg.log.Error("Error encountered while gathering Docker process status.")
		return types.Docker{Containers: []types.DockerContainer{}}
	}
	defer resp.Body.Close()

	var entries []DockerAPIContainer
	err = json.NewDecoder(resp.Body).Decode(&entries)
	if err != nil {
		dg.log.Error("Failed to decode Docker JSON response")
		return types.Docker{Containers: []types.DockerContainer{}}
	}

	var containers []types.DockerContainer
	for _, entry := range entries {
		containerDetails, err := dg.inspectContainer(entry.ID)
		if err != nil {
			dg.log.Errorf("Failed to inspect container %s: %v", entry.ID, err)
			continue
		}

		container := types.DockerContainer{
			Name:       containerDetails.Name,
			Image:      entry.Image,
			State:      entry.State,
			Status:     entry.Status,
			Running:    containerDetails.State.Running,
			Restarting: fmt.Sprintf("%v", containerDetails.State.Restarting),
			StartedAt:  containerDetails.State.StartedAt,
			FinishedAt: containerDetails.State.FinishedAt,
		}
		containers = append(containers, container)
	}

	return types.Docker{Containers: containers}
}

func (dg *DockerGatherer) inspectContainer(containerID string) (*DockerAPIContainerDetails, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost/containers/%s/json", containerID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := dg.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var details DockerAPIContainerDetails
	err = json.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		return nil, err
	}

	return &details, nil
}

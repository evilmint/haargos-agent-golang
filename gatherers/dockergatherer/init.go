package dockergatherer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

type DockerGatherer struct {
	log *logrus.Logger
}

func NewDockerGatherer() *DockerGatherer {
	return &DockerGatherer{
		log: logrus.New(),
	}
}

func (dg *DockerGatherer) GatherDocker() types.Docker {
	// Simulating shell command execution
	dockerPs, err := dg.executeShellCommand("docker ps --format json")
	if err != nil {
		dg.log.Error("Failed to gather Docker process status")
		return types.Docker{}
	}

	entries := dg.parseDockerPsOutput(dockerPs)

	var containers []types.DockerContainer
	for _, entry := range entries {
		inspectString, err := dg.executeShellCommand("docker inspect " + entry.ID)
		if err != nil {
			dg.log.Errorf("Failed to inspect entry %s", entry.Names)
			continue
		}

		container, err := dg.containerFromInspect(inspectString, entry)
		if err == nil {
			containers = append(containers, container)
		}
	}

	return types.Docker{Containers: containers}
}

func (dg *DockerGatherer) parseDockerPsOutput(output string) []types.DockerPsEntry {
	jsonStringArray := strings.Split(output, "\n")

	var entries []types.DockerPsEntry
	for _, jsonStr := range jsonStringArray {
		if strings.TrimSpace(jsonStr) == "" {
			continue
		}

		var entry types.DockerPsEntry
		err := json.Unmarshal([]byte(jsonStr), &entry)
		if err == nil {
			entries = append(entries, entry)
		} else {
			dg.log.Errorf("Failed to decode docker JSON: [%s] %s\n", jsonStr, err)
		}
	}

	return entries
}

func (dg *DockerGatherer) containerFromInspect(inspectString string, entry types.DockerPsEntry) (types.DockerContainer, error) {
	inspectData := []types.DockerInspectResult{}
	err := json.Unmarshal([]byte(inspectString), &inspectData)
	if err != nil || len(inspectData) == 0 {
		return types.DockerContainer{}, err
	}
	inspect := inspectData[0]

	return types.DockerContainer{
		Name:       inspect.Name,
		Image:      entry.Image,
		State:      entry.State,
		Status:     entry.Status,
		Running:    inspect.State.IsRunning,
		Restarting: fmt.Sprintf("%v", inspect.State.IsRestarting),
		StartedAt:  inspect.State.StartedAt,
		FinishedAt: inspect.State.FinishedAt,
	}, nil
}

func (dg *DockerGatherer) executeShellCommand(command string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", command)
	output, err := cmd.Output()
	return string(output), err
}

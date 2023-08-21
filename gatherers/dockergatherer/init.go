package dockergatherer

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

type DockerGatherer struct{}

var log = logrus.New()

func (dg *DockerGatherer) GatherDocker() types.Docker {
	// Simulating shell command execution
	dockerPs := executeShellCommand("docker ps --format json")
	entries := dg.parseDockerPsOutput(dockerPs)

	var containers = make([]types.DockerContainer, 0)
	for _, entry := range entries {
		inspectString := executeShellCommand("docker inspect " + entry.ID)
		inspectData := []types.DockerInspectResult{}
		err := json.Unmarshal([]byte(inspectString), &inspectData)
		if err != nil || len(inspectData) == 0 {
			log.Errorf("Failed to inspect entry %s %s", entry.Names, err)
			continue
		}
		inspect := inspectData[0]

		containers = append(containers, types.DockerContainer{
			Name:       inspect.Name,
			Image:      entry.Image,
			State:      entry.State,
			Status:     entry.Status,
			Running:    inspect.State.IsRunning,
			Restarting: inspect.State.IsRestarting,
			StartedAt:  inspect.State.StartedAt,
			FinishedAt: inspect.State.FinishedAt,
		})
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
			log.Errorf("Failed to decode docker JSON: [%s] %s\n", jsonStr, err)
		}
	}

	return entries
}

// Note: You might want to define or import the `executeShellCommand` function in this package.
func executeShellCommand(command string) string {
	cmd := exec.Command("/bin/bash", "-c", command)

	// Get the output of the command
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to execute command: %s, error: %v", command, err)
		return ""
	}

	return string(output)
}

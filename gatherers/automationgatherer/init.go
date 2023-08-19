package automationgatherer

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/evilmint/haargos-agent-golang/types"
	"gopkg.in/yaml.v3"
)

type AutomationGatherer struct{}

func findRestoreStateByID(
	restoreState types.RestoreStateResponse,
	id string,
) (*types.RestoreStateData, bool) {
	for _, data := range restoreState.Data {
		if data.State.Attributes.ID != nil {
			if *data.State.Attributes.ID == id {
				return &data, true
			}
		}
	}

	return nil, false
}

func (a *AutomationGatherer) GatherAutomations(configPath string, restoreState types.RestoreStateResponse) []types.Automation {
	// Read automations from YAML file
	automationsData, err := os.ReadFile(filepath.Join(configPath, "automations.yaml"))
	if err != nil {
		return []types.Automation{}
	}

	var automations []types.Automation
	if err := yaml.Unmarshal(automationsData, &automations); err != nil {
		log.Printf("Error parsing automations.yaml: %v", err)
		return []types.Automation{}
	}

	// Map automations based on the restore state
	for i, automation := range automations {
		if restoreStateForAutomation, ok := findRestoreStateByID(restoreState, automation.ID); ok {
			var lastTriggered time.Time
			var err error

			if restoreStateForAutomation.State.Attributes.LastTriggered != nil {
				lastTriggered, err = time.Parse(
					time.RFC3339,
					*restoreStateForAutomation.State.Attributes.LastTriggered,
				)
			}

			if err != nil {
				log.Printf(
					"Error parsing lastTriggered for automation ID %s: %v",
					automation.ID,
					err,
				)
			} else {
				automations[i].LastTriggered = lastTriggered
			}
			automations[i].FriendlyName = *restoreStateForAutomation.State.Attributes.FriendlyName
			automations[i].State = restoreStateForAutomation.State.State
		}
	}

	return automations
}

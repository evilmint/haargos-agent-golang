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

func (a *AutomationGatherer) GatherAutomations(configPath string, restoreState types.RestoreStateResponse) []types.Automation {
	automationsData, err := a.readAutomationsFile(configPath)
	if err != nil {
		log.Printf("Error reading automations.yaml: %v", err)
		return []types.Automation{}
	}

	return a.processAutomations(automationsData, restoreState)
}

func (a *AutomationGatherer) readAutomationsFile(configPath string) ([]byte, error) {
	return os.ReadFile(filepath.Join(configPath, "automations.yaml"))
}

func (a *AutomationGatherer) processAutomations(automationsData []byte, restoreState types.RestoreStateResponse) []types.Automation {
	var automations []types.Automation
	if err := yaml.Unmarshal(automationsData, &automations); err != nil {
		log.Printf("Error parsing automations.yaml: %v", err)
		return []types.Automation{}
	}

	for i, automation := range automations {
		if restoreStateForAutomation, ok := findRestoreStateByID(restoreState, automation.ID); ok {
			a.updateAutomationFromRestoreState(&automations[i], restoreStateForAutomation)
		}
	}

	return automations
}

func (a *AutomationGatherer) updateAutomationFromRestoreState(automation *types.Automation, restoreState *types.RestoreStateData) {
	var lastTriggered time.Time
	if restoreState.State.Attributes.LastTriggered != nil {
		var err error
		lastTriggered, err = time.Parse(time.RFC3339, *restoreState.State.Attributes.LastTriggered)
		if err != nil {
			log.Printf("Error parsing lastTriggered for automation ID %s: %v", automation.ID, err)
		}
	}

	automation.LastTriggered = lastTriggered
	automation.FriendlyName = *restoreState.State.Attributes.FriendlyName
	automation.State = restoreState.State.State
}

func findRestoreStateByID(
	restoreState types.RestoreStateResponse,
	id string,
) (*types.RestoreStateData, bool) {
	for _, data := range restoreState.Data {
		if data.State.Attributes.ID != nil && *data.State.Attributes.ID == id {
			return &data, true
		}
	}
	return nil, false
}

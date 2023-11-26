package scriptgatherer

import (
	"os"
	"time"

	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type ScriptGatherer struct {
	Logger *logrus.Logger
}

func NewScriptGatherer(logger *logrus.Logger) *ScriptGatherer {
	return &ScriptGatherer{
		Logger: logger,
	}
}

func (s *ScriptGatherer) GatherScripts(configPath string, restoreState types.RestoreStateResponse) []types.Script {

	// Read scripts data from file
	scriptsFilePath := configPath + "scripts.yaml"
	scriptsData, err := os.ReadFile(scriptsFilePath)
	if err != nil {
		s.Logger.Println("Error reading scripts file:", err)
		return []types.Script{}
	}

	// Unmarshal scripts from YAML
	var scriptsMap map[string]types.Script
	if err := yaml.Unmarshal(scriptsData, &scriptsMap); err != nil {
		s.Logger.Println("Error unmarshaling scripts data:", err)
		return []types.Script{}
	}

	// Convert map to slice
	scripts := make([]types.Script, 0, len(scriptsMap))
	for _, script := range scriptsMap {
		scripts = append(scripts, script)
	}

	// Modify scripts based on restore state
	for i, script := range scripts {
		if restoreStateForScript := s.findRestoreStateForScript("script."+script.Alias, restoreState); restoreStateForScript != nil {
			var lastTriggered time.Time
			var err error

			if restoreStateForScript.State.Attributes.LastTriggered != nil {
				lastTriggered, err = time.Parse(
					time.RFC3339,
					*restoreStateForScript.State.Attributes.LastTriggered,
				)
			}
			if err != nil {
			} else {
				scripts[i].LastTriggered = lastTriggered
			}

			scripts[i].FriendlyName = *restoreStateForScript.State.Attributes.FriendlyName
			scripts[i].State = restoreStateForScript.State.State // Assuming state is of a compatible type
		}
	}

	return scripts
}

func (s *ScriptGatherer) findRestoreStateForScript(scriptAlias string, restoreState types.RestoreStateResponse) *types.RestoreStateData {
	for _, data := range restoreState.Data {
		if data.State.EntityID == scriptAlias {
			return &data
		}
	}
	return nil
}

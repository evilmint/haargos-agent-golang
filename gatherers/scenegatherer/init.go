package scenegatherer

import (
	"os"
	"time"

	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var log = logrus.New()

type SceneGatherer struct{}

func (sg *SceneGatherer) GatherScenes(configPath string, restoreState types.RestoreStateResponse) []types.Scene {
	// Read scenes data from file
	scenesFilePath := configPath + "scenes.yaml"
	scenesData, err := os.ReadFile(scenesFilePath)
	if err != nil {
		log.Println("Error reading scenes file:", err)
		return []types.Scene{}
	}

	// Unmarshal scenes from YAML
	var scenes []types.Scene
	if err := yaml.Unmarshal(scenesData, &scenes); err != nil {
		log.Println("Error unmarshaling scenes data:", err)
		return []types.Scene{}
	}

	// Modify scenes based on restore state
	for i, scene := range scenes {
		if restoreStateForScene := findRestoreStateForScene(scene.ID, restoreState); restoreStateForScene != nil {
			scenes[i].FriendlyName = *restoreStateForScene.State.Attributes.FriendlyName

			var lastTriggered time.Time
			var err error

			if restoreStateForScene.State.State != "" {
				lastTriggered, err = time.Parse(time.RFC3339, restoreStateForScene.State.State)
			}
			if err == nil {
				scenes[i].State = lastTriggered
			}
		}
	}

	return scenes
}

// Helper function to find corresponding restore state for a scene
func findRestoreStateForScene(
	sceneID string,
	restoreState types.RestoreStateResponse,
) *types.RestoreStateData {
	for _, data := range restoreState.Data {
		if data.State.Attributes.ID != nil {
			if *data.State.Attributes.ID == sceneID {
				return &data
			}
		}
	}
	return nil
}

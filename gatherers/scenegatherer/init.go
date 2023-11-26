package scenegatherer

import (
	"os"
	"time"

	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type SceneGatherer struct {
	Logger *logrus.Logger
}

func NewSceneGatherer(logger *logrus.Logger) *SceneGatherer {
	return &SceneGatherer{
		Logger: logger,
	}
}

func (sg *SceneGatherer) GatherScenes(configPath string, restoreState types.RestoreStateResponse) []types.Scene {
	scenes, err := sg.readScenesFromFile(configPath + "scenes.yaml")
	if err != nil {
		sg.Logger.Println("Error reading scenes file:", err)
		return []types.Scene{}
	}

	scenes = sg.updateFriendlyNameAndState(scenes, restoreState)
	return scenes
}

func (sg *SceneGatherer) readScenesFromFile(filePath string) ([]types.Scene, error) {
	scenesData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var scenes []types.Scene
	if err := yaml.Unmarshal(scenesData, &scenes); err != nil {
		return nil, err
	}

	return scenes, nil
}

func (sg *SceneGatherer) updateFriendlyNameAndState(scenes []types.Scene, restoreState types.RestoreStateResponse) []types.Scene {
	for i, scene := range scenes {
		restoreStateForScene := findRestoreStateForScene(scene.ID, restoreState)
		if restoreStateForScene != nil {
			scenes[i] = sg.updateSceneFriendlyNameAndState(scene, restoreStateForScene)
		}
	}
	return scenes
}

func (sg *SceneGatherer) updateSceneFriendlyNameAndState(scene types.Scene, restoreState *types.RestoreStateData) types.Scene {
	if restoreState.State.Attributes.FriendlyName != nil {
		scene.FriendlyName = *restoreState.State.Attributes.FriendlyName
	}

	if restoreState.State.State != "" {
		lastTriggered, err := time.Parse(time.RFC3339, restoreState.State.State)
		if err == nil {
			scene.State = lastTriggered
		}
	}

	return scene
}

func findRestoreStateForScene(sceneID string, restoreState types.RestoreStateResponse) *types.RestoreStateData {
	for _, data := range restoreState.Data {
		if data.State.Attributes.ID != nil && *data.State.Attributes.ID == sceneID {
			return &data
		}
	}
	return nil
}

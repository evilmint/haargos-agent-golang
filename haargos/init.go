package haargos

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	"time"

	"github.com/evilmint/haargos-agent-golang/client"
	"github.com/evilmint/haargos-agent-golang/gatherers/automationgatherer"
	"github.com/evilmint/haargos-agent-golang/gatherers/dockergatherer"
	"github.com/evilmint/haargos-agent-golang/gatherers/environmentgatherer"
	"github.com/evilmint/haargos-agent-golang/gatherers/loggatherer"
	"github.com/evilmint/haargos-agent-golang/gatherers/scenegatherer"
	"github.com/evilmint/haargos-agent-golang/gatherers/scriptgatherer"
	"github.com/evilmint/haargos-agent-golang/gatherers/zigbeedevicegatherer"
	"github.com/evilmint/haargos-agent-golang/registry"
	"github.com/evilmint/haargos-agent-golang/repositories/commandrepository"
	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

type Haargos struct {
	EnvironmentGatherer *environmentgatherer.EnvironmentGatherer
	Logger              *logrus.Logger
}

func NewHaargos(logger *logrus.Logger, debugEnabled bool) *Haargos {
	return &Haargos{
		EnvironmentGatherer: environmentgatherer.NewEnvironmentGatherer(logger, commandrepository.NewCommandRepository(logger)),
		Logger:              logger,
	}
}

type RunParams struct {
	AgentToken   string
	AgentType    string
	HaConfigPath string
	Z2MPath      string
	ZHAPath      string
}

func (h *Haargos) fetchLogs(haConfigPath string, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	gatherer := loggatherer.NewLogGatherer(h.Logger)
	logContent := gatherer.GatherLogs(haConfigPath)
	h.Logger.Debugf("Collected logs.")
	ch <- logContent
}

func (h *Haargos) calculateDocker(ch chan types.Docker, wg *sync.WaitGroup) {
	defer wg.Done()
	h.Logger.Debugf("Analyzing Docker environment.")
	gatherer := dockergatherer.NewDockerGatherer("/var/run/docker.sock")
	dockerInfo := gatherer.GatherDocker()
	ch <- dockerInfo
}

func (h *Haargos) calculateEnvironment(ch chan types.Environment, wg *sync.WaitGroup) {
	defer wg.Done()
	h.EnvironmentGatherer.PausePeriodicTasks()
	environment := h.EnvironmentGatherer.CalculateEnvironment()
	h.EnvironmentGatherer.ResumePeriodicTasks()
	h.Logger.Debugf("Retrieved environment data.")
	ch <- environment
}

func (h *Haargos) readRestoreStateResponse(filePath string) (types.RestoreStateResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return types.RestoreStateResponse{}, fmt.Errorf("Error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	var response types.RestoreStateResponse
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&response); err != nil {
		return types.RestoreStateResponse{}, fmt.Errorf(
			"Error decoding JSON from file %s: %w",
			filePath,
			err,
		)
	}

	return response, nil
}

func (h *Haargos) calculateZigbee(haConfigPath string, z2mPath *string, zhaPath *string, ch chan types.ZigbeeStatus, wg *sync.WaitGroup) {
	defer wg.Done()
	gatherer := zigbeedevicegatherer.NewZigbeeDeviceGatherer(h.Logger)
	deviceRegistry, _ := registry.ReadDeviceRegistry(h.Logger, haConfigPath)
	entityRegistry, _ := registry.ReadEntityRegistry(haConfigPath)
	devices, err := gatherer.GatherDevices(z2mPath, zhaPath, &deviceRegistry, &entityRegistry, haConfigPath)

	if err != nil {
		h.Logger.Errorf("Error while gathering zigbee devices: %s", err)
		ch <- types.ZigbeeStatus{Devices: []types.ZigbeeDevice{}}
		return
	}

	ch <- types.ZigbeeStatus{Devices: devices}
}

func (h *Haargos) calculateHAConfig(haConfigPath string, ch chan types.HAConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	versionFilePath := path.Join(haConfigPath, ".HA_VERSION")
	versionBytes, err := os.ReadFile(versionFilePath)
	if err != nil {
		h.Logger.Errorf("Error reading HA_VERSION file: %v", err)
		ch <- types.HAConfig{}
		return
	}

	haConfig := types.HAConfig{Version: strings.TrimSpace(string(versionBytes))}
	h.Logger.Debugf("Retrieved Home Assistant configuration.")
	ch <- haConfig
}

func (h *Haargos) calculateAutomations(
	configPath string,
	restoreState types.RestoreStateResponse,
	ch chan []types.Automation,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	gatherer := automationgatherer.AutomationGatherer{}
	automations := gatherer.GatherAutomations(configPath, restoreState)

	h.Logger.Debugf("Retrieved HomeAssistant Automations (%d).", len(automations))
	ch <- automations
}

func (h *Haargos) calculateScripts(
	configPath string,
	restoreState types.RestoreStateResponse,
	ch chan []types.Script,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	gatherer := scriptgatherer.NewScriptGatherer(h.Logger)
	scripts := gatherer.GatherScripts(configPath, restoreState)

	h.Logger.Debugf("Retrieved HomeAssistant Scripts (%d).", len(scripts))
	ch <- scripts
}

func (h *Haargos) calculateScenes(
	configPath string,
	restoreState types.RestoreStateResponse,
	ch chan []types.Scene,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	gatherer := scenegatherer.NewSceneGatherer(h.Logger)
	scenes := gatherer.GatherScenes(configPath, restoreState)

	h.Logger.Debugf("Retrieved HomeAssistant Scenes (%d).", len(scenes))
	ch <- scenes
}

type AgentType string

// Define constants for AgentType.
const (
	AgentTypeBin   AgentType = "bin"
	AgentTypeAddon AgentType = "addon"
)

func (h *Haargos) Run(params RunParams) {
	var interval time.Duration

	validAgentTypes := []string{"bin", "addon", "docker"}

	isAgentTypeValid := false
	for _, t := range validAgentTypes {
		if params.AgentType == t {
			isAgentTypeValid = true
			break
		}
	}

	if !isAgentTypeValid {
		h.Logger.Fatalf("Invalid agent type.")
	}

	client := client.NewClient(params.AgentToken)
	agentConfig, err := client.FetchAgentConfig()

	if err != nil {
		h.Logger.Fatalf("Failed to fetch agent config: %s", err)
		return
	}

	// Check the environment variable for debug mode
	// if os.Getenv("DEBUG") == "true" {
	// 	interval = 1 * time.Minute
	// } else {
	// 	interval = time.Duration(agentConfig.CycleInterval) * time.Second
	// }

	// log.Errorf("cycle interval: %d", agentConfig.CycleInterval)

	interval = time.Duration(agentConfig.CycleInterval) * time.Second

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	handleTick := func() {
		var wg sync.WaitGroup
		var observation types.Observation

		restoreStateResponse, err := h.readRestoreStateResponse(
			params.HaConfigPath + ".storage/core.restore_state",
		)
		if err != nil {
			fmt.Println(err)
			return
		}

		dockerCh := make(chan types.Docker, 1)
		environmentCh := make(chan types.Environment, 1)
		zigbeeCh := make(chan types.ZigbeeStatus, 1)
		haConfigCh := make(chan types.HAConfig, 1)
		automationsCh := make(chan []types.Automation, 1)
		scriptsCh := make(chan []types.Script, 1)
		scenesCh := make(chan []types.Scene, 1)
		logsCh := make(chan string, 1)

		wg.Add(8)
		go h.calculateDocker(dockerCh, &wg)
		go h.calculateEnvironment(environmentCh, &wg)
		go h.calculateZigbee(params.HaConfigPath, &params.Z2MPath, &params.ZHAPath, zigbeeCh, &wg)
		go h.calculateHAConfig(params.HaConfigPath, haConfigCh, &wg)
		go h.calculateAutomations(params.HaConfigPath, restoreStateResponse, automationsCh, &wg)
		go h.calculateScripts(params.HaConfigPath, restoreStateResponse, scriptsCh, &wg)
		go h.calculateScenes(params.HaConfigPath, restoreStateResponse, scenesCh, &wg)
		go h.fetchLogs(params.HaConfigPath, logsCh, &wg)

		wg.Wait()

		observation.Docker = <-dockerCh
		observation.Environment = <-environmentCh
		observation.Zigbee = <-zigbeeCh
		observation.HAConfig = <-haConfigCh
		observation.Automations = <-automationsCh
		observation.Scripts = <-scriptsCh
		observation.Scenes = <-scenesCh
		observation.AgentVersion = "Release 1.0.0"
		observation.Logs = <-logsCh
		observation.AgentType = params.AgentType

		response, err := client.SendObservation(observation)

		if err != nil || response.Status != "200 OK" {
			h.Logger.Errorf("Error sending request request: %v", err)

			bodyBytes, err := io.ReadAll(response.Body)
			if err != nil {
				h.Logger.Errorf("Sending request failed: %v", err)
				return
			}

			bodyString := string(bodyBytes)
			if bodyString != "" {
				h.Logger.Errorf("Response body: %s\n", bodyString)
			}
		}

		response.Body.Close()
	}

	handleTick()

	for range ticker.C {
		handleTick()
	}
}

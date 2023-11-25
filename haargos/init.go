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
}

func NewHaargos() *Haargos {
	return &Haargos{
		EnvironmentGatherer: environmentgatherer.NewEnvironmentGatherer(&commandrepository.CommandRepository{}),
	}
}

var log = logrus.New()

type RunParams struct {
	AgentToken   string
	AgentType    string
	HaConfigPath string
	Z2MPath      string
	ZHAPath      string
}

func (h *Haargos) fetchLogs(haConfigPath string, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Infof("Calculating logs")
	gatherer := loggatherer.LogGatherer{}
	logContent := gatherer.GatherLogs(haConfigPath)

	log.Infof("Got logs")
	ch <- logContent
}

func (h *Haargos) calculateDocker(ch chan types.Docker, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Infof("Calculating docker")
	gatherer := dockergatherer.NewDockerGatherer("/var/run/docker.sock")
	dockerInfo := gatherer.GatherDocker()
	log.Infof("Got docker")
	ch <- dockerInfo
}

func (h *Haargos) calculateEnvironment(ch chan types.Environment, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Infof("Calculating environment")
	h.EnvironmentGatherer.PausePeriodicTasks()
	environment := h.EnvironmentGatherer.CalculateEnvironment()
	h.EnvironmentGatherer.ResumePeriodicTasks()
	log.Infof("Got environment")
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
	gatherer := zigbeedevicegatherer.ZigbeeDeviceGatherer{}
	deviceRegistry, _ := registry.ReadDeviceRegistry(haConfigPath)
	entityRegistry, _ := registry.ReadEntityRegistry(haConfigPath)
	devices, err := gatherer.GatherDevices(z2mPath, zhaPath, &deviceRegistry, &entityRegistry, haConfigPath)

	if err != nil {
		log.Errorf("Error while gathering zigbee devices: %s", err)
		ch <- types.ZigbeeStatus{Devices: []types.ZigbeeDevice{}}
		return
	}
	log.Infof("Got zigbee status")

	ch <- types.ZigbeeStatus{Devices: devices}
}

func (h *Haargos) calculateHAConfig(haConfigPath string, ch chan types.HAConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	versionFilePath := path.Join(haConfigPath, ".HA_VERSION")
	versionBytes, err := os.ReadFile(versionFilePath)
	if err != nil {
		log.Errorf("Error reading HA_VERSION file: %v", err)
		ch <- types.HAConfig{} // Send an empty config if there's an error
		return
	}

	// Create the HAConfig structure
	haConfig := types.HAConfig{Version: strings.TrimSpace(string(versionBytes))}
	log.Infof("Got HA config")
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

	log.Infof("Got automations")
	ch <- automations
}

func (h *Haargos) calculateScripts(
	configPath string,
	restoreState types.RestoreStateResponse,
	ch chan []types.Script,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	gatherer := scriptgatherer.ScriptGatherer{}
	scripts := gatherer.GatherScripts(configPath, restoreState)

	log.Infof("Got scripts")
	ch <- scripts
}

func (h *Haargos) calculateScenes(
	configPath string,
	restoreState types.RestoreStateResponse,
	ch chan []types.Scene,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	gatherer := scenegatherer.SceneGatherer{}
	scenes := gatherer.GatherScenes(configPath, restoreState)

	log.Infof("Got scenes")
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

	if params.AgentType != "bin" && params.AgentType != "addon" {
		log.Fatalf("Invalid agent type.")
	}

	// Check the environment variable for debug mode
	if os.Getenv("DEBUG") == "true" {
		interval = 1 * time.Minute
	} else {
		interval = 15 * time.Minute
	}

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

		client := client.NewClient()
		response, err := client.SendObservation(observation, params.AgentToken)

		if err != nil || response.Status != "200 OK" {
			log.Errorf("Error sending request request: %v", err)

			bodyBytes, err := io.ReadAll(response.Body)
			if err != nil {
				log.Errorf("Sending request failed: %v", err)
				return
			}

			bodyString := string(bodyBytes)
			if bodyString != "" {
				log.Errorf("Response body: %s\n", bodyString)
			}
		}

		response.Body.Close()
	}

	handleTick() // Call the function once before starting the ticker

	for range ticker.C {
		handleTick() // Call the function on each tick
	}
}

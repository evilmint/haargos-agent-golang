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
	websocketclient "github.com/evilmint/haargos-agent-golang/websocket-client"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
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

const (
	Production string = "production"
	Dev               = "dev"
)

type RunParams struct {
	AgentToken   string
	AgentType    string
	HaConfigPath string
	Z2MPath      string
	ZHAPath      string
	Stage        string
}

func (h *Haargos) fetchLogs(haConfigPath string, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	gatherer := loggatherer.NewLogGatherer(h.Logger)
	logContent := gatherer.GatherCoreLogs(haConfigPath)
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

	var apiURL string

	if params.Stage == Dev {
		apiURL = "https://api.dev.haargos.com/"
	} else {
		apiURL = "https://api.haargos.com/"
	}

	supervisorEndpoint := "http://supervisor/"

	supervisorToken := os.Getenv("SUPERVISOR_TOKEN")
	haargosClient := client.NewClient(apiURL, params.AgentToken)
	supervisorClient := client.NewClient(supervisorEndpoint, params.AgentToken)

	if supervisorToken != "" {
		h.Logger.Info("Supervisor token is set.")
	} else {
		h.Logger.Info("Supervisor token is not set.")
	}

	agentConfig, err := haargosClient.FetchAgentConfig()

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

	go h.sendLogsTick(params.HaConfigPath, haargosClient, supervisorClient, supervisorToken, interval)

	accessToken := os.Getenv("HA_ACCESS_TOKEN")
	haEndpoint := os.Getenv("HA_ENDPOINT")

	if accessToken != "" {
		if haEndpoint == "" {
			port := 8123

			configuration, err := h.readConfiguration(params.HaConfigPath)
			if err != nil && configuration.Http.ServerPort != nil {
				port = *configuration.Http.ServerPort
			}

			haEndpoint = fmt.Sprintf("homeassistant:%d", port)
		}

		go h.sendNotificationsTick(params.HaConfigPath, haargosClient, interval, accessToken, haEndpoint)
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

		wg.Add(7)
		go h.calculateDocker(dockerCh, &wg)
		go h.calculateEnvironment(environmentCh, &wg)
		go h.calculateZigbee(params.HaConfigPath, &params.Z2MPath, &params.ZHAPath, zigbeeCh, &wg)
		go h.calculateHAConfig(params.HaConfigPath, haConfigCh, &wg)
		go h.calculateAutomations(params.HaConfigPath, restoreStateResponse, automationsCh, &wg)
		go h.calculateScripts(params.HaConfigPath, restoreStateResponse, scriptsCh, &wg)
		go h.calculateScenes(params.HaConfigPath, restoreStateResponse, scenesCh, &wg)

		wg.Wait()

		observation.Docker = <-dockerCh
		observation.Environment = <-environmentCh
		observation.Zigbee = <-zigbeeCh
		observation.HAConfig = <-haConfigCh
		observation.Automations = <-automationsCh
		observation.Scripts = <-scriptsCh
		observation.Scenes = <-scenesCh
		observation.AgentVersion = "Release 1.0.0"
		observation.AgentType = params.AgentType

		response, err := haargosClient.SendObservation(observation)

		if err != nil || !strings.HasPrefix(response.Status, "2") {
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

func (h *Haargos) sendLogs(haConfigPath string, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {

	gatherer := loggatherer.NewLogGatherer(h.Logger)
	logContent := gatherer.GatherCoreLogs(haConfigPath)
	h.Logger.Debugf("Collected core logs.")

	coreLogs := types.Logs{Type: "core", Content: logContent}
	h.sendLogsToClient(client, coreLogs)

	if supervisorToken != "" {
		supervisorLogContent, err := gatherer.GatherSupervisorLogs(supervisorClient, supervisorToken)

		if err != nil {
			h.Logger.Errorf("Failed collecting supervisor logs")
		} else {
			h.Logger.Debugf("Collected supervisor logs.")

			supervisorLogs := types.Logs{Type: "supervisor", Content: supervisorLogContent}
			h.sendLogsToClient(client, supervisorLogs)
		}

		hostLogContent, err := gatherer.GatherHostLogs(supervisorClient, supervisorToken)

		if err != nil {
			h.Logger.Errorf("Failed collecting host logs")
		} else {
			h.Logger.Debugf("Collected host logs.")

			hostLogs := types.Logs{Type: "host", Content: hostLogContent}
			h.sendLogsToClient(client, hostLogs)
		}
	}
}

func (h *Haargos) sendLogsToClient(client *client.HaargosClient, logs types.Logs) {
	// Send the logs
	response, err := client.SendLogs(logs)

	h.Logger.Infof(fmt.Sprintf("Sending logs of type %s to %s", logs.Type, logs.Content))

	if err != nil || !strings.HasPrefix(response.Status, "2") {
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

func (h *Haargos) sendLogsTick(haConfigPath string, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string, logInterval time.Duration) {
	logTicker := time.NewTicker(logInterval)
	defer logTicker.Stop()

	h.sendLogs(haConfigPath, client, supervisorClient, supervisorToken)
	for range logTicker.C {
		h.sendLogs(haConfigPath, client, supervisorClient, supervisorToken)
	}
}

type Configuration struct {
	Http *ConfigurationHttp `yaml:"http"`
}

type ConfigurationHttp struct {
	ServerPort *int `yaml:"server_port"`
}

func (h *Haargos) readConfiguration(haConfigPath string) (*Configuration, error) {
	haConfigData, err := os.ReadFile(haConfigPath + "/configuration.yaml")

	if err != nil {
		return nil, err
	}

	haConfig := Configuration{}
	err = yaml.Unmarshal(haConfigData, &haConfig)

	return &haConfig, err
}

func (h *Haargos) sendNotifications(haConfigPath string, client *client.HaargosClient, accessToken string, endpoint string) {
	wsClient := websocketclient.NewWebSocketClient(fmt.Sprintf("ws://%s/api/websocket", endpoint))
	notification, err := wsClient.FetchNotifications(accessToken)

	if err != nil {
		h.Logger.Fatalf("Error fetching notifications: %v", err)
	} else {
		h.Logger.Infof("Read %d notifications", len(notification.Event.Notifications))

		// Send notifications
		notifications := make([]websocketclient.WSAPINotificationDetails, 0, len(notification.Event.Notifications))

		for _, notification := range notification.Event.Notifications {
			notifications = append(notifications, notification)
		}

		response, err := client.SendNotifications(notifications)
		if err != nil || !strings.HasPrefix(response.Status, "2") {
			h.Logger.Errorf("Error sending request request: %v", err)

			bodyBytes, err := io.ReadAll(response.Body)
			if err != nil {
				h.Logger.Errorf("Sending request failed: %v", err)
				return
			}

			bodyString := string(bodyBytes)
			if bodyString != "" {
				h.Logger.Errorf("Response body: %s", bodyString)
			}
		}

		response.Body.Close()

		h.Logger.Infof("Sent notifications")
	}
}

func (h *Haargos) sendNotificationsTick(haConfigPath string, client *client.HaargosClient, notificationsInterval time.Duration, accessToken string, endpoint string) {
	notificationsTicker := time.NewTicker(notificationsInterval)
	defer notificationsTicker.Stop()

	h.sendNotifications(haConfigPath, client, accessToken, endpoint)
	for range notificationsTicker.C {
		h.sendNotifications(haConfigPath, client, accessToken, endpoint)
	}
}

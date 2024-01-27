package haargos

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"time"

	"github.com/evilmint/haargos-agent-golang/client"
	"github.com/evilmint/haargos-agent-golang/gatherers/automationgatherer"
	"github.com/evilmint/haargos-agent-golang/gatherers/dockergatherer"
	"github.com/evilmint/haargos-agent-golang/gatherers/environmentgatherer"
	jobrunner "github.com/evilmint/haargos-agent-golang/gatherers/job-runner"
	"github.com/evilmint/haargos-agent-golang/gatherers/loggatherer"
	"github.com/evilmint/haargos-agent-golang/gatherers/scenegatherer"
	"github.com/evilmint/haargos-agent-golang/gatherers/scriptgatherer"
	"github.com/evilmint/haargos-agent-golang/gatherers/zigbeedevicegatherer"
	"github.com/evilmint/haargos-agent-golang/ingress"
	"github.com/evilmint/haargos-agent-golang/registry"
	"github.com/evilmint/haargos-agent-golang/repositories/commandrepository"
	"github.com/evilmint/haargos-agent-golang/statistics"
	"github.com/evilmint/haargos-agent-golang/types"
	websocketclient "github.com/evilmint/haargos-agent-golang/websocket-client"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Haargos struct {
	environmentGatherer *environmentgatherer.EnvironmentGatherer
	logger              *logrus.Logger
	ingress             *ingress.Ingress
	statistics          *statistics.Statistics
	jobRunner           *jobrunner.JobRunner
}

func NewHaargos(logger *logrus.Logger, debugEnabled bool) *Haargos {
	return &Haargos{
		environmentGatherer: environmentgatherer.NewEnvironmentGatherer(logger, commandrepository.NewCommandRepository(logger)),
		logger:              logger,
		statistics:          statistics.NewStatistics(),
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

	gatherer := loggatherer.NewLogGatherer(h.logger)
	logContent := gatherer.GatherCoreLogs(haConfigPath)
	h.logger.Debugf("Collected logs.")
	ch <- logContent
}

func (h *Haargos) calculateDocker(ch chan types.Docker, wg *sync.WaitGroup) {
	defer wg.Done()
	h.logger.Debugf("Analyzing Docker environment.")
	gatherer := dockergatherer.NewDockerGatherer("/var/run/docker.sock")
	dockerInfo := gatherer.GatherDocker()
	ch <- dockerInfo
}

func (h *Haargos) calculateEnvironment(ch chan types.Environment, wg *sync.WaitGroup) {
	defer wg.Done()
	h.environmentGatherer.PausePeriodicTasks()
	environment := h.environmentGatherer.CalculateEnvironment()
	h.environmentGatherer.ResumePeriodicTasks()
	h.logger.Debugf("Retrieved environment data.")
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
	gatherer := zigbeedevicegatherer.NewZigbeeDeviceGatherer(h.logger)
	deviceRegistry, _ := registry.ReadDeviceRegistry(h.logger, haConfigPath)
	entityRegistry, _ := registry.ReadEntityRegistry(haConfigPath)
	devices, err := gatherer.GatherDevices(z2mPath, zhaPath, &deviceRegistry, &entityRegistry, haConfigPath)

	if err != nil {
		h.logger.Errorf("Error while gathering zigbee devices: %s", err)
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
		h.logger.Errorf("Error reading HA_VERSION file: %v", err)
		ch <- types.HAConfig{}
		return
	}

	haConfig := types.HAConfig{Version: strings.TrimSpace(string(versionBytes))}
	h.logger.Debugf("Retrieved Home Assistant configuration.")
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

	h.logger.Debugf("Retrieved HomeAssistant Automations (%d).", len(automations))
	ch <- automations
}

func (h *Haargos) calculateScripts(
	configPath string,
	restoreState types.RestoreStateResponse,
	ch chan []types.Script,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	gatherer := scriptgatherer.NewScriptGatherer(h.logger)
	scripts := gatherer.GatherScripts(configPath, restoreState)

	h.logger.Debugf("Retrieved HomeAssistant Scripts (%d).", len(scripts))
	ch <- scripts
}

func (h *Haargos) calculateScenes(
	configPath string,
	restoreState types.RestoreStateResponse,
	ch chan []types.Scene,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	gatherer := scenegatherer.NewSceneGatherer(h.logger)
	scenes := gatherer.GatherScenes(configPath, restoreState)

	h.logger.Debugf("Retrieved HomeAssistant Scenes (%d).", len(scenes))
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

	h.validateAgentType(params.AgentType)

	var apiURL string

	if params.Stage == Dev {
		apiURL = "https://api.dev.haargos.com/"
	} else {
		apiURL = "https://api.haargos.com/"
	}

	supervisorEndpoint := "http://supervisor/"

	supervisorToken := os.Getenv("SUPERVISOR_TOKEN")
	haargosClient := client.NewClient(apiURL, params.AgentToken, func(number int) {
		h.statistics.AddDataSentInKB(number)
	})
	supervisorClient := client.NewClient(supervisorEndpoint, params.AgentToken, func(number int) {
		h.statistics.AddDataSentInKB(number)
	})

	h.jobRunner = jobrunner.NewJobRunner(h.logger, haargosClient, supervisorClient, h.statistics)

	if supervisorToken != "" {
		h.logger.Info("Supervisor token is set.")
	} else {
		h.logger.Info("Supervisor token is not set.")
	}

	agentConfig, err := haargosClient.FetchAgentConfig()

	if err != nil {
		h.logger.Fatalf("Failed to fetch agent config: %s", err)
		return
	}

	version := h.getAgentVersion()

	interval = time.Duration(agentConfig.CycleInterval) * time.Second

	runTicker(interval, func() {
		h.sendLogs(params.HaConfigPath, haargosClient, supervisorClient, supervisorToken)
	})
	runTicker(interval, func() {
		h.sendAddons(params.HaConfigPath, haargosClient, supervisorClient, supervisorToken)
	})
	runTicker(interval, func() {
		h.sendOS(params.HaConfigPath, haargosClient, supervisorClient, supervisorToken)
	})
	runTicker(interval, func() {
		h.sendSupervisor(params.HaConfigPath, haargosClient, supervisorClient, supervisorToken)
	})
	runTicker(3*time.Minute, func() {
		h.jobRunner.HandleJobs(params.HaConfigPath, supervisorToken)
	})

	accessToken := os.Getenv("HA_ACCESS_TOKEN")
	haEndpoint := os.Getenv("HA_ENDPOINT")

	isAccessTokenSet := accessToken != ""
	h.statistics.SetHAAccessTokenSet(isAccessTokenSet)

	if isAccessTokenSet {
		if haEndpoint == "" {
			port := 8123

			configuration, err := h.readConfiguration(params.HaConfigPath)
			if err != nil && configuration.Http.ServerPort != nil {
				port = *configuration.Http.ServerPort
			}

			haEndpoint = fmt.Sprintf("homeassistant:%d", port)
		}

		runTicker(interval, func() {
			h.sendNotifications(params.HaConfigPath, haargosClient, accessToken, haEndpoint)
		})
	}
	h.ingress = ingress.NewIngress(h.statistics)
	go h.ingress.Run()

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
		observation.AgentVersion = version
		observation.AgentType = params.AgentType

		response, err := haargosClient.SendObservation(observation)
		h.handleHttpResponse(response, err, h.logger, "sending observation")

		if err == nil {
			h.statistics.IncrementObservationsSentCount()
		} else {
			h.statistics.IncrementFailedRequestCount()
		}
	}

	handleTick()

	for range ticker.C {
		handleTick()
	}
}

func (h *Haargos) getAgentVersion() string {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		h.logger.Fatalf("Failed to read agent version.")
	}

	version := string(data)
	return version
}

func (h *Haargos) validateAgentType(agentType string) {
	validAgentTypes := []string{"bin", "addon", "docker"}

	isAgentTypeValid := false
	for _, t := range validAgentTypes {
		if agentType == t {
			isAgentTypeValid = true
			break
		}
	}

	if !isAgentTypeValid {
		h.logger.Fatalf("Invalid agent type.")
	}
}

type LogFetchType struct {
	logType string
}

func (h *Haargos) sendLogs(haConfigPath string, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {

	gatherer := loggatherer.NewLogGatherer(h.logger)
	logContent := gatherer.GatherCoreLogs(haConfigPath)
	h.logger.Debugf("Collected core logs.")

	coreLogs := types.Logs{Type: "core", Content: logContent}
	h.sendLogsToClient(client, coreLogs)

	if supervisorToken != "" {
		var fetchTypes = [6]LogFetchType{
			{logType: "core"},
			{logType: "host"},
			{logType: "supervisor"},
			{logType: "multicast"},
			{logType: "audio"},
			{logType: "dns"},
		}

		for _, fetchType := range fetchTypes {
			supervisorLogContent, err := gatherer.GatherHassioLogs(supervisorClient, supervisorToken, fetchType.logType)

			if err != nil {
				h.logger.Errorf("Failed collecting %s logs", fetchType.logType)
			} else {
				h.logger.Debugf("Collected %s logs.", fetchType.logType)

				logs := types.Logs{Type: fetchType.logType, Content: supervisorLogContent}
				h.sendLogsToClient(client, logs)
			}
		}
	}
}

func (h *Haargos) sendLogsToClient(client *client.HaargosClient, logs types.Logs) {
	response, err := client.SendLogs(logs)
	h.handleHttpResponse(response, err, h.logger, "sending logs")

	if err != nil {
		h.statistics.IncrementFailedRequestCount()
	}
}

func (h *Haargos) sendSupervisor(haConfigPath string, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {
	supervisor, err := supervisorClient.FetchSupervisor(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)})

	if err != nil || supervisor == nil {
		h.logger.Errorf("Failed collecting supervisor %s", err)
	} else {
		h.logger.Debugf("Collected supervisor.")

		response, err := client.SendSupervisor(*supervisor)
		h.handleHttpResponse(response, err, h.logger, "sending supervisor")

		if err != nil {
			h.statistics.IncrementFailedRequestCount()
		}
	}
}

func (h *Haargos) sendOS(haConfigPath string, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {
	osContent, err := supervisorClient.FetchOS(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)})

	if err != nil || osContent == nil {
		h.logger.Errorf("Failed collecting os %s", err)
	} else {
		h.logger.Debugf("Collected os.")

		response, err := client.SendOS(*osContent)
		h.handleHttpResponse(response, err, h.logger, "sending os")

		if err != nil {
			h.statistics.IncrementFailedRequestCount()
		}
	}
}

func (h *Haargos) sendAddons(haConfigPath string, client *client.HaargosClient, supervisorClient *client.HaargosClient, supervisorToken string) {
	addonContent, err := supervisorClient.FetchAddons(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", supervisorToken)})

	if err != nil || addonContent == nil {
		h.logger.Errorf("Failed collecting addons %s", err)
	} else {
		h.logger.Debugf("Collected %d addons.", len(*addonContent))

		response, err := client.SendAddons(*addonContent)
		h.handleHttpResponse(response, err, h.logger, "sending addons")

		if err != nil {
			h.statistics.IncrementFailedRequestCount()
		}
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
		h.logger.Fatalf("Error fetching notifications: %v", err)
	} else {
		h.logger.Infof("Read %d notifications", len(notification.Event.Notifications))

		notifications := make([]websocketclient.WSAPINotificationDetails, 0, len(notification.Event.Notifications))

		for _, notification := range notification.Event.Notifications {
			notifications = append(notifications, notification)
		}

		response, err := client.SendNotifications(notifications)
		h.handleHttpResponse(response, err, h.logger, "sending notifications")

		if err != nil {
			h.statistics.IncrementFailedRequestCount()
		}

		h.logger.Infof("Sent notifications")
	}
}

func (h *Haargos) handleHttpResponse(response *http.Response, err error, logger *logrus.Logger, context string) {
	if err != nil || !strings.HasPrefix(response.Status, "2") {
		logger.Errorf("Error in %s: %v", context, err)

		if response != nil {
			bodyBytes, err := io.ReadAll(response.Body)
			if err == nil {
				logger.Errorf("Response body in %s: %s", context, string(bodyBytes))
			}
		}
	} else {
		h.statistics.SetLastSuccessfulConnection(time.Now())
	}

	if response != nil {
		response.Body.Close()
	}
}

func runTicker(interval time.Duration, action func()) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		action() // Execute the action once immediately before starting the ticker loop

		for range ticker.C {
			action()
		}
	}()
}

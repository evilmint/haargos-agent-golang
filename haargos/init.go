package haargos

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"

	"time"

	"github.com/evilmint/haargos-agent-golang/client"
	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

type Haargos struct {
	// Additional fields can be added here if needed
}

var log = logrus.New()

type RunParams struct {
	UserID         string
	InstallationID string
	Token          string
	HaConfigPath   string
}

func fetchLogs(haConfigPath string, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open(haConfigPath + "home-assistant.log")
	if err != nil {
		log.Errorf("Error reading log file: %v", err)
		ch <- ""
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var logLines []string
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) >= 3 && (parts[2] == "WARNING" || parts[2] == "ERROR") {
			logLines = append(logLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("Error scanning log file: %v", err)
		ch <- ""
		return
	}

	if len(logLines) > 200 {
		logLines = logLines[len(logLines)-200:]
	}

	// Join them by newline
	logContent := strings.Join(logLines, "\n")
	ch <- logContent
}

func calculateDocker(ch chan types.Docker, wg *sync.WaitGroup) {
	defer wg.Done()
	// Calculate Docker information here
	docker := types.Docker{Containers: []types.DockerContainer{}}
	ch <- docker
}

func getMemoryInfo() (types.Memory, error) {
	out, err := exec.Command("bash", "-c", "free").Output()
	if err != nil {
		return types.Memory{}, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) >= 3 {
		memInfo := strings.Fields(lines[1])
		if len(memInfo) >= 7 {
			total, err := strconv.Atoi(memInfo[1])
			if err != nil {
				return types.Memory{}, err
			}
			used, err := strconv.Atoi(memInfo[2])
			if err != nil {
				log.Errorf("ER1")
				return types.Memory{}, err
			}
			free, err := strconv.Atoi(memInfo[3])
			if err != nil {
				log.Errorf("ER2")
				return types.Memory{}, err
			}
			shared, err := strconv.Atoi(memInfo[4])
			if err != nil {
				log.Errorf("ER3")
				return types.Memory{}, err
			}
			buffCache, err := strconv.Atoi(memInfo[5])
			if err != nil {
				log.Errorf("ER4")
				return types.Memory{}, err
			}
			available, err := strconv.Atoi(memInfo[6])
			if err != nil {
				log.Errorf("ER5")
				return types.Memory{}, err
			}

			return types.Memory{
				Total:     total,
				Used:      used,
				Free:      free,
				Shared:    shared,
				BuffCache: buffCache,
				Available: available,
			}, nil
		}
	}

	return types.Memory{}, fmt.Errorf("Failed to parse memory info")
}

func getFileSystems() ([]types.Storage, error) {
	out, err := exec.Command("bash", "-c", "df -h").Output()
	if err != nil {
		return nil, err
	}

	var fileSystems []types.Storage
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, line := range lines {
		if i == 0 {
			continue
		}

		components := strings.Fields(line)
		if len(components) >= 6 {
			fileSystem := types.Storage{
				Name:          components[0],
				Size:          components[1],
				Used:          components[2],
				Available:     components[3],
				UsePercentage: components[4],
				MountedOn:     components[5],
			}
			fileSystems = append(fileSystems, fileSystem)
		}
	}

	return fileSystems, nil
}

func getCPUDetails() (types.CPU, error) {
	topOut, err := exec.Command("bash", "-c", "top -bn 1 | awk 'NR == 3 {printf \"%.2f\", 100 - $8}'").Output()
	if err != nil {
		return types.CPU{}, err
	}

	load, _ := strconv.ParseFloat(strings.TrimSpace(string(topOut)), 64)

	cpuOut, err := exec.Command("bash", "-c", "lscpu | grep -E 'Architecture|Model name|CPU MHz|CPU(s)' | sed 's/   *//g'").Output()
	if err != nil {
		return types.CPU{}, err
	}

	var architecture, modelName, cpuMHz string

	lines := strings.Split(strings.TrimSpace(string(cpuOut)), "\n")
	for _, line := range lines {
		components := strings.SplitN(line, ":", 2)
		if len(components) == 2 {
			key := strings.TrimSpace(components[0])
			value := strings.TrimSpace(components[1])

			switch key {
			case "Architecture":
				architecture = value
			case "Model name":
				modelName = value
			case "CPU MHz":
				cpuMHz = value
			}
		}
	}

	return types.CPU{
		Architecture: architecture,
		ModelName:    modelName,
		CPUMHz:       cpuMHz,
		Load:         load,
	}, nil
}

func calculateEnvironment(ch chan types.Environment, wg *sync.WaitGroup) {
	defer wg.Done()

	memory, err := getMemoryInfo()
	if err != nil {
		log.Errorf("Error getting memory info: %v", err)
		return
	}

	fileSystems, err := getFileSystems()
	if err != nil {
		log.Errorf("Error getting file systems: %v", err)
		return
	}

	cpuDetails, err := getCPUDetails()
	if err != nil {
		log.Errorf("Error getting CPU details: %v", err)
		return
	}

	environment := types.Environment{
		Memory:  memory,
		CPU:     cpuDetails,
		Storage: fileSystems,
	}

	ch <- environment
}

func calculateZigbee(ch chan types.ZigbeeStatus, wg *sync.WaitGroup) {
	defer wg.Done()
	// Calculate ZigbeeStatus information here
	var nameByUser = "sd"
	var powerSource = "Battery"
	var entityName = "entity.name"
	var batteryLevel = "87"
	zigbee := types.ZigbeeStatus{Devices: []types.ZigbeeDevice{
		{
			Ieee:            "84:fd:27:ff:fe:6d:be:fa",
			Lqi:             0,
			LastUpdated:     time.Now(),
			NameByUser:      &nameByUser,
			PowerSource:     &powerSource,
			EntityName:      entityName,
			Brand:           "Yes",
			IntegrationType: "Z2M",
			BatteryLevel:    &batteryLevel,
		},
	}}
	ch <- zigbee
}

func calculateHAConfig(haConfigPath string, ch chan types.HAConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	// Read the contents of the ".HA_VERSION" file
	versionFilePath := path.Join(haConfigPath, ".HA_VERSION")
	versionBytes, err := os.ReadFile(versionFilePath)
	if err != nil {
		log.Errorf("Error reading HA_VERSION file: %v", err)
		ch <- types.HAConfig{} // Send an empty config if there's an error
		return
	}

	// Create the HAConfig structure
	haConfig := types.HAConfig{Version: strings.TrimSpace(string(versionBytes))}
	ch <- haConfig
}

func calculateAutomations(ch chan []types.Automation, wg *sync.WaitGroup) {
	defer wg.Done()
	// Calculate Automations information here
	automations := []types.Automation{
		{
			LastTriggered: time.Now(),
			Description:   "hi",
			ID:            "Some id",
			State:         "on",
			Alias:         "alias",
			FriendlyName:  "Friendly name",
		},
	}
	ch <- automations
}

func calculateScripts(ch chan []types.Script, wg *sync.WaitGroup) {
	defer wg.Done()
	// Calculate Scripts information here
	scripts := []types.Script{
		{LastTriggered: time.Now(), State: "on", Alias: "alias", FriendlyName: "Friendly name"},
	}
	ch <- scripts
}

func calculateScenes(ch chan []types.Scene, wg *sync.WaitGroup) {
	defer wg.Done()
	// Calculate Scenes information here
	scenes := []types.Scene{
		{Name: "Scene name", ID: "Some id", FriendlyName: "Friendly name"},
	}
	ch <- scenes
}

func (h *Haargos) Run(params RunParams) {
	var interval time.Duration

	// Check the environment variable for debug mode
	if os.Getenv("DEBUG") == "true" {
		interval = 1 * time.Minute
	} else {
		interval = 1 * time.Hour
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	handleTick := func() {
		var wg sync.WaitGroup
		var observation types.Observation

		dockerCh := make(chan types.Docker, 1)
		environmentCh := make(chan types.Environment, 1)
		zigbeeCh := make(chan types.ZigbeeStatus, 1)
		haConfigCh := make(chan types.HAConfig, 1)
		automationsCh := make(chan []types.Automation, 1)
		scriptsCh := make(chan []types.Script, 1)
		scenesCh := make(chan []types.Scene, 1)
		logsCh := make(chan string, 1)

		wg.Add(8)
		go calculateDocker(dockerCh, &wg)
		go calculateEnvironment(environmentCh, &wg)
		go calculateZigbee(zigbeeCh, &wg)
		go calculateHAConfig(params.HaConfigPath, haConfigCh, &wg)
		go calculateAutomations(automationsCh, &wg)
		go calculateScripts(scriptsCh, &wg)
		go calculateScenes(scenesCh, &wg)
		go fetchLogs(params.HaConfigPath, logsCh, &wg)

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
		observation.InstallationID = params.InstallationID

		response, err := client.SendObservation(observation, params.UserID, params.Token)

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

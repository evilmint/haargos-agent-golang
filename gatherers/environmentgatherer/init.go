package environmentgatherer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/evilmint/haargos-agent-golang/repositories/commandrepository"
	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

type EnvironmentGatherer struct {
	Logger            *logrus.Logger
	commandRepository *commandrepository.CommandRepository
	cpuLoadManager    *CPULoadManager
}

func NewEnvironmentGatherer(logger *logrus.Logger, commandRepo *commandrepository.CommandRepository) *EnvironmentGatherer {
	gatherer := &EnvironmentGatherer{
		Logger:            logger,
		commandRepository: commandRepo,
		cpuLoadManager:    NewCPULoadManager(logger, commandRepo),
	}

	return gatherer
}

func (e *EnvironmentGatherer) getMemoryInfo() (*types.Memory, error) {
	out, err := e.commandRepository.GetMemory()
	if err != nil {
		return nil, fmt.Errorf("Error getting memory info: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(*out), "\n")
	if len(lines) < 2 {
		return nil, errors.New("Failed to parse memory info: insufficient data")
	}

	// Parsing RAM info
	memInfo := strings.Fields(lines[1])
	if len(memInfo) < 7 {
		return nil, errors.New("Failed to parse memory info: unexpected format")
	}

	memory := &types.Memory{}
	fields := []*int{&memory.Total, &memory.Used, &memory.Free, &memory.Shared, &memory.BuffCache, &memory.Available}
	for i, field := range fields {
		if *field, err = strconv.Atoi(memInfo[i+1]); err != nil {
			return nil, fmt.Errorf("Failed to parse memory info at field %d: %v", i+1, err)
		}
	}

	// Parsing SWAP info
	swapInfo := strings.Fields(lines[2])
	if len(swapInfo) >= 3 {
		memory.SwapTotal, err = strconv.Atoi(swapInfo[1])
		if err != nil {
			e.Logger.Errorf("Failed to parse swap total: %v", err)
		}
		memory.SwapUsed, err = strconv.Atoi(swapInfo[2])
		if err != nil {
			e.Logger.Errorf("Failed to parse swap used: %v", err)
		}
	}

	return memory, nil
}

func (e *EnvironmentGatherer) getFileSystems() ([]types.Storage, error) {
	storage, err := e.commandRepository.GetStorage()
	if err != nil {
		return nil, fmt.Errorf("Error getting storage info: %v", err)
	}

	var fileSystems []types.Storage
	lines := strings.Split(strings.TrimSpace(*storage), "\n")
	if len(lines) < 2 {
		return nil, errors.New("Insufficient data in storage info")
	}

	// Parse header to get the index of each column
	header := strings.Fields(lines[0])
	columnIndices := map[string]int{}
	for i, columnName := range header {
		lowerColumnName := strings.ToLower(columnName)
		switch lowerColumnName {
		case "filesystem":
			columnIndices["name"] = i
		case "size":
			columnIndices["size"] = i
		case "used":
			columnIndices["used"] = i
		case "avail":
			columnIndices["available"] = i
		case "use%", "capacity":
			columnIndices["usepercentage"] = i
		case "mounted":
			columnIndices["mountedon"] = i
		}
	}

	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		fileSystem := types.Storage{
			Name:          fields[columnIndices["name"]],
			Size:          fields[columnIndices["size"]],
			Used:          fields[columnIndices["used"]],
			Available:     fields[columnIndices["available"]],
			UsePercentage: fields[columnIndices["usepercentage"]],
			MountedOn:     fields[columnIndices["mountedon"]],
		}

		fileSystems = append(fileSystems, fileSystem)
	}

	return fileSystems, nil
}

func (e *EnvironmentGatherer) getCPUDetails() (*types.CPU, error) {
	load := e.cpuLoadManager.GetLastCPULoad()

	cpuInfo, err := e.commandRepository.GetCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("Error getting CPU info: %v", err)
	}

	cpuDetails := &types.CPU{ModelName: cpuInfo.Model, Architecture: cpuInfo.Architecture, Load: load, CPUMHz: cpuInfo.MHz}

	return cpuDetails, nil
}

func (e *EnvironmentGatherer) getCPUTemperature() (float64, error) {
	tempStr, err := e.commandRepository.GetCPUTemperature()
	if err != nil {
		return 0, fmt.Errorf("Error getting CPU temperature: %v", err)
	}

	temp, err := strconv.ParseFloat(strings.TrimSpace(*tempStr), 64)
	if err != nil {
		return 0, fmt.Errorf("Error parsing CPU temperature: %v", err)
	}

	return temp, nil
}

func (e *EnvironmentGatherer) getLastBootTime() (string, error) {
	bootTime, err := e.commandRepository.GetLastBootTime()
	if err != nil {
		return "", fmt.Errorf("Error getting last boot time: %v", err)
	}

	return strings.TrimSpace(*bootTime), nil
}

func (e *EnvironmentGatherer) CalculateEnvironment() types.Environment {
	environment := types.Environment{}

	memory, err := e.getMemoryInfo()
	if err != nil {
		e.Logger.Error(err)
	} else {
		environment.Memory = memory
	}

	fileSystems, err := e.getFileSystems()
	if err != nil {
		e.Logger.Error(err)
	} else {
		environment.Storage = fileSystems
	}

	bootTime, err := e.getLastBootTime()
	if err != nil {
		e.Logger.Error(err)
	} else {
		environment.BootTime = bootTime
	}

	cpuDetails, err := e.getCPUDetails()
	if err != nil {
		e.Logger.Error(err)
	} else {
		environment.CPU = cpuDetails

		cpuTemp, err := e.getCPUTemperature()
		if err != nil {
			e.Logger.Error(err)
		} else {
			environment.CPU.Temperature = cpuTemp
		}
	}

	network, err := e.getNetworkDetails()
	if err != nil {
		e.Logger.Error(err)
	} else {
		environment.Network = network
	}

	return environment
}

func (e *EnvironmentGatherer) PausePeriodicTasks() {
	e.cpuLoadManager.Stop()
}

func (e *EnvironmentGatherer) ResumePeriodicTasks() {
	e.cpuLoadManager.Start()
}

func (e *EnvironmentGatherer) getNetworkDetails() (*types.Network, error) {
	networkInterfaces, err := e.commandRepository.GetNetworkInterfaces()
	if err != nil {
		return nil, fmt.Errorf("Error getting network interfaces: %v", err)
	}

	var networks *types.Network = &types.Network{Interfaces: []types.NetworkInterface{}}

	interfaces := strings.Split(strings.TrimSpace(*networkInterfaces), "\n")

	for _, iface := range interfaces {
		rxBytes, err := e.commandRepository.GetRXTXBytes(iface, "rx")
		if err != nil {
			e.Logger.Errorf("Failed to fetch RX bytes for interface %s: %v", iface, err)
			continue
		}

		txBytes, err := e.commandRepository.GetRXTXBytes(iface, "tx")
		if err != nil {
			e.Logger.Errorf("Failed to fetch TX bytes for interface %s: %v", iface, err)
			continue
		}
		rxPackets, err := e.commandRepository.GetRXTXPackets(iface, "rx")
		if err != nil {
			e.Logger.Errorf("Failed to fetch RX packets for interface %s: %v", iface, err)
			continue
		}

		txPackets, err := e.commandRepository.GetRXTXPackets(iface, "tx")
		if err != nil {
			e.Logger.Errorf("Failed to fetch TX packets for interface %s: %v", iface, err)
			continue
		}

		networks.Interfaces = append(networks.Interfaces, types.NetworkInterface{
			Name: iface,
			Rx: &types.NetworkInterfaceData{
				Bytes:   *rxBytes,
				Packets: *rxPackets,
			},
			Tx: &types.NetworkInterfaceData{
				Bytes:   *txBytes,
				Packets: *txPackets,
			},
		})
	}

	return networks, nil
}

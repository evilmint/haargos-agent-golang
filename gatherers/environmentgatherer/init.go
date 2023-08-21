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

var log = logrus.New()

type EnvironmentGatherer struct {
	commandRepository *commandrepository.CommandRepository
}

func (e *EnvironmentGatherer) getMemoryInfo() (*types.Memory, error) {
	out, err := e.commandRepository.GetMemory()
	if err != nil {
		return nil, fmt.Errorf("Error getting memory info: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(*out), "\n")
	if len(lines) < 3 {
		return nil, errors.New("Failed to parse memory info: insufficient data")
	}

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
	top, err := e.commandRepository.GetTopBatch()
	if err != nil {
		return nil, fmt.Errorf("Error getting CPU load: %v", err)
	}

	load, err := strconv.ParseFloat(strings.TrimSpace(*top), 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing CPU load: %v", err)
	}

	cpuInfo, err := e.commandRepository.GetCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("Error getting CPU info: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(*cpuInfo), "\n")
	cpuDetails := &types.CPU{Load: load}
	for _, line := range lines {
		components := strings.SplitN(line, ":", 2)
		if len(components) != 2 {
			continue
		}

		key := strings.TrimSpace(components[0])
		value := strings.TrimSpace(components[1])
		switch key {
		case "Architecture":
			cpuDetails.Architecture = value
		case "Model name":
			cpuDetails.ModelName = value
		case "CPU MHz":
			cpuDetails.CPUMHz = value
		}
	}

	return cpuDetails, nil
}

func (e *EnvironmentGatherer) CalculateEnvironment() types.Environment {
	environment := types.Environment{}

	memory, err := e.getMemoryInfo()
	if err != nil {
		log.Error(err)
	} else {
		environment.Memory = memory
	}

	fileSystems, err := e.getFileSystems()
	if err != nil {
		log.Error(err)
	} else {
		environment.Storage = fileSystems
	}

	cpuDetails, err := e.getCPUDetails()
	if err != nil {
		log.Error(err)
	} else {
		environment.CPU = cpuDetails
	}

	return environment
}

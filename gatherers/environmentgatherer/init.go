package environmentgatherer

import (
	"errors"
	"fmt"
	"regexp"
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
		return make([]types.Storage, 0), fmt.Errorf("Error getting storage info: %v", err)
	}

	var fileSystems []types.Storage
	lines := strings.Split(strings.TrimSpace(*storage), "\n")
	for i, line := range lines {
		if i == 0 {
			continue
		}

		re := regexp.MustCompile(`(.*?)\s+(\d+\S*)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.*)`)
		matches := re.FindStringSubmatch(line)

		if len(matches) >= 7 {
			fileSystems = append(fileSystems, types.Storage{
				Name:          matches[1],
				Size:          matches[2],
				Used:          matches[3],
				Available:     matches[4],
				UsePercentage: matches[5],
				MountedOn:     matches[6],
			})
		} else {
			log.Errorf("Invalid number of matches when collecting file systems.")
		}
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

package environmentgatherer

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type EnvironmentGatherer struct{}

func (e *EnvironmentGatherer) getMemoryInfo() (*types.Memory, error) {
	out, err := exec.Command("bash", "-c", "free").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) >= 3 {
		memInfo := strings.Fields(lines[1])
		if len(memInfo) >= 7 {
			total, err := strconv.Atoi(memInfo[1])
			if err != nil {
				return nil, err
			}
			used, err := strconv.Atoi(memInfo[2])
			if err != nil {
				return nil, err
			}
			free, err := strconv.Atoi(memInfo[3])
			if err != nil {
				return nil, err
			}
			shared, err := strconv.Atoi(memInfo[4])
			if err != nil {
				return nil, err
			}
			buffCache, err := strconv.Atoi(memInfo[5])
			if err != nil {
				return nil, err
			}
			available, err := strconv.Atoi(memInfo[6])
			if err != nil {
				return nil, err
			}

			return &types.Memory{
				Total:     total,
				Used:      used,
				Free:      free,
				Shared:    shared,
				BuffCache: buffCache,
				Available: available,
			}, nil
		}
	}

	return nil, fmt.Errorf("Failed to parse memory info")
}

func (e *EnvironmentGatherer) getFileSystems() ([]types.Storage, error) {
	out, err := exec.Command("bash", "-c", "df -h").Output()
	if err != nil {
		return make([]types.Storage, 0), err
	}

	var fileSystems = make([]types.Storage, 0)

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, line := range lines {
		if i == 0 {
			continue
		}

		re := regexp.MustCompile(`(.*?)\s+(\d+\S*)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.*)`)
		matches := re.FindStringSubmatch(line)

		if len(matches) >= 10 {
			fileSystem := types.Storage{
				Name:          matches[1],
				Size:          matches[2],
				Used:          matches[3],
				Available:     matches[4],
				UsePercentage: matches[5],
				MountedOn:     matches[9],
			}
			fileSystems = append(fileSystems, fileSystem)
		}
	}

	return fileSystems, nil
}

func (e *EnvironmentGatherer) getCPUDetails() (*types.CPU, error) {
	topOut, err := exec.Command("bash", "-c", "top -bn 1 | awk 'NR == 3 {printf \"%.2f\", 100 - $8}'").
		Output()
	if err != nil {
		return nil, err
	}

	load, _ := strconv.ParseFloat(strings.TrimSpace(string(topOut)), 64)

	cpuOut, err := exec.Command("bash", "-c", "lscpu | grep -E 'Architecture|Model name|CPU MHz|CPU(s)' | sed 's/   *//g'").
		Output()
	if err != nil || string(cpuOut) == "" {
		return nil, err
	}

	log.Infof("Yeahs %s", string(cpuOut))

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

	return &types.CPU{
		Architecture: architecture,
		ModelName:    modelName,
		CPUMHz:       cpuMHz,
		Load:         load,
	}, nil
}

func (e *EnvironmentGatherer) CalculateEnvironment() types.Environment {
	memory, err := e.getMemoryInfo()
	if err != nil {
		log.Errorf("Error getting memory info: %v", err)
	}

	fileSystems, err := e.getFileSystems()
	if err != nil {
		log.Errorf("Error getting file systems: %v", err)
	}

	cpuDetails, err := e.getCPUDetails()
	if err != nil {
		log.Errorf("Error getting CPU details: %v", err)
	}

	environment := types.Environment{
		Memory:  memory,
		CPU:     cpuDetails,
		Storage: fileSystems,
	}

	return environment
}

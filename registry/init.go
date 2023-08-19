package registry

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/evilmint/haargos-agent-golang/types"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func ReadDeviceRegistry(haConfigPath string) (types.DeviceRegistry, error) {
	path := haConfigPath + ".storage/core.device_registry"
	file, err := os.Open(path)
	if err != nil {
		log.Errorf("Failed %s", err)
		return types.DeviceRegistry{}, fmt.Errorf("Error opening file %s: %w", path, err)
	}
	defer file.Close()

	var response types.DeviceRegistry
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&response); err != nil {
		log.Errorf("Failed %s", err)
		return types.DeviceRegistry{}, fmt.Errorf(
			"Error decoding JSON from file %s: %w",
			path,
			err,
		)
	}

	return response, nil
}

func ReadEntityRegistry(haConfigPath string) (types.EntityRegistry, error) {
	path := haConfigPath + ".storage/core.entity_registry"
	file, err := os.Open(path)
	if err != nil {
		return types.EntityRegistry{}, fmt.Errorf("Error opening file %s: %w", path, err)
	}
	defer file.Close()

	var response types.EntityRegistry
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&response); err != nil {
		return types.EntityRegistry{}, fmt.Errorf(
			"Error decoding JSON from file %s: %w",
			path,
			err,
		)
	}

	return response, nil
}

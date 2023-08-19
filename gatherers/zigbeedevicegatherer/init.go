package zigbeedevicegatherer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/evilmint/haargos-agent-golang/types"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

type ZigbeeDeviceGatherer struct {
}

var log = logrus.New()

func copyFile(src, dst string) error {
	// Open the source file for reading
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file for writing
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func queryStatesMeta(db *sql.DB, entityIDs []string) (map[string]int, error) {
	result := make(map[string]int)

	query := "SELECT entity_id, metadata_id FROM states_meta WHERE entity_id IN ("
	params := make([]interface{}, len(entityIDs))
	for i, id := range entityIDs {
		params[i] = id
		if i > 0 {
			query += ", "
		}
		query += "?"
	}
	query += ")"

	log.Infof("Query: %s", query)

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entityID string
		var metadataID int
		err = rows.Scan(&entityID, &metadataID)
		if err != nil {
			return nil, err
		}
		result[entityID] = metadataID
	}

	return result, rows.Err()
}

func (z *ZigbeeDeviceGatherer) GatherDevices(z2mPath *string, zhaPath *string, deviceRegistry *types.DeviceRegistry, entityRegistry *types.EntityRegistry, configPath string) ([]types.ZigbeeDevice, error) {
	nameByIEEE := make(map[string]string)
	ieeeByDeviceId := make(map[string]string)

	for _, device := range deviceRegistry.Data.Devices {
		for _, connection := range device.Connections {
			if len(connection) == 2 && (connection[0] == "zigbee" || connection[0] == "zha") {
				nameByUser := device.Name
				if device.NameByUser != nil {
					nameByUser = *device.NameByUser
				}
				nameByIEEE[connection[1]] = nameByUser
				ieeeByDeviceId[device.ID] = connection[1]
			}
		}
	}

	stateByIeee := make(map[string]*string)
	deviceIDToEntityIDMap := make(map[string]string)
	var entityIds = []string{}

	for _, entity := range entityRegistry.Data.Entities {
		if entity.DeviceID == nil || entity.OriginalDeviceClass == nil || (entity.OriginalDeviceClass != nil && *entity.OriginalDeviceClass != "battery") {
			continue
		}

		deviceIDToEntityIDMap[*entity.DeviceID] = entity.EntityID
		entityIds = append(entityIds, entity.EntityID)
	}

	log.Infof("Opening zigbee")

	dbPath := configPath + "home-assistant_v2.db"

	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "home-assistant-")
	if err != nil {
		log.Fatal(err)
	}

	// Copy main DB, SHM, and WAL files if they exist
	for _, ext := range []string{"", "-shm", "-wal"} {
		src := dbPath + ext
		dst := filepath.Join(tempDir, filepath.Base(dbPath)+ext)
		if err := copyFile(src, dst); err != nil {
			log.Fatal(err)
		}
	}

	// Open the SQLite database from the copied temporary path
	tempDbPath := filepath.Join(tempDir, filepath.Base(dbPath))
	db, err := sql.Open("sqlite3", tempDbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	results, err := queryStatesMeta(db, entityIds)
	if err != nil {
		log.Fatal(err)
	}

	for entityID, metadataID := range results {
		fmt.Println(entityID, metadataID)
	}

	// Handle SQLite database connection, query execution, and data retrieval
	// Assuming you have the appropriate Go packages and logic for handling SQLite
	// You'll have to implement this part based on your specific database structure and requirements

	if z2mPath != nil {
		return z.gatherFromZ2M("", nameByIEEE, stateByIeee), nil
	}

	return []types.ZigbeeDevice{}, nil
}

func (z *ZigbeeDeviceGatherer) gatherFromZ2M(path string, nameByIEEE map[string]string, stateByIeee map[string]*string) []types.ZigbeeDevice {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Can not parse z2m database at: %s", path)
		return nil
	}

	lineString := string(data)
	var devices []types.Z2MDevice

	for _, line := range strings.Split(lineString, "\n") {
		var z2mDevice types.Z2MDevice
		if err := json.Unmarshal([]byte(line), &z2mDevice); err != nil {
			continue
		}
		devices = append(devices, z2mDevice)
	}

	var zigbeeDevices []types.ZigbeeDevice
	for _, device := range devices {
		batteryLevelStr, ok := stateByIeee[device.IEEEAddr]
		batteryLevel := 0
		if ok && batteryLevelStr != nil {
			batteryLevel, err = strconv.Atoi(*batteryLevelStr)
			if err != nil {
				log.Errorf("Failed converting battery level to integer.")
			}
		}
		zigbeeDevices = append(zigbeeDevices, types.NewZigbeeDevice(
			device,
			nameByIEEE[device.IEEEAddr],
			batteryLevel,
		))
	}
	return zigbeeDevices
}

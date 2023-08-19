package zigbeedevicegatherer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
				if device.NameByUser != nil && *device.NameByUser != "" {
					nameByUser = *device.NameByUser
				}
				nameByIEEE[connection[1]] = nameByUser
				ieeeByDeviceId[device.ID] = connection[1]
			}
		}
	}

	stateByIeee := make(map[string]string)
	deviceIDToEntityIDMap := make(map[string]string)
	var entityIds = []string{}

	for _, entity := range entityRegistry.Data.Entities {
		if entity.DeviceID == nil || entity.OriginalDeviceClass == nil || (entity.OriginalDeviceClass != nil && *entity.OriginalDeviceClass != "battery") {
			continue
		}

		deviceIDToEntityIDMap[*entity.DeviceID] = entity.EntityID
		// log.Infof("Setting deviceid of %s with entityid %s", *entity.DeviceID, entity.EntityID)
		entityIds = append(entityIds, entity.EntityID)
	}

	dbPath := configPath + "home-assistant_v2.db"

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "home-assistant-")
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

	var metadataIDs []int
	for _, metadataID := range results {
		metadataIDs = append(metadataIDs, metadataID)
	}

	// Construct stateQuery with ? placeholders
	placeholders := strings.Repeat(",?", len(metadataIDs)-1)
	stateQuery := "SELECT metadata_id, state FROM states WHERE state IS NOT NULL AND metadata_id IN (?" + placeholders + ")"

	// Convert metadataIDs to []interface{} for db.Query
	stateParams := make([]interface{}, len(metadataIDs))
	for i, v := range metadataIDs {
		stateParams[i] = v
	}

	stateRows, err := db.Query(stateQuery, stateParams...)
	if err != nil {
		log.Fatal(err)
	}
	defer stateRows.Close()

	stateByMetadataId := make(map[int]string)
	for stateRows.Next() {
		var metadataId2 int
		var state sql.NullString
		err = stateRows.Scan(&metadataId2, &state)
		if err != nil {
			log.Fatal(err)
		}
		if state.Valid {
			stateByMetadataId[metadataId2] = state.String
		}
	}

	for metaId, stateValue := range stateByMetadataId {
		for entityID, metadataID := range results {
			if metadataID == metaId {
				for deviceID, entityIDInMap := range deviceIDToEntityIDMap {
					if entityIDInMap == entityID {
						if ieee, exists := ieeeByDeviceId[deviceID]; exists {
							stateByIeee[ieee] = stateValue
						}
					}
				}
			}
		}
	}

	var zigbeeDevices []types.ZigbeeDevice

	if z2mPath != nil && *z2mPath != "" {
		log.Infof("Gathering Z2M from %s", *z2mPath)
		zigbeeDevices = append(zigbeeDevices, z.gatherFromZ2M(*z2mPath, nameByIEEE, stateByIeee)...)
	}

	if zhaPath != nil && *zhaPath != "" {
		log.Infof("Gathering ZHA from %s", *zhaPath)
		zigbeeDevices = append(zigbeeDevices, z.gatherFromZHA(*zhaPath, nameByIEEE, stateByIeee)...)
	}

	log.Infof("Devices count: %d", len(zigbeeDevices))

	return zigbeeDevices, nil
}

func convertHex(hexStr string) string {
	// Remove the 0x prefix if present
	if strings.HasPrefix(hexStr, "0x") {
		hexStr = hexStr[2:]
	}

	var result strings.Builder
	for i := 0; i < len(hexStr); i += 2 {
		if i > 0 {
			result.WriteRune(':')
		}
		result.WriteString(hexStr[i : i+2])
	}

	return result.String()
}

func (z *ZigbeeDeviceGatherer) gatherFromZ2M(path string, nameByIEEE map[string]string, stateByIeee map[string]string) []types.ZigbeeDevice {
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
		z2mDevice.IEEEAddr = convertHex(z2mDevice.IEEEAddr)
		devices = append(devices, z2mDevice)
	}

	var zigbeeDevices []types.ZigbeeDevice
	for _, device := range devices {
		batteryLevelStr, ok := stateByIeee[device.IEEEAddr]
		batteryLevel := 0
		if ok && batteryLevelStr != "" {
			batteryLevel, err = strconv.Atoi(batteryLevelStr)
			if err != nil {
				log.Errorf("Failed converting battery level to integer.")
				batteryLevel = 0
			}
		}

		var nameByUser *string
		if name, ok := nameByIEEE[device.IEEEAddr]; ok {
			nameByUser = &name
		}

		zigbeeDevices = append(zigbeeDevices, types.NewZigbeeDevice(
			device,
			nameByUser,
			batteryLevel,
		))
	}

	log.Infof("Gathered %d Z2M devices", len(zigbeeDevices))
	return zigbeeDevices
}

func (z *ZigbeeDeviceGatherer) gatherFromZHA(databasePath string, nameByIEEE map[string]string, stateByIeee map[string]string) []types.ZigbeeDevice {
	db, err := sql.Open("sqlite3", databasePath)
	defer db.Close()

	if err != nil {
		log.Errorf("Error: %s", err)
		return []types.ZigbeeDevice{}
	}

	attributesTable := "attributes_cache_v12"
	devicesTable := "devices_v12"
	neighborsTable := "neighbors_v12"
	nodeDescriptorsTable := "node_descriptors_v12"
	ieee := "ieee"
	lastSeen := "last_seen"
	lqi := "lqi"
	logicalType := "logical_type"

	var deviceMap = map[string]types.ZigbeeDevice{}

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", attributesTable))
	if err != nil {
		log.Errorf("Error: %s", err)
		return []types.ZigbeeDevice{}
	}
	defer rows.Close()

	for rows.Next() {
		var deviceIeee string
		var attridValue int
		var valueStr string
		if err := rows.Scan(&deviceIeee, &attridValue, &valueStr); err != nil {
			return []types.ZigbeeDevice{}
		}

		batteryLevelStr := stateByIeee[deviceIeee]
		batteryLevel := 0
		if batteryLevelStr != "" {
			batteryLevel, err = strconv.Atoi(batteryLevelStr)
			if err != nil {
				batteryLevel = 0
			}
		}

		defaultDevice := types.NewZigbeeDevice(types.Z2MDevice{
			IEEEAddr:    deviceIeee,
			LastSeen:    0,
			PowerSource: "Battery",
		}, nil, batteryLevel)

		device := deviceMap[deviceIeee]
		if (types.ZigbeeDevice{}) == device {
			device = defaultDevice
		}

		if attridValue == 4 {
			device.Brand = valueStr
		} else if attridValue == 5 {
			device.EntityName = valueStr
		}

		deviceRow := db.QueryRow(fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", lastSeen, devicesTable, ieee), deviceIeee)
		var timestamp int64
		if err := deviceRow.Scan(&timestamp); err == nil {
			lastUpdated := time.Unix(int64(timestamp/1000), 0)
			if lastUpdated.After(device.LastUpdated) {
				device.LastUpdated = lastUpdated
			}
		}

		nodeDescriptorRow := db.QueryRow(fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", logicalType, nodeDescriptorsTable, ieee), deviceIeee)
		var logicalTypeValue int
		if err := nodeDescriptorRow.Scan(&logicalTypeValue); err == nil {
			powerSource := "Battery"
			if logicalTypeValue == 1 {
				powerSource = "Mains"
			}
			device.PowerSource = &powerSource
		}

		lqiRow := db.QueryRow(fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", lqi, neighborsTable, ieee), deviceIeee)
		if err := lqiRow.Scan(&device.Lqi); err != nil {
			device.Lqi = 0
		}

		if name, ok := nameByIEEE[deviceIeee]; ok {
			device.NameByUser = &name
		}

		deviceMap[deviceIeee] = device
	}

	var zigbeeDevices []types.ZigbeeDevice
	for _, device := range deviceMap {
		zigbeeDevices = append(zigbeeDevices, device)
	}

	log.Infof("Gathered %d ZHA devices", len(zigbeeDevices))
	return zigbeeDevices
}

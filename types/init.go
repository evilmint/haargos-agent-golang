package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// Define the structs to match the provided JSON structure
type Memory struct {
	Used      int `json:"used"`
	Total     int `json:"total"`
	Available int `json:"available"`
	Free      int `json:"free"`
	Shared    int `json:"shared"`
	BuffCache int `json:"buff_cache"`
	SwapTotal int `json:"swap_total"`
	SwapUsed  int `json:"swap_used"`
}

type CPU struct {
	ModelName    string  `json:"model_name"`
	Architecture string  `json:"architecture"`
	Load         float64 `json:"load"`
	CPUMHz       string  `json:"cpu_mhz"`
	Temperature  float64 `json:"temp"`
}

type Storage struct {
	Size          string `json:"size"`
	Used          string `json:"used"`
	UsePercentage string `json:"use_percentage"`
	MountedOn     string `json:"mounted_on"`
	Name          string `json:"name"`
	Available     string `json:"available"`
}

type Environment struct {
	Memory   *Memory   `json:"memory"`
	CPU      *CPU      `json:"cpu"`
	Storage  []Storage `json:"storage"`
	Network  *Network  `json:"network"`
	BootTime string    `json:"boot_time"`
}

type Network struct {
	Interfaces []NetworkInterface `json:"interfaces"`
}

type NetworkInterface struct {
	Name string                `json:"name"`
	Rx   *NetworkInterfaceData `json:"rx"`
	Tx   *NetworkInterfaceData `json:"tx"`
}

type NetworkInterfaceData struct {
	Bytes   int `json:"bytes"`
	Packets int `json:"packets"`
}

type DockerContainer struct {
	FinishedAt string `json:"finished_at"`
	Image      string `json:"image"`
	State      string `json:"state"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Restarting string `json:"restarting"`
	Running    bool   `json:"running"`
	StartedAt  string `json:"started_at"`
}

type Docker struct {
	Containers []DockerContainer `json:"containers"`
}

type Auth struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

type Observation struct {
	Docker       Docker       `json:"docker"`
	AgentType    string       `json:"agent_type"`
	AgentVersion string       `json:"agent_version"`
	Environment  Environment  `json:"environment"`
	Zigbee       ZigbeeStatus `json:"zigbee"`
	HAConfig     HAConfig     `json:"ha_config"`
	Automations  []Automation `json:"automations"`
	Scripts      []Script     `json:"scripts"`
	Scenes       []Scene      `json:"scenes"`
}

type MemoryStatus struct {
	Used      int `json:"used"`
	Total     int `json:"total"`
	Free      int `json:"free"`
	Shared    int `json:"shared"`
	BuffCache int `json:"buff_cache"`
	Available int `json:"available"`
}

type CPUStatus struct {
	ModelName    string  `json:"model_name"`
	CpuMHz       string  `json:"cpu_mhz"`
	Architecture string  `json:"architecture"`
	Load         float64 `json:"load"`
}

type ZigbeeStatus struct {
	Devices []ZigbeeDevice `json:"devices"`
}

type GenericJob struct {
	CreatedAt            string      `json:"created_at"`
	StatusInstallationID string      `json:"status_installation_id"`
	InstallationID       string      `json:"installation_id"`
	ID                   string      `json:"id"`
	Type                 string      `json:"type"`
	Context              interface{} `json:"context"`
}

type JobsResponse struct {
	Body []GenericJob `json:"body"`
}

type ZigbeeDevice struct {
	Ieee            string    `json:"ieee"`
	Lqi             int       `json:"lqi"`
	LastUpdated     time.Time `json:"last_updated"`
	DeviceID        string    `json:"device_id"`
	NameByUser      *string   `json:"name_by_user"` // can be string or null
	PowerSource     *string   `json:"power_source"` // can be string or null
	EntityName      string    `json:"entity_name"`
	Brand           string    `json:"brand"`
	IntegrationType string    `json:"integration_type"`
	BatteryLevel    int       `json:"battery_level"` // can be string or null
}

type HAConfig struct {
	Version string `json:"version"`
}

type OSInfo struct {
	Version         string `json:"version"`
	VersionLatest   string `json:"version_latest"`
	UpdateAvailable bool   `json:"update_available"`
	Board           string `json:"board"`
	Boot            string `json:"boot"`
	DataDisk        string `json:"data_disk"`
}

type OSInfoResponse struct {
	Data OSInfo `json:"data"`
}

type SupervisorInfo struct {
	Version            string                      `json:"version"`
	VersionLatest      string                      `json:"version_latest"`
	UpdateAvailable    bool                        `json:"update_available"`
	Arch               string                      `json:"arch"`
	Channel            string                      `json:"channel"`
	Timezone           string                      `json:"timezone"`
	Healthy            bool                        `json:"healthy"`
	Supported          bool                        `json:"supported"`
	Logging            string                      `json:"logging"`
	IPAddress          string                      `json:"ip_address"`
	WaitBoot           int                         `json:"wait_boot"`
	Debug              bool                        `json:"debug"`
	DebugBlock         bool                        `json:"debug_block"`
	Diagnostics        *json.RawMessage            `json:"diagnostics,omitempty"` // Use json.RawMessage for arbitrary JSON
	AddonsRepositories []SupervisorAddonRepository `json:"addons_repositories"`
	AutoUpdate         bool                        `json:"auto_update"`
}

type SupervisorAddonRepository struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type SupervisorInfoResponse struct {
	Data SupervisorInfo `json:"data"`
}

type Automation struct {
	LastTriggered time.Time `json:"last_triggered" yaml:"last_triggered"`
	Description   string    `json:"description"    yaml:"description"`
	ID            string    `json:"id"             yaml:"id"`
	State         string    `json:"state"          yaml:"state"`
	Alias         string    `json:"alias"          yaml:"alias"`
	FriendlyName  string    `json:"friendly_name"  yaml:"friendly_name"`
}

type Script struct {
	FriendlyName  string    `json:"friendly_name"  yaml:"friendly_name"`
	State         string    `json:"state"          yaml:"state"`
	Alias         string    `json:"alias"          yaml:"alias"`
	UniqueId      string    `json:"unique_id"      yaml:"unique_id"`
	LastTriggered time.Time `json:"last_triggered" yaml:"last_triggered"`
}

type Scene struct {
	Name         string    `json:"name"          yaml:"name"`
	ID           string    `json:"id"            yaml:"id"`
	State        time.Time `json:"state"         yaml:"state"`
	FriendlyName string    `json:"friendly_name" yaml:"friendly_name"`
}

type EntityRegistry struct {
	Version      int                     `json:"version"`
	MinorVersion int                     `json:"minor_version"`
	Key          string                  `json:"key"`
	Data         EntityRegistryDataClass `json:"data"`
}

type EntityRegistryDataClass struct {
	Entities []EntityRegistryEntity `json:"entities"`
}

type EntityRegistryEntity struct {
	DeviceClass         *string `json:"device_class"`
	DeviceID            *string `json:"device_id"`
	EntityID            string  `json:"entity_id"`
	ID                  string  `json:"id"`
	Name                *string `json:"name"`
	OriginalDeviceClass *string `json:"original_device_class"`
}

type DeviceRegistry struct {
	Version      int                     `json:"version"`
	MinorVersion int                     `json:"minor_version"`
	Key          string                  `json:"key"`
	Data         DeviceRegistryDataClass `json:"data"`
}

type DeviceRegistryDataClass struct {
	Devices []DeviceRegistryDevice `json:"devices"`
}

type DeviceRegistryDevice struct {
	AreaId           *string    `json:"area_id"`
	ConfigEntries    []string   `json:"config_entries"`
	ConfigurationUrl *string    `json:"configuration_url"`
	Connections      [][]string `json:"connections"`
	DisabledBy       *string    `json:"disabled_by"`
	EntryType        *string    `json:"entry_type"`
	HwVersion        *string    `json:"hw_version"`
	ID               string     `json:"id"`
	Identifiers      [][]string
	Manufacturer     *string `json:"manufacturer"`
	Model            *string `json:"model"`
	NameByUser       *string `json:"name_by_user"`
	Name             string  `json:"name"`
}

func (d *DeviceRegistryDevice) UnmarshalJSON(data []byte) error {
	type alias DeviceRegistryDevice // Define an alias with the same structure but without custom unmarshaler
	aux := &struct {
		Identifiers [][]interface{} `yaml:"identifiers"`
		*alias
	}{
		alias: (*alias)(d), // Point to the receiver to avoid stack overflow
	}

	// Use the unmarshal func to decode into the auxiliary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert identifiers from interface{} to string
	for _, identifierPair := range aux.Identifiers {
		var stringPair []string
		for _, identifier := range identifierPair {
			switch v := identifier.(type) {
			case string:
				stringPair = append(stringPair, v)
			case int:
				stringPair = append(stringPair, fmt.Sprintf("%d", v))
			case float64:
				stringPair = append(stringPair, fmt.Sprintf("%.0f", v)) // handle unmarshaling into float64, common with YAML/JSON
			}
		}
		d.Identifiers = append(d.Identifiers, stringPair)
	}

	return nil
}

type RestoreStateAttributes struct {
	ID            *string `json:"id"`
	FriendlyName  *string `json:"friendly_name"`
	LastTriggered *string `json:"last_triggered"`
}

type RestoreStateContext struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parent_id"`
	UserID   *string `json:"user_id"`
}

type RestoreStateState struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  RestoreStateAttributes `json:"attributes"`
	LastChanged string                 `json:"last_changed"`
	LastUpdated string                 `json:"last_updated"`
	Context     RestoreStateContext    `json:"context"`
}

type RestoreStateData struct {
	State    RestoreStateState `json:"state"`
	LastSeen string            `json:"last_seen"`
}

type RestoreStateResponse struct {
	Data []RestoreStateData `json:"data"`
}

type DockerInspectState struct {
	Status       string `json:"Status"`
	IsRunning    bool   `json:"Running"`
	IsRestarting bool   `json:"Restarting"`
	StartedAt    string `json:"StartedAt"`
	FinishedAt   string `json:"FinishedAt"`
}

type DockerInspectLogConfig struct {
	Type   string            `json:"Type"`
	Config map[string]string `json:"Config"`
}

type DockerInspectRestartPolicy struct {
	Name              string `json:"Name"`
	MaximumRetryCount int    `json:"MaximumRetryCount"`
}

type DockerInspectDevice struct {
	PathOnHost        string `json:"PathOnHost"`
	PathInContainer   string `json:"PathInContainer"`
	CgroupPermissions string `json:"CgroupPermissions"`
}

type DockerInspectMount struct {
	Type        string `json:"Type"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Mode        string `json:"Mode"`
	RW          bool   `json:"RW"`
	Propagation string `json:"Propagation"`
}

type DockerInspectConfig struct {
	Image      string                       `json:"Image"`
	Volumes    map[string]map[string]string `json:"Volumes"`
	Entrypoint []string                     `json:"Entrypoint"`
	// Add other fields as needed from the "Config" section
}

type DockerInspectHostConfig struct {
	Binds         []string                   `json:"Binds"`
	LogConfig     DockerInspectLogConfig     `json:"LogConfig"`
	NetworkMode   string                     `json:"NetworkMode"`
	RestartPolicy DockerInspectRestartPolicy `json:"RestartPolicy"`
	// Add other fields as needed from the "HostConfig" section
}

type DockerInspectResult struct {
	State      DockerInspectState      `json:"State"`
	HostConfig DockerInspectHostConfig `json:"HostConfig"`
	Mounts     []DockerInspectMount    `json:"Mounts"`
	Config     DockerInspectConfig     `json:"Config"`
	Name       string                  `json:"Name"`
}

type DockerPsEntry struct {
	Command   string `json:"Command"`
	CreatedAt string `json:"CreatedAt"`
	ID        string `json:"ID"`
	Image     string `json:"Image"`
	Mounts    string `json:"Mounts"`
	Names     string `json:"Names"`
	Networks  string `json:"Networks"`
	Size      string `json:"Size"`
	State     string `json:"State"`
	Status    string `json:"Status"`
}

type Z2MDevice struct {
	ID                 int
	Type               string
	IEEEAddr           string
	NwkAddr            int
	ManufId            int
	ManufName          string
	PowerSource        string
	ModelId            string
	AppVersion         int
	StackVersion       int
	HWVersion          int
	DateCode           string
	SWBuildId          string
	ZclVersion         int
	InterviewCompleted bool
	LastSeen           int64
}

type Logs struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

func ifEmpty(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func NewZigbeeDevice(z2mDevice Z2MDevice, nameByUser *string, batteryLevel int) ZigbeeDevice {
	return ZigbeeDevice{
		Ieee:            z2mDevice.IEEEAddr,
		EntityName:      ifEmpty(z2mDevice.ModelId, "-"),
		Brand:           ifEmpty(z2mDevice.ManufName, "-"),
		LastUpdated:     time.Unix(int64(z2mDevice.LastSeen/1000), 0),
		Lqi:             0,
		IntegrationType: "z2m",
		NameByUser:      nameByUser,
		PowerSource:     &z2mDevice.PowerSource,
		BatteryLevel:    batteryLevel,
	}
}

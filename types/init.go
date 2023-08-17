package types

import "time"

// Define the structs to match the provided JSON structure
type Memory struct {
	Used      int `json:"used"`
	Total     int `json:"total"`
	Available int `json:"available"`
	Free      int `json:"free"`
	Shared    int `json:"shared"`
	BuffCache int `json:"buff_cache"`
}

type CPU struct {
	ModelName    string  `json:"model_name"`
	Architecture string  `json:"architecture"`
	Load         float64 `json:"load"`
	CPUMHz       string  `json:"cpu_mhz"`
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
	Memory  Memory    `json:"memory"`
	CPU     CPU       `json:"cpu"`
	Storage []Storage `json:"storage"`
}

type DockerContainer struct {
	FinishedAt string `json:"finished_at"`
	Image      string `json:"image"`
	State      string `json:"state"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Restarting bool   `json:"restarting"`
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
	Docker         Docker       `json:"docker"`
	AgentVersion   string       `json:"agent_version"`
	Environment    Environment  `json:"environment"`
	Logs           string       `json:"logs"`
	InstallationID string       `json:"installation_id"`
	Zigbee         ZigbeeStatus `json:"zigbee"`
	HAConfig       HAConfig     `json:"ha_config"`
	Automations    []Automation `json:"automations"`
	Scripts        []Script     `json:"scripts"`
	Scenes         []Scene      `json:"scenes"`
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

type ZigbeeDevice struct {
	Ieee            string    `json:"ieee"`
	Lqi             int       `json:"lqi"`
	LastUpdated     time.Time `json:"last_updated"`
	NameByUser      *string   `json:"name_by_user"` // can be string or null
	PowerSource     *string   `json:"power_source"` // can be string or null
	EntityName      string    `json:"entity_name"`
	Brand           string    `json:"brand"`
	IntegrationType string    `json:"integration_type"`
	BatteryLevel    *string   `json:"battery_level"` // can be string or null
}

type HAConfig struct {
	Version string `json:"version"`
}

type Automation struct {
	LastTriggered time.Time `json:"last_triggered"`
	Description   string    `json:"description"`
	ID            string    `json:"id"`
	State         string    `json:"state"`
	Alias         string    `json:"alias"`
	FriendlyName  string    `json:"friendly_name"`
}

type Script struct {
	FriendlyName  string    `json:"friendlyName,omitempty"`
	State         string    `json:"state,omitempty"`
	Alias         string    `json:"alias"`
	LastTriggered time.Time `json:"lastTriggered,omitempty"`
}

type Scene struct {
	Name         string    `json:"name"`
	ID           string    `json:"id"`
	State        time.Time `json:"state"`
	FriendlyName string    `json:"friendly_name"`
}

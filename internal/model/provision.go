package model

// BulkProvisionConfig represents the bulk provisioning configuration file.
type BulkProvisionConfig struct {
	WiFi    *ProvisionWiFiConfig    `yaml:"wifi,omitempty" json:"wifi,omitempty"`
	Devices []DeviceProvisionConfig `yaml:"devices" json:"devices"`
}

// ProvisionWiFiConfig represents shared WiFi settings.
type ProvisionWiFiConfig struct {
	SSID     string `yaml:"ssid" json:"ssid"`
	Password string `yaml:"password" json:"password"`
}

// DeviceProvisionConfig represents per-device settings.
type DeviceProvisionConfig struct {
	Name    string               `yaml:"name" json:"name"`
	Address string               `yaml:"address,omitempty" json:"address,omitempty"`
	WiFi    *ProvisionWiFiConfig `yaml:"wifi,omitempty" json:"wifi,omitempty"`
	DevName string               `yaml:"device_name,omitempty" json:"device_name,omitempty"`
}

// ProvisionResult holds the result of provisioning a single device.
type ProvisionResult struct {
	Device string
	Err    error
}

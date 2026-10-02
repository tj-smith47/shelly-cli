package model

// BulkProvisionConfig represents the bulk provisioning configuration file.
type BulkProvisionConfig struct {
	WiFi    *ProvisionWiFiConfig    `yaml:"wifi,omitempty" json:"wifi,omitempty"`
	Devices []DeviceProvisionConfig `yaml:"devices" json:"devices"`
}

// ProvisionWiFiConfig represents shared WiFi settings.
// An empty Password with Open false means the key is not known: a device on
// the same network keeps its own, and a different network takes the
// passphrase stored on this host or is refused.
type ProvisionWiFiConfig struct {
	SSID     string `yaml:"ssid" json:"ssid"`
	Password string `yaml:"password" json:"password"`
	// Open joins a network that has no password.
	Open bool `yaml:"open,omitempty" json:"open,omitempty"`
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
	// Warnings are SetWiFiConfig's, such as a static address kept on a
	// changed network.
	Warnings []string
	Err      error
}

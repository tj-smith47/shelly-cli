// Package plugins provides plugin discovery, loading, and manifest management.
package plugins

// ComponentInfo describes a device component discovered by a plugin.
type ComponentInfo struct {
	Type string `json:"type"`           // switch, light, cover, sensor
	ID   int    `json:"id"`             // Component index (0-based)
	Name string `json:"name,omitempty"` // Human-readable name
}

// DeviceDetectionResult is returned by the detect hook.
type DeviceDetectionResult struct {
	// Detected indicates whether the plugin recognized this device.
	Detected bool `json:"detected"`

	// Platform is the platform identifier (e.g., "tasmota", "esphome").
	Platform string `json:"platform"`

	// DeviceID is a unique identifier for the device (e.g., hostname, MAC).
	DeviceID string `json:"device_id,omitempty"`

	// DeviceName is a human-readable name for the device.
	DeviceName string `json:"device_name,omitempty"`

	// Model is the device model string.
	Model string `json:"model,omitempty"`

	// Firmware is the current firmware version.
	Firmware string `json:"firmware,omitempty"`

	// MAC is the device MAC address.
	MAC string `json:"mac,omitempty"`

	// Components lists controllable components on the device.
	Components []ComponentInfo `json:"components,omitempty"`
}

// DeviceStatusResult is returned by the status hook.
type DeviceStatusResult struct {
	// Online indicates whether the device is reachable.
	Online bool `json:"online"`

	// Components contains component-specific status (keyed by "type:id", e.g., "switch:0").
	Components map[string]any `json:"components,omitempty"`

	// Sensors contains sensor readings (keyed by sensor type).
	Sensors map[string]any `json:"sensors,omitempty"`

	// Energy contains power/energy metrics if available.
	Energy *EnergyStatus `json:"energy,omitempty"`
}

// EnergyStatus contains power/energy metrics.
type EnergyStatus struct {
	// Power is the current power consumption in Watts.
	Power float64 `json:"power,omitempty"`

	// Voltage is the current voltage in Volts.
	Voltage float64 `json:"voltage,omitempty"`

	// Current is the current amperage in Amps.
	Current float64 `json:"current,omitempty"`

	// Total is the total energy consumed in kWh.
	Total float64 `json:"total,omitempty"`

	// ApparentPower is the apparent power in VA.
	ApparentPower float64 `json:"apparent_power,omitempty"`

	// ReactivePower is the reactive power in VAR.
	ReactivePower float64 `json:"reactive_power,omitempty"`

	// PowerFactor is the power factor (0-1).
	PowerFactor float64 `json:"power_factor,omitempty"`
}

// ControlResult is returned by the control hook.
type ControlResult struct {
	// Success indicates whether the control operation succeeded.
	Success bool `json:"success"`

	// State is the resulting state after the operation (e.g., "on", "off", "opening").
	State string `json:"state,omitempty"`

	// Error contains an error message if Success is false.
	Error string `json:"error,omitempty"`
}

// FirmwareUpdateInfo is returned by the check_updates hook.
type FirmwareUpdateInfo struct {
	// CurrentVersion is the firmware version currently installed.
	CurrentVersion string `json:"current_version"`

	// LatestStable is the latest stable firmware version available.
	LatestStable string `json:"latest_stable,omitempty"`

	// LatestBeta is the latest beta/development firmware version available.
	LatestBeta string `json:"latest_beta,omitempty"`

	// HasUpdate indicates whether a stable update is available.
	HasUpdate bool `json:"has_update"`

	// HasBetaUpdate indicates whether a beta update is available.
	HasBetaUpdate bool `json:"has_beta_update,omitempty"`

	// OTAURLStable is the OTA URL for the stable firmware.
	OTAURLStable string `json:"ota_url_stable,omitempty"`

	// OTAURLBeta is the OTA URL for the beta firmware.
	OTAURLBeta string `json:"ota_url_beta,omitempty"`

	// ChipType is the device's chip type (e.g., "ESP8266", "ESP32").
	ChipType string `json:"chip_type,omitempty"`

	// Variant is the firmware variant (e.g., "tasmota", "tasmota-lite").
	Variant string `json:"variant,omitempty"`

	// ReleaseNotesURL is a URL to the release notes.
	ReleaseNotesURL string `json:"release_notes_url,omitempty"`
}

// UpdateResult is returned by the apply_update hook.
type UpdateResult struct {
	// Success indicates whether the update was initiated successfully.
	Success bool `json:"success"`

	// Message contains a status message.
	Message string `json:"message,omitempty"`

	// Error contains an error message if Success is false.
	Error string `json:"error,omitempty"`

	// Rebooting indicates the device is rebooting to apply the update.
	Rebooting bool `json:"rebooting,omitempty"`
}

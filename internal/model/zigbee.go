// Package model defines core domain types for the Shelly CLI.
package model

// ZigbeeDevice represents a Zigbee-capable device.
type ZigbeeDevice struct {
	Name         string `json:"name"`
	Address      string `json:"address"`
	Model        string `json:"model,omitempty"`
	Enabled      bool   `json:"zigbee_enabled"`
	NetworkState string `json:"network_state,omitempty"`
	EUI64        string `json:"eui64,omitempty"`
}

// ZigbeeStatus represents the full Zigbee status.
type ZigbeeStatus struct {
	Enabled          bool   `json:"enabled"`
	NetworkState     string `json:"network_state"`
	EUI64            string `json:"eui64,omitempty"`
	PANID            uint16 `json:"pan_id,omitempty"`
	Channel          int    `json:"channel,omitempty"`
	CoordinatorEUI64 string `json:"coordinator_eui64,omitempty"`
}

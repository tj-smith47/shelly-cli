// Package model defines core domain types for the Shelly CLI.
package model

// LoRaFullStatus combines config and status for LoRa add-on.
type LoRaFullStatus struct {
	Config *LoRaConfig `json:"config,omitempty"`
	Status *LoRaStatus `json:"status,omitempty"`
}

// LoRaConfig represents LoRa configuration.
type LoRaConfig struct {
	ID   int   `json:"id"`
	Freq int64 `json:"freq"`
	BW   int   `json:"bw"`
	DR   int   `json:"dr"`
	TxP  int   `json:"txp"`
}

// LoRaStatus represents LoRa status from the last received packet.
type LoRaStatus struct {
	ID   int     `json:"id"`
	RSSI int     `json:"rssi"`
	SNR  float64 `json:"snr"`
}

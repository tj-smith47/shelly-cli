package model

// BTHomeDeviceInfo holds information about a BTHome device.
type BTHomeDeviceInfo struct {
	ID         int     `json:"id"`
	Name       string  `json:"name,omitempty"`
	Addr       string  `json:"addr"`
	RSSI       *int    `json:"rssi,omitempty"`
	Battery    *int    `json:"battery,omitempty"`
	LastUpdate float64 `json:"last_update,omitempty"`
}

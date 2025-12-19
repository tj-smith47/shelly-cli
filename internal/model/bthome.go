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

// BTHomeComponentStatus represents BTHome component status.
type BTHomeComponentStatus struct {
	Discovery *BTHomeDiscoveryStatus `json:"discovery,omitempty"`
	Errors    []string               `json:"errors,omitempty"`
}

// BTHomeDiscoveryStatus represents active discovery scan status.
type BTHomeDiscoveryStatus struct {
	StartedAt float64 `json:"started_at"`
	Duration  int     `json:"duration"`
}

// BTHomeDeviceStatus represents a BTHome device status.
type BTHomeDeviceStatus struct {
	ID           int              `json:"id"`
	Name         string           `json:"name,omitempty"`
	Addr         string           `json:"addr"`
	RSSI         *int             `json:"rssi,omitempty"`
	Battery      *int             `json:"battery,omitempty"`
	PacketID     *int             `json:"packet_id,omitempty"`
	LastUpdateTS float64          `json:"last_updated_ts"`
	KnownObjects []BTHomeKnownObj `json:"known_objects,omitempty"`
	Errors       []string         `json:"errors,omitempty"`
}

// BTHomeKnownObj represents a known BTHome object.
type BTHomeKnownObj struct {
	ObjID     int     `json:"obj_id"`
	Idx       int     `json:"idx"`
	Component *string `json:"component,omitempty"`
}

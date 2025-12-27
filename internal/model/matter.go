// Package model defines core domain types for the Shelly CLI.
package model

// MatterStatus represents full Matter status.
type MatterStatus struct {
	Enabled        bool `json:"enabled"`
	Commissionable bool `json:"commissionable"`
	FabricsCount   int  `json:"fabrics_count"`
}

// CommissioningInfo holds Matter pairing information.
type CommissioningInfo struct {
	ManualCode    string `json:"manual_code,omitempty"`
	QRCode        string `json:"qr_code,omitempty"`
	Discriminator int    `json:"discriminator,omitempty"`
	SetupPINCode  int    `json:"setup_pin_code,omitempty"`
	Available     bool   `json:"available"`
}

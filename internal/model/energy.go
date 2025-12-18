// Package model defines core domain types for the Shelly CLI.
package model

import "time"

// DashboardData represents aggregated energy dashboard data.
type DashboardData struct {
	Timestamp     time.Time              `json:"timestamp"`
	TotalPower    float64                `json:"total_power_w"`
	TotalEnergy   float64                `json:"total_energy_wh,omitempty"`
	DeviceCount   int                    `json:"device_count"`
	OnlineCount   int                    `json:"online_count"`
	OfflineCount  int                    `json:"offline_count"`
	Devices       []DashboardDeviceEntry `json:"devices"`
	EstimatedCost *float64               `json:"estimated_cost,omitempty"`
	CostCurrency  string                 `json:"cost_currency,omitempty"`
	CostPerKwh    float64                `json:"cost_per_kwh,omitempty"`
}

// DashboardDeviceEntry represents energy status for a single device in the dashboard.
type DashboardDeviceEntry struct {
	Device      string           `json:"device"`
	Online      bool             `json:"online"`
	Error       string           `json:"error,omitempty"`
	TotalPower  float64          `json:"total_power_w"`
	TotalEnergy float64          `json:"total_energy_wh,omitempty"`
	Components  []ComponentPower `json:"components,omitempty"`
}

// ComponentPower represents power data for a single component.
type ComponentPower struct {
	Type    string  `json:"type"`
	ID      int     `json:"id"`
	Power   float64 `json:"power_w"`
	Voltage float64 `json:"voltage_v,omitempty"`
	Current float64 `json:"current_a,omitempty"`
	Energy  float64 `json:"energy_wh,omitempty"`
}

// ComparisonData represents energy comparison results.
type ComparisonData struct {
	Period      string         `json:"period"`
	From        time.Time      `json:"from"`
	To          time.Time      `json:"to"`
	Devices     []DeviceEnergy `json:"devices"`
	TotalEnergy float64        `json:"total_energy_kwh"`
	MaxEnergy   float64        `json:"max_energy_kwh"`
	MinEnergy   float64        `json:"min_energy_kwh"`
}

// DeviceEnergy represents energy data for a single device.
type DeviceEnergy struct {
	Device     string  `json:"device"`
	Energy     float64 `json:"energy_kwh"`
	AvgPower   float64 `json:"avg_power_w"`
	PeakPower  float64 `json:"peak_power_w"`
	DataPoints int     `json:"data_points"`
	Online     bool    `json:"online"`
	Error      string  `json:"error,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
}

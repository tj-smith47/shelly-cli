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

// EMStatus represents the status of an Energy Monitor (EM) component (3-phase).
type EMStatus struct {
	ID               int      `json:"id"`
	AVoltage         float64  `json:"a_voltage"`
	ACurrent         float64  `json:"a_current"`
	AActivePower     float64  `json:"a_act_power"`
	AApparentPower   float64  `json:"a_aprt_power"`
	APowerFactor     *float64 `json:"a_pf,omitempty"`
	AFreq            *float64 `json:"a_freq,omitempty"`
	BVoltage         float64  `json:"b_voltage"`
	BCurrent         float64  `json:"b_current"`
	BActivePower     float64  `json:"b_act_power"`
	BApparentPower   float64  `json:"b_aprt_power"`
	BPowerFactor     *float64 `json:"b_pf,omitempty"`
	BFreq            *float64 `json:"b_freq,omitempty"`
	CVoltage         float64  `json:"c_voltage"`
	CCurrent         float64  `json:"c_current"`
	CActivePower     float64  `json:"c_act_power"`
	CApparentPower   float64  `json:"c_aprt_power"`
	CPowerFactor     *float64 `json:"c_pf,omitempty"`
	CFreq            *float64 `json:"c_freq,omitempty"`
	NCurrent         *float64 `json:"n_current,omitempty"`
	TotalCurrent     float64  `json:"total_current"`
	TotalActivePower float64  `json:"total_act_power"`
	TotalAprtPower   float64  `json:"total_aprt_power"`
	Errors           []string `json:"errors,omitempty"`
}

// EM1Status represents the status of a single-phase Energy Monitor (EM1) component.
type EM1Status struct {
	ID        int      `json:"id"`
	Voltage   float64  `json:"voltage"`
	Current   float64  `json:"current"`
	ActPower  float64  `json:"act_power"`
	AprtPower float64  `json:"aprt_power"`
	PF        *float64 `json:"pf,omitempty"`
	Freq      *float64 `json:"freq,omitempty"`
	Errors    []string `json:"errors,omitempty"`
}

// PMStatus represents the status of a Power Meter (PM/PM1) component.
type PMStatus struct {
	ID         int               `json:"id"`
	Voltage    float64           `json:"voltage"`
	Current    float64           `json:"current"`
	APower     float64           `json:"apower"`
	Freq       *float64          `json:"freq,omitempty"`
	AEnergy    *PMEnergyCounters `json:"aenergy,omitempty"`
	RetAEnergy *PMEnergyCounters `json:"ret_aenergy,omitempty"`
	Errors     []string          `json:"errors,omitempty"`
}

// PMEnergyCounters represents accumulated energy measurements for power meters.
type PMEnergyCounters struct {
	Total    float64   `json:"total"`
	ByMinute []float64 `json:"by_minute,omitempty"`
	MinuteTs *int64    `json:"minute_ts,omitempty"`
}

// MonitoringSnapshot represents a complete device monitoring snapshot.
type MonitoringSnapshot struct {
	Device    string      `json:"device"`
	Timestamp time.Time   `json:"timestamp"`
	EM        []EMStatus  `json:"em,omitempty"`
	EM1       []EM1Status `json:"em1,omitempty"`
	PM        []PMStatus  `json:"pm,omitempty"`
	Online    bool        `json:"online"`
	Error     string      `json:"error,omitempty"`
}

// DeviceEvent represents a real-time event from a device.
type DeviceEvent struct {
	Device      string         `json:"device"`
	Timestamp   time.Time      `json:"timestamp"`
	Event       string         `json:"event"`
	Component   string         `json:"component"`
	ComponentID int            `json:"component_id"`
	Data        map[string]any `json:"data,omitempty"`
}

// ComponentReading is the unified intermediate format for all meter types.
// Type is one of: pm, pm1, em, em1.
// Phase is only set for EM (a, b, c, total).
type ComponentReading struct {
	Device  string   `json:"device"`
	Type    string   `json:"type"`
	ID      int      `json:"id"`
	Phase   string   `json:"phase,omitempty"`
	Power   float64  `json:"power"`
	Voltage float64  `json:"voltage"`
	Current float64  `json:"current"`
	Energy  *float64 `json:"energy,omitempty"`
	Freq    *float64 `json:"freq,omitempty"`
}

// MeterReading is a common interface for all meter types (PM, PM1, EM, EM1).
type MeterReading interface {
	GetPower() float64
	GetVoltage() float64
	GetCurrent() float64
	GetEnergy() *float64
	GetFreq() *float64
}

// GetPower returns the active power.
func (s *PMStatus) GetPower() float64 { return s.APower }

// GetVoltage returns the voltage.
func (s *PMStatus) GetVoltage() float64 { return s.Voltage }

// GetCurrent returns the current.
func (s *PMStatus) GetCurrent() float64 { return s.Current }

// GetEnergy returns the total energy if available.
func (s *PMStatus) GetEnergy() *float64 {
	if s.AEnergy == nil {
		return nil
	}
	return &s.AEnergy.Total
}

// GetFreq returns the frequency if available.
func (s *PMStatus) GetFreq() *float64 { return s.Freq }

// GetPower returns the total active power.
func (s *EMStatus) GetPower() float64 { return s.TotalActivePower }

// GetVoltage returns phase A voltage.
func (s *EMStatus) GetVoltage() float64 { return s.AVoltage }

// GetCurrent returns total current.
func (s *EMStatus) GetCurrent() float64 { return s.TotalCurrent }

// GetEnergy returns nil (EM uses EMData for historical energy).
func (s *EMStatus) GetEnergy() *float64 { return nil }

// GetFreq returns phase A frequency if available.
func (s *EMStatus) GetFreq() *float64 { return s.AFreq }

// GetPower returns the active power.
func (s *EM1Status) GetPower() float64 { return s.ActPower }

// GetVoltage returns the voltage.
func (s *EM1Status) GetVoltage() float64 { return s.Voltage }

// GetCurrent returns the current.
func (s *EM1Status) GetCurrent() float64 { return s.Current }

// GetEnergy returns nil (EM1 uses EM1Data for historical energy).
func (s *EM1Status) GetEnergy() *float64 { return nil }

// GetFreq returns the frequency if available.
func (s *EM1Status) GetFreq() *float64 { return s.Freq }

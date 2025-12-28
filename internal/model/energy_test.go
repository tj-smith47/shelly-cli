package model

import (
	"testing"
	"time"
)

const testDeviceKitchenSwitch = "kitchen-switch"

// TestDashboardData tests DashboardData struct.
func TestDashboardData(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cost := 1.50
	data := DashboardData{
		Timestamp:    now,
		TotalPower:   150.5,
		TotalEnergy:  1250.75,
		DeviceCount:  5,
		OnlineCount:  4,
		OfflineCount: 1,
		Devices: []DashboardDeviceEntry{
			{
				Device:     "switch1",
				Online:     true,
				TotalPower: 50.5,
			},
		},
		EstimatedCost: &cost,
		CostCurrency:  "USD",
		CostPerKwh:    0.12,
	}

	if data.Timestamp != now {
		t.Errorf("Timestamp = %v, want %v", data.Timestamp, now)
	}
	if data.TotalPower != 150.5 {
		t.Errorf("TotalPower = %f, want 150.5", data.TotalPower)
	}
	if data.TotalEnergy != 1250.75 {
		t.Errorf("TotalEnergy = %f, want 1250.75", data.TotalEnergy)
	}
	if data.DeviceCount != 5 {
		t.Errorf("DeviceCount = %d, want 5", data.DeviceCount)
	}
	if data.OnlineCount != 4 {
		t.Errorf("OnlineCount = %d, want 4", data.OnlineCount)
	}
	if data.OfflineCount != 1 {
		t.Errorf("OfflineCount = %d, want 1", data.OfflineCount)
	}
	if len(data.Devices) != 1 {
		t.Errorf("Devices len = %d, want 1", len(data.Devices))
	}
	if data.EstimatedCost == nil || *data.EstimatedCost != 1.50 {
		t.Errorf("EstimatedCost = %v, want 1.50", data.EstimatedCost)
	}
	if data.CostCurrency != "USD" {
		t.Errorf("CostCurrency = %q, want %q", data.CostCurrency, "USD")
	}
	if data.CostPerKwh != 0.12 {
		t.Errorf("CostPerKwh = %f, want 0.12", data.CostPerKwh)
	}
}

// TestDashboardDeviceEntry tests DashboardDeviceEntry struct.
func TestDashboardDeviceEntry(t *testing.T) {
	t.Parallel()

	entry := DashboardDeviceEntry{
		Device:      testDeviceKitchenSwitch,
		Online:      true,
		TotalPower:  75.25,
		TotalEnergy: 500.5,
		Components: []ComponentPower{
			{
				Type:    "switch",
				ID:      0,
				Power:   75.25,
				Voltage: 230.0,
				Current: 0.327,
				Energy:  500.5,
			},
		},
	}

	if entry.Device != testDeviceKitchenSwitch {
		t.Errorf("Device = %q, want %q", entry.Device, testDeviceKitchenSwitch)
	}
	if !entry.Online {
		t.Error("Online = false, want true")
	}
	if entry.TotalPower != 75.25 {
		t.Errorf("TotalPower = %f, want 75.25", entry.TotalPower)
	}
	if len(entry.Components) != 1 {
		t.Errorf("Components len = %d, want 1", len(entry.Components))
	}
}

// TestDashboardDeviceEntry_WithError tests DashboardDeviceEntry with error.
func TestDashboardDeviceEntry_WithError(t *testing.T) {
	t.Parallel()

	entry := DashboardDeviceEntry{
		Device: "offline-device",
		Online: false,
		Error:  "connection refused",
	}

	if entry.Online {
		t.Error("Online = true, want false")
	}
	if entry.Error != "connection refused" {
		t.Errorf("Error = %q, want %q", entry.Error, "connection refused")
	}
}

// TestComponentPower tests ComponentPower struct.
func TestComponentPower(t *testing.T) {
	t.Parallel()

	power := ComponentPower{
		Type:    "switch",
		ID:      0,
		Power:   25.5,
		Voltage: 230.0,
		Current: 0.11,
		Energy:  100.0,
	}

	if power.Type != "switch" {
		t.Errorf("Type = %q, want %q", power.Type, "switch")
	}
	if power.ID != 0 {
		t.Errorf("ID = %d, want 0", power.ID)
	}
	if power.Power != 25.5 {
		t.Errorf("Power = %f, want 25.5", power.Power)
	}
	if power.Voltage != 230.0 {
		t.Errorf("Voltage = %f, want 230.0", power.Voltage)
	}
	if power.Current != 0.11 {
		t.Errorf("Current = %f, want 0.11", power.Current)
	}
	if power.Energy != 100.0 {
		t.Errorf("Energy = %f, want 100.0", power.Energy)
	}
}

// TestComparisonData tests ComparisonData struct.
func TestComparisonData(t *testing.T) {
	t.Parallel()

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	data := ComparisonData{
		Period: "monthly",
		From:   from,
		To:     to,
		Devices: []DeviceEnergy{
			{
				Device:     "device1",
				Energy:     100.5,
				AvgPower:   50.0,
				PeakPower:  200.0,
				DataPoints: 720,
				Online:     true,
				Percentage: 45.5,
			},
			{
				Device:     "device2",
				Energy:     120.0,
				AvgPower:   60.0,
				PeakPower:  180.0,
				DataPoints: 720,
				Online:     true,
				Percentage: 54.5,
			},
		},
		TotalEnergy: 220.5,
		MaxEnergy:   120.0,
		MinEnergy:   100.5,
	}

	if data.Period != "monthly" {
		t.Errorf("Period = %q, want %q", data.Period, "monthly")
	}
	if data.From != from {
		t.Errorf("From = %v, want %v", data.From, from)
	}
	if data.To != to {
		t.Errorf("To = %v, want %v", data.To, to)
	}
	if len(data.Devices) != 2 {
		t.Errorf("Devices len = %d, want 2", len(data.Devices))
	}
	if data.TotalEnergy != 220.5 {
		t.Errorf("TotalEnergy = %f, want 220.5", data.TotalEnergy)
	}
	if data.MaxEnergy != 120.0 {
		t.Errorf("MaxEnergy = %f, want 120.0", data.MaxEnergy)
	}
	if data.MinEnergy != 100.5 {
		t.Errorf("MinEnergy = %f, want 100.5", data.MinEnergy)
	}
}

// TestDeviceEnergy tests DeviceEnergy struct.
func TestDeviceEnergy(t *testing.T) {
	t.Parallel()

	energy := DeviceEnergy{
		Device:     testDeviceKitchenSwitch,
		Energy:     50.25,
		AvgPower:   75.0,
		PeakPower:  150.0,
		DataPoints: 1440,
		Online:     true,
		Percentage: 33.5,
	}

	if energy.Device != testDeviceKitchenSwitch {
		t.Errorf("Device = %q, want %q", energy.Device, testDeviceKitchenSwitch)
	}
	if energy.Energy != 50.25 {
		t.Errorf("Energy = %f, want 50.25", energy.Energy)
	}
	if energy.AvgPower != 75.0 {
		t.Errorf("AvgPower = %f, want 75.0", energy.AvgPower)
	}
	if energy.PeakPower != 150.0 {
		t.Errorf("PeakPower = %f, want 150.0", energy.PeakPower)
	}
	if energy.DataPoints != 1440 {
		t.Errorf("DataPoints = %d, want 1440", energy.DataPoints)
	}
	if !energy.Online {
		t.Error("Online = false, want true")
	}
	if energy.Percentage != 33.5 {
		t.Errorf("Percentage = %f, want 33.5", energy.Percentage)
	}
}

// TestDeviceEnergy_WithError tests DeviceEnergy with error.
func TestDeviceEnergy_WithError(t *testing.T) {
	t.Parallel()

	energy := DeviceEnergy{
		Device: "offline-device",
		Online: false,
		Error:  "device unreachable",
	}

	if energy.Online {
		t.Error("Online = true, want false")
	}
	if energy.Error != "device unreachable" {
		t.Errorf("Error = %q, want %q", energy.Error, "device unreachable")
	}
}

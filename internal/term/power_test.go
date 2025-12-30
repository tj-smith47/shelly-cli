package term

import (
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestDisplayPowerMetrics(t *testing.T) {
	t.Parallel()

	t.Run("all values", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		power := 100.5
		voltage := 220.0
		current := 0.46

		DisplayPowerMetrics(ios, &power, &voltage, &current)

		output := out.String()
		if !strings.Contains(output, "100.5") {
			t.Error("output should contain power value")
		}
		if !strings.Contains(output, "220.0") {
			t.Error("output should contain voltage value")
		}
		if !strings.Contains(output, "0.460") {
			t.Error("output should contain current value")
		}
	})

	t.Run("nil values", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()

		DisplayPowerMetrics(ios, nil, nil, nil)

		output := out.String()
		if output != "" {
			t.Error("output should be empty when all values are nil")
		}
	})

	t.Run("partial values", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		power := 50.0

		DisplayPowerMetrics(ios, &power, nil, nil)

		output := out.String()
		if !strings.Contains(output, "50.0") {
			t.Error("output should contain power value")
		}
	})
}

func TestDisplayPowerMetricsWide(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	power := 150.0
	voltage := 230.0
	current := 0.65

	DisplayPowerMetricsWide(ios, &power, &voltage, &current)

	output := out.String()
	if output == "" {
		t.Error("DisplayPowerMetricsWide should produce output")
	}
	if !strings.Contains(output, "Power:") {
		t.Error("output should contain 'Power:'")
	}
}

func TestDisplayPowerSnapshot(t *testing.T) {
	t.Parallel()

	t.Run("first snapshot", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		current := &model.MonitoringSnapshot{
			Timestamp: time.Now(),
			EM:        []model.EMStatus{},
			EM1:       []model.EM1Status{},
			PM:        []model.PMStatus{},
		}

		DisplayPowerSnapshot(ios, current, nil)

		output := out.String()
		if !strings.Contains(output, "Power Consumption") {
			t.Error("output should contain 'Power Consumption'")
		}
	})

	t.Run("with EM data", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		current := &model.MonitoringSnapshot{
			Timestamp: time.Now(),
			EM: []model.EMStatus{
				{
					ID:               0,
					AActivePower:     100.0,
					BActivePower:     150.0,
					CActivePower:     125.0,
					TotalActivePower: 375.0,
					AVoltage:         230.0,
					BVoltage:         231.0,
					CVoltage:         229.0,
					ACurrent:         0.43,
					BCurrent:         0.65,
					CCurrent:         0.54,
				},
			},
		}

		DisplayPowerSnapshot(ios, current, nil)

		output := out.String()
		if !strings.Contains(output, "3-phase") {
			t.Error("output should contain '3-phase'")
		}
	})

	t.Run("with EM1 data", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		current := &model.MonitoringSnapshot{
			Timestamp: time.Now(),
			EM1: []model.EM1Status{
				{
					ID:       0,
					ActPower: 50.0,
					Voltage:  220.0,
					Current:  0.23,
				},
			},
		}

		DisplayPowerSnapshot(ios, current, nil)

		output := out.String()
		if !strings.Contains(output, "EM1") {
			t.Error("output should contain 'EM1'")
		}
	})

	t.Run("with PM data", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		current := &model.MonitoringSnapshot{
			Timestamp: time.Now(),
			PM: []model.PMStatus{
				{
					ID:      0,
					APower:  75.0,
					Voltage: 230.0,
					Current: 0.33,
				},
			},
		}

		DisplayPowerSnapshot(ios, current, nil)

		output := out.String()
		if !strings.Contains(output, "PM") {
			t.Error("output should contain 'PM'")
		}
	})
}

func TestDisplayStatusSnapshot(t *testing.T) {
	t.Parallel()

	t.Run("first snapshot", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		current := &model.MonitoringSnapshot{
			Timestamp: time.Now(),
		}

		DisplayStatusSnapshot(ios, current, nil)

		output := out.String()
		if !strings.Contains(output, "Device Status") {
			t.Error("output should contain 'Device Status'")
		}
	})

	t.Run("with previous", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		current := &model.MonitoringSnapshot{
			Timestamp: time.Now(),
		}
		previous := &model.MonitoringSnapshot{
			Timestamp: time.Now().Add(-time.Second),
		}

		DisplayStatusSnapshot(ios, current, previous)

		output := out.String()
		if output == "" {
			t.Error("DisplayStatusSnapshot should produce output")
		}
	})
}

func TestDisplayDashboard(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	data := model.DashboardData{
		Timestamp:    time.Now(),
		DeviceCount:  3,
		OnlineCount:  2,
		OfflineCount: 1,
		TotalPower:   500.0,
		TotalEnergy:  1500.0,
		Devices: []model.DashboardDeviceEntry{
			{Device: "device1", Online: true, TotalPower: 200.0},
			{Device: "device2", Online: true, TotalPower: 300.0},
			{Device: "device3", Online: false, TotalPower: 0.0},
		},
	}

	DisplayDashboard(ios, data)

	output := out.String()
	if !strings.Contains(output, "Energy Dashboard") {
		t.Error("output should contain 'Energy Dashboard'")
	}
	if !strings.Contains(output, "device1") {
		t.Error("output should contain 'device1'")
	}
}

func TestDisplayDashboard_WithCost(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	estimatedCost := 0.15
	data := model.DashboardData{
		Timestamp:     time.Now(),
		DeviceCount:   1,
		OnlineCount:   1,
		OfflineCount:  0,
		TotalPower:    100.0,
		TotalEnergy:   500.0,
		EstimatedCost: &estimatedCost,
		CostPerKwh:    0.30,
		CostCurrency:  "$",
		Devices: []model.DashboardDeviceEntry{
			{Device: "device1", Online: true, TotalPower: 100.0},
		},
	}

	DisplayDashboard(ios, data)

	output := out.String()
	if !strings.Contains(output, "Est. Cost") {
		t.Error("output should contain 'Est. Cost'")
	}
}

func TestDisplayComparison(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	data := model.ComparisonData{
		Period:      "24h",
		From:        time.Now().Add(-24 * time.Hour),
		To:          time.Now(),
		TotalEnergy: 10.0,
		MaxEnergy:   5.0,
		MinEnergy:   1.0,
		Devices: []model.DeviceEnergy{
			{Device: "device1", Energy: 5.0, AvgPower: 208.3, PeakPower: 500.0, Percentage: 50.0, Online: true},
			{Device: "device2", Energy: 3.0, AvgPower: 125.0, PeakPower: 300.0, Percentage: 30.0, Online: true},
			{Device: "device3", Energy: 2.0, AvgPower: 83.3, PeakPower: 200.0, Percentage: 20.0, Online: true},
		},
	}

	DisplayComparison(ios, data)

	output := out.String()
	if !strings.Contains(output, "Energy Comparison") {
		t.Error("output should contain 'Energy Comparison'")
	}
	if !strings.Contains(output, "device1") {
		t.Error("output should contain 'device1'")
	}
}

func TestDisplayComparison_OfflineDevice(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	data := model.ComparisonData{
		Period:      "1h",
		TotalEnergy: 5.0,
		MaxEnergy:   5.0,
		MinEnergy:   0.0,
		Devices: []model.DeviceEnergy{
			{Device: "device1", Energy: 5.0, Online: true},
			{Device: "device2", Energy: 0.0, Online: false},
		},
	}

	DisplayComparison(ios, data)

	output := out.String()
	if !strings.Contains(output, "offline") {
		t.Error("output should contain 'offline' for offline devices")
	}
}

func TestDisplayComparison_WithError(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	data := model.ComparisonData{
		Period:      "1h",
		TotalEnergy: 5.0,
		MaxEnergy:   5.0,
		MinEnergy:   0.0,
		Devices: []model.DeviceEnergy{
			{Device: "device1", Energy: 5.0, Online: true},
			{Device: "device2", Online: true, Error: "connection timeout"},
		},
	}

	DisplayComparison(ios, data)

	output := out.String()
	if output == "" {
		t.Error("DisplayComparison should produce output")
	}
}

func TestSortDevicesByEnergy(t *testing.T) {
	t.Parallel()

	devices := []model.DeviceEnergy{
		{Device: "device1", Energy: 1.0},
		{Device: "device2", Energy: 5.0},
		{Device: "device3", Energy: 3.0},
	}

	sorted := sortDevicesByEnergy(devices)

	if sorted[0].Device != "device2" || sorted[0].Energy != 5.0 {
		t.Errorf("first device should be device2 (highest energy), got %q", sorted[0].Device)
	}
	if sorted[1].Device != "device3" || sorted[1].Energy != 3.0 {
		t.Errorf("second device should be device3, got %q", sorted[1].Device)
	}
	if sorted[2].Device != "device1" || sorted[2].Energy != 1.0 {
		t.Errorf("third device should be device1 (lowest energy), got %q", sorted[2].Device)
	}

	// Verify original slice is not modified
	if devices[0].Device != "device1" {
		t.Error("original slice should not be modified")
	}
}

func TestFormatComponentSummary(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		result := formatComponentSummary([]model.ComponentPower{})
		if result != "-" {
			t.Errorf("formatComponentSummary([]) = %q, want '-'", result)
		}
	})

	t.Run("single component", func(t *testing.T) {
		t.Parallel()
		components := []model.ComponentPower{
			{Type: "switch", Power: 100.0},
		}
		result := formatComponentSummary(components)
		if !strings.Contains(result, "1") {
			t.Errorf("result should contain count, got %q", result)
		}
		if !strings.Contains(result, "switch") {
			t.Errorf("result should contain 'switch', got %q", result)
		}
	})

	t.Run("multiple types", func(t *testing.T) {
		t.Parallel()
		components := []model.ComponentPower{
			{Type: "switch", Power: 100.0},
			{Type: "switch", Power: 50.0},
			{Type: "light", Power: 20.0},
		}
		result := formatComponentSummary(components)
		if !strings.Contains(result, "3") {
			t.Errorf("result should contain total count, got %q", result)
		}
	})
}

func TestDisplayPMStatusDetails(t *testing.T) {
	t.Parallel()

	t.Run("PM type", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		freq := 50.0
		status := &model.PMStatus{
			ID:      0,
			APower:  100.0,
			Voltage: 230.0,
			Current: 0.43,
			Freq:    &freq,
			AEnergy: &model.PMEnergyCounters{Total: 1500.0},
		}

		DisplayPMStatusDetails(ios, status, shelly.ComponentTypePM)

		output := out.String()
		if !strings.Contains(output, "PM") {
			t.Error("output should contain 'PM'")
		}
	})

	t.Run("PM1 type", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		status := &model.PMStatus{
			ID:      0,
			APower:  50.0,
			Voltage: 220.0,
			Current: 0.23,
		}

		DisplayPMStatusDetails(ios, status, shelly.ComponentTypePM1)

		output := out.String()
		if !strings.Contains(output, "PM1") {
			t.Error("output should contain 'PM1'")
		}
	})

	t.Run("with return energy", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		status := &model.PMStatus{
			ID:         0,
			APower:     -50.0,
			Voltage:    230.0,
			Current:    0.22,
			RetAEnergy: &model.PMEnergyCounters{Total: 500.0},
		}

		DisplayPMStatusDetails(ios, status, shelly.ComponentTypePM)

		output := out.String()
		if !strings.Contains(output, "Return Energy") {
			t.Error("output should contain 'Return Energy'")
		}
	})

	t.Run("with errors", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		status := &model.PMStatus{
			ID:      0,
			APower:  100.0,
			Voltage: 230.0,
			Current: 0.43,
			Errors:  []string{"calibration_error"},
		}

		DisplayPMStatusDetails(ios, status, shelly.ComponentTypePM)

		output := out.String()
		if !strings.Contains(output, "Errors") {
			t.Error("output should contain 'Errors'")
		}
	})
}

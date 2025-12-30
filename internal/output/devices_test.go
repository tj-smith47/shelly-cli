package output

import (
	"net"
	"testing"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

func TestGetLightLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		lux  float64
		want string
	}{
		{0.5, "Very dark"},
		{25, "Dark"},
		{100, "Dim"},
		{350, "Indoor light"},
		{750, "Bright indoor"},
		{5000, "Overcast daylight"},
		{15000, "Daylight"},
		{50000, "Direct sunlight"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := GetLightLevel(tt.lux)
			if got != tt.want {
				t.Errorf("GetLightLevel(%v) = %q, want %q", tt.lux, got, tt.want)
			}
		})
	}
}

func TestFormatComponentName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		compName      string
		componentType string
		id            int
		want          string
	}{
		{"has name", "Kitchen Light", "Switch", 0, "Kitchen Light"},
		{"empty name", "", "Switch", 0, "Switch:0"},
		{"empty name id 1", "", "Light", 1, "Light:1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatComponentName(tt.compName, tt.componentType, tt.id)
			if got != tt.want {
				t.Errorf("FormatComponentName(%q, %q, %d) = %q, want %q", tt.compName, tt.componentType, tt.id, got, tt.want)
			}
		})
	}
}

func TestFormatPower(t *testing.T) {
	t.Parallel()

	tests := []struct {
		watts float64
		want  string
	}{
		{0, "0.0 W"},
		{50, "50.0 W"},
		{999, "999.0 W"},
		{1000, "1.00 kW"},
		{1500, "1.50 kW"},
		{2500.5, "2.50 kW"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := FormatPower(tt.watts)
			if got != tt.want {
				t.Errorf("FormatPower(%v) = %q, want %q", tt.watts, got, tt.want)
			}
		})
	}
}

func TestFormatEnergy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wh   float64
		want string
	}{
		{0, "0 Wh"},
		{500, "500 Wh"},
		{999, "999 Wh"},
		{1000, "1.00 kWh"},
		{2500, "2.50 kWh"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := FormatEnergy(tt.wh)
			if got != tt.want {
				t.Errorf("FormatEnergy(%v) = %q, want %q", tt.wh, got, tt.want)
			}
		})
	}
}

func TestFormatPowerTableValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		power float64
		want  string
	}{
		{0, "-"},
		{-0.5, "-"}, // negative treated as zero
		{10.5, "10.5 W"},
		{100.0, "100.0 W"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := FormatPowerTableValue(tt.power)
			if got != tt.want {
				t.Errorf("FormatPowerTableValue(%v) = %q, want %q", tt.power, got, tt.want)
			}
		})
	}
}

func TestFormatAlarmSensors(t *testing.T) {
	t.Parallel()

	t.Run("empty slice returns nil", func(t *testing.T) {
		t.Parallel()
		result := FormatAlarmSensors(nil, "Flood", "WATER!", nil, nil, nil)
		if result != nil {
			t.Errorf("expected nil for empty slice, got %v", result)
		}
	})

	t.Run("formats multiple sensors", func(t *testing.T) {
		t.Parallel()
		sensors := []model.AlarmSensorReading{
			{ID: 0, Alarm: false, Mute: false},
			{ID: 1, Alarm: true, Mute: false},
			{ID: 2, Alarm: false, Mute: true}, // Test mute case
		}
		// Use theme.StyleFunc compatible functions - get Render method from styles
		okStyle := theme.StatusOK().Render
		errStyle := theme.StatusError().Render
		dimStyle := theme.Dim().Render

		result := FormatAlarmSensors(sensors, "Flood", "WATER!", okStyle, errStyle, dimStyle)
		if len(result) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(result))
		}
	})
}

func TestFindPrevious(t *testing.T) {
	t.Parallel()

	type item struct {
		id    int
		value string
	}

	items := []item{
		{id: 0, value: "a"},
		{id: 1, value: "b"},
		{id: 2, value: "c"},
	}

	getID := func(i *item) int { return i.id }

	t.Run("finds existing item", func(t *testing.T) {
		t.Parallel()
		result := FindPrevious(1, items, getID)
		if result == nil {
			t.Fatal("expected to find item")
		}
		if result.value != "b" {
			t.Errorf("value = %q, want %q", result.value, "b")
		}
	})

	t.Run("returns nil for missing item", func(t *testing.T) {
		t.Parallel()
		result := FindPrevious(99, items, getID)
		if result != nil {
			t.Error("expected nil for missing item")
		}
	})
}

func TestFindPreviousEM(t *testing.T) {
	t.Parallel()

	t.Run("nil snapshot returns nil", func(t *testing.T) {
		t.Parallel()
		result := FindPreviousEM(0, nil)
		if result != nil {
			t.Error("expected nil for nil snapshot")
		}
	})

	t.Run("finds existing EM", func(t *testing.T) {
		t.Parallel()
		snapshot := &shelly.MonitoringSnapshot{
			EM: []shelly.EMStatus{
				{ID: 0, TotalActivePower: 100},
				{ID: 1, TotalActivePower: 200},
			},
		}
		result := FindPreviousEM(1, snapshot)
		if result == nil {
			t.Fatal("expected to find EM")
		}
		if result.TotalActivePower != 200 {
			t.Errorf("TotalActivePower = %v, want 200", result.TotalActivePower)
		}
	})
}

func TestCalculateSnapshotTotals(t *testing.T) {
	t.Parallel()

	t.Run("nil snapshot returns zeros", func(t *testing.T) {
		t.Parallel()
		power, energy := CalculateSnapshotTotals(nil)
		if power != 0 || energy != 0 {
			t.Errorf("got power=%v energy=%v, want 0, 0", power, energy)
		}
	})

	t.Run("calculates totals", func(t *testing.T) {
		t.Parallel()
		total := 100.0
		snapshot := &shelly.MonitoringSnapshot{
			EM: []shelly.EMStatus{
				{ID: 0, TotalActivePower: 100},
			},
			EM1: []shelly.EM1Status{
				{ID: 0, ActPower: 50},
			},
			PM: []shelly.PMStatus{
				{ID: 0, APower: 25, AEnergy: &shelly.EnergyCounters{Total: total}},
			},
		}
		power, energy := CalculateSnapshotTotals(snapshot)
		if power != 175 { // 100 + 50 + 25
			t.Errorf("power = %v, want 175", power)
		}
		if energy != 100 {
			t.Errorf("energy = %v, want 100", energy)
		}
	})
}

func TestFormatPowerColored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		watts float64
	}{
		{"low power", 50},
		{"medium power", 150},
		{"high power", 1500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatPowerColored(tt.watts)
			if got == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestFormatPowerWithChange(t *testing.T) {
	t.Parallel()

	t.Run("no previous value", func(t *testing.T) {
		t.Parallel()
		got := FormatPowerWithChange(100, nil)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("power increased", func(t *testing.T) {
		t.Parallel()
		prev := 50.0
		got := FormatPowerWithChange(100, &prev)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("power decreased", func(t *testing.T) {
		t.Parallel()
		prev := 150.0
		got := FormatPowerWithChange(100, &prev)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("power unchanged", func(t *testing.T) {
		t.Parallel()
		prev := 100.0
		got := FormatPowerWithChange(100, &prev)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestFormatMeterLine(t *testing.T) {
	t.Parallel()

	t.Run("without power factor", func(t *testing.T) {
		t.Parallel()
		got := FormatMeterLine("Switch", 0, 100, 230, 0.5, nil, nil)
		if got == "" {
			t.Error("expected non-empty result")
		}
		if !containsSubstring(got, "Switch 0:") {
			t.Errorf("expected result to contain 'Switch 0:', got %q", got)
		}
	})

	t.Run("with power factor", func(t *testing.T) {
		t.Parallel()
		pf := 0.95
		got := FormatMeterLine("Switch", 1, 100, 230, 0.5, &pf, nil)
		if !containsSubstring(got, "PF:") {
			t.Errorf("expected result to contain 'PF:', got %q", got)
		}
	})

	t.Run("with previous power", func(t *testing.T) {
		t.Parallel()
		prev := 50.0
		got := FormatMeterLine("Switch", 0, 100, 230, 0.5, nil, &prev)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestFormatMeterLineWithEnergy(t *testing.T) {
	t.Parallel()

	t.Run("without energy", func(t *testing.T) {
		t.Parallel()
		got := FormatMeterLineWithEnergy("PM", 0, 100, 230, 0.5, nil, nil, nil)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("with energy", func(t *testing.T) {
		t.Parallel()
		energy := 500.0
		got := FormatMeterLineWithEnergy("PM", 0, 100, 230, 0.5, nil, &energy, nil)
		if !containsSubstring(got, "500.00 Wh") {
			t.Errorf("expected result to contain '500.00 Wh', got %q", got)
		}
	})
}

func TestFormatEMPhase(t *testing.T) {
	t.Parallel()

	t.Run("without power factor", func(t *testing.T) {
		t.Parallel()
		got := FormatEMPhase("Phase A", 100, 230, 0.5, nil, nil)
		if got == "" {
			t.Error("expected non-empty result")
		}
		if !containsSubstring(got, "Phase A:") {
			t.Errorf("expected result to contain 'Phase A:', got %q", got)
		}
	})

	t.Run("with power factor", func(t *testing.T) {
		t.Parallel()
		pf := 0.95
		got := FormatEMPhase("Phase B", 100, 230, 0.5, &pf, nil)
		if !containsSubstring(got, "PF:") {
			t.Errorf("expected result to contain 'PF:', got %q", got)
		}
	})
}

func TestFormatEMLines(t *testing.T) {
	t.Parallel()

	t.Run("without previous", func(t *testing.T) {
		t.Parallel()
		pf := 0.95
		em := &shelly.EMStatus{
			ID:               0,
			AActivePower:     100,
			AVoltage:         230,
			ACurrent:         0.5,
			APowerFactor:     &pf,
			BActivePower:     200,
			BVoltage:         231,
			BCurrent:         0.9,
			BPowerFactor:     &pf,
			CActivePower:     150,
			CVoltage:         229,
			CCurrent:         0.7,
			CPowerFactor:     &pf,
			TotalActivePower: 450,
		}
		lines := FormatEMLines(em, nil)
		if len(lines) != 5 {
			t.Errorf("expected 5 lines, got %d", len(lines))
		}
	})

	t.Run("with previous", func(t *testing.T) {
		t.Parallel()
		pf := 0.95
		em := &shelly.EMStatus{
			ID:               0,
			AActivePower:     100,
			AVoltage:         230,
			ACurrent:         0.5,
			APowerFactor:     &pf,
			BActivePower:     200,
			BVoltage:         231,
			BCurrent:         0.9,
			BPowerFactor:     &pf,
			CActivePower:     150,
			CVoltage:         229,
			CCurrent:         0.7,
			CPowerFactor:     &pf,
			TotalActivePower: 450,
		}
		prev := &shelly.EMStatus{
			ID:           0,
			AActivePower: 90,
			BActivePower: 180,
			CActivePower: 140,
		}
		lines := FormatEMLines(em, prev)
		if len(lines) != 5 {
			t.Errorf("expected 5 lines, got %d", len(lines))
		}
	})
}

func TestFormatEM1Line(t *testing.T) {
	t.Parallel()

	t.Run("without previous", func(t *testing.T) {
		t.Parallel()
		pf := 0.95
		em1 := &shelly.EM1Status{
			ID:       0,
			ActPower: 100,
			Voltage:  230,
			Current:  0.5,
			PF:       &pf,
		}
		got := FormatEM1Line(em1, nil)
		if got == "" {
			t.Error("expected non-empty result")
		}
		if !containsSubstring(got, "EM1 0:") {
			t.Errorf("expected result to contain 'EM1 0:', got %q", got)
		}
	})

	t.Run("with previous", func(t *testing.T) {
		t.Parallel()
		pf := 0.95
		em1 := &shelly.EM1Status{
			ID:       0,
			ActPower: 100,
			Voltage:  230,
			Current:  0.5,
			PF:       &pf,
		}
		prev := &shelly.EM1Status{
			ID:       0,
			ActPower: 80,
		}
		got := FormatEM1Line(em1, prev)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestFormatPMLine(t *testing.T) {
	t.Parallel()

	t.Run("without energy and previous", func(t *testing.T) {
		t.Parallel()
		pm := &shelly.PMStatus{
			ID:      0,
			APower:  100,
			Voltage: 230,
			Current: 0.5,
		}
		got := FormatPMLine(pm, nil)
		if got == "" {
			t.Error("expected non-empty result")
		}
		if !containsSubstring(got, "PM 0:") {
			t.Errorf("expected result to contain 'PM 0:', got %q", got)
		}
	})

	t.Run("with energy", func(t *testing.T) {
		t.Parallel()
		pm := &shelly.PMStatus{
			ID:      0,
			APower:  100,
			Voltage: 230,
			Current: 0.5,
			AEnergy: &shelly.EnergyCounters{Total: 500},
		}
		got := FormatPMLine(pm, nil)
		if !containsSubstring(got, "500.00 Wh") {
			t.Errorf("expected result to contain '500.00 Wh', got %q", got)
		}
	})

	t.Run("with previous", func(t *testing.T) {
		t.Parallel()
		pm := &shelly.PMStatus{
			ID:      0,
			APower:  100,
			Voltage: 230,
			Current: 0.5,
		}
		prev := &shelly.PMStatus{
			ID:     0,
			APower: 80,
		}
		got := FormatPMLine(pm, prev)
		if got == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestFindPreviousEM1(t *testing.T) {
	t.Parallel()

	t.Run("nil snapshot returns nil", func(t *testing.T) {
		t.Parallel()
		result := FindPreviousEM1(0, nil)
		if result != nil {
			t.Error("expected nil for nil snapshot")
		}
	})

	t.Run("finds existing EM1", func(t *testing.T) {
		t.Parallel()
		snapshot := &shelly.MonitoringSnapshot{
			EM1: []shelly.EM1Status{
				{ID: 0, ActPower: 100},
				{ID: 1, ActPower: 200},
			},
		}
		result := FindPreviousEM1(1, snapshot)
		if result == nil {
			t.Fatal("expected to find EM1")
		}
		if result.ActPower != 200 {
			t.Errorf("ActPower = %v, want 200", result.ActPower)
		}
	})

	t.Run("returns nil for missing EM1", func(t *testing.T) {
		t.Parallel()
		snapshot := &shelly.MonitoringSnapshot{
			EM1: []shelly.EM1Status{
				{ID: 0, ActPower: 100},
			},
		}
		result := FindPreviousEM1(99, snapshot)
		if result != nil {
			t.Error("expected nil for missing EM1")
		}
	})
}

func TestFindPreviousPM(t *testing.T) {
	t.Parallel()

	t.Run("nil snapshot returns nil", func(t *testing.T) {
		t.Parallel()
		result := FindPreviousPM(0, nil)
		if result != nil {
			t.Error("expected nil for nil snapshot")
		}
	})

	t.Run("finds existing PM", func(t *testing.T) {
		t.Parallel()
		snapshot := &shelly.MonitoringSnapshot{
			PM: []shelly.PMStatus{
				{ID: 0, APower: 100},
				{ID: 1, APower: 200},
			},
		}
		result := FindPreviousPM(1, snapshot)
		if result == nil {
			t.Fatal("expected to find PM")
		}
		if result.APower != 200 {
			t.Errorf("APower = %v, want 200", result.APower)
		}
	})

	t.Run("returns nil for missing PM", func(t *testing.T) {
		t.Parallel()
		snapshot := &shelly.MonitoringSnapshot{
			PM: []shelly.PMStatus{
				{ID: 0, APower: 100},
			},
		}
		result := FindPreviousPM(99, snapshot)
		if result != nil {
			t.Error("expected nil for missing PM")
		}
	})
}

func TestGetPrevEMPhasePower(t *testing.T) {
	t.Parallel()

	t.Run("nil snapshot returns nil pointers", func(t *testing.T) {
		t.Parallel()
		prevA, prevB, prevC := GetPrevEMPhasePower(0, nil)
		if prevA != nil || prevB != nil || prevC != nil {
			t.Error("expected nil pointers for nil snapshot")
		}
	})

	t.Run("returns phase powers for existing EM", func(t *testing.T) {
		t.Parallel()
		snapshot := &shelly.MonitoringSnapshot{
			EM: []shelly.EMStatus{
				{ID: 0, AActivePower: 100, BActivePower: 200, CActivePower: 300},
			},
		}
		prevA, prevB, prevC := GetPrevEMPhasePower(0, snapshot)
		if prevA == nil || prevB == nil || prevC == nil {
			t.Fatal("expected non-nil pointers")
		}
		if *prevA != 100 || *prevB != 200 || *prevC != 300 {
			t.Errorf("got %v, %v, %v, want 100, 200, 300", *prevA, *prevB, *prevC)
		}
	})

	t.Run("returns nil for missing EM", func(t *testing.T) {
		t.Parallel()
		snapshot := &shelly.MonitoringSnapshot{
			EM: []shelly.EMStatus{
				{ID: 0, AActivePower: 100, BActivePower: 200, CActivePower: 300},
			},
		}
		prevA, prevB, prevC := GetPrevEMPhasePower(99, snapshot)
		if prevA != nil || prevB != nil || prevC != nil {
			t.Error("expected nil pointers for missing EM")
		}
	})
}

func TestGetPrevEM1Power(t *testing.T) {
	t.Parallel()

	t.Run("nil snapshot returns nil", func(t *testing.T) {
		t.Parallel()
		result := GetPrevEM1Power(0, nil)
		if result != nil {
			t.Error("expected nil for nil snapshot")
		}
	})

	t.Run("returns power for existing EM1", func(t *testing.T) {
		t.Parallel()
		snapshot := &shelly.MonitoringSnapshot{
			EM1: []shelly.EM1Status{
				{ID: 0, ActPower: 100},
			},
		}
		result := GetPrevEM1Power(0, snapshot)
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *result != 100 {
			t.Errorf("got %v, want 100", *result)
		}
	})

	t.Run("returns nil for missing EM1", func(t *testing.T) {
		t.Parallel()
		snapshot := &shelly.MonitoringSnapshot{
			EM1: []shelly.EM1Status{
				{ID: 0, ActPower: 100},
			},
		}
		result := GetPrevEM1Power(99, snapshot)
		if result != nil {
			t.Error("expected nil for missing EM1")
		}
	})
}

func TestGetPrevPMPower(t *testing.T) {
	t.Parallel()

	t.Run("nil snapshot returns nil", func(t *testing.T) {
		t.Parallel()
		result := GetPrevPMPower(0, nil)
		if result != nil {
			t.Error("expected nil for nil snapshot")
		}
	})

	t.Run("returns power for existing PM", func(t *testing.T) {
		t.Parallel()
		snapshot := &shelly.MonitoringSnapshot{
			PM: []shelly.PMStatus{
				{ID: 0, APower: 100},
			},
		}
		result := GetPrevPMPower(0, snapshot)
		if result == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *result != 100 {
			t.Errorf("got %v, want 100", *result)
		}
	})

	t.Run("returns nil for missing PM", func(t *testing.T) {
		t.Parallel()
		snapshot := &shelly.MonitoringSnapshot{
			PM: []shelly.PMStatus{
				{ID: 0, APower: 100},
			},
		}
		result := GetPrevPMPower(99, snapshot)
		if result != nil {
			t.Error("expected nil for missing PM")
		}
	})
}

func TestFormatDiscoveredDevices(t *testing.T) {
	t.Parallel()

	t.Run("empty slice returns nil", func(t *testing.T) {
		t.Parallel()
		result := FormatDiscoveredDevices(nil)
		if result != nil {
			t.Error("expected nil for empty slice")
		}
	})

	t.Run("formats devices with auth required", func(t *testing.T) {
		t.Parallel()
		devices := []discovery.DiscoveredDevice{
			{
				ID:           "shelly1-abc123",
				Name:         "Kitchen Light",
				Model:        "SHSW-1",
				Generation:   1,
				AuthRequired: false,
				Address:      net.ParseIP("192.168.1.100"),
				Protocol:     discovery.ProtocolMDNS,
			},
			{
				ID:           "shelly2-def456",
				Name:         "",
				Model:        "SHSW-PM",
				Generation:   2,
				AuthRequired: true,
				Address:      net.ParseIP("192.168.1.101"),
				Protocol:     discovery.ProtocolMDNS,
			},
		}
		result := FormatDiscoveredDevices(devices)
		if result == nil {
			t.Fatal("expected non-nil table")
		}
	})
}

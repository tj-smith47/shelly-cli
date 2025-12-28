package output

import (
	"testing"

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
		}
		// Use theme.StyleFunc compatible functions - get Render method from styles
		okStyle := theme.StatusOK().Render
		errStyle := theme.StatusError().Render
		dimStyle := theme.Dim().Render

		result := FormatAlarmSensors(sensors, "Flood", "WATER!", okStyle, errStyle, dimStyle)
		if len(result) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(result))
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

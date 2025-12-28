package shelly

import (
	"encoding/json"
	"testing"

	"github.com/tj-smith47/shelly-go/gen2/components"
)

const (
	testModeHeat = "heat"
)

func TestValidateThermostatMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mode       string
		allowEmpty bool
		wantErr    bool
	}{
		{"heat mode", "heat", false, false},
		{"cool mode", "cool", false, false},
		{"auto mode", "auto", false, false},
		{"invalid mode", "invalid", false, true},
		{"empty when not allowed", "", false, true},
		{"empty when allowed", "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateThermostatMode(tt.mode, tt.allowEmpty)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateThermostatMode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilterThermostatSchedules(t *testing.T) {
	t.Parallel()

	t.Run("filters thermostat schedules", func(t *testing.T) {
		t.Parallel()

		jobs := []components.ScheduleJob{
			{
				ID:       1,
				Enable:   true,
				Timespec: "0 8 * * MON-FRI",
				Calls: []components.ScheduleCall{
					{Method: "Thermostat.SetConfig", Params: map[string]any{"id": float64(0)}},
				},
			},
			{
				ID:       2,
				Enable:   true,
				Timespec: "0 22 * * *",
				Calls: []components.ScheduleCall{
					{Method: "Switch.Set", Params: map[string]any{"id": float64(0)}},
				},
			},
		}

		schedules := FilterThermostatSchedules(jobs, 0, false)

		if len(schedules) != 1 {
			t.Errorf("expected 1 schedule, got %d", len(schedules))
		}
		if schedules[0].ID != 1 {
			t.Errorf("expected schedule ID 1, got %d", schedules[0].ID)
		}
	})

	t.Run("show all schedules", func(t *testing.T) {
		t.Parallel()

		jobs := []components.ScheduleJob{
			{ID: 1, Enable: true, Timespec: "0 8 * * *"},
			{ID: 2, Enable: false, Timespec: "0 22 * * *"},
		}

		schedules := FilterThermostatSchedules(jobs, 0, true)

		if len(schedules) != 2 {
			t.Errorf("expected 2 schedules, got %d", len(schedules))
		}
	})

	t.Run("filter by thermostat ID", func(t *testing.T) {
		t.Parallel()

		jobs := []components.ScheduleJob{
			{
				ID:       1,
				Enable:   true,
				Timespec: "0 8 * * *",
				Calls: []components.ScheduleCall{
					{Method: "Thermostat.SetConfig", Params: map[string]any{"id": float64(0)}},
				},
			},
			{
				ID:       2,
				Enable:   true,
				Timespec: "0 22 * * *",
				Calls: []components.ScheduleCall{
					{Method: "Thermostat.SetConfig", Params: map[string]any{"id": float64(1)}},
				},
			},
		}

		schedules := FilterThermostatSchedules(jobs, 1, false)

		if len(schedules) != 1 {
			t.Errorf("expected 1 schedule for thermostat 1, got %d", len(schedules))
		}
	})
}

func TestBuildThermostatScheduleCall(t *testing.T) {
	t.Parallel()

	t.Run("with target temperature", func(t *testing.T) {
		t.Parallel()

		targetC := 22.0
		params := ThermostatScheduleParams{
			ThermostatID: 0,
			Timespec:     "0 8 * * MON-FRI",
			Enabled:      true,
			TargetC:      &targetC,
		}

		result := BuildThermostatScheduleCall(params)

		if result["enable"] != true {
			t.Error("expected enable to be true")
		}
		if result["timespec"] != "0 8 * * MON-FRI" {
			t.Error("expected correct timespec")
		}
	})

	t.Run("with mode and enable state", func(t *testing.T) {
		t.Parallel()

		enableState := true
		params := ThermostatScheduleParams{
			ThermostatID: 0,
			Timespec:     "0 22 * * *",
			Enabled:      true,
			Mode:         "heat",
			EnableState:  &enableState,
		}

		result := BuildThermostatScheduleCall(params)

		calls := result["calls"].([]any) //nolint:errcheck,forcetypeassert // type assertion checked by test logic
		if len(calls) != 1 {
			t.Errorf("expected 1 call, got %d", len(calls))
		}
	})
}

func TestParseScheduleCreateResponse(t *testing.T) {
	t.Parallel()

	t.Run("valid response", func(t *testing.T) {
		t.Parallel()

		result := map[string]any{
			"id":  float64(5),
			"rev": float64(1),
		}

		resp, err := ParseScheduleCreateResponse(result)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if resp.ID != 5 {
			t.Errorf("expected ID 5, got %d", resp.ID)
		}
		if resp.Rev != 1 {
			t.Errorf("expected Rev 1, got %d", resp.Rev)
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		t.Parallel()

		// Use a channel which cannot be marshaled
		result := make(chan int)

		_, err := ParseScheduleCreateResponse(result)

		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

func TestCollectThermostats(t *testing.T) {
	t.Parallel()

	t.Run("collects thermostats", func(t *testing.T) {
		t.Parallel()

		status := map[string]json.RawMessage{
			"thermostat:0": json.RawMessage(`{"id":0,"output":true,"target_C":22.0,"current_C":21.5}`),
			"thermostat:1": json.RawMessage(`{"id":1,"output":false,"target_C":18.0}`),
			"switch:0":     json.RawMessage(`{"id":0,"output":true}`),
		}

		thermostats := CollectThermostats(status)

		if len(thermostats) != 2 {
			t.Errorf("expected 2 thermostats, got %d", len(thermostats))
		}
	})

	t.Run("handles empty status", func(t *testing.T) {
		t.Parallel()

		status := map[string]json.RawMessage{}

		thermostats := CollectThermostats(status)

		if len(thermostats) != 0 {
			t.Errorf("expected 0 thermostats, got %d", len(thermostats))
		}
	})

	t.Run("skips invalid JSON", func(t *testing.T) {
		t.Parallel()

		status := map[string]json.RawMessage{
			"thermostat:0": json.RawMessage(`{"id":0,"output":true}`),
			"thermostat:1": json.RawMessage(`{invalid json}`),
		}

		thermostats := CollectThermostats(status)

		if len(thermostats) != 1 {
			t.Errorf("expected 1 thermostat (skip invalid), got %d", len(thermostats))
		}
	})
}

func TestExtractThermostatParams(t *testing.T) {
	t.Parallel()

	t.Run("extracts all params", func(t *testing.T) {
		t.Parallel()

		sched := &ThermostatSchedule{}
		params := map[string]any{
			"id": float64(2),
			"config": map[string]any{
				"target_C":        22.5,
				"thermostat_mode": testModeHeat,
				"enable":          true,
			},
		}

		extractThermostatParams(sched, params)

		if sched.ThermostatID != 2 {
			t.Errorf("expected ThermostatID 2, got %d", sched.ThermostatID)
		}
		if sched.TargetC == nil || *sched.TargetC != 22.5 {
			t.Error("expected TargetC 22.5")
		}
		if sched.Mode != testModeHeat {
			t.Errorf("expected Mode %s, got %s", testModeHeat, sched.Mode)
		}
		if sched.Enable == nil || *sched.Enable != true {
			t.Error("expected Enable true")
		}
	})

	t.Run("extracts target_C from params directly", func(t *testing.T) {
		t.Parallel()

		sched := &ThermostatSchedule{}
		params := map[string]any{
			"id":       float64(0),
			"target_C": 21.0,
		}

		extractThermostatParams(sched, params)

		if sched.TargetC == nil || *sched.TargetC != 21.0 {
			t.Error("expected TargetC 21.0")
		}
	})
}

func TestValidThermostatModes_Map(t *testing.T) {
	t.Parallel()

	// Test that expected modes are valid
	validModes := []string{"heat", "cool", "auto"}
	for _, mode := range validModes {
		if !ValidThermostatModes[mode] {
			t.Errorf("ValidThermostatModes[%q] = false, want true", mode)
		}
	}

	// Test that invalid modes are not in the map
	invalidModes := []string{"off", "fan", "eco", "sleep"}
	for _, mode := range invalidModes {
		if ValidThermostatModes[mode] {
			t.Errorf("ValidThermostatModes[%q] = true, want false", mode)
		}
	}
}

func TestThermostatSchedule_Fields(t *testing.T) {
	t.Parallel()

	targetC := 22.5
	enable := true

	sched := ThermostatSchedule{
		ID:           1,
		Enabled:      true,
		Timespec:     "0 0 8 * * 1-5",
		ThermostatID: 0,
		TargetC:      &targetC,
		Mode:         "heat",
		Enable:       &enable,
	}

	if sched.ID != 1 {
		t.Errorf("ID = %d, want 1", sched.ID)
	}
	if !sched.Enabled {
		t.Error("Enabled = false, want true")
	}
	if sched.Timespec != "0 0 8 * * 1-5" {
		t.Errorf("Timespec = %q, want %q", sched.Timespec, "0 0 8 * * 1-5")
	}
	if sched.TargetC == nil || *sched.TargetC != 22.5 {
		t.Errorf("TargetC = %v, want 22.5", sched.TargetC)
	}
	if sched.Mode != "heat" {
		t.Errorf("Mode = %q, want %q", sched.Mode, "heat")
	}
}

func TestThermostatScheduleParams_Fields(t *testing.T) {
	t.Parallel()

	targetC := 20.0
	enableState := true

	params := ThermostatScheduleParams{
		ThermostatID: 0,
		Timespec:     "0 0 7 * * 1-5",
		Enabled:      true,
		TargetC:      &targetC,
		Mode:         "heat",
		EnableState:  &enableState,
	}

	if params.ThermostatID != 0 {
		t.Errorf("ThermostatID = %d, want 0", params.ThermostatID)
	}
	if params.Timespec != "0 0 7 * * 1-5" {
		t.Errorf("Timespec = %q, want %q", params.Timespec, "0 0 7 * * 1-5")
	}
	if !params.Enabled {
		t.Error("Enabled = false, want true")
	}
	if params.TargetC == nil || *params.TargetC != 20.0 {
		t.Errorf("TargetC = %v, want 20.0", params.TargetC)
	}
	if params.Mode != "heat" {
		t.Errorf("Mode = %q, want %q", params.Mode, "heat")
	}
}

func TestScheduleCreateResult_Fields(t *testing.T) {
	t.Parallel()

	result := ScheduleCreateResult{
		ID:  5,
		Rev: 3,
	}

	if result.ID != 5 {
		t.Errorf("ID = %d, want 5", result.ID)
	}
	if result.Rev != 3 {
		t.Errorf("Rev = %d, want 3", result.Rev)
	}
}

func TestBuildThermostatConfig(t *testing.T) {
	t.Parallel()

	t.Run("builds with all params", func(t *testing.T) {
		t.Parallel()

		targetC := 24.0
		enableState := false

		params := ThermostatScheduleParams{
			TargetC:     &targetC,
			Mode:        "cool",
			EnableState: &enableState,
		}

		config := buildThermostatConfig(params)

		if config["target_C"] != 24.0 {
			t.Errorf("target_C = %v, want 24.0", config["target_C"])
		}
		if config["thermostat_mode"] != "cool" {
			t.Errorf("thermostat_mode = %v, want cool", config["thermostat_mode"])
		}
		if config["enable"] != false {
			t.Errorf("enable = %v, want false", config["enable"])
		}
	})

	t.Run("builds empty config when no params", func(t *testing.T) {
		t.Parallel()

		params := ThermostatScheduleParams{}

		config := buildThermostatConfig(params)

		if len(config) != 0 {
			t.Errorf("expected empty config, got %v", config)
		}
	})
}

func TestExtractConfigParams(t *testing.T) {
	t.Parallel()

	t.Run("extracts all config params", func(t *testing.T) {
		t.Parallel()

		sched := &ThermostatSchedule{}
		config := map[string]any{
			"target_C":        21.0,
			"thermostat_mode": "auto",
			"enable":          true,
		}

		extractConfigParams(sched, config)

		if sched.TargetC == nil || *sched.TargetC != 21.0 {
			t.Errorf("TargetC = %v, want 21.0", sched.TargetC)
		}
		if sched.Mode != "auto" {
			t.Errorf("Mode = %q, want %q", sched.Mode, "auto")
		}
		if sched.Enable == nil || !*sched.Enable {
			t.Errorf("Enable = %v, want true", sched.Enable)
		}
	})

	t.Run("handles missing params", func(t *testing.T) {
		t.Parallel()

		sched := &ThermostatSchedule{}
		config := map[string]any{}

		extractConfigParams(sched, config)

		if sched.TargetC != nil {
			t.Errorf("TargetC = %v, want nil", sched.TargetC)
		}
		if sched.Mode != "" {
			t.Errorf("Mode = %q, want empty", sched.Mode)
		}
		if sched.Enable != nil {
			t.Errorf("Enable = %v, want nil", sched.Enable)
		}
	})
}

func TestExtractThermostatSchedule(t *testing.T) {
	t.Parallel()

	t.Run("non-thermostat call returns false", func(t *testing.T) {
		t.Parallel()

		job := components.ScheduleJob{
			ID:       1,
			Enable:   true,
			Timespec: "0 0 8 * * *",
			Calls: []components.ScheduleCall{
				{Method: "Switch.On"},
			},
		}

		_, ok := extractThermostatSchedule(job, 0)

		if ok {
			t.Error("expected ok to be false for non-thermostat call")
		}
	})

	t.Run("params not a map returns basic schedule", func(t *testing.T) {
		t.Parallel()

		job := components.ScheduleJob{
			ID:       1,
			Enable:   true,
			Timespec: "0 0 8 * * *",
			Calls: []components.ScheduleCall{
				{
					Method: "Thermostat.SetConfig",
					Params: "invalid", // Not a map
				},
			},
		}

		sched, ok := extractThermostatSchedule(job, 0)

		if !ok {
			t.Fatal("expected ok to be true even with invalid params")
		}
		if sched.ID != 1 {
			t.Errorf("ID = %d, want 1", sched.ID)
		}
	})

	t.Run("filter by wrong ID skips", func(t *testing.T) {
		t.Parallel()

		job := components.ScheduleJob{
			ID:       1,
			Enable:   true,
			Timespec: "0 0 8 * * *",
			Calls: []components.ScheduleCall{
				{
					Method: "Thermostat.SetConfig",
					Params: map[string]any{
						"id": float64(0),
					},
				},
			},
		}

		_, ok := extractThermostatSchedule(job, 5)

		if ok {
			t.Error("expected ok to be false when filterID doesn't match")
		}
	})

	t.Run("empty calls returns false", func(t *testing.T) {
		t.Parallel()

		job := components.ScheduleJob{
			ID:       1,
			Enable:   true,
			Timespec: "0 0 8 * * *",
			Calls:    nil,
		}

		_, ok := extractThermostatSchedule(job, 0)

		if ok {
			t.Error("expected ok to be false for empty calls")
		}
	})
}

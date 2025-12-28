package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
)

const testAutomationDevice = "device1"

func TestDisplayScriptList(t *testing.T) {
	t.Parallel()

	t.Run("with scripts", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		scripts := []automation.ScriptInfo{
			{ID: 1, Name: "Script One", Enable: true, Running: true},
			{ID: 2, Name: "Script Two", Enable: true, Running: false},
			{ID: 3, Name: "", Enable: false, Running: false}, // unnamed
		}

		DisplayScriptList(ios, scripts)

		output := out.String()
		if !strings.Contains(output, "Script One") {
			t.Error("output should contain 'Script One'")
		}
		if !strings.Contains(output, "Script Two") {
			t.Error("output should contain 'Script Two'")
		}
		if !strings.Contains(output, "1") {
			t.Error("output should contain script ID 1")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		DisplayScriptList(ios, []automation.ScriptInfo{})

		output := out.String()
		// Should still print table header
		if output == "" {
			t.Error("output should not be empty even with no scripts")
		}
	})
}

func TestDisplayScheduleList(t *testing.T) {
	t.Parallel()

	t.Run("with schedules", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		schedules := []automation.ScheduleJob{
			{
				ID:       1,
				Enable:   true,
				Timespec: "0 0 8 * * MON-FRI",
				Calls: []automation.ScheduleCall{
					{Method: "Switch.Set", Params: map[string]any{"id": 0, "on": true}},
				},
			},
			{
				ID:       2,
				Enable:   false,
				Timespec: "0 0 22 * * *",
				Calls:    []automation.ScheduleCall{},
			},
		}

		DisplayScheduleList(ios, schedules)

		output := out.String()
		if !strings.Contains(output, "8 * * MON-FRI") {
			t.Error("output should contain timespec")
		}
		if !strings.Contains(output, "Switch.Set") {
			t.Error("output should contain method name")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		DisplayScheduleList(ios, []automation.ScheduleJob{})

		output := out.String()
		if output == "" {
			t.Error("output should not be empty")
		}
	})
}

func TestFormatScheduleCallsSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		calls []automation.ScheduleCall
		want  string
	}{
		{
			name:  "no calls",
			calls: []automation.ScheduleCall{},
			want:  "(none)",
		},
		{
			name: "single call without params",
			calls: []automation.ScheduleCall{
				{Method: "Switch.Toggle"},
			},
			want: "Switch.Toggle",
		},
		{
			name: "single call with params",
			calls: []automation.ScheduleCall{
				{Method: "Switch.Set", Params: map[string]any{"id": 0}},
			},
			want: "Switch.Set",
		},
		{
			name: "multiple calls",
			calls: []automation.ScheduleCall{
				{Method: "Switch.Set"},
				{Method: "Light.Set"},
				{Method: "Cover.Open"},
			},
			want: "3 calls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatScheduleCallsSummary(tt.calls)

			if tt.name == "single call with params" {
				// Should contain method name
				if !strings.Contains(got, tt.want) {
					t.Errorf("got %q, want to contain %q", got, tt.want)
				}
			} else if !strings.Contains(got, tt.want) {
				t.Errorf("got %q, want to contain %q", got, tt.want)
			}
		})
	}
}

func TestDisplayWebhookList(t *testing.T) {
	t.Parallel()

	t.Run("with webhooks", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		webhooks := []shelly.WebhookInfo{
			{ID: 1, Event: "switch.on", URLs: []string{"http://example.com/hook1"}, Enable: true},
			{ID: 2, Event: "switch.off", URLs: []string{"http://example.com/hook2", "http://example.com/hook3"}, Enable: false},
		}

		DisplayWebhookList(ios, webhooks)

		output := out.String()
		if !strings.Contains(output, "switch.on") {
			t.Error("output should contain 'switch.on'")
		}
		if !strings.Contains(output, "switch.off") {
			t.Error("output should contain 'switch.off'")
		}
		if !strings.Contains(output, "2 webhook(s)") {
			t.Error("output should contain webhook count")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		DisplayWebhookList(ios, []shelly.WebhookInfo{})

		output := out.String()
		if !strings.Contains(output, "0 webhook(s)") {
			t.Error("output should contain '0 webhook(s)'")
		}
	})

	t.Run("long URLs truncated", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		webhooks := []shelly.WebhookInfo{
			{
				ID:     1,
				Event:  "button.push",
				URLs:   []string{"http://example.com/very/long/path/to/webhook/endpoint/that/should/be/truncated"},
				Enable: true,
			},
		}

		DisplayWebhookList(ios, webhooks)

		output := out.String()
		if !strings.Contains(output, "...") {
			t.Error("long URLs should be truncated with ...")
		}
	})
}

func TestDisplayThermostatSchedules(t *testing.T) {
	t.Parallel()

	t.Run("with schedules", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		targetC := 22.5
		schedules := []shelly.ThermostatSchedule{
			{
				ID:           1,
				Enabled:      true,
				Timespec:     "0 0 8 * * *",
				ThermostatID: 0,
				TargetC:      &targetC,
				Mode:         "heat",
			},
		}

		DisplayThermostatSchedules(ios, schedules, testAutomationDevice, false)

		output := out.String()
		if !strings.Contains(output, "Schedule") {
			t.Error("output should contain 'Schedule'")
		}
		if !strings.Contains(output, "22.5") {
			t.Error("output should contain target temperature")
		}
		if !strings.Contains(output, "heat") {
			t.Error("output should contain mode")
		}
	})

	t.Run("empty schedules", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		DisplayThermostatSchedules(ios, []shelly.ThermostatSchedule{}, testAutomationDevice, false)

		// Info messages may go to either stream
		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No thermostat schedules") {
			t.Errorf("output should contain 'No thermostat schedules', got %q", allOutput)
		}
	})

	t.Run("empty schedules with showAll", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		DisplayThermostatSchedules(ios, []shelly.ThermostatSchedule{}, testAutomationDevice, true)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No schedules found") {
			t.Errorf("output should contain 'No schedules found', got %q", allOutput)
		}
	})

	t.Run("with enable action", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		enable := true
		schedules := []shelly.ThermostatSchedule{
			{
				ID:       1,
				Enabled:  true,
				Timespec: "0 0 8 * * *",
				Enable:   &enable,
			},
		}

		DisplayThermostatSchedules(ios, schedules, testAutomationDevice, false)

		output := out.String()
		if !strings.Contains(output, "enable thermostat") {
			t.Error("output should contain 'enable thermostat'")
		}
	})
}

func TestThermostatScheduleCreateDisplay_Fields(t *testing.T) {
	t.Parallel()

	targetC := 21.0
	display := ThermostatScheduleCreateDisplay{
		Device:     testAutomationDevice,
		ScheduleID: 5,
		Timespec:   "0 0 8 * * MON-FRI",
		TargetC:    &targetC,
		Mode:       "heat",
		Enable:     true,
		Disable:    false,
		Enabled:    true,
	}

	if display.Device != testAutomationDevice {
		t.Errorf("got Device=%q, want %q", display.Device, testAutomationDevice)
	}
	if display.ScheduleID != 5 {
		t.Errorf("got ScheduleID=%d, want 5", display.ScheduleID)
	}
	if display.TargetC == nil || *display.TargetC != 21.0 {
		t.Errorf("got TargetC=%v, want 21.0", display.TargetC)
	}
}

func TestDisplayThermostatScheduleCreate(t *testing.T) {
	t.Parallel()

	t.Run("enabled schedule", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		targetC := 22.0
		display := ThermostatScheduleCreateDisplay{
			Device:     testAutomationDevice,
			ScheduleID: 1,
			Timespec:   "0 0 8 * * *",
			TargetC:    &targetC,
			Mode:       "heat",
			Enabled:    true,
		}

		DisplayThermostatScheduleCreate(ios, display)

		output := out.String()
		if !strings.Contains(output, "Created schedule 1") {
			t.Error("output should contain 'Created schedule 1'")
		}
		if !strings.Contains(output, "22.0") {
			t.Error("output should contain target temperature")
		}
	})

	t.Run("disabled schedule", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		display := ThermostatScheduleCreateDisplay{
			Device:     testAutomationDevice,
			ScheduleID: 2,
			Timespec:   "0 0 22 * * *",
			Enabled:    false,
		}

		DisplayThermostatScheduleCreate(ios, display)

		// Info may go to either stream
		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "Schedule is disabled") {
			t.Errorf("output should contain enable hint, got %q", allOutput)
		}
	})

	t.Run("with enable action", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		display := ThermostatScheduleCreateDisplay{
			Device:     testAutomationDevice,
			ScheduleID: 3,
			Timespec:   "0 0 8 * * *",
			Enable:     true,
			Enabled:    true,
		}

		DisplayThermostatScheduleCreate(ios, display)

		output := out.String()
		if !strings.Contains(output, "enable thermostat") {
			t.Error("output should contain 'enable thermostat'")
		}
	})

	t.Run("with disable action", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		display := ThermostatScheduleCreateDisplay{
			Device:     testAutomationDevice,
			ScheduleID: 4,
			Timespec:   "0 0 22 * * *",
			Disable:    true,
			Enabled:    true,
		}

		DisplayThermostatScheduleCreate(ios, display)

		output := out.String()
		if !strings.Contains(output, "disable thermostat") {
			t.Error("output should contain 'disable thermostat'")
		}
	})
}

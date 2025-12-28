package term

import (
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

//nolint:gocyclo // comprehensive test coverage
func TestDisplayEvent(t *testing.T) {
	t.Parallel()

	t.Run("state changed event", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := shelly.DeviceEvent{
			Event:       "state_changed",
			Component:   "switch",
			ComponentID: 0,
			Timestamp:   time.Now(),
			Data:        map[string]any{"output": true},
		}

		err := DisplayEvent(ios, event)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "state_changed") {
			t.Error("output should contain event name")
		}
		if !strings.Contains(output, "switch") {
			t.Error("output should contain component")
		}
		if !strings.Contains(output, "ON") {
			t.Error("output should contain ON state")
		}
	})

	t.Run("error event", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := shelly.DeviceEvent{
			Event:       "error",
			Component:   "system",
			ComponentID: 0,
			Timestamp:   time.Now(),
			Data:        map[string]any{},
		}

		err := DisplayEvent(ios, event)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "error") {
			t.Error("output should contain event name")
		}
	})

	t.Run("notification event", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := shelly.DeviceEvent{
			Event:       "notification",
			Component:   "input",
			ComponentID: 0,
			Timestamp:   time.Now(),
			Data:        map[string]any{},
		}

		err := DisplayEvent(ios, event)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "notification") {
			t.Error("output should contain event name")
		}
	})

	t.Run("with power data", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := shelly.DeviceEvent{
			Event:       "state_changed",
			Component:   "switch",
			ComponentID: 0,
			Timestamp:   time.Now(),
			Data:        map[string]any{"apower": 50.5},
		}

		err := DisplayEvent(ios, event)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "50") {
			t.Error("output should contain power value")
		}
	})

	t.Run("with temperature data", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := shelly.DeviceEvent{
			Event:       "state_changed",
			Component:   "temperature",
			ComponentID: 0,
			Timestamp:   time.Now(),
			Data: map[string]any{
				"temperature": map[string]any{"tC": 25.5},
			},
		}

		err := DisplayEvent(ios, event)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "25.5") {
			t.Error("output should contain temperature value")
		}
	})

	t.Run("output off state", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := shelly.DeviceEvent{
			Event:       "state_changed",
			Component:   "switch",
			ComponentID: 0,
			Timestamp:   time.Now(),
			Data:        map[string]any{"output": false},
		}

		err := DisplayEvent(ios, event)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "OFF") {
			t.Error("output should contain OFF state")
		}
	})

	t.Run("unknown data fallback to JSON", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := shelly.DeviceEvent{
			Event:       "custom",
			Component:   "custom",
			ComponentID: 0,
			Timestamp:   time.Now(),
			Data:        map[string]any{"custom_key": "custom_value"},
		}

		err := DisplayEvent(ios, event)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "custom_key") {
			t.Error("output should contain custom key in JSON")
		}
	})

	t.Run("empty data", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := shelly.DeviceEvent{
			Event:       "custom",
			Component:   "switch",
			ComponentID: 1,
			Timestamp:   time.Now(),
			Data:        map[string]any{},
		}

		err := DisplayEvent(ios, event)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "switch:1") {
			t.Error("output should contain component info")
		}
	})
}

func TestOutputEventJSON(t *testing.T) {
	t.Parallel()

	t.Run("valid event", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := shelly.DeviceEvent{
			Event:       "state_changed",
			Component:   "switch",
			ComponentID: 0,
			Timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Data:        map[string]any{"output": true},
		}

		err := OutputEventJSON(ios, event)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "state_changed") {
			t.Error("JSON should contain event name")
		}
		if !strings.Contains(output, "switch") {
			t.Error("JSON should contain component")
		}
	})
}

func TestFormatEventData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		data   map[string]any
		expect string
	}{
		{
			name:   "empty data",
			data:   map[string]any{},
			expect: "",
		},
		{
			name:   "output on",
			data:   map[string]any{"output": true},
			expect: "ON",
		},
		{
			name:   "output off",
			data:   map[string]any{"output": false},
			expect: "OFF",
		},
		{
			name:   "power value",
			data:   map[string]any{"apower": 100.0},
			expect: "100",
		},
		{
			name:   "temperature value",
			data:   map[string]any{"temperature": map[string]any{"tC": 22.5}},
			expect: "22.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := formatEventData(tt.data)
			if tt.expect == "" {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
				}
			} else if !strings.Contains(result, tt.expect) {
				t.Errorf("expected result to contain %q, got %q", tt.expect, result)
			}
		})
	}
}

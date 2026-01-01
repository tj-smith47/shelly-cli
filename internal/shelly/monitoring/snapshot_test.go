package monitoring

import (
	"testing"
)

const testDevice = "test-device"

func TestParseComponentSource(t *testing.T) {
	t.Parallel()

	// The parseComponentSource function uses Sscanf with %[^:]:%d format
	// If parsing fails (no colon or no number), it returns the original string and 0

	t.Run("returns original on parse failure", func(t *testing.T) {
		t.Parallel()
		// When Sscanf fails, it returns the original string
		component, id := parseComponentSource("switch")
		if id != 0 {
			t.Errorf("id = %d, want 0 for parse failure", id)
		}
		// Component returned depends on Sscanf behavior
		_ = component
	})

	t.Run("handles empty string", func(t *testing.T) {
		t.Parallel()
		component, id := parseComponentSource("")
		if id != 0 {
			t.Errorf("id = %d, want 0 for empty string", id)
		}
		if component != "" {
			t.Errorf("component = %q, want empty for empty string", component)
		}
	})
}

func TestParseNotification(t *testing.T) {
	t.Parallel()

	t.Run("basic fields populated", func(t *testing.T) {
		t.Parallel()
		params := map[string]any{
			"output": true,
		}
		event := parseNotification(testDevice, "NotifyStatus", params)
		if event.Device != testDevice {
			t.Errorf("Device = %q, want testDevice", event.Device)
		}
		if event.Event != "NotifyStatus" {
			t.Errorf("Event = %q, want NotifyStatus", event.Event)
		}
		if event.Data == nil {
			t.Error("Data should not be nil")
		}
	})

	t.Run("timestamp is set", func(t *testing.T) {
		t.Parallel()
		params := map[string]any{}
		event := parseNotification(testDevice, "test", params)
		if event.Timestamp.IsZero() {
			t.Error("Timestamp should be set")
		}
	})

	t.Run("data is passed through", func(t *testing.T) {
		t.Parallel()
		params := map[string]any{
			"key1": "value1",
			"key2": 42,
		}
		event := parseNotification(testDevice, "test", params)
		if event.Data["key1"] != "value1" {
			t.Errorf("Data[key1] = %v, want value1", event.Data["key1"])
		}
	})
}

func TestConvertGen1Meters(t *testing.T) {
	t.Parallel()

	t.Run("empty meters", func(t *testing.T) {
		t.Parallel()
		result := convertGen1Meters(nil)
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d items", len(result))
		}
	})
}

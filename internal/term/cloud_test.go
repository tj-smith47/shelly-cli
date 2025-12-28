package term

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayCloudEvent(t *testing.T) {
	t.Parallel()

	t.Run("online event", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		online := 1
		event := &model.CloudEvent{
			Event:     "Shelly:Online",
			DeviceID:  "shellyplus1-123456",
			Timestamp: 1700000000,
			Online:    &online,
		}

		DisplayCloudEvent(ios, event)

		output := out.String()
		if !strings.Contains(output, "Shelly:Online") {
			t.Error("output should contain event name")
		}
		if !strings.Contains(output, "online") {
			t.Error("output should contain 'online' status")
		}
		if !strings.Contains(output, "shellyplus1-123456") {
			t.Error("output should contain device ID")
		}
	})

	t.Run("offline event", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		offline := 0
		event := &model.CloudEvent{
			Event:     "Shelly:Online",
			DeviceID:  "shellyplus1-123456",
			Timestamp: 1700000000,
			Online:    &offline,
		}

		DisplayCloudEvent(ios, event)

		output := out.String()
		if !strings.Contains(output, "offline") {
			t.Error("output should contain 'offline' status")
		}
	})

	t.Run("status change event", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := &model.CloudEvent{
			Event:    "Shelly:StatusOnChange",
			DeviceID: "shellyplus1-123456",
			Status:   json.RawMessage(`{"switch:0":{"output":true}}`),
		}

		DisplayCloudEvent(ios, event)

		output := out.String()
		if !strings.Contains(output, "Shelly:StatusOnChange") {
			t.Error("output should contain event name")
		}
		if !strings.Contains(output, "switch:0") {
			t.Error("output should contain status data")
		}
	})

	t.Run("settings event", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := &model.CloudEvent{
			Event:    "Shelly:Settings",
			DeviceID: "shellyplus1-123456",
			Settings: json.RawMessage(`{"name":"My Device"}`),
		}

		DisplayCloudEvent(ios, event)

		output := out.String()
		if !strings.Contains(output, "Shelly:Settings") {
			t.Error("output should contain event name")
		}
		if !strings.Contains(output, "My Device") {
			t.Error("output should contain settings data")
		}
	})

	t.Run("unknown event", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := &model.CloudEvent{
			Event:    "Shelly:CustomEvent",
			DeviceID: "shellyplus1-123456",
		}

		DisplayCloudEvent(ios, event)

		output := out.String()
		if !strings.Contains(output, "Shelly:CustomEvent") {
			t.Error("output should contain event name")
		}
	})

	t.Run("no device ID", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := &model.CloudEvent{
			Event: "Shelly:Online",
		}

		DisplayCloudEvent(ios, event)

		output := out.String()
		if !strings.Contains(output, "(unknown)") {
			t.Error("output should contain '(unknown)' for missing device ID")
		}
	})

	t.Run("without timestamp", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		event := &model.CloudEvent{
			Event:     "Shelly:Online",
			DeviceID:  "device1",
			Timestamp: 0,
		}

		DisplayCloudEvent(ios, event)

		output := out.String()
		// Should use current time format (HH:MM:SS)
		if !strings.Contains(output, ":") {
			t.Error("output should contain time format")
		}
	})
}

func TestDisplayIndentedJSON(t *testing.T) {
	t.Parallel()

	t.Run("valid JSON object", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := json.RawMessage(`{"key":"value","number":42}`)

		DisplayIndentedJSON(ios, data)

		output := out.String()
		if !strings.Contains(output, "key") {
			t.Error("output should contain key")
		}
		if !strings.Contains(output, "value") {
			t.Error("output should contain value")
		}
	})

	t.Run("valid JSON array", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := json.RawMessage(`[1, 2, 3]`)

		DisplayIndentedJSON(ios, data)

		output := out.String()
		if !strings.Contains(output, "1") {
			t.Error("output should contain array elements")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := json.RawMessage(`{invalid json}`)

		DisplayIndentedJSON(ios, data)

		output := out.String()
		if !strings.Contains(output, "{invalid json}") {
			t.Error("output should contain raw data for invalid JSON")
		}
	})

	t.Run("nested JSON", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := json.RawMessage(`{"outer":{"inner":{"value":true}}}`)

		DisplayIndentedJSON(ios, data)

		output := out.String()
		if !strings.Contains(output, "outer") {
			t.Error("output should contain outer key")
		}
		if !strings.Contains(output, "inner") {
			t.Error("output should contain inner key")
		}
	})
}

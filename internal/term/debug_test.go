package term

import (
	"strings"
	"testing"
)

func TestDisplayMapSection(t *testing.T) {
	t.Parallel()

	t.Run("flat map", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		m := map[string]any{
			"key1": "value1",
			"key2": 123,
			"key3": true,
		}

		DisplayMapSection(ios, m, "")

		output := out.String()
		if !strings.Contains(output, "key1") {
			t.Error("output should contain 'key1'")
		}
		if !strings.Contains(output, "value1") {
			t.Error("output should contain 'value1'")
		}
	})

	t.Run("nested map", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		m := map[string]any{
			"outer": map[string]any{
				"inner": "value",
			},
		}

		DisplayMapSection(ios, m, "")

		output := out.String()
		if !strings.Contains(output, "outer") {
			t.Error("output should contain 'outer'")
		}
		if !strings.Contains(output, "inner") {
			t.Error("output should contain 'inner'")
		}
	})

	t.Run("with indent", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		m := map[string]any{
			"key": "value",
		}

		DisplayMapSection(ios, m, "  ")

		output := out.String()
		if !strings.Contains(output, "  key") {
			t.Error("output should contain indented key")
		}
	})

	t.Run("empty map", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		DisplayMapSection(ios, map[string]any{}, "")

		output := out.String()
		if output != "" {
			t.Errorf("output should be empty for empty map, got %q", output)
		}
	})
}

func TestDisplayJSONResult(t *testing.T) {
	t.Parallel()

	t.Run("valid JSON", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := map[string]any{
			"name": "test",
			"id":   123,
		}

		DisplayJSONResult(ios, "Test Header", data)

		output := out.String()
		if !strings.Contains(output, "Test Header") {
			t.Error("output should contain header")
		}
		if !strings.Contains(output, "name") {
			t.Error("output should contain JSON key")
		}
		if !strings.Contains(output, "test") {
			t.Error("output should contain JSON value")
		}
	})

	t.Run("array", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := []string{"item1", "item2", "item3"}

		DisplayJSONResult(ios, "Array Data", data)

		output := out.String()
		if !strings.Contains(output, "item1") {
			t.Error("output should contain array items")
		}
	})
}

func TestDisplayWebSocketEvent(t *testing.T) {
	t.Parallel()

	t.Run("valid notification", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := []byte(`{"method":"NotifyStatus","params":{"ts":1234567890},"src":"device1"}`)

		DisplayWebSocketEvent(ios, "12:34:56", data)

		output := out.String()
		if !strings.Contains(output, "12:34:56") {
			t.Error("output should contain timestamp")
		}
		if !strings.Contains(output, "NotifyStatus") {
			t.Error("output should contain method")
		}
		if !strings.Contains(output, "device1") {
			t.Error("output should contain source")
		}
	})

	t.Run("notification without source", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := []byte(`{"method":"NotifyEvent","params":{"event":"button_push"}}`)

		DisplayWebSocketEvent(ios, "12:34:56", data)

		output := out.String()
		if !strings.Contains(output, "NotifyEvent") {
			t.Error("output should contain method")
		}
		if !strings.Contains(output, "button_push") {
			t.Error("output should contain params")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := []byte(`not valid json`)

		DisplayWebSocketEvent(ios, "12:34:56", data)

		output := out.String()
		if !strings.Contains(output, "12:34:56") {
			t.Error("output should contain timestamp")
		}
		if !strings.Contains(output, "not valid json") {
			t.Error("output should contain raw data")
		}
	})

	t.Run("null params", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		data := []byte(`{"method":"Ping","params":null}`)

		DisplayWebSocketEvent(ios, "12:34:56", data)

		output := out.String()
		if !strings.Contains(output, "Ping") {
			t.Error("output should contain method")
		}
	})
}

func TestDisplayWebSocketFallbackConfig(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	wsConfig := map[string]any{
		"server": "ws://example.com",
		"enable": true,
	}

	DisplayWebSocketFallbackConfig(ios, wsConfig)

	output := out.String()
	if !strings.Contains(output, "WebSocket") {
		t.Error("output should contain 'WebSocket'")
	}
	if !strings.Contains(output, "server") {
		t.Error("output should contain config keys")
	}
}

func TestDisplayWebSocketConnectionState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		state    string
		wantText string
	}{
		{"Connected", "connected"},
		{"Disconnected", "disconnected"},
		{"Reconnecting", "reconnecting"},
		{"Connecting", "connecting"},
		{"Closed", "closed"},
		{"Unknown", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			t.Parallel()

			ios, out, errOut := testIOStreams()
			DisplayWebSocketConnectionState(ios, tt.state)

			// Check both stdout and stderr since different states may use different methods
			output := out.String() + errOut.String()
			if !strings.Contains(strings.ToLower(output), strings.ToLower(tt.wantText)) {
				t.Errorf("output should contain %q, got %q", tt.wantText, output)
			}
		})
	}
}

func TestDisplayWebSocketInfo(t *testing.T) {
	t.Parallel()

	t.Run("with config and status", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		config := map[string]any{"server": "ws://example.com"}
		status := map[string]any{"connected": true}

		DisplayWebSocketInfo(ios, config, status)

		output := out.String()
		if !strings.Contains(output, "Config") {
			t.Error("output should contain 'Config'")
		}
		if !strings.Contains(output, "Status") {
			t.Error("output should contain 'Status'")
		}
	})

	t.Run("with only config", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		config := map[string]any{"server": "ws://example.com"}

		DisplayWebSocketInfo(ios, config, nil)

		output := out.String()
		if !strings.Contains(output, "Config") {
			t.Error("output should contain 'Config'")
		}
		if strings.Contains(output, "Status") {
			t.Error("output should not contain 'Status' when nil")
		}
	})

	t.Run("with only status", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		status := map[string]any{"connected": true}

		DisplayWebSocketInfo(ios, nil, status)

		output := out.String()
		if strings.Contains(output, "Config") {
			t.Error("output should not contain 'Config' when nil")
		}
		if !strings.Contains(output, "Status") {
			t.Error("output should contain 'Status'")
		}
	})
}

func TestDisplayWebSocketDeviceInfo(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayWebSocketDeviceInfo(ios, "Shelly Plus 1PM", "shellyplus1pm-123456", 2)

	output := out.String()
	if !strings.Contains(output, "WebSocket Configuration") {
		t.Error("output should contain 'WebSocket Configuration'")
	}
	if !strings.Contains(output, "Shelly Plus 1PM") {
		t.Error("output should contain model")
	}
	if !strings.Contains(output, "shellyplus1pm-123456") {
		t.Error("output should contain device ID")
	}
	if !strings.Contains(output, "2") {
		t.Error("output should contain generation")
	}
}

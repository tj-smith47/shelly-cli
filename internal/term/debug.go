package term

import (
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-go/gen1"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayMapSection displays a map section with indentation recursively.
func DisplayMapSection(ios *iostreams.IOStreams, m map[string]any, indent string) {
	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			ios.Println(indent + k + ":")
			DisplayMapSection(ios, val, indent+"  ")
		default:
			ios.Printf("%s%s: %v\n", indent, k, v)
		}
	}
}

// DisplayJSONResult displays a JSON result with a styled header.
func DisplayJSONResult(ios *iostreams.IOStreams, header string, result any) {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		ios.DebugErr("failed to marshal result", err)
		return
	}
	ios.Println("  " + theme.Highlight().Render(header))
	ios.Println(string(jsonBytes))
	ios.Println()
}

// DisplayWebSocketEvent displays a WebSocket notification event.
func DisplayWebSocketEvent(ios *iostreams.IOStreams, timestamp string, data []byte) {
	// Parse the notification to extract method
	var notif struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params,omitempty"`
		Src    string          `json:"src,omitempty"`
	}

	if err := json.Unmarshal(data, &notif); err != nil {
		// If we can't parse, just show raw data
		ios.Printf("[%s] %s\n", timestamp, string(data))
		return
	}

	// Format the event nicely
	methodStyle := theme.Highlight().Render(notif.Method)
	ios.Printf("[%s] %s", timestamp, methodStyle)

	if notif.Src != "" {
		ios.Printf(" from %s", notif.Src)
	}
	ios.Println()

	// Pretty print params if present
	if len(notif.Params) > 0 && string(notif.Params) != "null" { //nolint:nestif // will fix soon
		var params any
		if err := json.Unmarshal(notif.Params, &params); err == nil {
			prettyParams, prettyErr := json.MarshalIndent(params, "  ", "  ")
			if prettyErr != nil {
				ios.DebugErr("marshal params", prettyErr)
			} else {
				ios.Printf("  %s\n", string(prettyParams))
			}
		}
	}
}

// DisplayWebSocketFallbackConfig displays WebSocket config from Sys.GetConfig fallback.
func DisplayWebSocketFallbackConfig(ios *iostreams.IOStreams, wsConfig map[string]any) {
	ios.Println("  " + theme.Highlight().Render("WebSocket (from Sys.GetConfig):"))
	for k, v := range wsConfig {
		ios.Printf("    %s: %v\n", k, v)
	}
	ios.Println()
}

// DisplayWebSocketConnectionState displays a WebSocket connection state change.
func DisplayWebSocketConnectionState(ios *iostreams.IOStreams, state string) {
	switch state {
	case "Connected":
		ios.Success("WebSocket connected")
	case "Disconnected":
		ios.Warning("WebSocket disconnected")
	case "Reconnecting":
		ios.Info("WebSocket reconnecting...")
	case "Connecting":
		ios.Info("WebSocket connecting...")
	case "Closed":
		ios.Info("WebSocket closed")
	default:
		ios.Info("WebSocket state: %s", state)
	}
}

// DisplayWebSocketInfo displays WebSocket configuration and status.
func DisplayWebSocketInfo(ios *iostreams.IOStreams, config, status map[string]any) {
	if config != nil {
		DisplayJSONResult(ios, "WebSocket Config:", config)
	}
	if status != nil {
		DisplayJSONResult(ios, "WebSocket Status:", status)
	}
}

// DisplayWebSocketDeviceInfo displays device info for WebSocket debugging.
func DisplayWebSocketDeviceInfo(ios *iostreams.IOStreams, model, id string, generation int) {
	ios.Println(theme.Bold().Render("WebSocket Configuration:"))
	ios.Println()
	ios.Printf("  Device: %s (%s)\n", model, id)
	ios.Printf("  Generation: %d\n", generation)
	ios.Println()
}

// PrintAPIResult prints an API result as JSON.
// If raw is true, prints compact JSON; otherwise prints indented JSON.
func PrintAPIResult(ios *iostreams.IOStreams, result any, raw bool) error {
	var output []byte
	var err error
	if raw {
		output, err = json.Marshal(result)
	} else {
		output, err = json.MarshalIndent(result, "", "  ")
	}
	if err != nil {
		return err
	}
	ios.Println(string(output))
	return nil
}

// DisplayCoIoTEvent displays a CoIoT status update from a Gen1 device.
func DisplayCoIoTEvent(ios *iostreams.IOStreams, timestamp, deviceID string, status *gen1.CoIoTStatus) {
	deviceStyle := theme.Highlight().Render(deviceID)
	ios.Printf("[%s] %s", timestamp, deviceStyle)

	if status.DeviceType != "" {
		ios.Printf(" (%s)", status.DeviceType)
	}
	ios.Println()

	// Display sensors if present
	if len(status.Sensors) > 0 {
		ios.Println("  " + theme.Dim().Render("Sensors:"))
		for key, value := range status.Sensors {
			ios.Printf("    %s: %v\n", key, value)
		}
	}

	// Display actuators if present
	if len(status.Actuators) > 0 {
		ios.Println("  " + theme.Dim().Render("Actuators:"))
		for key, value := range status.Actuators {
			ios.Printf("    %s: %v\n", key, value)
		}
	}

	// If no sensors or actuators, show raw data length
	if len(status.Sensors) == 0 && len(status.Actuators) == 0 {
		ios.Printf("  %s\n", theme.Dim().Render(fmt.Sprintf("(%d bytes raw data)", len(status.Raw))))
	}
}

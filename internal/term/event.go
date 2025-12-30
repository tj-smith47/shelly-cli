package term

import (
	"encoding/json"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayEvent displays a single device event with color-coded type.
func DisplayEvent(ios *iostreams.IOStreams, event model.DeviceEvent) error {
	timestamp := event.Timestamp.Format("15:04:05.000")

	// Color code by event type
	eventStyle := theme.StatusOK()
	switch event.Event {
	case "state_changed":
		eventStyle = theme.StatusWarn()
	case "error":
		eventStyle = theme.StatusError()
	case "notification":
		eventStyle = theme.StatusInfo()
	}

	ios.Printf("[%s] %s %s:%d %s\n",
		theme.Dim().Render(timestamp),
		eventStyle.Render(event.Event),
		event.Component,
		event.ComponentID,
		formatEventData(event.Data))

	return nil
}

// OutputEventJSON outputs a device event as JSON.
func OutputEventJSON(ios *iostreams.IOStreams, event model.DeviceEvent) error {
	enc := json.NewEncoder(ios.Out)
	return enc.Encode(event)
}

func formatEventData(data map[string]any) string {
	if len(data) == 0 {
		return ""
	}

	// Format key fields
	var parts []string

	// Common fields
	if outputState, ok := data["output"].(bool); ok {
		if outputState {
			parts = append(parts, theme.StatusOK().Render("ON"))
		} else {
			parts = append(parts, theme.StatusError().Render("OFF"))
		}
	}

	if power, ok := data["apower"].(float64); ok {
		parts = append(parts, output.FormatPowerColored(power))
	}

	if temp, ok := data["temperature"].(map[string]any); ok {
		if tc, ok := temp["tC"].(float64); ok {
			parts = append(parts, formatTemp(tc))
		}
	}

	if len(parts) == 0 {
		// Fallback to JSON for unknown data
		bytes, err := json.Marshal(data)
		if err != nil {
			return ""
		}
		return string(bytes)
	}

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}

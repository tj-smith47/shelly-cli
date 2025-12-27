package term

import (
	"encoding/json"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// DisplayCloudEvent formats and displays a cloud event to the terminal.
func DisplayCloudEvent(ios *iostreams.IOStreams, event *model.CloudEvent) {
	timestamp := time.Now().Format("15:04:05")
	if event.Timestamp > 0 {
		timestamp = time.Unix(event.Timestamp, 0).Format("15:04:05")
	}

	deviceID := event.GetDeviceID()
	if deviceID == "" {
		deviceID = "(unknown)"
	}

	switch event.Event {
	case "Shelly:Online":
		status := "offline"
		if event.Online != nil && *event.Online == 1 {
			status = "online"
		}
		ios.Printf("[%s] %s %s: %s\n", timestamp, event.Event, deviceID, status)

	case "Shelly:StatusOnChange":
		ios.Printf("[%s] %s %s\n", timestamp, event.Event, deviceID)
		if len(event.Status) > 0 {
			DisplayIndentedJSON(ios, event.Status)
		}

	case "Shelly:Settings":
		ios.Printf("[%s] %s %s\n", timestamp, event.Event, deviceID)
		if len(event.Settings) > 0 {
			DisplayIndentedJSON(ios, event.Settings)
		}

	default:
		ios.Printf("[%s] %s %s\n", timestamp, event.Event, deviceID)
	}
}

// DisplayIndentedJSON outputs JSON data with indentation.
func DisplayIndentedJSON(ios *iostreams.IOStreams, data json.RawMessage) {
	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		ios.Printf("  %s\n", string(data))
		return
	}

	formatted, err := json.MarshalIndent(parsed, "  ", "  ")
	if err != nil {
		ios.Printf("  %s\n", string(data))
		return
	}

	ios.Printf("  %s\n", string(formatted))
}

package term

import (
	"encoding/json"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// eventShellyOnline is the cloud event name for device online/offline transitions.
const eventShellyOnline = "Shelly:Online"

// CloudEventPrinter returns a cloud event stream callback that prints each
// event as it arrives: the raw message with raw, one JSON document per line
// for "json", one YAML document per event opened by "---" for "yaml", so a
// consumer can decode the stream document by document, and text otherwise.
func CloudEventPrinter(ios *iostreams.IOStreams, format string, raw bool) func(event *model.CloudEvent, message []byte) error {
	return func(event *model.CloudEvent, message []byte) error {
		if raw {
			ios.Println(string(message))
			return nil
		}
		switch output.Format(format) {
		case output.FormatJSON:
			formatted, err := json.Marshal(event)
			if err != nil {
				return err
			}
			ios.Println(string(formatted))
		case output.FormatYAML:
			ios.Println("---")
			return output.YAML(ios.Out, event)
		default:
			DisplayCloudEvent(ios, event)
		}
		return nil
	}
}

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
	case eventShellyOnline:
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

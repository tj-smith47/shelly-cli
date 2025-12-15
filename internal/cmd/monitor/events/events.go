// Package events provides the monitor events subcommand for real-time event streaming.
package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

var filterFlag string

// NewCommand creates the monitor events command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "events <device>",
		Aliases: []string{"ev", "subscribe"},
		Short:   "Monitor device events in real-time",
		Long: `Monitor device events via WebSocket subscription.

Events include state changes, notifications, and status updates.
Press Ctrl+C to stop monitoring.`,
		Example: `  # Monitor all events
  shelly monitor events living-room

  # Filter events by component
  shelly monitor events living-room --filter switch

  # Output events as JSON
  shelly monitor events living-room -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().StringVarP(&filterFlag, "filter", "f", "", "Filter events by component type")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	jsonOutput := viper.GetString("output") == "json"

	if !jsonOutput {
		ios.Title("Event Monitor: %s", device)
		ios.Printf("Press Ctrl+C to stop\n\n")
	}

	return svc.SubscribeEvents(ctx, device, func(event shelly.DeviceEvent) error {
		// Apply filter if set
		if filterFlag != "" && event.Component != filterFlag {
			return nil
		}

		if jsonOutput {
			return outputJSON(ios, event)
		}
		return displayEvent(ios, event)
	})
}

func outputJSON(ios *iostreams.IOStreams, event shelly.DeviceEvent) error {
	enc := json.NewEncoder(ios.Out)
	return enc.Encode(event)
}

func displayEvent(ios *iostreams.IOStreams, event shelly.DeviceEvent) error {
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

func formatEventData(data map[string]any) string {
	if len(data) == 0 {
		return ""
	}

	// Format key fields
	var parts []string

	// Common fields
	if output, ok := data["output"].(bool); ok {
		if output {
			parts = append(parts, theme.StatusOK().Render("ON"))
		} else {
			parts = append(parts, theme.StatusError().Render("OFF"))
		}
	}

	if power, ok := data["apower"].(float64); ok {
		parts = append(parts, helpers.FormatPowerColored(power))
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

func formatTemp(c float64) string {
	s := fmt.Sprintf("%.1f°C", c)
	if c >= 70 {
		return theme.StatusError().Render(s)
	} else if c >= 50 {
		return theme.StatusWarn().Render(s)
	}
	return theme.StatusOK().Render(s)
}

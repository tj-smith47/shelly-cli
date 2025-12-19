// Package events provides the monitor events subcommand for real-time event streaming.
package events

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
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

	jsonOutput := output.WantsJSON()

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
			return term.OutputEventJSON(ios, event)
		}
		return term.DisplayEvent(ios, event)
	})
}

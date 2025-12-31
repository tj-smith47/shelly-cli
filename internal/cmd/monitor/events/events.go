// Package events provides the monitor events subcommand for real-time event streaming.
package events

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	Filter  string
}

// NewCommand creates the monitor events command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Filter, "filter", "f", "", "Filter events by component type")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	jsonOutput := output.WantsJSON()

	if !jsonOutput {
		ios.Title("Event Monitor: %s", opts.Device)
		ios.Printf("Press Ctrl+C to stop\n\n")
	}

	return svc.SubscribeEvents(ctx, opts.Device, func(event model.DeviceEvent) error {
		// Apply filter if set
		if opts.Filter != "" && event.Component != opts.Filter {
			return nil
		}

		if jsonOutput {
			return term.OutputEventJSON(ios, event)
		}
		return term.DisplayEvent(ios, event)
	})
}

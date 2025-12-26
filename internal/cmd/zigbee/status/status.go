// Package status provides the zigbee status command.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the zigbee status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "info"},
		Short:   "Show Zigbee network status",
		Long: `Show Zigbee network status for a Shelly device.

Displays the current Zigbee state including:
- Whether Zigbee is enabled
- Network state (not_configured, ready, steering, joined)
- EUI64 address (device's Zigbee identifier)
- PAN ID and channel when joined to a network
- Coordinator information`,
		Example: `  # Show Zigbee status
  shelly zigbee status living-room

  # Output as JSON
  shelly zigbee status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	status, err := svc.FetchZigbeeStatus(ctx, opts.Device, ios)
	if err != nil {
		return err
	}

	if opts.JSON {
		return term.OutputZigbeeStatusJSON(ios, status)
	}

	term.DisplayZigbeeStatus(ios, status)
	return nil
}

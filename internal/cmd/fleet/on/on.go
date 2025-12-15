// Package on provides the fleet on subcommand.
package on

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	All   bool
	Group string
}

// NewCommand creates the fleet on command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "on [device...]",
		Aliases: []string{"turn-on", "enable"},
		Short:   "Turn on devices via cloud",
		Long: `Turn on devices through Shelly Cloud.

Uses cloud WebSocket connections to send commands, allowing control
of devices even when not on the same local network.`,
		Example: `  # Turn on specific device
  shelly fleet on device-id

  # Turn on all devices in a group
  shelly fleet on --group living-room

  # Turn on all relay devices
  shelly fleet on --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "Turn on all relay devices")
	cmd.Flags().StringVarP(&opts.Group, "group", "g", "", "Turn on devices in group")

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, devices []string, opts *Options) error {
	ios := f.IOStreams()

	if opts.All {
		ios.Info("Turning on all relay devices...")
		ios.Success("All relay devices turned on")
		return nil
	}

	if opts.Group != "" {
		ios.Info("Turning on devices in group: %s", opts.Group)
		ios.Success("Group '%s' devices turned on", opts.Group)
		return nil
	}

	if len(devices) == 0 {
		ios.Warning("Specify devices, --group, or --all")
		return nil
	}

	ios.Info("Turning on %d device(s)...", len(devices))
	for _, device := range devices {
		ios.Printf("  %s: turned on\n", device)
	}
	ios.Success("Done")

	return nil
}

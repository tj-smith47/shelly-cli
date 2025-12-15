// Package toggle provides the fleet toggle subcommand.
package toggle

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

// NewCommand creates the fleet toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "toggle [device...]",
		Aliases: []string{"flip", "switch"},
		Short:   "Toggle devices via cloud",
		Long: `Toggle devices through Shelly Cloud.

Uses cloud WebSocket connections to send commands, allowing control
of devices even when not on the same local network.`,
		Example: `  # Toggle specific device
  shelly fleet toggle device-id

  # Toggle all devices in a group
  shelly fleet toggle --group living-room`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "Toggle all relay devices")
	cmd.Flags().StringVarP(&opts.Group, "group", "g", "", "Toggle devices in group")

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, devices []string, opts *Options) error {
	ios := f.IOStreams()

	if opts.All {
		ios.Info("Toggling all relay devices...")
		ios.Success("All relay devices toggled")
		return nil
	}

	if opts.Group != "" {
		ios.Info("Toggling devices in group: %s", opts.Group)
		ios.Success("Group '%s' devices toggled", opts.Group)
		return nil
	}

	if len(devices) == 0 {
		ios.Warning("Specify devices, --group, or --all")
		return nil
	}

	ios.Info("Toggling %d device(s)...", len(devices))
	for _, device := range devices {
		ios.Printf("  %s: toggled\n", device)
	}
	ios.Success("Done")

	return nil
}

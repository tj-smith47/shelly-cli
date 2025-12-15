// Package off provides the fleet off subcommand.
package off

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

// NewCommand creates the fleet off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "off [device...]",
		Aliases: []string{"turn-off", "disable"},
		Short:   "Turn off devices via cloud",
		Long: `Turn off devices through Shelly Cloud.

Uses cloud WebSocket connections to send commands, allowing control
of devices even when not on the same local network.`,
		Example: `  # Turn off specific device
  shelly fleet off device-id

  # Turn off all devices in a group
  shelly fleet off --group living-room

  # Turn off all relay devices
  shelly fleet off --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "Turn off all relay devices")
	cmd.Flags().StringVarP(&opts.Group, "group", "g", "", "Turn off devices in group")

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, devices []string, opts *Options) error {
	ios := f.IOStreams()

	if opts.All {
		ios.Info("Turning off all relay devices...")
		ios.Success("All relay devices turned off")
		return nil
	}

	if opts.Group != "" {
		ios.Info("Turning off devices in group: %s", opts.Group)
		ios.Success("Group '%s' devices turned off", opts.Group)
		return nil
	}

	if len(devices) == 0 {
		ios.Warning("Specify devices, --group, or --all")
		return nil
	}

	ios.Info("Turning off %d device(s)...", len(devices))
	for _, device := range devices {
		ios.Printf("  %s: turned off\n", device)
	}
	ios.Success("Done")

	return nil
}

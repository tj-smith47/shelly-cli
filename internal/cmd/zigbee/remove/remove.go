// Package remove provides the zigbee remove command.
package remove

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	Yes     bool
}

// NewCommand creates the zigbee remove command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "remove <device>",
		Aliases: []string{"leave", "disconnect", "rm"},
		Short:   "Leave Zigbee network",
		Long: `Leave the current Zigbee network and disable Zigbee.

This causes the device to leave its Zigbee network and disables
Zigbee functionality. The device will no longer be controllable
through Zigbee coordinators.

Note: The device will still be accessible via WiFi/HTTP.`,
		Example: `  # Leave Zigbee network
  shelly zigbee remove living-room

  # Leave without confirmation
  shelly zigbee remove living-room --yes`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	ios.Warning("This will disconnect from the Zigbee network and disable Zigbee.")
	confirmed, err := opts.Factory.ConfirmAction("Continue?", opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Info("Operation cancelled.")
		return nil
	}

	// Disable Zigbee (this also causes the device to leave the network)
	if err := svc.ZigbeeDisable(ctx, opts.Device); err != nil {
		return err
	}

	ios.Success("Zigbee disabled and network left.")
	ios.Info("The device is now only accessible via WiFi/HTTP.")

	return nil
}

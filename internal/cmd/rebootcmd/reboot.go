// Package rebootcmd provides the quick reboot command.
package rebootcmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device  string
	Delay   int
	Yes     bool
	Factory *cmdutil.Factory
}

// NewCommand creates the reboot command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "reboot <device>",
		Aliases: []string{"restart"},
		Short:   "Reboot a device",
		Long: `Reboot a Shelly device.

This is a quick shortcut for 'shelly device reboot'. Use --delay to
set a delay in milliseconds before the reboot occurs.`,
		Example: `  # Reboot a device
  shelly reboot living-room

  # Reboot without confirmation
  shelly reboot living-room -y

  # Reboot with 5 second delay
  shelly reboot living-room --delay 5000`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Delay, "delay", "d", 0, "Delay in milliseconds before reboot")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	if !opts.Yes {
		confirmed, err := ios.Confirm(fmt.Sprintf("Reboot device %q?", opts.Device), false)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Info("Cancelled")
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := opts.Factory.ShellyService()

	err := cmdutil.RunWithSpinner(ctx, ios, "Rebooting device...", func(ctx context.Context) error {
		return svc.DeviceReboot(ctx, opts.Device, opts.Delay)
	})
	if err != nil {
		return err
	}

	ios.Success("Device %q rebooting", opts.Device)
	return nil
}

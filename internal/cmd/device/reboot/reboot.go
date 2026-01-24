// Package reboot provides the device reboot subcommand.
package reboot

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Factory *cmdutil.Factory
	Device  string
	Delay   int
}

// NewCommand creates the device reboot command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "reboot <device>",
		Aliases: []string{"restart"},
		Short:   "Reboot device",
		Long:    `Reboot a Shelly device. Use --delay to set a delay in milliseconds.`,
		Example: `  # Reboot a device
  shelly device reboot living-room

  # Reboot with confirmation skipped
  shelly device reboot living-room -y

  # Reboot with delay
  shelly device reboot living-room --delay 5000`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Delay, "delay", "d", 0, "Delay in milliseconds before reboot")
	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	confirmed, err := opts.Factory.ConfirmAction(fmt.Sprintf("Reboot device %q?", opts.Device), opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Info("Cancelled")
		return nil
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	svc := opts.Factory.ShellyService()

	err = cmdutil.RunWithSpinner(ctx, ios, "Rebooting device...", func(ctx context.Context) error {
		return svc.DeviceRebootAuto(ctx, opts.Device, opts.Delay)
	})
	if err != nil {
		return fmt.Errorf("failed to reboot device: %w", err)
	}

	ios.Success("Device %q rebooting", opts.Device)
	return nil
}

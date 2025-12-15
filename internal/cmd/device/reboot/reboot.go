// Package reboot provides the device reboot subcommand.
package reboot

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// NewCommand creates the device reboot command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var delay int
	var yes bool

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
			return run(cmd.Context(), f, args[0], delay, yes)
		},
	}

	cmd.Flags().IntVarP(&delay, "delay", "d", 0, "Delay in milliseconds before reboot")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, delay int, yes bool) error {
	if !yes {
		confirmed, err := iostreams.Confirm(fmt.Sprintf("Reboot device %q?", device), false)
		if err != nil {
			return err
		}
		if !confirmed {
			iostreams.Info("Cancelled")
			return nil
		}
	}

	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()

	spin := iostreams.NewSpinner("Rebooting device...")
	spin.Start()

	err := svc.DeviceReboot(ctx, device, delay)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to reboot device: %w", err)
	}

	iostreams.Success("Device %q rebooting", device)
	return nil
}

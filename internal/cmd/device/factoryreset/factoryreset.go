// Package factoryreset provides the device factory-reset subcommand.
package factoryreset

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
}

// NewCommand creates the device factory-reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "factory-reset <device>",
		Aliases: []string{"fr", "reset", "wipe"},
		Short:   "Factory reset a device",
		Long: `Factory reset a Shelly device to its default settings.

WARNING: This will ERASE ALL settings on the device including:
- WiFi configuration
- Device name
- Authentication settings
- Schedules
- Scripts
- Webhooks

The device will return to AP mode and need to be reconfigured.

This command requires both --yes and --confirm flags for safety.`,
		Example: `  # Factory reset with double confirmation
  shelly device factory-reset living-room --yes --confirm

  # Using aliases
  shelly device fr living-room --yes --confirm
  shelly device reset living-room --yes --confirm
  shelly device wipe living-room --yes --confirm

  # This will fail (safety measure)
  shelly device factory-reset living-room --yes`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddConfirmFlags(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Require both flags for safety
	if !opts.Yes || !opts.Confirm {
		ios.Error("Factory reset requires both --yes and --confirm flags for safety")
		ios.Info("")
		ios.Info("WARNING: Factory reset will ERASE ALL settings on the device!")
		ios.Info("The device will return to AP mode and need to be reconfigured.")
		ios.Info("")
		ios.Info("To proceed, use:")
		ios.Info("  shelly device factory-reset %s --yes --confirm", opts.Device)
		return fmt.Errorf("missing required confirmation flags")
	}

	// Final interactive confirmation (default to false for safety)
	confirmed, err := opts.Factory.ConfirmAction(fmt.Sprintf("FINAL WARNING: Factory reset device %q? This cannot be undone!", opts.Device), false)
	if err != nil {
		return fmt.Errorf("confirmation failed: %w", err)
	}
	if !confirmed {
		ios.Info("Factory reset cancelled")
		return nil
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	svc := opts.Factory.ShellyService()

	err = cmdutil.RunWithSpinner(ctx, ios, "Factory resetting device...", func(ctx context.Context) error {
		return svc.DeviceFactoryResetAuto(ctx, opts.Device)
	})
	if err != nil {
		return fmt.Errorf("failed to factory reset device: %w", err)
	}

	ios.Success("Device %q has been factory reset", opts.Device)
	ios.Info("The device is now in AP mode and needs to be reconfigured")

	return nil
}

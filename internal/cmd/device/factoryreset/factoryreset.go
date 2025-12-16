// Package factoryreset provides the device factory-reset subcommand.
package factoryreset

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the device factory-reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		yes     bool
		confirm bool
	)

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
			return run(cmd.Context(), f, args[0], yes, confirm)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Confirm you want to proceed")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Double-confirm factory reset")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, yes, confirm bool) error {
	ios := f.IOStreams()

	// Require both flags for safety
	if !yes || !confirm {
		ios.Error("Factory reset requires both --yes and --confirm flags for safety")
		ios.Info("")
		ios.Info("WARNING: Factory reset will ERASE ALL settings on the device!")
		ios.Info("The device will return to AP mode and need to be reconfigured.")
		ios.Info("")
		ios.Info("To proceed, use:")
		ios.Info("  shelly device factory-reset %s --yes --confirm", device)
		return fmt.Errorf("missing required confirmation flags")
	}

	// Final interactive confirmation (default to false for safety)
	confirmed, err := f.ConfirmAction(fmt.Sprintf("FINAL WARNING: Factory reset device %q? This cannot be undone!", device), false)
	if err != nil {
		return fmt.Errorf("confirmation failed: %w", err)
	}
	if !confirmed {
		ios.Info("Factory reset cancelled")
		return nil
	}

	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()

	ios.StartProgress("Factory resetting device...")

	err = svc.DeviceFactoryReset(ctx, device)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to factory reset device: %w", err)
	}

	ios.Success("Device %q has been factory reset", device)
	ios.Info("The device is now in AP mode and needs to be reconfigured")

	return nil
}

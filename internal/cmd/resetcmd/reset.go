// Package resetcmd provides the quick reset (factory reset) command.
package resetcmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	Device  string
	Yes     bool
	Confirm bool
	Factory *cmdutil.Factory
}

// NewCommand creates the reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "reset <device>",
		Aliases: []string{"factory-reset", "wipe"},
		Short:   "Factory reset a device",
		Long: `Factory reset a Shelly device to its default settings.

WARNING: This will ERASE ALL settings on the device including:
- WiFi configuration
- Device name
- Authentication settings
- Schedules, Scripts, Webhooks

The device will return to AP mode and need to be reconfigured.

This command requires both --yes and --confirm flags for safety.`,
		Example: `  # Factory reset with double confirmation
  shelly reset living-room --yes --confirm

  # This will fail (safety measure)
  shelly reset living-room --yes`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Confirm you want to proceed")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Double-confirm factory reset")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Require both flags for safety
	if !opts.Yes || !opts.Confirm {
		ios.Error("Factory reset requires both --yes and --confirm flags for safety")
		ios.Println()
		ios.Warning("Factory reset will ERASE ALL settings on the device!")
		ios.Info("The device will return to AP mode and need to be reconfigured.")
		ios.Println()
		ios.Info("To proceed, use:")
		ios.Info("  shelly reset %s --yes --confirm", opts.Device)
		return fmt.Errorf("missing required confirmation flags")
	}

	// Final interactive confirmation (--yes flags already checked above)
	confirmed, err := ios.Confirm(fmt.Sprintf("FINAL WARNING: Factory reset device %q? This cannot be undone!", opts.Device), false)
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
		return svc.DeviceFactoryReset(ctx, opts.Device)
	})
	if err != nil {
		return err
	}

	ios.Success("Device %q has been factory reset", opts.Device)
	ios.Info("The device is now in AP mode and needs to be reconfigured")

	return nil
}

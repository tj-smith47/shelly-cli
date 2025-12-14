// Package reset provides the matter reset command.
package reset

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
	Factory *cmdutil.Factory
	Device  string
	Yes     bool
}

// NewCommand creates the matter reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "reset <device>",
		Aliases: []string{"factory-reset", "unpair"},
		Short:   "Reset Matter configuration",
		Long: `Reset all Matter settings on a Shelly device.

This command:
- Unpairs the device from all Matter fabrics
- Erases all Matter credentials and data
- Returns Matter to factory default state

Unlike 'shelly device factory-reset', this only affects Matter settings.
WiFi, device name, and other configurations are preserved.

After reset, the device must be re-commissioned to any Matter fabrics.`,
		Example: `  # Reset Matter (with confirmation)
  shelly matter reset living-room

  # Reset without confirmation
  shelly matter reset living-room --yes`,
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
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	if !opts.Yes {
		ios.Warning("This will unpair the device from all Matter fabrics.")
		ios.Printf("Continue? [y/N]: ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			ios.Debug("failed to read response: %v", err)
			return fmt.Errorf("operation cancelled")
		}

		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			ios.Info("Operation cancelled.")
			return nil
		}
	}

	if err := svc.MatterReset(ctx, opts.Device); err != nil {
		return err
	}

	ios.Success("Matter configuration reset.")
	ios.Info("The device has been unpaired from all fabrics.")
	ios.Info("Enable and re-commission with:")
	ios.Info("  shelly matter enable %s", opts.Device)
	ios.Info("  shelly matter code %s", opts.Device)

	return nil
}

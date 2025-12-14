package ota

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// UpdateOptions holds update command options.
type UpdateOptions struct {
	Factory *cmdutil.Factory
	Device  string
	URL     string
	Force   bool
}

func newUpdateCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &UpdateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "update <device>",
		Short: "Update device firmware",
		Long: `Update the firmware on a Gen1 Shelly device.

By default, updates to the latest available firmware from Shelly.
You can specify a custom firmware URL to update to a specific version.

The device will reboot after the update is applied.`,
		Example: `  # Update to latest firmware
  shelly gen1 ota update living-room

  # Update to specific firmware URL
  shelly gen1 ota update living-room --url http://archive.shelly-tools.de/...

  # Force update (even if no update detected)
  shelly gen1 ota update living-room --force`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runUpdate(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.URL, "url", "", "Custom firmware URL")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force update even if no update available")

	return cmd
}

func runUpdate(ctx context.Context, opts *UpdateOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	// If no custom URL, check for available update first
	if opts.URL == "" && !opts.Force {
		ios.StartProgress("Checking for updates...")
		info, err := gen1Client.CheckForUpdate(ctx)
		ios.StopProgress()

		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		if !info.HasUpdate {
			ios.Info("No update available. Current firmware: %s", gen1Client.Info().Firmware)
			ios.Info("Use --force to update anyway, or --url to specify a custom firmware.")
			return nil
		}

		ios.Info("Update available: %s -> %s", gen1Client.Info().Firmware, info.NewVersion)
	}

	ios.StartProgress("Starting firmware update...")
	err = gen1Client.Update(ctx, opts.URL)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to start update: %w", err)
	}

	ios.Success("Firmware update started on %s", opts.Device)
	ios.Info("The device will reboot during the update process.")
	ios.Info("Check progress with: shelly gen1 ota check %s", opts.Device)

	return nil
}

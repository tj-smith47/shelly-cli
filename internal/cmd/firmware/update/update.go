// Package update provides the firmware update subcommand.
package update

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Factory     *cmdutil.Factory
	Device      string
	Beta        bool
	URL         string
	All         bool
	List        bool
	Parallelism int
	Staged      int
}

// NewCommand creates the firmware update command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:     f,
		Parallelism: 3,
		Staged:      100,
	}

	cmd := &cobra.Command{
		Use:     "update [device]",
		Aliases: []string{"up"},
		Short:   "Update device firmware",
		Long: `Update device firmware to the latest version.

By default, updates to the latest stable version. Use --beta for beta firmware
or --url for a custom firmware file.

Use --list to show available updates before prompting for confirmation.
This is useful for reviewing what version will be installed.

Supports both native Shelly devices and plugin-managed devices (Tasmota, etc.).
Plugin devices are automatically detected and updated using the appropriate plugin.

Use --all to update all registered devices. The --staged flag allows percentage-based
rollouts (e.g., --staged 25 updates 25% of devices).`,
		Example: `  # Update to latest stable
  shelly firmware update living-room

  # Show update info before prompting
  shelly firmware update living-room --list

  # Update to beta
  shelly firmware update living-room --beta

  # Update from custom URL
  shelly firmware update living-room --url http://example.com/firmware.zip

  # Update plugin-managed device (Tasmota)
  shelly firmware update tasmota-plug --url http://ota.tasmota.com/tasmota.bin.gz

  # Update all devices
  shelly firmware update --all

  # Staged rollout (25% of devices)
  shelly firmware update --all --staged 25`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Device = args[0]
			}
			if !opts.All && opts.Device == "" {
				return fmt.Errorf("device name required (or use --all)")
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Beta, "beta", false, "Update to beta firmware")
	cmd.Flags().StringVar(&opts.URL, "url", "", "Custom firmware URL")
	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)
	cmd.Flags().BoolVar(&opts.All, "all", false, "Update all registered devices")
	cmd.Flags().BoolVarP(&opts.List, "list", "l", false, "Show available updates before prompting")
	cmd.Flags().IntVar(&opts.Parallelism, "parallel", 3, "Number of devices to update in parallel")
	cmd.Flags().IntVar(&opts.Staged, "staged", 100, "Percentage of devices to update (for staged rollouts)")

	return cmd
}

//nolint:gocyclo,nestif // Complexity from handling both --all batch mode and single device mode in one function
func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ios := f.IOStreams()
	svc := f.ShellyService()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Handle --all mode
	if opts.All {
		if len(cfg.Devices) == 0 {
			ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
			return nil
		}

		// Get device names
		deviceNames := make([]string, 0, len(cfg.Devices))
		for name := range cfg.Devices {
			deviceNames = append(deviceNames, name)
		}

		// Check all devices for updates
		toUpdate := svc.CheckDevicesForUpdates(ctx, ios, deviceNames, opts.Staged)

		if len(toUpdate) == 0 {
			ios.Info("All devices are up to date")
			return nil
		}

		// Display devices to update
		term.DisplayDevicesToUpdate(ios, toUpdate)

		// Confirm
		confirmed, confirmErr := f.ConfirmAction(fmt.Sprintf("Update %d device(s)?", len(toUpdate)), opts.Yes)
		if confirmErr != nil {
			return confirmErr
		}
		if !confirmed {
			ios.Warning("Update cancelled")
			return nil
		}

		// Perform updates
		results := svc.UpdateDevices(ctx, ios, toUpdate, shelly.UpdateOpts{
			Beta:        opts.Beta,
			CustomURL:   opts.URL,
			Parallelism: opts.Parallelism,
		})

		// Convert to term.UpdateResult for display
		termResults := make([]term.UpdateResult, len(results))
		for i, r := range results {
			termResults[i] = term.UpdateResult{
				Name:    r.Name,
				Success: r.Success,
				Err:     r.Err,
			}
		}

		term.DisplayUpdateResults(ios, termResults)
		return nil
	}

	// Single device mode
	device, ok := cfg.Devices[opts.Device]
	if !ok {
		// Ad-hoc device (IP address) - treat as Shelly
		device = model.Device{
			Name:     opts.Device,
			Address:  opts.Device,
			Platform: model.PlatformShelly,
		}
	} else {
		device.Name = opts.Device
	}

	// Check for updates first (platform-aware)
	var info *shelly.FirmwareInfo
	err = cmdutil.RunWithSpinner(ctx, ios, "Checking for updates...", func(ctx context.Context) error {
		var checkErr error
		info, checkErr = svc.CheckDeviceFirmware(ctx, device)
		return checkErr
	})
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	// If --list flag, show detailed update info
	if opts.List {
		term.DisplayFirmwareUpdateInfo(ios, info, device.DisplayName(), device.GetPlatform())
	}

	if !info.HasUpdate && opts.URL == "" && !opts.Beta {
		ios.Info("Device %s is already up to date (version %s)", opts.Device, info.Current)
		return nil
	}

	// Show what will be updated
	term.DisplayUpdateTarget(ios, term.UpdateTarget{
		DeviceID:    info.DeviceID,
		DeviceModel: info.DeviceModel,
		Current:     info.Current,
		Available:   info.Available,
		Beta:        info.Beta,
		CustomURL:   opts.URL,
		UseBeta:     opts.Beta,
	})

	// Confirm unless --yes
	confirmed, confirmErr := f.ConfirmAction("Proceed with firmware update?", opts.Yes)
	if confirmErr != nil {
		return confirmErr
	}
	if !confirmed {
		ios.Warning("Update cancelled")
		return nil
	}

	// Extended timeout for firmware updates
	updateCtx, cancel := f.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	err = cmdutil.RunWithSpinner(updateCtx, ios, "Updating firmware...", func(ctx context.Context) error {
		if updateErr := svc.UpdateDeviceFirmware(ctx, device, opts.Beta, opts.URL); updateErr != nil {
			return fmt.Errorf("failed to start update: %w", updateErr)
		}
		ios.Success("Firmware update started on %s", opts.Device)
		ios.Info("The device will reboot automatically. Use 'shelly firmware status %s' to check progress.", opts.Device)
		return nil
	})
	if err != nil {
		return err
	}

	// Invalidate all cached data for this device since firmware updates affect everything
	cmdutil.InvalidateDeviceCache(f, opts.Device)
	return nil
}

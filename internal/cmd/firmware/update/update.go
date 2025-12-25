// Package update provides the firmware update subcommand.
package update

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

var (
	betaFlag    bool
	urlFlag     string
	yesFlag     bool
	allFlag     bool
	listFlag    bool
	parallelism int
	stagedFlag  int
)

// NewCommand creates the firmware update command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			if allFlag {
				return runAll(cmd.Context(), f)
			}
			if len(args) == 0 {
				return fmt.Errorf("device name required (or use --all)")
			}
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().BoolVar(&betaFlag, "beta", false, "Update to beta firmware")
	cmd.Flags().StringVar(&urlFlag, "url", "", "Custom firmware URL")
	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&allFlag, "all", false, "Update all registered devices")
	cmd.Flags().BoolVarP(&listFlag, "list", "l", false, "Show available updates before prompting")
	cmd.Flags().IntVar(&parallelism, "parallel", 3, "Number of devices to update in parallel")
	cmd.Flags().IntVar(&stagedFlag, "staged", 100, "Percentage of devices to update (for staged rollouts)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, deviceName string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Resolve device to check if it's plugin-managed
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	device, ok := cfg.Devices[deviceName]
	if !ok {
		// Ad-hoc device (IP address) - treat as Shelly
		device = model.Device{
			Name:     deviceName,
			Address:  deviceName,
			Platform: model.PlatformShelly,
		}
	} else {
		device.Name = deviceName
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
	if listFlag {
		term.DisplayFirmwareUpdateInfo(ios, info, device.DisplayName(), device.GetPlatform())
	}

	if !info.HasUpdate && urlFlag == "" && !betaFlag {
		ios.Info("Device %s is already up to date (version %s)", deviceName, info.Current)
		return nil
	}

	// Show what will be updated
	term.DisplayUpdateTarget(ios, term.UpdateTarget{
		DeviceID:    info.DeviceID,
		DeviceModel: info.DeviceModel,
		Current:     info.Current,
		Available:   info.Available,
		Beta:        info.Beta,
		CustomURL:   urlFlag,
		UseBeta:     betaFlag,
	})

	// Confirm unless --yes
	confirmed, confirmErr := f.ConfirmAction("Proceed with firmware update?", yesFlag)
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

	return cmdutil.RunWithSpinner(updateCtx, ios, "Updating firmware...", func(ctx context.Context) error {
		if err := svc.UpdateDeviceFirmware(ctx, device, betaFlag, urlFlag); err != nil {
			return fmt.Errorf("failed to start update: %w", err)
		}
		ios.Success("Firmware update started on %s", deviceName)
		ios.Info("The device will reboot automatically. Use 'shelly firmware status %s' to check progress.", deviceName)
		return nil
	})
}

func runAll(ctx context.Context, f *cmdutil.Factory) error {
	ios := f.IOStreams()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Devices) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return nil
	}

	// Get device names
	deviceNames := make([]string, 0, len(cfg.Devices))
	for name := range cfg.Devices {
		deviceNames = append(deviceNames, name)
	}

	svc := f.ShellyService()

	// Check all devices for updates
	toUpdate := svc.CheckDevicesForUpdates(ctx, ios, deviceNames, stagedFlag)

	if len(toUpdate) == 0 {
		ios.Info("All devices are up to date")
		return nil
	}

	// Display devices to update
	term.DisplayDevicesToUpdate(ios, toUpdate)

	// Confirm
	confirmed, confirmErr := f.ConfirmAction(fmt.Sprintf("Update %d device(s)?", len(toUpdate)), yesFlag)
	if confirmErr != nil {
		return confirmErr
	}
	if !confirmed {
		ios.Warning("Update cancelled")
		return nil
	}

	// Perform updates
	results := svc.UpdateDevices(ctx, ios, toUpdate, shelly.UpdateOpts{
		Beta:        betaFlag,
		CustomURL:   urlFlag,
		Parallelism: parallelism,
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

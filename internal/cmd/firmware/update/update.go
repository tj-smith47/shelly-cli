// Package update provides the firmware update subcommand.
package update

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

var (
	betaFlag    bool
	urlFlag     string
	yesFlag     bool
	allFlag     bool
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

Use --all to update all registered devices. The --staged flag allows percentage-based
rollouts (e.g., --staged 25 updates 25% of devices).`,
		Example: `  # Update to latest stable
  shelly firmware update living-room

  # Update to beta
  shelly firmware update living-room --beta

  # Update from custom URL
  shelly firmware update living-room --url http://example.com/firmware.zip

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
	cmd.Flags().IntVar(&parallelism, "parallel", 3, "Number of devices to update in parallel")
	cmd.Flags().IntVar(&stagedFlag, "staged", 100, "Percentage of devices to update (for staged rollouts)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Check for updates first
	ios.StartProgress("Checking for updates...")
	info, err := svc.CheckFirmware(ctx, device)
	ios.StopProgress()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !info.HasUpdate && urlFlag == "" && !betaFlag {
		ios.Info("Device %s is already up to date (version %s)", device, info.Current)
		return nil
	}

	// Show what will be updated
	displayUpdateTarget(ios, info)

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
		if updateErr := doUpdate(ctx, svc, device); updateErr != nil {
			return fmt.Errorf("failed to start update: %w", updateErr)
		}
		ios.Success("Firmware update started on %s", device)
		ios.Info("The device will reboot automatically. Use 'shelly firmware status %s' to check progress.", device)
		return nil
	})
}

func displayUpdateTarget(ios *iostreams.IOStreams, info *shelly.FirmwareInfo) {
	ios.Println("")
	ios.Printf("  Device:  %s (%s)\n", info.DeviceID, info.DeviceModel)
	ios.Printf("  Current: %s\n", info.Current)
	switch {
	case urlFlag != "":
		ios.Printf("  Target:  %s\n", theme.StatusWarn().Render("custom URL"))
	case betaFlag:
		ios.Printf("  Target:  %s %s\n", info.Beta, theme.StatusWarn().Render("(beta)"))
	default:
		ios.Printf("  Target:  %s\n", info.Available)
	}
	ios.Println("")
}

func doUpdate(ctx context.Context, svc *shelly.Service, device string) error {
	switch {
	case urlFlag != "":
		return svc.UpdateFirmwareFromURL(ctx, device, urlFlag)
	case betaFlag:
		return svc.UpdateFirmwareBeta(ctx, device)
	default:
		return svc.UpdateFirmwareStable(ctx, device)
	}
}

type deviceStatus struct {
	name      string
	info      *shelly.FirmwareInfo
	hasUpdate bool
}

type updateResult struct {
	name    string
	success bool
	err     error
}

func runAll(ctx context.Context, f *cmdutil.Factory) error {
	ios := f.IOStreams()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	devices := cfg.Devices
	if len(devices) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return nil
	}

	svc := f.ShellyService()

	// Check all devices for updates
	toUpdate := checkDevicesForUpdates(ctx, ios, svc, devices, stagedFlag)

	if len(toUpdate) == 0 {
		ios.Info("All devices are up to date")
		return nil
	}

	// Display and confirm
	if err := displayAndConfirmUpdates(ios, f, toUpdate); err != nil {
		return err
	}

	// Perform updates
	return performUpdates(ctx, ios, svc, toUpdate)
}

func checkDevicesForUpdates(
	ctx context.Context,
	ios *iostreams.IOStreams,
	svc *shelly.Service,
	devices map[string]model.Device,
	staged int,
) []deviceStatus {
	ios.StartProgress("Checking devices for updates...")

	var (
		statuses []deviceStatus
		mu       sync.Mutex
	)

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	for name := range devices {
		deviceName := name
		g.Go(func() error {
			info, checkErr := svc.CheckFirmware(gctx, deviceName)
			hasUpdate := checkErr == nil && info != nil && info.HasUpdate
			mu.Lock()
			statuses = append(statuses, deviceStatus{
				name:      deviceName,
				info:      info,
				hasUpdate: hasUpdate,
			})
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("errgroup wait", err)
	}
	ios.StopProgress()

	// Filter devices with updates and apply staged percentage
	var toUpdate []deviceStatus
	for _, s := range statuses {
		if s.hasUpdate {
			toUpdate = append(toUpdate, s)
		}
	}

	// Apply staged percentage
	targetCount := len(toUpdate) * staged / 100
	if targetCount == 0 && staged > 0 && len(toUpdate) > 0 {
		targetCount = 1
	}
	if targetCount < len(toUpdate) {
		toUpdate = toUpdate[:targetCount]
	}

	return toUpdate
}

func displayAndConfirmUpdates(ios *iostreams.IOStreams, f *cmdutil.Factory, toUpdate []deviceStatus) error {
	ios.Println("")
	ios.Printf("%s\n", theme.Bold().Render("Devices to update:"))
	table := output.NewTable("Device", "Current", "Available")
	for _, s := range toUpdate {
		table.AddRow(s.name, s.info.Current, s.info.Available)
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	ios.Println("")

	confirmed, confirmErr := f.ConfirmAction(fmt.Sprintf("Update %d device(s)?", len(toUpdate)), yesFlag)
	if confirmErr != nil {
		return confirmErr
	}
	if !confirmed {
		ios.Warning("Update cancelled")
		return fmt.Errorf("cancelled")
	}
	return nil
}

func performUpdates(
	ctx context.Context,
	ios *iostreams.IOStreams,
	svc *shelly.Service,
	toUpdate []deviceStatus,
) error {
	ios.Println("")
	ios.StartProgress("Updating devices...")

	var (
		results []updateResult
		resMu   sync.Mutex
	)

	ug, ugctx := errgroup.WithContext(ctx)
	ug.SetLimit(parallelism)

	for _, s := range toUpdate {
		status := s
		ug.Go(func() error {
			var updateErr error
			if betaFlag {
				updateErr = svc.UpdateFirmwareBeta(ugctx, status.name)
			} else {
				updateErr = svc.UpdateFirmwareStable(ugctx, status.name)
			}
			resMu.Lock()
			results = append(results, updateResult{
				name:    status.name,
				success: updateErr == nil,
				err:     updateErr,
			})
			resMu.Unlock()
			return nil
		})
	}

	if err := ug.Wait(); err != nil {
		ios.DebugErr("errgroup wait", err)
	}
	ios.StopProgress()

	// Show results
	displayUpdateResults(ios, results)
	return nil
}

func displayUpdateResults(ios *iostreams.IOStreams, results []updateResult) {
	ios.Println("")
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.success {
			ios.Success("Updated: %s", r.name)
			successCount++
		} else {
			ios.Error("Failed: %s - %v", r.name, r.err)
			failCount++
		}
	}

	ios.Println("")
	if failCount > 0 {
		ios.Warning("Updated %d device(s), %d failed", successCount, failCount)
	} else {
		ios.Success("Updated %d device(s)", successCount)
	}
}

// Package all provides the monitor all subcommand for monitoring all registered devices.
package all

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

var intervalFlag time.Duration

// NewCommand creates the monitor all command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "all",
		Aliases: []string{"overview", "summary"},
		Short:   "Monitor all registered devices",
		Long: `Monitor all devices in the registry.

Shows a summary of power consumption and status for all registered devices.
Press Ctrl+C to stop monitoring.`,
		Example: `  # Monitor all devices
  shelly monitor all

  # Monitor with custom interval
  shelly monitor all --interval 5s`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(f, cmd.Context())
		},
	}

	cmd.Flags().DurationVarP(&intervalFlag, "interval", "i", 5*time.Second, "Refresh interval")

	return cmd
}

// deviceSnapshot holds the latest status for a device.
type deviceSnapshot struct {
	Device    string
	Address   string
	Info      *shelly.DeviceInfo
	Snapshot  *shelly.MonitoringSnapshot
	Error     error
	UpdatedAt time.Time
}

func run(f *cmdutil.Factory, ctx context.Context) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Load registered devices
	devices := config.ListDevices()
	if len(devices) == 0 {
		ios.NoResults("No devices registered. Use 'shelly device add' to add devices.")
		return nil
	}

	ios.Title("Monitoring %d devices", len(devices))
	ios.Printf("Press Ctrl+C to stop\n\n")

	// Create snapshot storage
	snapshots := make(map[string]*deviceSnapshot)
	var mu sync.Mutex

	// Initial fetch
	fetchAllSnapshots(ctx, svc, devices, snapshots, &mu)
	displayAllSnapshots(ios, snapshots, &mu)

	// Monitoring loop
	ticker := time.NewTicker(intervalFlag)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			fetchAllSnapshots(ctx, svc, devices, snapshots, &mu)
			displayAllSnapshots(ios, snapshots, &mu)
		}
	}
}

func fetchAllSnapshots(ctx context.Context, svc *shelly.Service, devices map[string]model.Device, snapshots map[string]*deviceSnapshot, mu *sync.Mutex) {
	var wg sync.WaitGroup
	for name, dev := range devices {
		wg.Go(func() {
			snapshot := &deviceSnapshot{
				Device:    name,
				Address:   dev.Address,
				UpdatedAt: time.Now(),
			}

			// Get device info
			info, err := svc.DeviceInfo(ctx, dev.Address)
			if err != nil {
				snapshot.Error = err
			} else {
				snapshot.Info = info
			}

			// Get monitoring snapshot
			if snapshot.Error == nil {
				snap, err := svc.GetMonitoringSnapshot(ctx, dev.Address)
				if err != nil {
					snapshot.Error = err
				} else {
					snapshot.Snapshot = snap
				}
			}

			mu.Lock()
			snapshots[name] = snapshot
			mu.Unlock()
		})
	}
	wg.Wait()
}

func displayAllSnapshots(ios *iostreams.IOStreams, snapshots map[string]*deviceSnapshot, mu *sync.Mutex) {
	// Clear screen
	clearScreen(ios)

	mu.Lock()
	defer mu.Unlock()

	ios.Title("Device Status Summary")
	ios.Printf("  Updated: %s\n\n", time.Now().Format("15:04:05"))

	totalPower := 0.0
	totalEnergy := 0.0
	onlineCount := 0
	offlineCount := 0

	// Display each device
	for name, snap := range snapshots {
		if snap.Error != nil {
			ios.Printf("%s %s: %s\n",
				theme.StatusError().Render("●"),
				name,
				theme.Dim().Render(snap.Error.Error()))
			offlineCount++
			continue
		}

		onlineCount++

		// Calculate device power
		devicePower := 0.0
		deviceEnergy := 0.0
		if snap.Snapshot != nil {
			for i := range snap.Snapshot.EM {
				devicePower += snap.Snapshot.EM[i].TotalActivePower
			}
			for i := range snap.Snapshot.EM1 {
				devicePower += snap.Snapshot.EM1[i].ActPower
			}
			for i := range snap.Snapshot.PM {
				pm := &snap.Snapshot.PM[i]
				devicePower += pm.APower
				if pm.AEnergy != nil {
					deviceEnergy += pm.AEnergy.Total
				}
			}
		}

		totalPower += devicePower
		totalEnergy += deviceEnergy

		// Display device line
		statusIcon := theme.StatusOK().Render("●")
		deviceModel := "Unknown"
		if snap.Info != nil {
			deviceModel = snap.Info.Model
		}

		powerStr := helpers.FormatPowerColored(devicePower)
		energyStr := ""
		if deviceEnergy > 0 {
			energyStr = fmt.Sprintf("  %.2f Wh", deviceEnergy)
		}
		ios.Printf("%s %s (%s): %s%s\n",
			statusIcon, name, deviceModel, powerStr, energyStr)
	}

	// Display summary
	ios.Println()
	ios.Printf("═══════════════════════════════════════\n")
	ios.Printf("  Online:       %s / %d devices\n",
		theme.StatusOK().Render(fmt.Sprintf("%d", onlineCount)),
		onlineCount+offlineCount)
	ios.Printf("  Total Power:  %s\n", theme.StatusOK().Render(helpers.FormatPower(totalPower)))
	if totalEnergy > 0 {
		ios.Printf("  Total Energy: %.2f Wh\n", totalEnergy)
	}
}

func clearScreen(ios *iostreams.IOStreams) {
	ios.Printf("\033[H\033[2J")
}

// Package all provides the monitor all subcommand for monitoring all registered devices.
package all

import (
	"context"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
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

	// Build device map for FetchAllSnapshots
	deviceAddrs := make(map[string]string, len(devices))
	for name, dev := range devices {
		deviceAddrs[name] = dev.Address
	}

	// Create snapshot storage
	snapshots := make(map[string]*shelly.DeviceSnapshot)
	var mu sync.Mutex

	// Initial fetch
	svc.FetchAllSnapshots(ctx, deviceAddrs, snapshots, &mu)
	mu.Lock()
	term.DisplayAllSnapshots(ios, snapshots)
	mu.Unlock()

	// Monitoring loop
	ticker := time.NewTicker(intervalFlag)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			svc.FetchAllSnapshots(ctx, deviceAddrs, snapshots, &mu)
			mu.Lock()
			term.DisplayAllSnapshots(ios, snapshots)
			mu.Unlock()
		}
	}
}

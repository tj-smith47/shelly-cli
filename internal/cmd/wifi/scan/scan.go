// Package scan provides the wifi scan subcommand.
package scan

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the wifi scan command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scan <device>",
		Aliases: []string{"search", "find"},
		Short:   "Scan for available WiFi networks",
		Long: `Scan for available WiFi networks using a device.

The device will scan for nearby WiFi networks and report their SSID,
signal strength (RSSI), channel, and authentication type.

Note: Scanning may take several seconds to complete.`,
		Example: `  # Scan for networks
  shelly wifi scan living-room

  # Output as JSON
  shelly wifi scan living-room -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	// Use longer timeout for scanning
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*2)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunList(ctx, ios, svc, device,
		"Scanning for WiFi networks...",
		"No WiFi networks found",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.WiFiScanResult, error) {
			results, err := svc.ScanWiFi(ctx, device)
			if err != nil {
				return nil, err
			}
			// Sort by signal strength (strongest first)
			sort.Slice(results, func(i, j int) bool {
				return results[i].RSSI > results[j].RSSI
			})
			return results, nil
		},
		displayResults)
}

func displayResults(ios *iostreams.IOStreams, results []shelly.WiFiScanResult) {
	ios.Title("Available WiFi Networks")
	ios.Println()

	table := output.NewTable("SSID", "Signal", "Channel", "Security")
	for _, r := range results {
		ssid := r.SSID
		if ssid == "" {
			ssid = "<hidden>"
		}
		signal := formatSignal(r.RSSI)
		table.AddRow(ssid, signal, fmt.Sprintf("%d", r.Channel), r.Auth)
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print wifi scan table", err)
	}

	ios.Printf("\n%d network(s) found\n", len(results))
}

func formatSignal(rssi int) string {
	bars := "▁▃▅▇"
	var strength int
	switch {
	case rssi >= -50:
		strength = 4
	case rssi >= -60:
		strength = 3
	case rssi >= -70:
		strength = 2
	default:
		strength = 1
	}
	return fmt.Sprintf("%s %d dBm", bars[:strength], rssi)
}

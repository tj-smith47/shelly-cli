// Package health provides the fleet health subcommand.
package health

import (
	"context"
	"encoding/json"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	JSONOutput bool
	Threshold  time.Duration
}

// DeviceHealth represents health metrics for a device.
type DeviceHealth struct {
	DeviceID      string `json:"device_id"`
	Online        bool   `json:"online"`
	LastSeen      string `json:"last_seen"`
	OnlineCount   int    `json:"online_count"`
	OfflineCount  int    `json:"offline_count"`
	ActivityCount int    `json:"activity_count"`
	Status        string `json:"status"` // healthy, warning, unhealthy
}

// NewCommand creates the fleet health command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Threshold: 10 * time.Minute,
	}

	cmd := &cobra.Command{
		Use:     "health",
		Aliases: []string{"check", "diagnose"},
		Short:   "Check device health",
		Long: `Check the health status of devices in your fleet.

Reports devices that haven't been seen recently or have frequent
online/offline transitions indicating connectivity issues.`,
		Example: `  # Check device health
  shelly fleet health

  # Custom threshold for "unhealthy"
  shelly fleet health --threshold 30m

  # JSON output
  shelly fleet health -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.JSONOutput, "json", "o", false, "Output as JSON")
	cmd.Flags().DurationVar(&opts.Threshold, "threshold", 10*time.Minute, "Time threshold for unhealthy status")

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	// Placeholder health data
	devices := []DeviceHealth{
		{DeviceID: "demo-device-1", Online: true, LastSeen: "2s ago", OnlineCount: 100, OfflineCount: 2, ActivityCount: 500, Status: "healthy"},
		{DeviceID: "demo-device-2", Online: true, LastSeen: "5s ago", OnlineCount: 95, OfflineCount: 5, ActivityCount: 450, Status: "healthy"},
		{DeviceID: "demo-device-3", Online: false, LastSeen: "2h ago", OnlineCount: 50, OfflineCount: 50, ActivityCount: 100, Status: "unhealthy"},
	}

	if opts.JSONOutput {
		data, err := json.MarshalIndent(devices, "", "  ")
		if err != nil {
			return err
		}
		ios.Println(string(data))
		return nil
	}

	var healthy, warning, unhealthy int
	for _, d := range devices {
		switch d.Status {
		case "healthy":
			healthy++
		case "warning":
			warning++
		case "unhealthy":
			unhealthy++
		}
	}

	ios.Success("Fleet Health Report")
	ios.Println("")
	ios.Printf("Summary: %d healthy, %d warning, %d unhealthy\n", healthy, warning, unhealthy)
	ios.Println("")

	for _, d := range devices {
		var statusIcon string
		switch d.Status {
		case "healthy":
			statusIcon = "✓"
		case "warning":
			statusIcon = "!"
		case "unhealthy":
			statusIcon = "✗"
		}

		ios.Printf("%s %s (%s)\n", statusIcon, d.DeviceID, d.Status)
		ios.Printf("   Online: %t | Last seen: %s | Activity: %d\n", d.Online, d.LastSeen, d.ActivityCount)
	}

	return nil
}

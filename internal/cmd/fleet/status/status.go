// Package status provides the fleet status subcommand.
package status

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	JSONOutput bool
	Online     bool
	Offline    bool
}

// DeviceStatus represents a device in the fleet.
type DeviceStatus struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name,omitempty"`
	Host     string `json:"host"`
	Online   bool   `json:"online"`
	LastSeen string `json:"last_seen,omitempty"`
}

// NewCommand creates the fleet status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"st", "list", "ls"},
		Short:   "View fleet device status",
		Long: `View the status of all devices in your fleet.

Shows online/offline status and last seen time for each device
connected through Shelly Cloud.`,
		Example: `  # View all device status
  shelly fleet status

  # Show only online devices
  shelly fleet status --online

  # Show only offline devices
  shelly fleet status --offline

  # JSON output
  shelly fleet status -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.JSONOutput, "json", "o", false, "Output as JSON")
	cmd.Flags().BoolVar(&opts.Online, "online", false, "Show only online devices")
	cmd.Flags().BoolVar(&opts.Offline, "offline", false, "Show only offline devices")

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	// Placeholder status data
	devices := []DeviceStatus{
		{DeviceID: "demo-device-1", Name: "Living Room", Host: "shelly-13-eu.shelly.cloud", Online: true, LastSeen: "2s ago"},
		{DeviceID: "demo-device-2", Name: "Kitchen", Host: "shelly-13-eu.shelly.cloud", Online: true, LastSeen: "5s ago"},
		{DeviceID: "demo-device-3", Name: "Bedroom", Host: "shelly-13-eu.shelly.cloud", Online: false, LastSeen: "2h ago"},
	}

	// Filter devices
	filtered := make([]DeviceStatus, 0, len(devices))
	for _, d := range devices {
		if opts.Online && !d.Online {
			continue
		}
		if opts.Offline && d.Online {
			continue
		}
		filtered = append(filtered, d)
	}

	if opts.JSONOutput {
		data, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return err
		}
		ios.Println(string(data))
		return nil
	}

	if len(filtered) == 0 {
		ios.Warning("No devices found matching criteria")
		return nil
	}

	ios.Success("Fleet Status (%d devices)", len(filtered))
	ios.Println("")

	for _, d := range filtered {
		status := "●"
		if !d.Online {
			status = "○"
		}
		ios.Printf("%s %s (%s)\n", status, d.Name, d.DeviceID)
		ios.Printf("   Host: %s | Last seen: %s\n", d.Host, d.LastSeen)
	}

	return nil
}

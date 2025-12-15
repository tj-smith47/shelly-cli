// Package stats provides the fleet stats subcommand.
package stats

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the command options.
type Options struct {
	JSONOutput bool
}

// FleetStats represents fleet-wide statistics.
type FleetStats struct {
	TotalDevices     int `json:"total_devices"`
	OnlineDevices    int `json:"online_devices"`
	OfflineDevices   int `json:"offline_devices"`
	TotalConnections int `json:"total_connections"`
	TotalGroups      int `json:"total_groups"`
}

// NewCommand creates the fleet stats command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "stats",
		Aliases: []string{"statistics", "summary"},
		Short:   "View fleet statistics",
		Long:    `View aggregate statistics for your device fleet.`,
		Example: `  # View fleet statistics
  shelly fleet stats

  # JSON output
  shelly fleet stats -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.JSONOutput, "json", "o", false, "Output as JSON")

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	// Placeholder stats
	stats := FleetStats{
		TotalDevices:     3,
		OnlineDevices:    2,
		OfflineDevices:   1,
		TotalConnections: 1,
		TotalGroups:      0,
	}

	if opts.JSONOutput {
		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return err
		}
		ios.Println(string(data))
		return nil
	}

	ios.Success("Fleet Statistics")
	ios.Println("")
	ios.Printf("Total Devices:  %d\n", stats.TotalDevices)
	ios.Printf("  Online:       %d\n", stats.OnlineDevices)
	ios.Printf("  Offline:      %d\n", stats.OfflineDevices)
	ios.Printf("Connections:    %d\n", stats.TotalConnections)
	ios.Printf("Groups:         %d\n", stats.TotalGroups)

	return nil
}

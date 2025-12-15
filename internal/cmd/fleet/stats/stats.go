// Package stats provides the fleet stats subcommand.
package stats

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// Options holds the command options.
type Options struct{}

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
  shelly fleet stats --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, _ *Options) error {
	ios := f.IOStreams()

	// Placeholder stats
	stats := FleetStats{
		TotalDevices:     3,
		OnlineDevices:    2,
		OfflineDevices:   1,
		TotalConnections: 1,
		TotalGroups:      0,
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, stats)
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

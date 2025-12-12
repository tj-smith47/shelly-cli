// Package monitor provides commands for real-time device monitoring.
package monitor

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/monitor/all"
	"github.com/tj-smith47/shelly-cli/internal/cmd/monitor/events"
	"github.com/tj-smith47/shelly-cli/internal/cmd/monitor/power"
	"github.com/tj-smith47/shelly-cli/internal/cmd/monitor/status"
)

// NewCommand creates the monitor command group.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "monitor",
		Aliases: []string{"mon"},
		Short:   "Real-time device monitoring",
		Long: `Real-time monitoring of Shelly devices.

Monitor device status, power consumption, and events in real-time
with automatic refresh and color-coded status changes.`,
	}

	cmd.AddCommand(status.NewCommand())
	cmd.AddCommand(power.NewCommand())
	cmd.AddCommand(events.NewCommand())
	cmd.AddCommand(all.NewCommand())

	return cmd
}

// Package batch provides the batch command group for bulk device operations.
package batch

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/batch/command"
	"github.com/tj-smith47/shelly-cli/internal/cmd/batch/off"
	"github.com/tj-smith47/shelly-cli/internal/cmd/batch/on"
	"github.com/tj-smith47/shelly-cli/internal/cmd/batch/toggle"
)

// NewCommand creates the batch command group.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "batch",
		Aliases: []string{"b"},
		Short:   "Execute commands on multiple devices",
		Long: `Execute commands on multiple devices simultaneously.

Batch operations can target:
- A device group (created with 'shelly group create')
- Multiple devices specified by name or IP
- All registered devices

Failed operations are reported but don't stop the batch.`,
		Example: `  # Turn on all devices in a group
  shelly batch on --group living-room

  # Turn off specific devices
  shelly batch off light-1 light-2 switch-1

  # Toggle all devices
  shelly batch toggle --all

  # Send custom RPC command to a group
  shelly batch command --group office "Switch.Set" '{"id":0,"on":true}'`,
	}

	cmd.AddCommand(on.NewCommand())
	cmd.AddCommand(off.NewCommand())
	cmd.AddCommand(toggle.NewCommand())
	cmd.AddCommand(command.NewCommand())

	return cmd
}

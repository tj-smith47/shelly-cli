// Package schedule provides the schedule management command group.
package schedule

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/schedule/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/schedule/deleteall"
	"github.com/tj-smith47/shelly-cli/internal/cmd/schedule/deletecmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/schedule/disable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/schedule/enable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/schedule/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/schedule/update"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the schedule command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "schedule",
		Aliases: []string{"sched"},
		Short:   "Manage device schedules",
		Long: `Manage time-based schedules on Gen2+ Shelly devices.

Schedules allow you to execute RPC calls at specified times using
cron-like timespec expressions. Supports wildcards, ranges, and
special values like @sunrise and @sunset.

Note: Maximum 20 schedules per device.`,
		Example: `  # List schedules
  shelly schedule list living-room

  # Create a schedule to turn on at 8:00 AM every day
  shelly schedule create living-room --timespec "0 0 8 * *" \
    --calls '[{"method":"Switch.Set","params":{"id":0,"on":true}}]'

  # Create a schedule for sunset
  shelly schedule create living-room --timespec "@sunset" \
    --calls '[{"method":"Switch.Set","params":{"id":0,"on":false}}]'

  # Enable/disable a schedule
  shelly schedule enable living-room 1
  shelly schedule disable living-room 1

  # Delete a schedule
  shelly schedule delete living-room 1`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(update.NewCommand(f))
	cmd.AddCommand(deletecmd.NewCommand(f))
	cmd.AddCommand(deleteall.NewCommand(f))
	cmd.AddCommand(enable.NewCommand(f))
	cmd.AddCommand(disable.NewCommand(f))

	return cmd
}

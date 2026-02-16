// Package disable provides the schedule disable subcommand.
package disable

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the schedule disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewScheduleToggleCommand(f, factories.ScheduleToggleOpts{
		Enable:  false,
		Aliases: []string{"off", "deactivate"},
		Long:    `Disable a schedule on a Gen2+ Shelly device.`,
		Example: `  # Disable a schedule
  shelly schedule disable living-room 1`,
		ValidArgsFunc: completion.DeviceThenScheduleID(),
		ServiceFunc: func(ctx context.Context, f *cmdutil.Factory, device string, id int) error {
			return f.AutomationService().DisableSchedule(ctx, device, id)
		},
	})
}

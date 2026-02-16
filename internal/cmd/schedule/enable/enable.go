// Package enable provides the schedule enable subcommand.
package enable

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the schedule enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewScheduleToggleCommand(f, factories.ScheduleToggleOpts{
		Enable:  true,
		Aliases: []string{"on", "activate"},
		Long:    `Enable a schedule on a Gen2+ Shelly device.`,
		Example: `  # Enable a schedule
  shelly schedule enable living-room 1`,
		ValidArgsFunc: completion.DeviceThenScheduleID(),
		ServiceFunc: func(ctx context.Context, f *cmdutil.Factory, device string, id int) error {
			return f.AutomationService().EnableSchedule(ctx, device, id)
		},
	})
}

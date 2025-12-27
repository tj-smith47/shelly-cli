// Package del provides the schedule delete subcommand.
package del

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
)

// NewCommand creates the schedule delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewDeviceDeleteCommand(f, factories.DeviceDeleteOpts{
		Resource:      "schedule",
		Long:          "Delete a schedule from a Gen2+ Shelly device.",
		ShowWarning:   true,
		ValidArgsFunc: completion.DeviceThenScheduleID(),
		AutomationServiceFunc: func(ctx context.Context, svc *automation.Service, device string, id int) error {
			return svc.DeleteSchedule(ctx, device, id)
		},
	})
}

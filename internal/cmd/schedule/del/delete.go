// Package del provides the schedule delete subcommand.
package del

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the schedule delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return cmdutil.NewDeviceDeleteCommand(f, cmdutil.DeviceDeleteOpts{
		Resource:      "schedule",
		Long:          "Delete a schedule from a Gen2+ Shelly device.",
		ShowWarning:   true,
		ValidArgsFunc: completion.DeviceThenScheduleID(),
		ServiceFunc: func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.DeleteSchedule(ctx, device, id)
		},
	})
}

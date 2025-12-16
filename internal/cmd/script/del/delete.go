// Package del provides the script delete subcommand.
package del

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the script delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return cmdutil.NewDeviceDeleteCommand(f, cmdutil.DeviceDeleteOpts{
		Resource: "script",
		Long: `Delete a script from a Gen2+ Shelly device.

This permanently removes the script and its code from the device.`,
		ShowWarning:   true,
		ValidArgsFunc: completion.DeviceThenScriptID(),
		ServiceFunc: func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.DeleteScript(ctx, device, id)
		},
	})
}

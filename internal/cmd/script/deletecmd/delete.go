// Package deletecmd provides the script delete subcommand.
package deletecmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
)

// NewCommand creates the script delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewDeviceDeleteCommand(f, factories.DeviceDeleteOpts{
		Resource: "script",
		Long: `Delete a script from a Gen2+ Shelly device.

This permanently removes the script and its code from the device.`,
		ShowWarning:   true,
		ValidArgsFunc: completion.DeviceThenScriptID(),
		AutomationServiceFunc: func(ctx context.Context, svc *automation.Service, device string, id int) error {
			return svc.DeleteScript(ctx, device, id)
		},
	})
}

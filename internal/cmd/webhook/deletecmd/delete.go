// Package del provides the webhook delete subcommand.
package del

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the webhook delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewDeviceDeleteCommand(f, factories.DeviceDeleteOpts{
		Resource: "webhook",
		Aliases:  []string{"rm", "remove"},
		Long: `Delete a webhook by ID.

Use 'shelly webhook list' to see webhook IDs.`,
		ServiceFunc: func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.DeleteWebhook(ctx, device, id)
		},
	})
}

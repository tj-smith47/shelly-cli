// Package toggle provides the batch toggle subcommand.
package toggle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the batch toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return cmdutil.NewBatchComponentCommand(f, cmdutil.BatchComponentOpts{
		Component: "Switch",
		Action:    cmdutil.ActionToggle,
		ServiceFunc: func(ctx context.Context, svc *shelly.Service, device string, componentID int) error {
			_, err := svc.SwitchToggle(ctx, device, componentID)
			return err
		},
	})
}

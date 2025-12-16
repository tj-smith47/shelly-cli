// Package on provides the switch on subcommand.
package on

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the switch on command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return cmdutil.NewComponentCommand(f, cmdutil.ComponentOpts{
		Component: "Switch",
		Action:    cmdutil.ActionOn,
		SimpleFunc: func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.SwitchOn(ctx, device, id)
		},
	})
}

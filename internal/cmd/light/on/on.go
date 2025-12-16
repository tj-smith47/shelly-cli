// Package on provides the light on subcommand.
package on

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the light on command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return cmdutil.NewComponentCommand(f, cmdutil.ComponentOpts{
		Component: "Light",
		Action:    cmdutil.ActionOn,
		SimpleFunc: func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.LightOn(ctx, device, id)
		},
	})
}

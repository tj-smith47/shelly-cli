// Package toggle provides the rgb toggle subcommand.
package toggle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the rgb toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return cmdutil.NewComponentCommand(f, cmdutil.ComponentOpts{
		Component: "RGB",
		Action:    cmdutil.ActionToggle,
		ToggleFunc: func(ctx context.Context, svc *shelly.Service, device string, id int) (bool, error) {
			status, err := svc.RGBToggle(ctx, device, id)
			if err != nil {
				return false, err
			}
			return status.Output, nil
		},
	})
}

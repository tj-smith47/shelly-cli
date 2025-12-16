// Package toggle provides the switch toggle subcommand.
package toggle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the switch toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewComponentCommand(f, factories.ComponentOpts{
		Component: "Switch",
		Action:    factories.ActionToggle,
		ToggleFunc: func(ctx context.Context, svc *shelly.Service, device string, id int) (bool, error) {
			status, err := svc.SwitchToggle(ctx, device, id)
			if err != nil {
				return false, err
			}
			return status.Output, nil
		},
	})
}

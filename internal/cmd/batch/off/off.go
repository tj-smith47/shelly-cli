// Package off provides the batch off subcommand.
package off

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the batch off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewBatchComponentCommand(f, factories.BatchComponentOpts{
		Component: "Switch",
		Action:    factories.ActionOff,
		ServiceFunc: func(ctx context.Context, svc *shelly.Service, device string, componentID int) error {
			return svc.SwitchOff(ctx, device, componentID)
		},
	})
}

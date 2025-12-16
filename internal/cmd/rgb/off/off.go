// Package off provides the rgb off subcommand.
package off

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the rgb off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return cmdutil.NewComponentCommand(f, cmdutil.ComponentOpts{
		Component: "RGB",
		Action:    cmdutil.ActionOff,
		SimpleFunc: func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.RGBOff(ctx, device, id)
		},
	})
}

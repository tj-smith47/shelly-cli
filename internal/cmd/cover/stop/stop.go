// Package stop provides the cover stop subcommand.
package stop

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover stop command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return cmdutil.NewCoverCommand(f, cmdutil.CoverOpts{
		Action: cmdutil.CoverActionStop,
		ServiceFunc: func(ctx context.Context, svc *shelly.Service, device string, id int, _ *int) error {
			return svc.CoverStop(ctx, device, id)
		},
	})
}

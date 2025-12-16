// Package open provides the cover open subcommand.
package open

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover open command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return cmdutil.NewCoverCommand(f, cmdutil.CoverOpts{
		Action: cmdutil.CoverActionOpen,
		ServiceFunc: func(ctx context.Context, svc *shelly.Service, device string, id int, duration *int) error {
			return svc.CoverOpen(ctx, device, id, duration)
		},
	})
}

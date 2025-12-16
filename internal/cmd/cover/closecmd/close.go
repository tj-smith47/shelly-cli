// Package closecmd provides the cover close subcommand.
package closecmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover close command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewCoverCommand(f, factories.CoverOpts{
		Action: factories.CoverActionClose,
		ServiceFunc: func(ctx context.Context, svc *shelly.Service, device string, id int, duration *int) error {
			return svc.CoverClose(ctx, device, id, duration)
		},
	})
}

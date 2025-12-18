// Package status provides the input status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the input status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewStatusCommand(f, factories.StatusOpts[*model.InputStatus]{
		Component: "Input",
		Aliases:   []string{"st"},
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.InputStatus, error) {
			return svc.InputStatus(ctx, device, id)
		},
		Display: cmdutil.DisplayInputStatus,
	})
}

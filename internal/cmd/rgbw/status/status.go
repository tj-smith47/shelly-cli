// Package status provides the rgbw status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the rgbw status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewStatusCommand(f, factories.StatusOpts[*model.RGBWStatus]{
		Component: "RGBW",
		Aliases:   []string{"st", "s"},
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.RGBWStatus, error) {
			return svc.RGBWStatus(ctx, device, id)
		},
		Display: term.DisplayRGBWStatus,
	})
}

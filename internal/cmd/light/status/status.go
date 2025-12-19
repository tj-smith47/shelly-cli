// Package status provides the light status subcommand.
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

// NewCommand creates the light status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewStatusCommand(f, factories.StatusOpts[*model.LightStatus]{
		Component: "Light",
		Aliases:   []string{"st", "s"},
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.LightStatus, error) {
			return svc.LightStatus(ctx, device, id)
		},
		Display: term.DisplayLightStatus,
	})
}

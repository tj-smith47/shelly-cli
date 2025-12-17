// Package status provides the cover status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewStatusCommand(f, factories.StatusOpts[*model.CoverStatus]{
		Component: "Cover",
		Aliases:   []string{"st", "s"},
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.CoverStatus, error) {
			return svc.CoverStatus(ctx, device, id)
		},
		Display: displayStatus,
	})
}

func displayStatus(ios *iostreams.IOStreams, status *model.CoverStatus) {
	ios.Title("Cover %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:    %s\n", output.RenderCoverState(status.State))
	if status.CurrentPosition != nil {
		ios.Printf("  Position: %d%%\n", *status.CurrentPosition)
	}
	cmdutil.DisplayPowerMetricsWide(ios, status.Power, status.Voltage, status.Current)
}

// Package status provides the switch status subcommand.
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
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the switch status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewStatusCommand(f, factories.StatusOpts[*model.SwitchStatus]{
		Component: "Switch",
		Aliases:   []string{"st", "s"},
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.SwitchStatus, error) {
			return svc.SwitchStatus(ctx, device, id)
		},
		Display: displayStatus,
	})
}

func displayStatus(ios *iostreams.IOStreams, status *model.SwitchStatus) {
	ios.Title("Switch %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:   %s\n", output.RenderOnOff(status.Output, output.CaseUpper, theme.FalseError))
	cmdutil.DisplayPowerMetrics(ios, status.Power, status.Voltage, status.Current)
	if status.Energy != nil {
		ios.Printf("  Energy:  %.2f Wh\n", status.Energy.Total)
	}
}

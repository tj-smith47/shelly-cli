// Package status provides the light status subcommand.
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

// NewCommand creates the light status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewStatusCommand(f, factories.StatusOpts[*model.LightStatus]{
		Component: "Light",
		Aliases:   []string{"st", "s"},
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.LightStatus, error) {
			return svc.LightStatus(ctx, device, id)
		},
		Display: displayStatus,
	})
}

func displayStatus(ios *iostreams.IOStreams, status *model.LightStatus) {
	ios.Title("Light %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:      %s\n", output.RenderOnOff(status.Output, output.CaseUpper, theme.FalseError))
	if status.Brightness != nil {
		ios.Printf("  Brightness: %d%%\n", *status.Brightness)
	}
	if status.Power != nil {
		ios.Printf("  Power:      %.1f W\n", *status.Power)
	}
	if status.Voltage != nil {
		ios.Printf("  Voltage:    %.1f V\n", *status.Voltage)
	}
	if status.Current != nil {
		ios.Printf("  Current:    %.3f A\n", *status.Current)
	}
}

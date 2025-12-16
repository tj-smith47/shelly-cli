// Package status provides the rgb status subcommand.
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

// NewCommand creates the rgb status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewStatusCommand(f, factories.StatusOpts[*model.RGBStatus]{
		Component: "RGB",
		Aliases:   []string{"st", "s"},
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.RGBStatus, error) {
			return svc.RGBStatus(ctx, device, id)
		},
		Display: displayStatus,
	})
}

func displayStatus(ios *iostreams.IOStreams, status *model.RGBStatus) {
	ios.Title("RGB %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:      %s\n", output.RenderOnOffState(status.Output))
	if status.RGB != nil {
		ios.Printf("  Color:      R:%d G:%d B:%d\n",
			status.RGB.Red, status.RGB.Green, status.RGB.Blue)
	}
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

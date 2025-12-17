// Package status provides the input status subcommand.
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

// NewCommand creates the input status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewStatusCommand(f, factories.StatusOpts[*model.InputStatus]{
		Component: "Input",
		Aliases:   []string{"st"},
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.InputStatus, error) {
			return svc.InputStatus(ctx, device, id)
		},
		Display: displayStatus,
	})
}

func displayStatus(ios *iostreams.IOStreams, status *model.InputStatus) {
	ios.Title("Input %d Status", status.ID)
	ios.Println()

	ios.Printf("  State: %s\n", output.RenderActive(status.State, output.CaseLower, theme.FalseError))
	if status.Type != "" {
		ios.Printf("  Type:  %s\n", status.Type)
	}
}

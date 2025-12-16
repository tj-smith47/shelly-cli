// Package status provides the cover status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var coverID int

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "s"},
		Short:   "Show cover status",
		Long:    `Show the current status of a cover component on the specified device.`,
		Example: `  # Show cover status
  shelly cover status bedroom

  # Show status with JSON output
  shelly cv st bedroom -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], coverID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &coverID, "Cover")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, coverID int) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunStatus(ctx, ios, svc, device, coverID,
		"Fetching cover status...",
		func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.CoverStatus, error) {
			return svc.CoverStatus(ctx, device, id)
		},
		displayStatus)
}

func displayStatus(ios *iostreams.IOStreams, status *model.CoverStatus) {
	ios.Title("Cover %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:    %s\n", output.RenderCoverState(status.State))
	if status.CurrentPosition != nil {
		ios.Printf("  Position: %d%%\n", *status.CurrentPosition)
	}
	cmdutil.PrintPowerMetricsWide(ios, status.Power, status.Voltage, status.Current)
}

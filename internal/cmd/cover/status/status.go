// Package status provides the cover status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the cover status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var coverID int

	cmd := &cobra.Command{
		Use:   "status <device>",
		Short: "Show cover status",
		Long:  `Show the current status of a cover component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], coverID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &coverID, "Cover")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, coverID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
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

	stateStyle := theme.StatusWarn()
	switch status.State {
	case "open":
		stateStyle = theme.StatusOK()
	case "closed":
		stateStyle = theme.StatusError()
	}
	ios.Printf("  State:    %s\n", stateStyle.Render(status.State))

	if status.CurrentPosition != nil {
		ios.Printf("  Position: %d%%\n", *status.CurrentPosition)
	}
	if status.Power != nil {
		ios.Printf("  Power:    %.1f W\n", *status.Power)
	}
	if status.Voltage != nil {
		ios.Printf("  Voltage:  %.1f V\n", *status.Voltage)
	}
	if status.Current != nil {
		ios.Printf("  Current:  %.3f A\n", *status.Current)
	}
}

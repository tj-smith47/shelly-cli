// Package status provides the rgb status subcommand.
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

// NewCommand creates the rgb status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var rgbID int

	cmd := &cobra.Command{
		Use:   "status <device>",
		Short: "Show RGB status",
		Long:  `Show the current status of an RGB light component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], rgbID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &rgbID, "RGB")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, rgbID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunStatus(ctx, ios, svc, device, rgbID,
		"Fetching RGB status...",
		func(ctx context.Context, svc *shelly.Service, device string, id int) (*model.RGBStatus, error) {
			return svc.RGBStatus(ctx, device, id)
		},
		displayStatus)
}

func displayStatus(ios *iostreams.IOStreams, status *model.RGBStatus) {
	ios.Title("RGB %d Status", status.ID)
	ios.Println()

	state := theme.StatusError().Render("OFF")
	if status.Output {
		state = theme.StatusOK().Render("ON")
	}
	ios.Printf("  State:      %s\n", state)

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

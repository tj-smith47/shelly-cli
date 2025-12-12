// Package off provides the rgb off subcommand.
package off

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the rgb off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var rgbID int

	cmd := &cobra.Command{
		Use:   "off <device>",
		Short: "Turn RGB off",
		Long:  `Turn off an RGB light component on the specified device.`,
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

	return cmdutil.RunSimple(ctx, ios, svc, device, rgbID,
		"Turning RGB off...",
		fmt.Sprintf("RGB %d turned off", rgbID),
		func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.RGBOff(ctx, device, id)
		})
}

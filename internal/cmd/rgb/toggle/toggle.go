// Package toggle provides the rgb toggle subcommand.
package toggle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the rgb toggle command.
func NewCommand() *cobra.Command {
	var rgbID int

	cmd := &cobra.Command{
		Use:   "toggle <device>",
		Short: "Toggle RGB on/off",
		Long:  `Toggle an RGB light component on or off on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], rgbID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &rgbID, "RGB")

	return cmd
}

func run(ctx context.Context, device string, rgbID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := iostreams.System()
	svc := shelly.NewService()

	return cmdutil.RunWithSpinner(ctx, ios, "Toggling RGB...", func(ctx context.Context) error {
		status, err := svc.RGBToggle(ctx, device, rgbID)
		if err != nil {
			return err
		}

		state := "off"
		if status.Output {
			state = "on"
		}
		ios.Success("RGB %d toggled %s", rgbID, state)
		return nil
	})
}

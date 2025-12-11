// Package toggle provides the light toggle subcommand.
package toggle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the light toggle command.
func NewCommand() *cobra.Command {
	var lightID int

	cmd := &cobra.Command{
		Use:   "toggle <device>",
		Short: "Toggle light on/off",
		Long:  `Toggle a light component on or off on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], lightID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &lightID, "Light")

	return cmd
}

func run(ctx context.Context, device string, lightID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := iostreams.System()
	svc := shelly.NewService()

	return cmdutil.RunWithSpinner(ctx, ios, "Toggling light...", func(ctx context.Context) error {
		status, err := svc.LightToggle(ctx, device, lightID)
		if err != nil {
			return err
		}

		state := "off"
		if status.Output {
			state = "on"
		}
		ios.Success("Light %d toggled %s", lightID, state)
		return nil
	})
}

// Package toggle provides the rgb toggle subcommand.
package toggle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the rgb toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var rgbID int

	cmd := &cobra.Command{
		Use:     "toggle <device>",
		Aliases: []string{"flip", "t"},
		Short:   "Toggle RGB on/off",
		Long:    `Toggle an RGB light component on or off on the specified device.`,
		Example: `  # Toggle RGB light
  shelly rgb toggle living-room

  # Toggle specific RGB ID
  shelly rgb flip living-room --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], rgbID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &rgbID, "RGB")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, rgbID int) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

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

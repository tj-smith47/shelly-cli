// Package toggle provides the light toggle subcommand.
package toggle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the light toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var lightID int

	cmd := &cobra.Command{
		Use:     "toggle <device>",
		Aliases: []string{"flip", "t"},
		Short:   "Toggle light on/off",
		Long:    `Toggle a light component on or off on the specified device.`,
		Example: `  # Toggle light
  shelly light toggle bedroom

  # Toggle specific light ID
  shelly lt flip bedroom --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], lightID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &lightID, "Light")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, lightID int) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

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

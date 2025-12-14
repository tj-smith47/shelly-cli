// Package off provides the light off subcommand.
package off

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the light off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var lightID int

	cmd := &cobra.Command{
		Use:     "off <device>",
		Aliases: []string{"disable", "0"},
		Short:   "Turn light off",
		Long:    `Turn off a light component on the specified device.`,
		Example: `  # Turn off light
  shelly light off living-room

  # Turn off specific light ID
  shelly light off living-room --id 1`,
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
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunSimple(ctx, ios, svc, device, lightID,
		"Turning light off...",
		fmt.Sprintf("Light %d turned off", lightID),
		func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.LightOff(ctx, device, id)
		})
}

// Package off provides the switch off subcommand.
package off

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the switch off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var switchID int

	cmd := &cobra.Command{
		Use:     "off <device>",
		Aliases: []string{"disable", "0"},
		Short:   "Turn switch off",
		Long:    `Turn off a switch component on the specified device.`,
		Example: `  # Turn off switch on living-room device
  shelly switch off living-room

  # Turn off specific switch ID
  shelly switch off living-room --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], switchID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &switchID, "Switch")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, switchID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunSimple(ctx, ios, svc, device, switchID,
		"Turning switch off...",
		fmt.Sprintf("Switch %d turned off", switchID),
		func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.SwitchOff(ctx, device, id)
		})
}

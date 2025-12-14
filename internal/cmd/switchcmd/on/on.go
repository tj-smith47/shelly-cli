// Package on provides the switch on subcommand.
package on

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the switch on command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var switchID int

	cmd := &cobra.Command{
		Use:     "on <device>",
		Aliases: []string{"enable", "1"},
		Short:   "Turn switch on",
		Long:    `Turn on a switch component on the specified device.`,
		Example: `  # Turn on switch on kitchen device
  shelly switch on kitchen

  # Turn on specific switch ID
  shelly switch on kitchen --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
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
		"Turning switch on...",
		fmt.Sprintf("Switch %d turned on", switchID),
		func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.SwitchOn(ctx, device, id)
		})
}

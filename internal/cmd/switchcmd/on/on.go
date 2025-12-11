// Package on provides the switch on subcommand.
package on

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the switch on command.
func NewCommand() *cobra.Command {
	var switchID int

	cmd := &cobra.Command{
		Use:   "on <device>",
		Short: "Turn switch on",
		Long:  `Turn on a switch component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], switchID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &switchID, "Switch")

	return cmd
}

func run(ctx context.Context, device string, switchID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := iostreams.System()
	svc := shelly.NewService()

	return cmdutil.RunSimple(ctx, ios, svc, device, switchID,
		"Turning switch on...",
		fmt.Sprintf("Switch %d turned on", switchID),
		func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.SwitchOn(ctx, device, id)
		})
}

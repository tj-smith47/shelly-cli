// Package off provides the light off subcommand.
package off

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the light off command.
func NewCommand() *cobra.Command {
	var lightID int

	cmd := &cobra.Command{
		Use:   "off <device>",
		Short: "Turn light off",
		Long:  `Turn off a light component on the specified device.`,
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

	return cmdutil.RunSimple(ctx, ios, svc, device, lightID,
		"Turning light off...",
		fmt.Sprintf("Light %d turned off", lightID),
		func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.LightOff(ctx, device, id)
		})
}

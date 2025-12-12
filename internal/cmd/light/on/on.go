// Package on provides the light on subcommand.
package on

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the light on command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var lightID int

	cmd := &cobra.Command{
		Use:   "on <device>",
		Short: "Turn light on",
		Long:  `Turn on a light component on the specified device.`,
		Args:  cobra.ExactArgs(1),
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
		"Turning light on...",
		fmt.Sprintf("Light %d turned on", lightID),
		func(ctx context.Context, svc *shelly.Service, device string, id int) error {
			return svc.LightOn(ctx, device, id)
		})
}

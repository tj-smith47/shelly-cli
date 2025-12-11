// Package on provides the light on subcommand.
package on

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// NewCommand creates the light on command.
func NewCommand() *cobra.Command {
	var lightID int

	cmd := &cobra.Command{
		Use:   "on <device>",
		Short: "Turn light on",
		Long:  `Turn on a light component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args[0], lightID)
		},
	}

	cmd.Flags().IntVarP(&lightID, "id", "i", 0, "Light ID (default 0)")

	return cmd
}

func run(device string, lightID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Turning light on...")
	spin.Start()

	err := svc.LightOn(ctx, device, lightID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to turn light on: %w", err)
	}

	iostreams.Success("Light %d turned on", lightID)
	return nil
}

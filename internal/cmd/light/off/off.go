// Package off provides the light off subcommand.
package off

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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

	cmd.Flags().IntVarP(&lightID, "id", "i", 0, "Light ID (default 0)")

	return cmd
}

func run(ctx context.Context, device string, lightID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Turning light off...")
	spin.Start()

	err := svc.LightOff(ctx, device, lightID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to turn light off: %w", err)
	}

	iostreams.Success("Light %d turned off", lightID)
	return nil
}

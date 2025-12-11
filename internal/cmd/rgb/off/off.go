// Package off provides the rgb off subcommand.
package off

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// NewCommand creates the rgb off command.
func NewCommand() *cobra.Command {
	var rgbID int

	cmd := &cobra.Command{
		Use:   "off <device>",
		Short: "Turn RGB off",
		Long:  `Turn off an RGB light component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args[0], rgbID)
		},
	}

	cmd.Flags().IntVarP(&rgbID, "id", "i", 0, "RGB ID (default 0)")

	return cmd
}

func run(device string, rgbID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Turning RGB off...")
	spin.Start()

	err := svc.RGBOff(ctx, device, rgbID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to turn RGB off: %w", err)
	}

	iostreams.Success("RGB %d turned off", rgbID)
	return nil
}

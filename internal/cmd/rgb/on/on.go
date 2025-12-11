// Package on provides the rgb on subcommand.
package on

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the rgb on command.
func NewCommand() *cobra.Command {
	var rgbID int

	cmd := &cobra.Command{
		Use:   "on <device>",
		Short: "Turn RGB on",
		Long:  `Turn on an RGB light component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], rgbID)
		},
	}

	cmd.Flags().IntVarP(&rgbID, "id", "i", 0, "RGB ID (default 0)")

	return cmd
}

func run(ctx context.Context, device string, rgbID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Turning RGB on...")
	spin.Start()

	err := svc.RGBOn(ctx, device, rgbID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to turn RGB on: %w", err)
	}

	iostreams.Success("RGB %d turned on", rgbID)
	return nil
}

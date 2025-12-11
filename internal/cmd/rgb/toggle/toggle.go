// Package toggle provides the rgb toggle subcommand.
package toggle

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the rgb toggle command.
func NewCommand() *cobra.Command {
	var rgbID int

	cmd := &cobra.Command{
		Use:   "toggle <device>",
		Short: "Toggle RGB on/off",
		Long:  `Toggle an RGB light component on or off on the specified device.`,
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

	spin := iostreams.NewSpinner("Toggling RGB...")
	spin.Start()

	status, err := svc.RGBToggle(ctx, device, rgbID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to toggle RGB: %w", err)
	}

	state := "off"
	if status.Output {
		state = "on"
	}

	iostreams.Success("RGB %d toggled %s", rgbID, state)
	return nil
}

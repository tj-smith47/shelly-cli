// Package toggle provides the light toggle subcommand.
package toggle

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the light toggle command.
func NewCommand() *cobra.Command {
	var lightID int

	cmd := &cobra.Command{
		Use:   "toggle <device>",
		Short: "Toggle light on/off",
		Long:  `Toggle a light component on or off on the specified device.`,
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

	spin := iostreams.NewSpinner("Toggling light...")
	spin.Start()

	status, err := svc.LightToggle(ctx, device, lightID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to toggle light: %w", err)
	}

	state := "off"
	if status.Output {
		state = "on"
	}

	iostreams.Success("Light %d toggled %s", lightID, state)
	return nil
}

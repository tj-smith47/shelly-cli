// Package toggle provides the switch toggle subcommand.
package toggle

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the switch toggle command.
func NewCommand() *cobra.Command {
	var switchID int

	cmd := &cobra.Command{
		Use:   "toggle <device>",
		Short: "Toggle switch state",
		Long:  `Toggle the state of a switch component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], switchID)
		},
	}

	cmd.Flags().IntVarP(&switchID, "id", "i", 0, "Switch ID (default 0)")

	return cmd
}

func run(ctx context.Context, device string, switchID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Toggling switch...")
	spin.Start()

	status, err := svc.SwitchToggle(ctx, device, switchID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to toggle switch: %w", err)
	}

	state := "off"
	if status.Output {
		state = "on"
	}
	iostreams.Success("Switch %d toggled to %s", switchID, state)
	return nil
}

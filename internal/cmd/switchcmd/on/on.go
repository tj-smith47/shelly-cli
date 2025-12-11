// Package on provides the switch on subcommand.
package on

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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
			return run(args[0], switchID)
		},
	}

	cmd.Flags().IntVarP(&switchID, "id", "i", 0, "Switch ID (default 0)")

	return cmd
}

func run(device string, switchID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Turning switch on...")
	spin.Start()

	err := svc.SwitchOn(ctx, device, switchID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to turn switch on: %w", err)
	}

	iostreams.Success("Switch %d turned on", switchID)
	return nil
}

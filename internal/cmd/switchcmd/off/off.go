// Package off provides the switch off subcommand.
package off

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the switch off command.
func NewCommand() *cobra.Command {
	var switchID int

	cmd := &cobra.Command{
		Use:   "off <device>",
		Short: "Turn switch off",
		Long:  `Turn off a switch component on the specified device.`,
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

	spin := iostreams.NewSpinner("Turning switch off...")
	spin.Start()

	err := svc.SwitchOff(ctx, device, switchID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to turn switch off: %w", err)
	}

	iostreams.Success("Switch %d turned off", switchID)
	return nil
}

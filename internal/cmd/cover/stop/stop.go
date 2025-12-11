// Package stop provides the cover stop subcommand.
package stop

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover stop command.
func NewCommand() *cobra.Command {
	var coverID int

	cmd := &cobra.Command{
		Use:   "stop <device>",
		Short: "Stop cover",
		Long:  `Stop a cover/roller component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args[0], coverID)
		},
	}

	cmd.Flags().IntVarP(&coverID, "id", "i", 0, "Cover ID (default 0)")

	return cmd
}

func run(device string, coverID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Stopping cover...")
	spin.Start()

	err := svc.CoverStop(ctx, device, coverID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to stop cover: %w", err)
	}

	iostreams.Success("Cover %d stopped", coverID)
	return nil
}

// Package closecmd provides the cover close subcommand.
package closecmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover close command.
func NewCommand() *cobra.Command {
	var coverID int
	var duration int

	cmd := &cobra.Command{
		Use:   "close <device>",
		Short: "Close cover",
		Long:  `Close a cover/roller component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], coverID, duration)
		},
	}

	cmd.Flags().IntVarP(&coverID, "id", "i", 0, "Cover ID (default 0)")
	cmd.Flags().IntVarP(&duration, "duration", "d", 0, "Duration in seconds (0 = full close)")

	return cmd
}

func run(ctx context.Context, device string, coverID, duration int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Closing cover...")
	spin.Start()

	var dur *int
	if duration > 0 {
		dur = &duration
	}

	err := svc.CoverClose(ctx, device, coverID, dur)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to close cover: %w", err)
	}

	iostreams.Success("Cover %d closing", coverID)
	return nil
}

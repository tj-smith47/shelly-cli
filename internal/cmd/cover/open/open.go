// Package open provides the cover open subcommand.
package open

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover open command.
func NewCommand() *cobra.Command {
	var coverID int
	var duration int

	cmd := &cobra.Command{
		Use:   "open <device>",
		Short: "Open cover",
		Long:  `Open a cover/roller component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], coverID, duration)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &coverID, "Cover")
	cmd.Flags().IntVarP(&duration, "duration", "d", 0, "Duration in seconds (0 = full open)")

	return cmd
}

func run(ctx context.Context, device string, coverID, duration int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Opening cover...")
	spin.Start()

	var dur *int
	if duration > 0 {
		dur = &duration
	}

	err := svc.CoverOpen(ctx, device, coverID, dur)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to open cover: %w", err)
	}

	iostreams.Success("Cover %d opening", coverID)
	return nil
}

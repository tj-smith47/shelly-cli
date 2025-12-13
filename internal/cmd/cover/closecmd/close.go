// Package closecmd provides the cover close subcommand.
package closecmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover close command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var coverID int
	var duration int

	cmd := &cobra.Command{
		Use:   "close <device>",
		Short: "Close cover",
		Long:  `Close a cover/roller component on the specified device.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], coverID, duration)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &coverID, "Cover")
	cmd.Flags().IntVarP(&duration, "duration", "d", 0, "Duration in seconds (0 = full close)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, coverID, duration int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	ios.StartProgress("Closing cover...")

	var dur *int
	if duration > 0 {
		dur = &duration
	}

	err := svc.CoverClose(ctx, device, coverID, dur)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to close cover: %w", err)
	}

	ios.Success("Cover %d closing", coverID)
	return nil
}

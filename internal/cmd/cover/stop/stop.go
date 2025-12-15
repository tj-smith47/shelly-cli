// Package stop provides the cover stop subcommand.
package stop

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the cover stop command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var coverID int

	cmd := &cobra.Command{
		Use:     "stop <device>",
		Aliases: []string{"halt", "pause"},
		Short:   "Stop cover",
		Long:    `Stop a cover/roller component on the specified device.`,
		Example: `  # Stop cover movement
  shelly cover stop bedroom

  # Stop specific cover ID
  shelly cover halt bedroom --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], coverID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &coverID, "Cover")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, coverID int) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	ios.StartProgress("Stopping cover...")

	err := svc.CoverStop(ctx, device, coverID)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to stop cover: %w", err)
	}

	ios.Success("Cover %d stopped", coverID)
	return nil
}

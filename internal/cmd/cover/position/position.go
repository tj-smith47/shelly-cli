// Package position provides the cover position subcommand.
package position

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover position command.
func NewCommand() *cobra.Command {
	var coverID int

	cmd := &cobra.Command{
		Use:     "position <device> <percent>",
		Aliases: []string{"pos", "set"},
		Short:   "Set cover position",
		Long: `Set a cover/roller component to a specific position.

Position is specified as a percentage from 0 (closed) to 100 (open).`,
		Example: `  # Set cover to 50% open
  shelly cover position my-cover 50

  # Set cover 1 to fully open
  shelly cover position my-cover 100 --id 1`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			position, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid position: %w", err)
			}
			if position < 0 || position > 100 {
				return fmt.Errorf("position must be between 0 and 100, got %d", position)
			}
			return run(cmd.Context(), args[0], coverID, position)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &coverID, "Cover")

	return cmd
}

func run(ctx context.Context, device string, coverID, position int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner(fmt.Sprintf("Setting cover to %d%%...", position))
	spin.Start()

	err := svc.CoverPosition(ctx, device, coverID, position)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to set cover position: %w", err)
	}

	iostreams.Success("Cover %d set to %d%%", coverID, position)
	return nil
}

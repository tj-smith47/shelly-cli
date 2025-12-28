// Package position provides the cover position subcommand.
package position

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
)

// Options holds command options.
type Options struct {
	flags.ComponentFlags
	Factory  *cmdutil.Factory
	Device   string
	Position int
}

// NewCommand creates the cover position command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Device = args[0]
			position, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid position: %w", err)
			}
			if position < 0 || position > 100 {
				return fmt.Errorf("position must be between 0 and 100, got %d", position)
			}
			opts.Position = position
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Cover")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	err := cmdutil.RunWithSpinner(ctx, ios, fmt.Sprintf("Setting cover to %d%%...", opts.Position), func(ctx context.Context) error {
		return svc.CoverPosition(ctx, opts.Device, opts.ID, opts.Position)
	})
	if err != nil {
		return fmt.Errorf("failed to set cover position: %w", err)
	}

	ios.Success("Cover %d set to %d%%", opts.ID, opts.Position)
	return nil
}

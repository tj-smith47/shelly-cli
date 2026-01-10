// Package devices provides the dash devices subcommand.
package devices

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/tui"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Filter  string
}

// NewCommand creates the dash devices command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "devices",
		Aliases: []string{"device", "dev", "d"},
		Short:   "Launch dashboard in devices view",
		Long: `Launch the TUI dashboard directly in the devices view.

The devices view shows all registered Shelly devices with their current
status, power consumption, and quick actions.`,
		Example: `  # Launch devices view
  shelly dash devices

  # With filter
  shelly dash devices --filter kitchen`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Filter, "filter", "", "Filter devices by name pattern")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	if !ios.IsStdoutTTY() {
		return errors.New("dashboard requires an interactive terminal (TTY)")
	}

	tuiOpts := tui.Options{
		Filter: opts.Filter,
	}

	return tui.Run(ctx, opts.Factory, tuiOpts)
}

// Package events provides the dash events subcommand.
package events

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/tui"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Refresh int
	Filter  string
}

// NewCommand creates the dash events command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "events",
		Aliases: []string{"event", "ev", "e"},
		Short:   "Launch dashboard in events view",
		Long: `Launch the TUI dashboard directly in the events view.

The events view shows a real-time stream of device events including
button presses, state changes, errors, and other notifications.`,
		Example: `  # Launch events view
  shelly dash events

  # With filter
  shelly dash events --filter switch

  # With custom refresh
  shelly dash events --refresh 1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.Refresh, "refresh", 5, "Data refresh interval in seconds")
	cmd.Flags().StringVar(&opts.Filter, "filter", "", "Filter devices by name pattern")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	if !ios.IsStdoutTTY() {
		return errors.New("dashboard requires an interactive terminal (TTY)")
	}

	tuiOpts := tui.Options{
		RefreshInterval: time.Duration(opts.Refresh) * time.Second,
		Filter:          opts.Filter,
	}

	return tui.Run(ctx, opts.Factory, tuiOpts)
}

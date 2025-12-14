// Package eventscmd provides the dash events subcommand.
package eventscmd

import (
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/tui"
)

// NewCommand creates the dash events command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var opts struct {
		refresh int
		filter  string
	}

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
			return run(cmd, f, opts.refresh, opts.filter)
		},
	}

	cmd.Flags().IntVar(&opts.refresh, "refresh", 5, "Data refresh interval in seconds")
	cmd.Flags().StringVar(&opts.filter, "filter", "", "Filter devices by name pattern")

	return cmd
}

func run(cmd *cobra.Command, f *cmdutil.Factory, refresh int, filter string) error {
	ios := f.IOStreams()

	if !ios.IsStdoutTTY() {
		return errors.New("dashboard requires an interactive terminal (TTY)")
	}

	opts := tui.Options{
		RefreshInterval: time.Duration(refresh) * time.Second,
		InitialView:     tui.ViewEvents,
		Filter:          filter,
	}

	return tui.Run(cmd.Context(), f, opts)
}

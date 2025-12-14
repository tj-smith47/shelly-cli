// Package monitor provides the dash monitor subcommand.
package monitor

import (
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/tui"
)

// NewCommand creates the dash monitor command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var opts struct {
		refresh int
		filter  string
	}

	cmd := &cobra.Command{
		Use:     "monitor",
		Aliases: []string{"mon", "m"},
		Short:   "Launch dashboard in monitor view",
		Long: `Launch the TUI dashboard directly in the monitor view.

The monitor view shows real-time status updates for all devices,
including power consumption, temperature, and connectivity.`,
		Example: `  # Launch monitor view
  shelly dash monitor

  # With filter
  shelly dash monitor --filter living

  # With fast refresh
  shelly dash monitor --refresh 2`,
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
		InitialView:     tui.ViewMonitor,
		Filter:          filter,
	}

	return tui.Run(cmd.Context(), f, opts)
}

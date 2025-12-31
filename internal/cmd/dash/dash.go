// Package dash provides the TUI dashboard command.
package dash

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/dash/devices"
	"github.com/tj-smith47/shelly-cli/internal/cmd/dash/events"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/tui"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Refresh int
	Filter  string
}

// NewCommand creates the dash command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "dash",
		Aliases: []string{"dashboard", "ui"},
		Short:   "Launch interactive TUI dashboard",
		Long: `Launch the interactive TUI dashboard for managing Shelly devices.

The dashboard provides real-time monitoring, device control, and energy
tracking in a full-screen terminal interface.

Navigation:
  j/k or arrows  Navigate up/down
  h/l            Select component within device
  t              Toggle device/component
  o/O            Turn on/off
  R              Reboot device
  /              Filter devices
  :              Command mode
  ?              Show keyboard shortcuts
  q              Quit`,
		Example: `  # Launch dashboard with default settings
  shelly dash

  # Launch with 10 second refresh interval
  shelly dash --refresh 10

  # Start with a device filter
  shelly dash --filter kitchen`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.Refresh, "refresh", 5, "Data refresh interval in seconds")
	cmd.Flags().StringVar(&opts.Filter, "filter", "", "Filter devices by name pattern")

	// Add subcommands
	cmd.AddCommand(devices.NewCommand(f))
	cmd.AddCommand(events.NewCommand(f))

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Check if we're in a TTY
	if !ios.IsStdoutTTY() {
		return fmt.Errorf("dashboard requires an interactive terminal (TTY)")
	}

	tuiOpts := tui.Options{
		RefreshInterval: time.Duration(opts.Refresh) * time.Second,
		Filter:          opts.Filter,
	}

	return tui.Run(ctx, opts.Factory, tuiOpts)
}

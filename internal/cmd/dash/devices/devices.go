// Package devices provides the dash devices subcommand.
package devices

import (
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/tui"
)

// NewCommand creates the dash devices command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var opts struct {
		refresh int
		filter  string
	}

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
  shelly dash devices --filter kitchen

  # With custom refresh
  shelly dash devices --refresh 10`,
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
		InitialView:     tui.ViewDevices,
		Filter:          filter,
	}

	return tui.Run(cmd.Context(), f, opts)
}

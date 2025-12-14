// Package dash provides the TUI dashboard command.
package dash

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/tui"
)

// NewCommand creates the dash command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var opts struct {
		refresh int
		view    string
		filter  string
	}

	cmd := &cobra.Command{
		Use:     "dash",
		Aliases: []string{"dashboard", "ui"},
		Short:   "Launch interactive TUI dashboard",
		Long: `Launch the interactive TUI dashboard for managing Shelly devices.

The dashboard provides real-time monitoring, device control, and energy
tracking in a full-screen terminal interface.

Views:
  1 - Devices   List and manage registered devices
  2 - Monitor   Real-time device status monitoring
  3 - Events    Device event stream
  4 - Energy    Power consumption and energy dashboard

Navigation:
  j/k or arrows  Navigate up/down
  tab            Switch between views
  1-4            Jump to specific view
  enter          Select/toggle device
  r              Refresh data
  ?              Show keyboard shortcuts
  q              Quit`,
		Example: `  # Launch dashboard with default settings
  shelly dash

  # Launch with 10 second refresh interval
  shelly dash --refresh 10

  # Start in monitor view
  shelly dash --view monitor

  # Start with a device filter
  shelly dash --filter kitchen`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, f, opts.refresh, opts.view, opts.filter)
		},
	}

	cmd.Flags().IntVar(&opts.refresh, "refresh", 5, "Data refresh interval in seconds")
	cmd.Flags().StringVar(&opts.view, "view", "devices", "Initial view (devices, monitor, events, energy)")
	cmd.Flags().StringVar(&opts.filter, "filter", "", "Filter devices by name pattern")

	return cmd
}

func run(cmd *cobra.Command, f *cmdutil.Factory, refresh int, view, filter string) error {
	ios := f.IOStreams()

	// Check if we're in a TTY
	if !ios.IsStdoutTTY() {
		return fmt.Errorf("dashboard requires an interactive terminal (TTY)")
	}

	// Parse view option
	initialView := tui.ViewDevices
	switch view {
	case "devices", "device", "d", "1":
		initialView = tui.ViewDevices
	case "monitor", "mon", "m", "2":
		initialView = tui.ViewMonitor
	case "events", "event", "e", "3":
		initialView = tui.ViewEvents
	case "energy", "power", "p", "4":
		initialView = tui.ViewEnergy
	}

	opts := tui.Options{
		RefreshInterval: time.Duration(refresh) * time.Second,
		InitialView:     initialView,
		Filter:          filter,
	}

	return tui.Run(cmd.Context(), f, opts)
}

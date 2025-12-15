// Package open provides the open command for opening device web UIs.
package open

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the open command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "open <device>",
		Aliases: []string{"ui", "web", "browse"},
		Short:   "Open device web UI in browser",
		Long: `Open a Shelly device's web interface in your default browser.

The device can be specified by:
  - Registered device name
  - IP address
  - Hostname

This command resolves the device address and opens http://<device-ip>
in your system's default web browser.`,
		Example: `  # Open by registered device name
  shelly open kitchen-light

  # Open by IP address
  shelly open 192.168.1.100

  # Using alias
  shelly ui living-room`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, identifier string) error {
	ios := f.IOStreams()

	// Resolve device address
	address := identifier
	if device, ok := config.GetDevice(identifier); ok {
		address = device.Address
	}

	ios.Info("Opening web UI for %s...", address)

	// Use the browser package to open the device UI
	b := browser.New()
	if err := browser.OpenDeviceUI(ctx, b, address); err != nil {
		return err
	}

	ios.Success("Opened http://%s in browser", address)
	return nil
}

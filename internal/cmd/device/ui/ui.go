// Package ui provides the device ui subcommand to open a device's web interface.
package ui

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the device ui command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ui <device>",
		Aliases: []string{"web", "open"},
		Short:   "Open device web interface in browser",
		Long: `Open a Shelly device's web interface in your default browser.

The device can be specified by name (from config) or by IP address/hostname.`,
		Example: `  # Open web interface by device name
  shelly device ui living-room

  # Open web interface by IP address
  shelly device ui 192.168.1.100

  # Using the 'web' alias
  shelly device web kitchen`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ios := f.IOStreams()

	// Resolve device to get its address
	addr := f.ResolveAddress(device)

	url := fmt.Sprintf("http://%s", addr)

	ios.Info("Opening %s in browser...", url)

	if err := f.Browser().Browse(ctx, url); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	return nil
}

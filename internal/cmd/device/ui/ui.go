// Package ui provides the device ui subcommand to open a device's web interface.
package ui

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	CopyURL bool
	Device  string
}

// NewCommand creates the device ui command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
  shelly device web kitchen

  # Copy URL to clipboard instead of opening
  shelly device ui living-room --copy-url`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.CopyURL, "copy-url", false, "Copy URL to clipboard instead of opening browser")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	br := opts.Factory.Browser()

	// Resolve device to get its address
	addr := opts.Factory.ResolveAddress(opts.Device)
	url := fmt.Sprintf("http://%s", addr)

	// If --copy-url flag is set, copy to clipboard directly
	if opts.CopyURL {
		if err := br.CopyToClipboard(url); err != nil {
			return fmt.Errorf("failed to copy URL to clipboard: %w", err)
		}
		ios.Success("URL copied to clipboard: %s", url)
		return nil
	}

	ios.Info("Opening %s in browser...", url)

	if err := br.OpenDeviceUI(ctx, addr); err != nil {
		// Check if URL was copied to clipboard as fallback
		var clipErr *browser.ClipboardFallbackError
		if errors.As(err, &clipErr) {
			ios.Warning("Could not open browser. URL copied to clipboard: %s", clipErr.URL)
			return nil
		}
		return fmt.Errorf("failed to open browser: %w", err)
	}

	return nil
}

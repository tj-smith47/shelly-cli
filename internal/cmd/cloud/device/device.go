// Package device provides the cloud device subcommand.
package device

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	DeviceID string
	Status   bool
}

// NewCommand creates the cloud device command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "device <id>",
		Aliases: []string{"get"},
		Short:   "Show cloud device details",
		Long: `Show details for a specific device from Shelly Cloud.

Displays device information including status, settings, and online state.`,
		Example: `  # Get device details
  shelly cloud device abc123

  # Get device with full status
  shelly cloud device abc123 --status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.DeviceID = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Status, "status", false, "Show full device status")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*2)
	defer cancel()

	ios := opts.Factory.IOStreams()

	// Check if logged in
	cfg := config.Get()
	if cfg.Cloud.AccessToken == "" {
		ios.Error("Not logged in to Shelly Cloud")
		ios.Info("Use 'shelly cloud login' to authenticate")
		return fmt.Errorf("not logged in")
	}

	// Create cloud client
	client := network.NewCloudClient(cfg.Cloud.AccessToken)

	return cmdutil.RunWithSpinner(ctx, ios, "Fetching device from cloud...", func(ctx context.Context) error {
		device, err := client.GetDevice(ctx, opts.DeviceID)
		if err != nil {
			return fmt.Errorf("failed to get device: %w", err)
		}

		term.DisplayCloudDevice(ios, device, opts.Status)
		return nil
	})
}

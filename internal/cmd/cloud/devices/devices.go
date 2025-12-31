// Package devices provides the cloud devices subcommand.
package devices

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
	Factory *cmdutil.Factory
}

// NewCommand creates the cloud devices command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "devices",
		Aliases: []string{"ls", "list"},
		Short:   "List cloud-registered devices",
		Long: `List all devices registered with your Shelly Cloud account.

Shows device ID, name, model, firmware version, and online status.`,
		Example: `  # List all cloud devices
  shelly cloud devices

  # Output as JSON
  shelly cloud devices -o json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*2) // Longer timeout for cloud
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

	return cmdutil.RunWithSpinner(ctx, ios, "Fetching devices from cloud...", func(ctx context.Context) error {
		devices, err := client.GetAllDevices(ctx)
		if err != nil {
			return fmt.Errorf("failed to get devices: %w", err)
		}

		term.DisplayCloudDevices(ios, devices)
		return nil
	})
}

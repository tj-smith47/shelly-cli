// Package control provides the cloud control subcommand.
package control

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
)

// NewCommand creates the cloud control command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var channel int

	cmd := &cobra.Command{
		Use:     "control <device-id> <action>",
		Aliases: []string{"ctrl", "cmd"},
		Short:   "Control a device via cloud",
		Long: `Control a Shelly device through the Shelly Cloud API.

Supported actions:
  switch:  on, off, toggle
  cover:   open, close, stop, position=<0-100>
  light:   on, off, toggle, brightness=<0-100>

This command requires authentication with 'shelly cloud login'.`,
		Example: `  # Turn on a switch
  shelly cloud control abc123 on

  # Turn off switch on channel 1
  shelly cloud control abc123 off --channel 1

  # Toggle a switch
  shelly cloud control abc123 toggle

  # Set cover to 50%
  shelly cloud control abc123 position=50

  # Open cover
  shelly cloud control abc123 open

  # Set light brightness
  shelly cloud control abc123 brightness=75`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], args[1], channel)
		},
	}

	cmd.Flags().IntVar(&channel, "channel", 0, "Device channel/relay number")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, deviceID, action string, channel int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*2)
	defer cancel()

	ios := f.IOStreams()

	// Check if logged in
	cfg := config.Get()
	if cfg.Cloud.AccessToken == "" {
		ios.Error("Not logged in to Shelly Cloud")
		ios.Info("Use 'shelly cloud login' to authenticate")
		return fmt.Errorf("not logged in")
	}

	// Create cloud client
	client := network.NewCloudClient(cfg.Cloud.AccessToken)

	return cmdutil.RunWithSpinner(ctx, ios, "Sending command...", func(ctx context.Context) error {
		result, err := client.ExecuteAction(ctx, deviceID, action, channel)
		if err != nil {
			return err
		}
		ios.Success("%s", result.Message)
		return nil
	})
}

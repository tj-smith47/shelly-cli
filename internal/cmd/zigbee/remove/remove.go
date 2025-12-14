// Package remove provides the zigbee remove command.
package remove

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	Yes     bool
}

// NewCommand creates the zigbee remove command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "remove <device>",
		Aliases: []string{"leave", "disconnect", "rm"},
		Short:   "Leave Zigbee network",
		Long: `Leave the current Zigbee network and disable Zigbee.

This causes the device to leave its Zigbee network and disables
Zigbee functionality. The device will no longer be controllable
through Zigbee coordinators.

Note: The device will still be accessible via WiFi/HTTP.`,
		Example: `  # Leave Zigbee network
  shelly zigbee remove living-room

  # Leave without confirmation
  shelly zigbee remove living-room --yes`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	if !opts.Yes {
		ios.Warning("This will disconnect from the Zigbee network and disable Zigbee.")
		ios.Printf("Continue? [y/N]: ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			ios.Debug("failed to read response: %v", err)
			return fmt.Errorf("operation cancelled")
		}

		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			ios.Info("Operation cancelled.")
			return nil
		}
	}

	err := svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
		// Disable Zigbee (this also causes the device to leave the network)
		_, err := conn.Call(ctx, "Zigbee.SetConfig", map[string]any{
			"config": map[string]any{
				"enable": false,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to disable Zigbee: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	ios.Success("Zigbee disabled and network left.")
	ios.Info("The device is now only accessible via WiFi/HTTP.")

	return nil
}

// Package pair provides the zigbee pair command.
package pair

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	Timeout int
}

// NewCommand creates the zigbee pair command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "pair <device>",
		Aliases: []string{"join", "connect"},
		Short:   "Start Zigbee network pairing",
		Long: `Start Zigbee network pairing on a Shelly device.

This initiates the network steering process, causing the device to
search for and attempt to join a Zigbee network. The coordinator
(e.g., Home Assistant ZHA, Zigbee2MQTT) must be in pairing mode.

The command enables Zigbee if not already enabled, then starts
network steering. Use --timeout to specify how long to wait for
the device to join a network.`,
		Example: `  # Start pairing with default 180s timeout
  shelly zigbee pair living-room

  # Start pairing with custom timeout
  shelly zigbee pair living-room --timeout 60`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.Timeout, "timeout", 180, "Pairing timeout in seconds")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(opts.Timeout+30)*time.Second)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	ios.Println(theme.Bold().Render("Starting Zigbee Pairing..."))
	ios.Println()

	// Enable Zigbee if not already enabled
	ios.Info("Enabling Zigbee...")
	if err := svc.ZigbeeEnable(ctx, opts.Device); err != nil {
		return err
	}

	// Wait a moment for Zigbee to initialize
	time.Sleep(2 * time.Second)

	// Start network steering
	ios.Info("Starting network steering...")
	if err := svc.ZigbeeStartNetworkSteering(ctx, opts.Device); err != nil {
		return err
	}

	ios.Println()
	ios.Success("Network steering started!")
	ios.Println()
	ios.Info("The device is now searching for Zigbee networks.")
	ios.Info("Make sure your coordinator is in pairing mode.")
	ios.Println()
	ios.Info("Check status with: shelly zigbee status %s", opts.Device)
	ios.Info("The network_state will change to 'joined' when successful.")

	return nil
}

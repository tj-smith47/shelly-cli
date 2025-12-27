// Package on provides the fleet on subcommand.
package on

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	All   bool
	Group string
}

// NewCommand creates the fleet on command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "on [device...]",
		Aliases: []string{"turn-on", "enable"},
		Short:   "Turn on devices via cloud",
		Long: `Turn on devices through Shelly Cloud.

Uses cloud WebSocket connections to send commands, allowing control
of devices even when not on the same local network.

Requires integrator credentials configured via environment variables or config:
  SHELLY_INTEGRATOR_TAG - Your integrator tag
  SHELLY_INTEGRATOR_TOKEN - Your integrator token`,
		Example: `  # Turn on specific device
  shelly fleet on device-id

  # Turn on all devices in a group
  shelly fleet on --group living-room

  # Turn on all relay devices
  shelly fleet on --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "Turn on all relay devices")
	cmd.Flags().StringVarP(&opts.Group, "group", "g", "", "Turn on devices in group")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, devices []string, opts *Options) error {
	ios := f.IOStreams()

	// Validate arguments
	if !opts.All && opts.Group == "" && len(devices) == 0 {
		ios.Warning("Specify devices, --group, or --all")
		return fmt.Errorf("no devices specified")
	}

	// Get credentials and connect
	cfg, cfgErr := f.Config()
	if cfgErr != nil {
		ios.DebugErr("load config", cfgErr)
	}

	creds, err := shelly.GetIntegratorCredentials(ios, cfg)
	if err != nil {
		return err
	}

	conn, err := shelly.ConnectFleet(ctx, ios, creds)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Execute the command based on options
	var results []integrator.BatchResult
	switch {
	case opts.All:
		results = conn.Manager.AllRelaysOn(ctx)
	case opts.Group != "":
		results = conn.Manager.GroupRelaysOn(ctx, opts.Group)
	default:
		commands := make([]integrator.BatchCommand, len(devices))
		for i, deviceID := range devices {
			commands[i] = integrator.BatchCommand{
				DeviceID: deviceID,
				Action:   "relay",
				Params:   map[string]any{"id": 0, "turn": "on"},
			}
		}
		results = conn.Manager.SendBatchCommands(ctx, commands)
	}

	// Report results
	return shelly.ReportBatchResults(ios, results, "turned on")
}

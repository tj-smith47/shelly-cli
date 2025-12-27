// Package off provides the fleet off subcommand.
package off

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	All   bool
	Group string
}

// NewCommand creates the fleet off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "off [device...]",
		Aliases: []string{"turn-off", "disable"},
		Short:   "Turn off devices via cloud",
		Long: `Turn off devices through Shelly Cloud.

Uses cloud WebSocket connections to send commands, allowing control
of devices even when not on the same local network.

Requires integrator credentials configured via environment variables or config:
  SHELLY_INTEGRATOR_TAG - Your integrator tag
  SHELLY_INTEGRATOR_TOKEN - Your integrator token`,
		Example: `  # Turn off specific device
  shelly fleet off device-id

  # Turn off all devices in a group
  shelly fleet off --group living-room

  # Turn off all relay devices
  shelly fleet off --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "Turn off all relay devices")
	cmd.Flags().StringVarP(&opts.Group, "group", "g", "", "Turn off devices in group")

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

	// Execute relay control
	results := shelly.FleetRelayControl(ctx, conn.Manager, shelly.RelayOff, devices, opts.All, opts.Group)
	return shelly.ReportBatchResults(ios, results, "turned off")
}

// Package toggle provides the fleet toggle subcommand.
package toggle

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	flags.BatchFlags
}

// NewCommand creates the fleet toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "toggle [device...]",
		Aliases: []string{"flip", "switch"},
		Short:   "Toggle devices via cloud",
		Long: `Toggle devices through Shelly Cloud.

Uses cloud WebSocket connections to send commands, allowing control
of devices even when not on the same local network.

Requires integrator credentials configured via environment variables or config:
  SHELLY_INTEGRATOR_TAG - Your integrator tag
  SHELLY_INTEGRATOR_TOKEN - Your integrator token`,
		Example: `  # Toggle specific device
  shelly fleet toggle device-id

  # Toggle all devices in a group
  shelly fleet toggle --group living-room

  # Toggle all relay devices
  shelly fleet toggle --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args, opts)
		},
	}

	flags.AddBatchFlags(cmd, &opts.BatchFlags)

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, devices []string, opts *Options) error {
	ios := f.IOStreams()

	// Validate arguments
	if !opts.All && opts.GroupName == "" && len(devices) == 0 {
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
	results := shelly.FleetRelayControl(ctx, conn.Manager, shelly.RelayToggle, devices, opts.All, opts.GroupName)
	return shelly.ReportBatchResults(ios, results, "toggled")
}

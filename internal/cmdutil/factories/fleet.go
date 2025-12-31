// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// FleetRelayOpts configures a fleet relay command (on/off/toggle for cloud-connected devices).
type FleetRelayOpts struct {
	// Action is the relay action to perform.
	Action shelly.RelayAction

	// Aliases are alternate command names.
	Aliases []string

	// Short is the short description.
	Short string

	// Long is the detailed description.
	Long string

	// Example shows usage examples.
	Example string

	// SuccessVerb is the verb for success messages (e.g., "turned on", "turned off", "toggled").
	SuccessVerb string
}

// fleetRelayOptions holds the runtime options for a fleet relay command.
type fleetRelayOptions struct {
	flags.BatchFlags
	Devices []string
	Factory *cmdutil.Factory
	Config  FleetRelayOpts
}

// NewFleetRelayCommand creates a fleet relay control command.
func NewFleetRelayCommand(f *cmdutil.Factory, config FleetRelayOpts) *cobra.Command {
	opts := &fleetRelayOptions{
		Factory: f,
		Config:  config,
	}

	cmd := &cobra.Command{
		Use:     string(config.Action) + " [device...]",
		Aliases: config.Aliases,
		Short:   config.Short,
		Long:    config.Long,
		Example: config.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Devices = args
			return runFleetRelay(cmd.Context(), opts)
		},
	}

	flags.AddBatchFlags(cmd, &opts.BatchFlags)

	return cmd
}

func runFleetRelay(ctx context.Context, opts *fleetRelayOptions) error {
	f := opts.Factory
	ios := f.IOStreams()

	// Validate arguments
	if !opts.All && opts.GroupName == "" && len(opts.Devices) == 0 {
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
	results := shelly.FleetRelayControl(ctx, conn.Manager, opts.Config.Action, opts.Devices, opts.All, opts.GroupName)
	return shelly.ReportBatchResults(ios, results, opts.Config.SuccessVerb)
}

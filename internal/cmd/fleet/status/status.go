// Package status provides the fleet status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Online  bool
	Offline bool
}

// NewCommand creates the fleet status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"st", "list", "ls"},
		Short:   "View fleet device status",
		Long: `View the status of all devices in your fleet.

Shows online/offline status and last seen time for each device
connected through Shelly Cloud.

Requires an active fleet connection. Run 'shelly fleet connect' first.`,
		Example: `  # View all device status
  shelly fleet status

  # Show only online devices
  shelly fleet status --online

  # Show only offline devices
  shelly fleet status --offline

  # JSON output
  shelly fleet status -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Online, "online", false, "Show only online devices")
	cmd.Flags().BoolVar(&opts.Offline, "offline", false, "Show only offline devices")

	return cmd
}

//nolint:gocyclo,nestif // Complexity from handling demo mode and filtering options
func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	var statuses []*integrator.DeviceStatus
	var org string

	// Check for demo mode
	if mock.IsDemoMode() && mock.HasFleetFixtures() {
		statuses = mock.GetFleetDeviceStatuses()
		org = mock.GetFleetOrganization()
	} else {
		// Get credentials
		cfg, cfgErr := opts.Factory.Config()
		if cfgErr != nil {
			ios.DebugErr("load config", cfgErr)
		}
		creds, err := shelly.GetIntegratorCredentials(ios, cfg)
		if err != nil {
			return err
		}

		// Connect to fleet
		fc, err := shelly.ConnectFleet(ctx, ios, creds)
		if err != nil {
			return err
		}
		defer fc.Close()

		statuses = fc.Manager.ListDeviceStatuses()
	}

	// Filter device statuses
	if opts.Online || opts.Offline {
		filtered := make([]*integrator.DeviceStatus, 0, len(statuses))
		for _, s := range statuses {
			if opts.Online && !s.Online {
				continue
			}
			if opts.Offline && s.Online {
				continue
			}
			filtered = append(filtered, s)
		}
		statuses = filtered
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, statuses)
	}

	if len(statuses) == 0 {
		ios.Warning("No devices found matching criteria")
		return nil
	}

	if org != "" {
		ios.Info("Organization: %s", org)
	}
	ios.Success("Fleet Status (%d devices)", len(statuses))
	ios.Println()
	term.DisplayFleetStatus(ios, statuses)

	return nil
}

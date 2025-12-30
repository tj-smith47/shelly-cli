// Package status provides the fleet status subcommand.
package status

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Online  bool
	Offline bool
}

// NewCommand creates the fleet status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

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
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Online, "online", false, "Show only online devices")
	cmd.Flags().BoolVar(&opts.Offline, "offline", false, "Show only offline devices")

	return cmd
}

//nolint:gocyclo // Complexity from handling demo mode and filtering options
func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	// Check for demo mode
	if mock.IsDemoMode() && mock.HasFleetFixtures() {
		return runDemoMode(f, opts)
	}

	// Get credentials
	cfg, cfgErr := f.Config()
	if cfgErr != nil {
		ios.DebugErr("load config", cfgErr)
	}
	tag, token, err := cfg.GetIntegratorCredentials()
	if err != nil {
		return fmt.Errorf("%w. Run 'shelly fleet connect' first", err)
	}

	// Create client and authenticate
	client := integrator.New(tag, token)
	if err := client.Authenticate(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Create fleet manager and connect
	fm := integrator.NewFleetManager(client)
	ios.Info("Connecting to fleet...")
	for host, connErr := range fm.ConnectAll(ctx, nil) {
		ios.Warning("Failed to connect to %s: %v", host, connErr)
	}

	// Get and filter device statuses
	statuses := fm.ListDeviceStatuses()
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

	ios.Success("Fleet Status (%d devices)", len(statuses))
	ios.Println()
	term.DisplayFleetStatus(ios, statuses)

	return nil
}

func runDemoMode(f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	// Get mock fleet devices
	mockDevices := mock.GetFleetDevices()

	// Convert to integrator.DeviceStatus for display compatibility
	statuses := make([]*integrator.DeviceStatus, 0, len(mockDevices))
	for _, d := range mockDevices {
		if opts.Online && !d.Online {
			continue
		}
		if opts.Offline && d.Online {
			continue
		}
		statuses = append(statuses, &integrator.DeviceStatus{
			DeviceID: d.DeviceID,
			Online:   d.Online,
			LastSeen: d.LastSeen,
			Host:     d.Host,
		})
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, statuses)
	}

	if len(statuses) == 0 {
		ios.Warning("No devices found matching criteria")
		return nil
	}

	org := mock.GetFleetOrganization()
	if org != "" {
		ios.Info("Organization: %s", org)
	}
	ios.Success("Fleet Status (%d devices)", len(statuses))
	ios.Println()
	term.DisplayFleetStatus(ios, statuses)

	return nil
}

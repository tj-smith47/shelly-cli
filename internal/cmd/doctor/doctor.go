// Package doctor provides the doctor command for system diagnostics.
package doctor

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Network bool
	Devices bool
	Full    bool
}

// NewCommand creates the doctor command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"diagnose", "check", "health"},
		Short:   "Check system health and diagnose issues",
		Long: `Run comprehensive diagnostics on the Shelly CLI setup.

Checks include:
  - CLI version and update availability
  - Configuration file validity
  - Registered devices and their reachability
  - Network connectivity
  - Cloud authentication status
  - Firmware update availability

Use --full for all checks including device connectivity tests.`,
		Example: `  # Run basic diagnostics
  shelly doctor

  # Check network connectivity
  shelly doctor --network

  # Test all registered devices
  shelly doctor --devices

  # Full diagnostic suite
  shelly doctor --full`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Full {
				opts.Network = true
				opts.Devices = true
			}
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Network, "network", false, "Check network connectivity")
	cmd.Flags().BoolVar(&opts.Devices, "devices", false, "Test device reachability")
	cmd.Flags().BoolVar(&opts.Full, "full", false, "Run all diagnostic checks")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	term.PrintDoctorHeader(ios)

	issues := 0
	warnings := 0

	// Check CLI version
	verIssues, verWarnings := term.CheckCLIVersion(ios)
	issues += verIssues
	warnings += verWarnings

	// Check config
	issues += term.CheckConfig(ios)

	// Check devices
	devIssues, devWarnings := term.CheckDevices(ios)
	issues += devIssues
	warnings += devWarnings

	// Check cloud auth
	warnings += term.CheckCloudAuth(ios)

	// Optional: Network checks
	if opts.Network {
		issues += term.CheckNetwork(ctx, ios)
	}

	// Optional: Device connectivity
	if opts.Devices {
		svc := f.ShellyService()
		issues += term.CheckDeviceConnectivity(ctx, ios, svc)
	}

	term.PrintDoctorSummary(ios, issues, warnings)

	return nil
}

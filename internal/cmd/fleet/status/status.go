// Package status provides the fleet status subcommand.
package status

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	// Get credentials
	cfg, cfgErr := f.Config()
	if cfgErr != nil {
		ios.DebugErr("load config", cfgErr)
	}
	tag, token, err := loadCredentials(cfg)
	if err != nil {
		return err
	}

	// Create client and authenticate
	client := integrator.New(tag, token)
	if err := client.Authenticate(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Create fleet manager and connect
	fm := integrator.NewFleetManager(client)
	ios.Info("Connecting to fleet...")
	reportConnectionErrors(ios, fm.ConnectAll(ctx, nil))

	// Get and filter device statuses
	statuses := filterDevices(fm.ListDeviceStatuses(), opts)

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

// loadCredentials loads integrator credentials from environment or config.
func loadCredentials(cfg *config.Config) (tag, token string, err error) {
	tag = os.Getenv("SHELLY_INTEGRATOR_TAG")
	token = os.Getenv("SHELLY_INTEGRATOR_TOKEN")

	if cfg != nil {
		if tag == "" {
			tag = cfg.Integrator.Tag
		}
		if token == "" {
			token = cfg.Integrator.Token
		}
	}

	if tag == "" || token == "" {
		return "", "", fmt.Errorf("integrator credentials required. Run 'shelly fleet connect' first")
	}
	return tag, token, nil
}

// reportConnectionErrors logs warnings for any connection failures.
func reportConnectionErrors(ios *iostreams.IOStreams, errors map[string]error) {
	for host, err := range errors {
		ios.Warning("Failed to connect to %s: %v", host, err)
	}
}

// filterDevices filters device statuses based on online/offline options.
func filterDevices(statuses []*integrator.DeviceStatus, opts *Options) []*integrator.DeviceStatus {
	if !opts.Online && !opts.Offline {
		return statuses
	}

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
	return filtered
}

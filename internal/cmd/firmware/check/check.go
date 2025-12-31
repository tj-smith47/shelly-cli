// Package check provides the firmware check subcommand.
package check

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	All     bool
	Devices []string
}

// NewCommand creates the firmware check command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "check [device]",
		Aliases: []string{"ck"},
		Short:   "Check for firmware updates",
		Long: `Check if firmware updates are available for a device.

Use --all to check all registered devices.`,
		Example: `  # Check a specific device
  shelly firmware check living-room

  # Check all registered devices
  shelly firmware check --all`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Devices = args
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "Check all registered devices")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	if opts.All {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(cfg.Devices) == 0 {
			ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
			return nil
		}

		// Use platform-aware checking that includes plugin-managed devices
		results := svc.CheckFirmwareAllPlatforms(ctx, ios, cfg.Devices)
		term.DisplayFirmwareCheckAll(ios, results)
		return nil
	}

	if len(opts.Devices) == 0 {
		return fmt.Errorf("device name required (or use --all)")
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	return cmdutil.RunDeviceStatus(ctx, ios, svc, opts.Devices[0],
		"Checking for updates...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.FirmwareInfo, error) {
			return svc.CheckFirmware(ctx, device)
		},
		term.DisplayFirmwareInfo)
}

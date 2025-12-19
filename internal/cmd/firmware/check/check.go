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

var allFlag bool

// NewCommand creates the firmware check command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			if allFlag {
				return runAll(f, cmd.Context())
			}
			if len(args) == 0 {
				return fmt.Errorf("device name required (or use --all)")
			}
			return run(f, cmd.Context(), args[0])
		},
	}

	cmd.Flags().BoolVar(&allFlag, "all", false, "Check all registered devices")

	return cmd
}

func run(f *cmdutil.Factory, ctx context.Context, device string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunDeviceStatus(ctx, ios, svc, device,
		"Checking for updates...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.FirmwareInfo, error) {
			return svc.CheckFirmware(ctx, device)
		},
		term.DisplayFirmwareInfo)
}

func runAll(f *cmdutil.Factory, ctx context.Context) error {
	ios := f.IOStreams()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Devices) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return nil
	}

	// Get device names
	deviceNames := make([]string, 0, len(cfg.Devices))
	for name := range cfg.Devices {
		deviceNames = append(deviceNames, name)
	}

	svc := f.ShellyService()
	results := svc.CheckFirmwareAll(ctx, ios, deviceNames)
	term.DisplayFirmwareCheckAll(ios, results)

	return nil
}

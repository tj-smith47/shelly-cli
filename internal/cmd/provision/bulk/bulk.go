// Package bulk provides bulk device provisioning.
package bulk

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	ConfigFile string
	Parallel   int
	DryRun     bool
	Factory    *cmdutil.Factory
}

// NewCommand creates the provision bulk command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "bulk <config-file>",
		Aliases: []string{"batch", "mass"},
		Short:   "Bulk provision from config file",
		Long: `Provision multiple devices from a YAML configuration file.

The config file specifies WiFi credentials and device-specific settings.
Provisioning is performed in parallel for efficiency.

Config file format:
  wifi:
    ssid: "MyNetwork"
    password: "secret"
  devices:
    - name: living-room
      address: 192.168.1.100  # optional, uses registered device if omitted
      device_name: "Living Room Light"  # optional device name to set
    - name: bedroom
      wifi:  # optional per-device WiFi override
        ssid: "OtherNetwork"
        password: "other-secret"`,
		Example: `  # Provision devices from config file
  shelly provision bulk devices.yaml

  # Dry run to validate config
  shelly provision bulk devices.yaml --dry-run

  # Limit parallel operations
  shelly provision bulk devices.yaml --parallel 2`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ConfigFile = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.Parallel, "parallel", 5, "Maximum parallel provisioning operations")
	flags.AddDryRunFlag(cmd, &opts.DryRun)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Load config file
	cfg, err := shelly.ParseBulkProvisionFile(opts.ConfigFile)
	if err != nil {
		return err
	}

	if len(cfg.Devices) == 0 {
		return fmt.Errorf("no devices specified in config file")
	}

	// Validate all device names before starting
	isRegistered := func(name string) bool {
		return opts.Factory.GetDevice(name) != nil
	}
	if err := shelly.ValidateBulkProvisionConfig(cfg, isRegistered); err != nil {
		return err
	}

	ios.Info("Found %d devices to provision", len(cfg.Devices))

	if opts.DryRun {
		term.DisplayBulkProvisionDryRun(ios, cfg)
		return nil
	}

	// Add timeout for entire bulk operation
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Provision devices in parallel
	results := svc.ProvisionDevices(ctx, cfg, opts.Parallel)

	// Display results
	failed := term.DisplayBulkProvisionResults(ios, results, len(cfg.Devices))
	if failed > 0 {
		return fmt.Errorf("%d devices failed to provision", failed)
	}

	return nil
}

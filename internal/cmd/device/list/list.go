// Package list provides the device list subcommand.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.DeviceListFlags
	Factory *cmdutil.Factory
}

// NewCommand creates the device list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List registered devices",
		Long: `List all devices registered in the local registry.

The registry stores device information including name, address, model,
generation, platform, and authentication credentials. Use filters to
narrow results by device generation, device type, or platform.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting and piping to tools like jq.

Columns: Name, Address, Platform, Type, Model, Generation, Auth`,
		Example: `  # List all registered devices
  shelly device list

  # List only Gen2 devices
  shelly device list --generation 2

  # List devices by type
  shelly device list --type SHSW-1

  # List only Shelly devices (exclude plugin-managed)
  shelly device list --platform shelly

  # List only Tasmota devices (from shelly-tasmota plugin)
  shelly device list --platform tasmota

  # Show firmware versions and sort updates first
  shelly device list --version --updates-first

  # Output as JSON for scripting
  shelly device list -o json

  # Pipe to jq to extract device names
  shelly device list -o json | jq -r '.[].name'

  # Parse table output in scripts (disable colors)
  shelly device list --no-color | tail -n +2 | while read name addr _; do
    echo "Device: $name at $addr"
  done

  # Export to CSV via jq
  shelly device list -o json | jq -r '.[] | [.name,.address,.model] | @csv'

  # Short form
  shelly dev ls`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	flags.AddDeviceListFlags(cmd, &opts.DeviceListFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	mgr, err := opts.Factory.ConfigManager()
	if err != nil {
		return err
	}
	devices := mgr.ListDevices()

	if len(devices) == 0 {
		ios.Info("No devices registered")
		ios.Info("Use 'shelly discover' to find devices or 'shelly device add' to register one")
		return nil
	}

	// Apply filters and build sorted list
	filterOpts := shelly.DeviceListFilterOptions{
		Generation: opts.Generation,
		DeviceType: opts.DeviceType,
		Platform:   opts.Platform,
	}
	filtered, platforms := shelly.FilterDeviceList(devices, filterOpts)

	if len(filtered) == 0 {
		ios.Info("No devices match the specified filters")
		return nil
	}

	// Populate firmware info if version display or updates-first sorting is requested
	if opts.ShowVersion || opts.UpdatesFirst {
		svc := opts.Factory.ShellyService()
		svc.PopulateDeviceListFirmware(ctx, filtered)
	}

	// Sort: updates first if requested, then by name
	shelly.SortDeviceList(filtered, opts.UpdatesFirst)

	// Handle structured output (JSON/YAML)
	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, filtered)
	}

	// Show Platform column only when there are multiple platforms
	showPlatform := len(platforms) > 1
	term.DisplayDeviceList(ios, filtered, showPlatform, opts.ShowVersion)

	return nil
}

// Package ansible provides the export ansible subcommand.
package ansible

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
)

// yamlExtensions defines valid YAML file extensions.
var yamlExtensions = []string{".yaml", ".yml"}

// Options holds command options.
type Options struct {
	Devices   []string
	File      string
	GroupName string
	Factory   *cmdutil.Factory
}

// NewCommand creates the export ansible command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "ansible <devices...> [file]",
		Aliases: []string{"ans"},
		Short:   "Export devices as Ansible inventory",
		Long: `Export devices as an Ansible inventory YAML file.

Creates an Ansible-compatible inventory with device groups based on
model type. Use @all to export all registered devices.`,
		Example: `  # Export to stdout
  shelly export ansible @all

  # Export to file
  shelly export ansible @all inventory.yaml

  # Export specific devices
  shelly export ansible living-room bedroom inventory.yaml

  # Specify group name
  shelly export ansible @all --group-name shelly_devices`,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: completion.DevicesWithGroups(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Devices, opts.File = shelly.SplitDevicesAndFile(args, yamlExtensions)
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.GroupName, "group-name", "shelly", "Ansible group name for devices")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, 2*shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Expand @all to all registered devices
	devices := completion.ExpandDeviceArgs(opts.Devices)
	if len(devices) == 0 {
		return fmt.Errorf("no devices specified")
	}

	// Collect device data using shared helper
	var deviceData []model.DeviceData
	err := cmdutil.RunWithSpinner(ctx, ios, "Fetching device data...", func(ctx context.Context) error {
		deviceData = svc.CollectDeviceData(ctx, devices)
		return nil
	})
	if err != nil {
		return err
	}

	// Build inventory using export builder
	_, data, err := export.BuildAnsibleInventory(deviceData, opts.GroupName)
	if err != nil {
		return fmt.Errorf("failed to build inventory: %w", err)
	}

	// Output
	if opts.File == "" {
		ios.Printf("%s", string(data))
		return nil
	}

	if err := afero.WriteFile(config.Fs(), opts.File, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	ios.Success("Exported %d devices to %s", len(deviceData), opts.File)
	return nil
}

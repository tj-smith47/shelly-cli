// Package terraform provides the export terraform subcommand.
package terraform

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
)

// tfExtensions defines valid Terraform file extensions.
var tfExtensions = []string{".tf"}

// Options holds command options.
type Options struct {
	Devices      []string
	File         string
	ResourceName string
	Factory      *cmdutil.Factory
}

// NewCommand creates the export terraform command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "terraform <devices...> [file]",
		Aliases: []string{"tf"},
		Short:   "Export devices as Terraform configuration",
		Long: `Export devices as Terraform local values configuration.

Creates a Terraform locals block with device information that can be
used as data source for infrastructure as code workflows.`,
		Example: `  # Export to stdout
  shelly export terraform @all

  # Export to file
  shelly export terraform @all shelly.tf

  # Export with custom resource name
  shelly export terraform @all --resource-name my_shelly_devices`,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: completion.DevicesWithGroups(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Devices, opts.File = shelly.SplitDevicesAndFile(args, tfExtensions)
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.ResourceName, "resource-name", "shelly_devices", "Terraform local variable name")

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
	var deviceData []shelly.DeviceData
	err := cmdutil.RunWithSpinner(ctx, ios, "Fetching device data...", func(ctx context.Context) error {
		deviceData = svc.CollectDeviceData(ctx, devices)
		return nil
	})
	if err != nil {
		return err
	}

	// Build Terraform config using export builder
	config, err := export.BuildTerraformConfig(deviceData, opts.ResourceName)
	if err != nil {
		return fmt.Errorf("failed to build terraform config: %w", err)
	}

	// Output
	if opts.File == "" {
		ios.Printf("%s", config)
		return nil
	}

	file, err := os.Create(opts.File)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer iostreams.CloseWithDebug("closing terraform export file", file)

	if _, err := file.WriteString(config); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	ios.Success("Exported %d devices to %s", len(deviceData), opts.File)
	return nil
}

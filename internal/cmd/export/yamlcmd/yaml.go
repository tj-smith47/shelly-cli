// Package yamlcmd provides the export yaml subcommand.
package yamlcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device  string
	File    string
	Full    bool
	Factory *cmdutil.Factory
}

// NewCommand creates the export yaml command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:   "yaml <device> [file]",
		Short: "Export device configuration as YAML",
		Long: `Export a device's configuration and status as YAML.

If no file is specified, outputs to stdout.
Use --full to include device status in addition to configuration.`,
		Example: `  # Export to stdout
  shelly export yaml living-room

  # Export to file
  shelly export yaml living-room device.yaml

  # Export with full status
  shelly export yaml living-room --full`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completeDevice(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			if len(args) > 1 {
				opts.File = args[1]
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Full, "full", false, "Include device status (not just config)")

	return cmd
}

// DeviceExport holds the exported device data.
type DeviceExport struct {
	Name       string         `yaml:"name"`
	Address    string         `yaml:"address"`
	Model      string         `yaml:"model"`
	Generation int            `yaml:"generation"`
	Config     map[string]any `yaml:"config"`
	Status     map[string]any `yaml:"status,omitempty"`
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Get device config for name/address
	deviceCfg, exists := config.GetDevice(opts.Device)
	if !exists {
		return fmt.Errorf("device %q not found in config", opts.Device)
	}

	// Get device export data
	var export DeviceExport
	export.Name = opts.Device
	export.Address = deviceCfg.Address

	err := cmdutil.RunWithSpinner(ctx, ios, "Fetching device data...", func(ctx context.Context) error {
		return svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
			info := conn.Info()
			export.Model = info.Model
			export.Generation = info.Generation

			// Get config
			cfg, err := conn.GetConfig(ctx)
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}
			export.Config = cfg

			// Get status if full export requested
			if opts.Full {
				status, err := conn.GetStatus(ctx)
				if err != nil {
					return fmt.Errorf("failed to get status: %w", err)
				}
				export.Status = status
			}

			return nil
		})
	})
	if err != nil {
		return err
	}

	// Serialize to YAML
	data, err := yaml.Marshal(export)
	if err != nil {
		return fmt.Errorf("failed to serialize YAML: %w", err)
	}

	// Output
	if opts.File == "" {
		ios.Printf("%s", string(data))
		return nil
	}

	if err := os.WriteFile(opts.File, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	ios.Success("Device %q exported to %s", opts.Device, opts.File)
	return nil
}

// completeDevice provides completion for device argument.
func completeDevice() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			devices := config.ListDevices()
			completions := make([]string, 0, len(devices))
			for name := range devices {
				completions = append(completions, name)
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			return nil, cobra.ShellCompDirectiveDefault
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

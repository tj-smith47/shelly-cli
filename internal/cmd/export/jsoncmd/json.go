// Package jsoncmd provides the export json subcommand.
package jsoncmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device  string
	File    string
	Pretty  bool
	Full    bool
	Factory *cmdutil.Factory
}

// NewCommand creates the export json command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f, Pretty: true}

	cmd := &cobra.Command{
		Use:   "json <device> [file]",
		Short: "Export device configuration as JSON",
		Long: `Export a device's configuration and status as JSON.

If no file is specified, outputs to stdout.
Use --full to include device status in addition to configuration.`,
		Example: `  # Export to stdout
  shelly export json living-room

  # Export to file
  shelly export json living-room device.json

  # Export with full status
  shelly export json living-room --full

  # Export compact (no pretty printing)
  shelly export json living-room --no-pretty`,
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

	cmd.Flags().BoolVar(&opts.Pretty, "pretty", true, "Pretty print JSON output")
	cmd.Flags().BoolVar(&opts.Full, "full", false, "Include device status (not just config)")

	return cmd
}

// DeviceExport holds the exported device data.
type DeviceExport struct {
	Name       string         `json:"name"`
	Address    string         `json:"address"`
	Model      string         `json:"model"`
	Generation int            `json:"generation"`
	Config     map[string]any `json:"config"`
	Status     map[string]any `json:"status,omitempty"`
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

	// Serialize to JSON
	var data []byte
	if opts.Pretty {
		data, err = json.MarshalIndent(export, "", "  ")
	} else {
		data, err = json.Marshal(export)
	}
	if err != nil {
		return fmt.Errorf("failed to serialize JSON: %w", err)
	}

	// Output
	if opts.File == "" {
		ios.Printf("%s\n", string(data))
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

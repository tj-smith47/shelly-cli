// Package info provides the device info subcommand.
package info

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the device info command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "info <device>",
		Aliases: []string{"details", "show"},
		Short:   "Show device information",
		Long: `Show detailed information about a device.

The device can be specified by its registered name or IP address.`,
		Example: `  # Show info for a registered device
  shelly device info living-room

  # Show info by IP address
  shelly device info 192.168.1.100

  # Output as JSON
  shelly device info living-room -o json

  # Output as YAML
  shelly device info living-room -o yaml

  # Short form
  shelly dev info office-switch`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	svc := opts.Factory.ShellyService()
	ios := opts.Factory.IOStreams()

	var info *shelly.DeviceInfo
	err := cmdutil.RunWithSpinner(ctx, ios, "Getting device info...", func(ctx context.Context) error {
		var err error
		// Use DeviceInfoAuto to support both Gen1 and Gen2 devices
		info, err = svc.DeviceInfoAuto(ctx, opts.Device)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to get device info: %w", err)
	}

	// Check output format
	if output.WantsJSON() {
		return output.PrintJSON(info)
	}
	if output.WantsYAML() {
		return output.PrintYAML(info)
	}

	// Display info as a formatted table
	builder := table.NewBuilder("Property", "Value")

	builder.AddRow("ID", info.ID)
	builder.AddRow("MAC", info.MAC)
	builder.AddRow("Model", info.Model)
	builder.AddRow("Generation", output.RenderGeneration(info.Generation))
	builder.AddRow("Firmware", info.Firmware)
	builder.AddRow("Application", info.App)
	builder.AddRow("Auth Enabled", output.RenderAuthRequired(info.AuthEn))

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print device info table", err)
	}

	return nil
}

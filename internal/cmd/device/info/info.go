// Package info provides the device info subcommand.
package info

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the device info command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	var info *shelly.DeviceInfo
	err := cmdutil.RunWithSpinner(ctx, ios, "Getting device info...", func(ctx context.Context) error {
		var err error
		// Use DeviceInfoAuto to support both Gen1 and Gen2 devices
		info, err = svc.DeviceInfoAuto(ctx, device)
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
	table := output.NewTable("Property", "Value")

	table.AddRow("ID", info.ID)
	table.AddRow("MAC", info.MAC)
	table.AddRow("Model", info.Model)
	table.AddRow("Generation", output.RenderGeneration(info.Generation))
	table.AddRow("Firmware", info.Firmware)
	table.AddRow("Application", info.App)
	table.AddRow("Auth Enabled", output.RenderAuthRequired(info.AuthEn))

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print device info table", err)
	}

	return nil
}

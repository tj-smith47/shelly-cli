// Package info provides the device info subcommand.
package info

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// NewCommand creates the device info command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <device>",
		Short: "Show device information",
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
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args[0])
		},
	}

	return cmd
}

func run(device string) error {
	ctx, cancel := context.WithTimeout(context.Background(), shelly.DefaultTimeout)
	defer cancel()

	svc := shelly.NewService()

	spin := iostreams.NewSpinner("Getting device info...")
	spin.Start()

	info, err := svc.DeviceInfo(ctx, device)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to get device info: %w", err)
	}

	// Check output format
	format := viper.GetString("output")
	if format == "json" {
		return output.PrintJSON(info)
	}
	if format == "yaml" {
		return output.PrintYAML(info)
	}

	// Display info as a formatted table
	table := output.NewTable("Property", "Value")

	table.AddRow("ID", info.ID)
	table.AddRow("MAC", info.MAC)
	table.AddRow("Model", info.Model)
	table.AddRow("Generation", helpers.FormatGeneration(info.Generation))
	table.AddRow("Firmware", info.Firmware)
	table.AddRow("Application", info.App)
	table.AddRow("Auth Enabled", helpers.FormatAuth(info.AuthEn))

	table.Print()

	return nil
}

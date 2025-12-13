// Package list provides the device list subcommand.
package list

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the device list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		generation int
		deviceType string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List registered devices",
		Long: `List all devices registered in the local registry.

Registered devices can be filtered by generation or device type.`,
		Example: `  # List all registered devices
  shelly device list

  # List only Gen2 devices
  shelly device list --generation 2

  # List devices by type
  shelly device list --type SHSW-1

  # Short form
  shelly dev ls`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(generation, deviceType)
		},
	}

	cmd.Flags().IntVarP(&generation, "generation", "g", 0, "Filter by generation (1, 2, or 3)")
	cmd.Flags().StringVarP(&deviceType, "type", "t", "", "Filter by device type")

	return cmd
}

func run(generation int, deviceType string) error {
	devices := config.ListDevices()

	if len(devices) == 0 {
		iostreams.Info("No devices registered")
		iostreams.Info("Use 'shelly discover' to find devices or 'shelly device add' to register one")
		return nil
	}

	// Apply filters
	filtered := make(map[string]config.Device)
	for name, dev := range devices {
		if generation > 0 && dev.Generation != generation {
			continue
		}
		if deviceType != "" && dev.Type != deviceType {
			continue
		}
		filtered[name] = dev
	}

	if len(filtered) == 0 {
		iostreams.Info("No devices match the specified filters")
		return nil
	}

	table := output.NewTable("Name", "Address", "Model", "Generation", "Auth")

	for name, dev := range filtered {
		gen := helpers.FormatGeneration(dev.Generation)
		auth := helpers.FormatAuth(dev.Auth != nil)
		table.AddRow(name, dev.Address, dev.Model, gen, auth)
	}

	table.Print()
	iostreams.Count("device", len(filtered))

	return nil
}

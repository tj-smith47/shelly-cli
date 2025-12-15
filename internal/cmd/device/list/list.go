// Package list provides the device list subcommand.
package list

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// DeviceInfo represents device information for JSON/YAML output.
type DeviceInfo struct {
	Name       string `json:"name" yaml:"name"`
	Address    string `json:"address" yaml:"address"`
	Model      string `json:"model" yaml:"model"`
	Type       string `json:"type,omitempty" yaml:"type,omitempty"`
	Generation int    `json:"generation" yaml:"generation"`
	Auth       bool   `json:"auth" yaml:"auth"`
}

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

The registry stores device information including name, address, model,
generation, and authentication credentials. Use filters to narrow results
by device generation (1, 2, or 3) or device type (e.g., SHSW-1, SHRGBW2).

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting and piping to tools like jq.

Columns: Name, Address, Model, Generation, Auth (yes/no)`,
		Example: `  # List all registered devices
  shelly device list

  # List only Gen2 devices
  shelly device list --generation 2

  # List devices by type
  shelly device list --type SHSW-1

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
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(f, generation, deviceType)
		},
	}

	cmd.Flags().IntVarP(&generation, "generation", "g", 0, "Filter by generation (1, 2, or 3)")
	cmd.Flags().StringVarP(&deviceType, "type", "t", "", "Filter by device type")

	return cmd
}

func run(f *cmdutil.Factory, generation int, deviceType string) error {
	ios := f.IOStreams()
	devices := config.ListDevices()

	if len(devices) == 0 {
		ios.Info("No devices registered")
		ios.Info("Use 'shelly discover' to find devices or 'shelly device add' to register one")
		return nil
	}

	// Apply filters and build sorted list
	filtered := make([]DeviceInfo, 0, len(devices))
	for name, dev := range devices {
		if generation > 0 && dev.Generation != generation {
			continue
		}
		if deviceType != "" && dev.Type != deviceType {
			continue
		}
		filtered = append(filtered, DeviceInfo{
			Name:       name,
			Address:    dev.Address,
			Model:      dev.Model,
			Type:       dev.Type,
			Generation: dev.Generation,
			Auth:       dev.Auth != nil,
		})
	}

	// Sort by name for consistent output
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})

	if len(filtered) == 0 {
		ios.Info("No devices match the specified filters")
		return nil
	}

	// Handle structured output (JSON/YAML)
	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, filtered)
	}

	// Table output
	table := output.NewTable("Name", "Address", "Model", "Generation", "Auth")
	for _, dev := range filtered {
		gen := helpers.FormatGeneration(dev.Generation)
		auth := helpers.FormatAuth(dev.Auth)
		table.AddRow(dev.Name, dev.Address, dev.Model, gen, auth)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	ios.Println()
	ios.Count("device", len(filtered))

	return nil
}

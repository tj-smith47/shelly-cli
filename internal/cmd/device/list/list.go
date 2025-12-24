// Package list provides the device list subcommand.
package list

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// DeviceInfo represents device information for JSON/YAML output.
type DeviceInfo struct {
	Name       string `json:"name" yaml:"name"`
	Address    string `json:"address" yaml:"address"`
	Platform   string `json:"platform" yaml:"platform"`
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
		platform   string
	)

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
			return run(f, generation, deviceType, platform)
		},
	}

	cmd.Flags().IntVarP(&generation, "generation", "g", 0, "Filter by generation (1, 2, or 3)")
	cmd.Flags().StringVarP(&deviceType, "type", "t", "", "Filter by device type")
	cmd.Flags().StringVarP(&platform, "platform", "p", "", "Filter by platform (e.g., shelly, tasmota)")

	return cmd
}

func run(f *cmdutil.Factory, generation int, deviceType, platform string) error {
	ios := f.IOStreams()
	devices := config.ListDevices()

	if len(devices) == 0 {
		ios.Info("No devices registered")
		ios.Info("Use 'shelly discover' to find devices or 'shelly device add' to register one")
		return nil
	}

	// Apply filters and build sorted list
	filtered, platforms := filterDevices(devices, generation, deviceType, platform)
	if len(filtered) == 0 {
		ios.Info("No devices match the specified filters")
		return nil
	}

	// Handle structured output (JSON/YAML)
	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, filtered)
	}

	// Show Platform column only when there are multiple platforms
	showPlatform := len(platforms) > 1
	printTable(ios, filtered, showPlatform)

	return nil
}

// filterDevices filters and sorts devices based on criteria.
// Returns filtered devices and set of unique platforms.
func filterDevices(devices map[string]model.Device, generation int, deviceType, platform string) (filtered []DeviceInfo, platforms map[string]struct{}) {
	filtered = make([]DeviceInfo, 0, len(devices))
	platforms = make(map[string]struct{})

	for name, dev := range devices {
		if !matchesFilters(dev, generation, deviceType, platform) {
			continue
		}
		devPlatform := dev.GetPlatform()
		platforms[devPlatform] = struct{}{}
		filtered = append(filtered, DeviceInfo{
			Name:       name,
			Address:    dev.Address,
			Platform:   devPlatform,
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

	return filtered, platforms
}

// matchesFilters checks if a device matches all filter criteria.
func matchesFilters(dev model.Device, generation int, deviceType, platform string) bool {
	if generation > 0 && dev.Generation != generation {
		return false
	}
	if deviceType != "" && dev.Type != deviceType {
		return false
	}
	if platform != "" && dev.GetPlatform() != platform {
		return false
	}
	return true
}

// printTable renders device list as a table.
func printTable(ios *iostreams.IOStreams, devices []DeviceInfo, showPlatform bool) {
	var table *output.Table
	if showPlatform {
		table = output.NewStyledTable(ios, "#", "Name", "Address", "Platform", "Type", "Model", "Generation", "Auth")
	} else {
		table = output.NewStyledTable(ios, "#", "Name", "Address", "Type", "Model", "Generation", "Auth")
	}

	for i, dev := range devices {
		gen := output.RenderGeneration(dev.Generation)
		auth := output.RenderAuthRequired(dev.Auth)
		if showPlatform {
			table.AddRow(fmt.Sprintf("%d", i+1), dev.Name, dev.Address, dev.Platform, dev.Type, dev.Model, gen, auth)
		} else {
			table.AddRow(fmt.Sprintf("%d", i+1), dev.Name, dev.Address, dev.Type, dev.Model, gen, auth)
		}
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	if _, err := fmt.Fprintln(ios.Out); err != nil {
		ios.DebugErr("print newline", err)
	}
}

// Package list provides the device list subcommand.
package list

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DeviceInfo represents device information for JSON/YAML output.
type DeviceInfo struct {
	Name             string `json:"name" yaml:"name"`
	Address          string `json:"address" yaml:"address"`
	Platform         string `json:"platform" yaml:"platform"`
	Model            string `json:"model" yaml:"model"`
	Type             string `json:"type,omitempty" yaml:"type,omitempty"`
	Generation       int    `json:"generation" yaml:"generation"`
	Auth             bool   `json:"auth" yaml:"auth"`
	CurrentVersion   string `json:"current_version,omitempty" yaml:"current_version,omitempty"`
	AvailableVersion string `json:"available_version,omitempty" yaml:"available_version,omitempty"`
	HasUpdate        bool   `json:"has_update,omitempty" yaml:"has_update,omitempty"`
}

// NewCommand creates the device list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		generation   int
		deviceType   string
		platform     string
		updatesFirst bool
		showVersion  bool
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

  # Show firmware versions and sort updates first
  shelly device list --version --updates-first

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
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, generation, deviceType, platform, updatesFirst, showVersion)
		},
	}

	cmd.Flags().IntVarP(&generation, "generation", "g", 0, "Filter by generation (1, 2, or 3)")
	cmd.Flags().StringVarP(&deviceType, "type", "t", "", "Filter by device type")
	cmd.Flags().StringVarP(&platform, "platform", "p", "", "Filter by platform (e.g., shelly, tasmota)")
	cmd.Flags().BoolVarP(&updatesFirst, "updates-first", "u", false, "Sort devices with available updates first")
	cmd.Flags().BoolVarP(&showVersion, "version", "V", false, "Show firmware version information")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, generation int, deviceType, platform string, updatesFirst, showVersion bool) error {
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

	// Populate firmware info if version display or updates-first sorting is requested
	if showVersion || updatesFirst {
		svc := f.ShellyService()
		populateFirmwareInfo(ctx, svc, filtered)
	}

	// Sort: updates first if requested, then by name
	sortDevices(filtered, updatesFirst)

	// Handle structured output (JSON/YAML)
	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, filtered)
	}

	// Show Platform column only when there are multiple platforms
	showPlatform := len(platforms) > 1
	printTable(ios, filtered, showPlatform, showVersion)

	return nil
}

// filterDevices filters devices based on criteria.
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

	return filtered, platforms
}

// populateFirmwareInfo fills in firmware version info from the cache.
// Uses a short cache validity period (5 minutes) so it doesn't trigger network calls during list.
func populateFirmwareInfo(ctx context.Context, svc *shelly.Service, devices []DeviceInfo) {
	const cacheMaxAge = 5 * time.Minute
	for i := range devices {
		entry := svc.GetCachedFirmware(ctx, devices[i].Name, cacheMaxAge)
		if entry != nil && entry.Info != nil {
			devices[i].CurrentVersion = entry.Info.Current
			devices[i].AvailableVersion = entry.Info.Available
			devices[i].HasUpdate = entry.Info.HasUpdate
		}
	}
}

// sortDevices sorts the device list. If updatesFirst is true, devices with
// available updates are sorted to the top. Within each group, devices are sorted by name.
func sortDevices(devices []DeviceInfo, updatesFirst bool) {
	sort.Slice(devices, func(i, j int) bool {
		if updatesFirst {
			// Updates first
			if devices[i].HasUpdate != devices[j].HasUpdate {
				return devices[i].HasUpdate // true sorts before false
			}
		}
		// Then by name
		return devices[i].Name < devices[j].Name
	})
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
func printTable(ios *iostreams.IOStreams, devices []DeviceInfo, showPlatform, showVersion bool) {
	var table *output.Table

	// Build headers based on options
	headers := []string{"#", "Name", "Address"}
	if showPlatform {
		headers = append(headers, "Platform")
	}
	headers = append(headers, "Type", "Model", "Generation")
	if showVersion {
		headers = append(headers, "Version", "Update")
	}
	headers = append(headers, "Auth")

	table = output.NewStyledTable(ios, headers...)

	for i, dev := range devices {
		gen := output.RenderGeneration(dev.Generation)
		auth := output.RenderAuthRequired(dev.Auth)

		// Build row based on options
		row := []string{fmt.Sprintf("%d", i+1), dev.Name, dev.Address}
		if showPlatform {
			row = append(row, dev.Platform)
		}
		row = append(row, dev.Type, dev.Model, gen)
		if showVersion {
			version := dev.CurrentVersion
			if version == "" {
				version = "-"
			}
			update := "-"
			if dev.HasUpdate && dev.AvailableVersion != "" {
				update = theme.StatusOK().Render("↑ " + dev.AvailableVersion)
			}
			row = append(row, version, update)
		}
		row = append(row, auth)

		table.AddRow(row...)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	if _, err := fmt.Fprintln(ios.Out); err != nil {
		ios.DebugErr("print newline", err)
	}
}

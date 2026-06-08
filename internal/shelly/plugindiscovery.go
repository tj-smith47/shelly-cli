// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"
	"net"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// PluginDiscoverer handles device detection through plugins.
type PluginDiscoverer struct {
	registry *plugins.Registry
}

// NewPluginDiscoverer creates a new plugin discoverer.
func NewPluginDiscoverer(registry *plugins.Registry) *PluginDiscoverer {
	return &PluginDiscoverer{registry: registry}
}

// PluginDetectionResult holds the result of a plugin detection along with the plugin info.
type PluginDetectionResult struct {
	Detection *plugins.DeviceDetectionResult
	Plugin    *plugins.Plugin
	Address   string
}

// DetectWithPlugins tries all detection-capable plugins on an address.
// Returns nil, nil if no plugin detected the device (not an error condition).
func (d *PluginDiscoverer) DetectWithPlugins(ctx context.Context, address string, auth *model.Auth) (*PluginDetectionResult, error) {
	capable, err := d.registry.ListDetectionCapable()
	if err != nil {
		return nil, err
	}

	for i := range capable {
		plugin := &capable[i]
		executor := plugins.NewHookExecutor(plugin)
		result, err := executor.ExecuteDetect(ctx, address, auth)
		if err != nil {
			// Detection errors are non-propagating (treated as "not this
			// platform"), but a crashing hook, malformed manifest, or JSON
			// parse failure must leave a diagnostic trail rather than being
			// silently identical to a clean non-detection.
			iostreams.DebugErr(fmt.Sprintf("plugin %s detect %s", plugin.Name, address), err)
			continue
		}
		if result.Detected {
			return &PluginDetectionResult{
				Detection: result,
				Plugin:    plugin,
				Address:   address,
			}, nil
		}
	}

	return nil, nil //nolint:nilnil // No detection is a valid result, not an error
}

// DetectWithPlatform tries only plugins that match the specified platform.
// Returns nil, nil if no matching plugin detected the device (not an error condition).
func (d *PluginDiscoverer) DetectWithPlatform(ctx context.Context, address string, auth *model.Auth, platform string) (*PluginDetectionResult, error) {
	plugin, err := d.registry.FindByPlatform(platform)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, nil //nolint:nilnil // No plugin for this platform is valid
	}

	executor := plugins.NewHookExecutor(plugin)
	result, detectErr := executor.ExecuteDetect(ctx, address, auth)
	if detectErr != nil {
		// Detection failure is expected for addresses that aren't this
		// platform, so it stays non-propagating; but a genuine exec/parse/
		// manifest failure must leave a diagnostic trail.
		iostreams.DebugErr(fmt.Sprintf("plugin %s detect %s", plugin.Name, address), detectErr)
		return nil, nil //nolint:nilnil // Expected when address is not this platform
	}
	if result == nil || !result.Detected {
		// Device not detected by this plugin
		return nil, nil //nolint:nilnil // No detection is a valid result
	}

	return &PluginDetectionResult{
		Detection: result,
		Plugin:    plugin,
		Address:   address,
	}, nil
}

// PluginDiscoveredDevice extends DiscoveredDevice with plugin information.
type PluginDiscoveredDevice struct {
	ID         string
	Name       string
	Model      string
	Address    net.IP
	Platform   string
	Generation int
	Firmware   string
	AuthEn     bool
	Added      bool
	Components []plugins.ComponentInfo
}

// ToPluginDiscoveredDevice converts a plugin detection result to PluginDiscoveredDevice.
func ToPluginDiscoveredDevice(result *PluginDetectionResult, added bool) PluginDiscoveredDevice {
	ip := net.ParseIP(result.Address)
	return PluginDiscoveredDevice{
		ID:         result.Detection.DeviceID,
		Name:       result.Detection.DeviceName,
		Model:      result.Detection.Model,
		Address:    ip,
		Platform:   result.Detection.Platform,
		Generation: 0, // Plugin devices don't have Shelly generations
		Firmware:   result.Detection.Firmware,
		AuthEn:     false, // Not available from detection
		Added:      added,
		Components: result.Detection.Components,
	}
}

// RunPluginDetection scans one or more subnets for plugin-managed devices.
// Returns discovered devices. The isRegistered function checks if an address is already registered.
func RunPluginDetection(ctx context.Context, registry *plugins.Registry, subnets []string, isRegistered func(string) bool) []PluginDiscoveredDevice {
	// Get detection-capable plugins
	capablePlugins, err := registry.ListDetectionCapable()
	if err != nil || len(capablePlugins) == 0 {
		return nil
	}

	addresses := generateSubnetsAddresses(subnets)
	if len(addresses) == 0 {
		return nil
	}

	pluginDiscoverer := NewPluginDiscoverer(registry)
	var pluginDevices []PluginDiscoveredDevice

	for _, addr := range addresses {
		if ctx.Err() != nil {
			break
		}

		result, detectErr := pluginDiscoverer.DetectWithPlugins(ctx, addr, nil)
		if detectErr == nil && result != nil {
			added := isRegistered(result.Address)
			pluginDevices = append(pluginDevices, ToPluginDiscoveredDevice(result, added))
		}
	}

	return pluginDevices
}

// generateSubnetsAddresses generates host addresses across all given subnets
// so plugin discovery covers the same address space as native Shelly discovery
// when multiple --subnet values are supplied.
func generateSubnetsAddresses(subnets []string) []string {
	// Preallocate to the common case of a single /24 (~254 hosts) so multi-subnet
	// scans append into existing capacity instead of growing from nil.
	addresses := make([]string, 0, len(subnets)*254)
	for _, subnet := range subnets {
		// Use the shelly-go SDK helper so plugin discovery enumerates exactly
		// the same host set as native Shelly discovery; a local reimplementation
		// would silently drift on edge masks.
		addresses = append(addresses, discovery.GenerateSubnetAddresses(subnet)...)
	}
	return addresses
}

// DiscoveryProgress represents progress during plugin discovery.
type DiscoveryProgress struct {
	Done    int                     // Number of addresses probed
	Total   int                     // Total addresses to probe
	Found   bool                    // Whether a device was found this iteration
	Device  *PluginDiscoveredDevice // The device found, if any
	Address string                  // Current address being probed
}

// ProgressCallback is called during discovery to report progress.
// Return false to stop discovery early.
type ProgressCallback func(DiscoveryProgress) bool

// RunPluginPlatformDiscoveryWithProgress scans one or more subnets for devices of a specific platform with progress.
func RunPluginPlatformDiscoveryWithProgress(ctx context.Context, registry *plugins.Registry, platform string, subnets []string, isRegistered func(string) bool, onProgress ProgressCallback) []PluginDiscoveredDevice {
	addresses := generateSubnetsAddresses(subnets)
	if len(addresses) == 0 {
		return nil
	}

	pluginDiscoverer := NewPluginDiscoverer(registry)
	var pluginDevices []PluginDiscoveredDevice

	for i, addr := range addresses {
		if ctx.Err() != nil {
			break
		}

		progress := DiscoveryProgress{
			Done:    i + 1,
			Total:   len(addresses),
			Address: addr,
		}

		result, detectErr := pluginDiscoverer.DetectWithPlatform(ctx, addr, nil, platform)
		if detectErr == nil && result != nil {
			added := isRegistered(result.Address)
			device := ToPluginDiscoveredDevice(result, added)
			pluginDevices = append(pluginDevices, device)
			progress.Found = true
			progress.Device = &device
		}

		if onProgress != nil && !onProgress(progress) {
			break
		}
	}

	return pluginDevices
}

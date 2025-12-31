// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"net"

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
			// Plugin didn't detect or errored, try next
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
		// Detection failed - this is expected for addresses that aren't this platform
		return nil, nil //nolint:nilnil,nilerr // Expected when address is not this platform
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

// RunPluginDetection scans a subnet for plugin-managed devices.
// Returns discovered devices. The isRegistered function checks if an address is already registered.
func RunPluginDetection(ctx context.Context, registry *plugins.Registry, subnet string, isRegistered func(string) bool) []PluginDiscoveredDevice {
	// Get detection-capable plugins
	capablePlugins, err := registry.ListDetectionCapable()
	if err != nil || len(capablePlugins) == 0 {
		return nil
	}

	addresses := generateSubnetAddresses(subnet)
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

// RunPluginPlatformDiscovery scans a subnet for devices of a specific platform.
// Returns discovered devices.
func RunPluginPlatformDiscovery(ctx context.Context, registry *plugins.Registry, platform, subnet string, isRegistered func(string) bool) []PluginDiscoveredDevice {
	addresses := generateSubnetAddresses(subnet)
	if len(addresses) == 0 {
		return nil
	}

	pluginDiscoverer := NewPluginDiscoverer(registry)
	var pluginDevices []PluginDiscoveredDevice

	for _, addr := range addresses {
		if ctx.Err() != nil {
			break
		}

		result, detectErr := pluginDiscoverer.DetectWithPlatform(ctx, addr, nil, platform)
		if detectErr == nil && result != nil {
			added := isRegistered(result.Address)
			pluginDevices = append(pluginDevices, ToPluginDiscoveredDevice(result, added))
		}
	}

	return pluginDevices
}

// generateSubnetAddresses generates all host addresses in a CIDR subnet.
func generateSubnetAddresses(subnet string) []string {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil
	}

	var addresses []string
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
		// Skip network and broadcast addresses for /24 and larger
		if isNetworkOrBroadcast(ip, ipNet) {
			continue
		}
		addresses = append(addresses, ip.String())
	}
	return addresses
}

// incIP increments an IP address by 1.
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// isNetworkOrBroadcast checks if an IP is a network or broadcast address.
func isNetworkOrBroadcast(ip net.IP, ipNet *net.IPNet) bool {
	ones, bits := ipNet.Mask.Size()
	if bits-ones < 2 {
		return false // /31 or /32 don't have network/broadcast
	}

	network := ipNet.IP.Mask(ipNet.Mask)
	if ip.Equal(network) {
		return true
	}

	// Calculate broadcast
	broadcast := make(net.IP, len(ip))
	for i := range ip {
		broadcast[i] = ip[i] | ^ipNet.Mask[i]
	}
	return ip.Equal(broadcast)
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

// RunPluginDetectionWithProgress scans a subnet for plugin-managed devices with progress reporting.
func RunPluginDetectionWithProgress(ctx context.Context, registry *plugins.Registry, subnet string, isRegistered func(string) bool, onProgress ProgressCallback) []PluginDiscoveredDevice {
	// Get detection-capable plugins
	capablePlugins, err := registry.ListDetectionCapable()
	if err != nil || len(capablePlugins) == 0 {
		return nil
	}

	addresses := generateSubnetAddresses(subnet)
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

		result, detectErr := pluginDiscoverer.DetectWithPlugins(ctx, addr, nil)
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

// RunPluginPlatformDiscoveryWithProgress scans a subnet for devices of a specific platform with progress.
func RunPluginPlatformDiscoveryWithProgress(ctx context.Context, registry *plugins.Registry, platform, subnet string, isRegistered func(string) bool, onProgress ProgressCallback) []PluginDiscoveredDevice {
	addresses := generateSubnetAddresses(subnet)
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

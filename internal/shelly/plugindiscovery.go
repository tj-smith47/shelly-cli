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

// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// ConfigResolver resolves device identifiers using the config package.
// It auto-detects device generation when not already known.
type ConfigResolver struct{}

// NewConfigResolver creates a new config-based device resolver.
func NewConfigResolver() *ConfigResolver {
	return &ConfigResolver{}
}

// Resolve resolves a device identifier to a model.Device.
func (r *ConfigResolver) Resolve(identifier string) (model.Device, error) {
	// config.ResolveDevice now returns model.Device directly
	return config.ResolveDevice(identifier)
}

// ResolveWithGeneration resolves a device identifier and auto-detects generation if needed.
// If the device generation is 0 (unknown), it probes the device to detect the generation
// and updates the config to cache the result for future calls.
func (r *ConfigResolver) ResolveWithGeneration(ctx context.Context, identifier string) (model.Device, error) {
	device, err := config.ResolveDevice(identifier)
	if err != nil {
		return model.Device{}, err
	}

	// If generation is already known, return as-is
	if device.Generation > 0 {
		return device, nil
	}

	// Auto-detect generation (best-effort, don't fail if detection fails)
	result := tryDetectGeneration(ctx, device.Address, device.Auth)
	if result == nil {
		// Detection failed - return device without generation info
		return device, nil
	}

	// Update device with detected info
	device.Generation = int(result.Generation)
	if device.Model == "" {
		device.Model = result.Model
	}
	if device.Type == "" {
		device.Type = result.DeviceType
	}

	// Cache the detected generation in config if this is a registered device
	if _, exists := config.GetDevice(identifier); exists {
		// Update the device in config with detected generation
		updateDeviceGeneration(identifier, device.Generation, device.Model, device.Type)
	}

	return device, nil
}

// updateDeviceGeneration updates a device's generation info in the config.
func updateDeviceGeneration(name string, generation int, deviceModel, deviceType string) {
	device, ok := config.GetDevice(name)
	if !ok {
		return
	}

	// Only update if values changed
	if device.Generation == generation &&
		(deviceModel == "" || device.Model == deviceModel) &&
		(deviceType == "" || device.Type == deviceType) {
		return
	}

	// Re-register with updated info
	var auth *model.Auth
	if device.Auth != nil {
		auth = device.Auth
	}

	finalModel := device.Model
	if finalModel == "" && deviceModel != "" {
		finalModel = deviceModel
	}

	finalType := device.Type
	if finalType == "" && deviceType != "" {
		finalType = deviceType
	}

	// Update by re-registering (this overwrites the existing entry)
	if err := config.RegisterDevice(name, device.Address, generation, finalType, finalModel, auth); err != nil {
		// Ignore errors - this is best-effort caching
		return
	}
}

// tryDetectGeneration attempts to detect device generation, returning nil on failure.
// This is a best-effort operation - errors are intentionally ignored.
func tryDetectGeneration(ctx context.Context, address string, auth *model.Auth) *client.DetectionResult {
	result, err := client.DetectGeneration(ctx, address, auth)
	if err != nil {
		return nil
	}
	return result
}

// NewService creates a new Shelly service with the default config resolver.
func NewService() *Service {
	return New(NewConfigResolver())
}

// NewServiceWithRateLimiting creates a new Shelly service with rate limiting from app config.
// This reads the rate limit config from the config package and applies it.
func NewServiceWithRateLimiting() *Service {
	cfg := config.Get()
	if cfg == nil {
		return NewService()
	}

	rateLimitCfg := cfg.GetRateLimitConfig()
	return New(NewConfigResolver(), WithRateLimiterFromAppConfig(rateLimitCfg))
}

// NewServiceWithPluginSupport creates a new Shelly service with rate limiting and plugin registry.
// This is the recommended constructor for CLI commands that need plugin support.
func NewServiceWithPluginSupport() *Service {
	cfg := config.Get()
	opts := []ServiceOption{}

	if cfg != nil {
		rateLimitCfg := cfg.GetRateLimitConfig()
		opts = append(opts, WithRateLimiterFromAppConfig(rateLimitCfg))
	}

	// Try to initialize plugin registry - ignore errors (plugins are optional)
	registry, err := plugins.NewRegistry()
	if err == nil {
		opts = append(opts, WithPluginRegistry(registry))
	}

	return New(NewConfigResolver(), opts...)
}

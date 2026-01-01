package shelly

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ProvisionDevice provisions a single device with WiFi configuration.
func (s *Service) ProvisionDevice(ctx context.Context, device model.DeviceProvisionConfig, globalWiFi *model.ProvisionWiFiConfig) error {
	// Get WiFi config (device-specific or global)
	wifi := globalWiFi
	if device.WiFi != nil {
		wifi = device.WiFi
	}

	if wifi == nil {
		return fmt.Errorf("no WiFi configuration")
	}

	// Apply WiFi settings
	enable := true
	if err := s.SetWiFiConfig(ctx, device.Name, wifi.SSID, wifi.Password, &enable); err != nil {
		return fmt.Errorf("failed to set WiFi: %w", err)
	}

	return nil
}

// ProvisionDevices provisions multiple devices in parallel.
// Returns results for each device indicating success or failure.
func (s *Service) ProvisionDevices(ctx context.Context, cfg *model.BulkProvisionConfig, parallel int) []model.ProvisionResult {
	results := make(chan model.ProvisionResult, len(cfg.Devices))
	sem := make(chan struct{}, parallel)

	var wg sync.WaitGroup
	for _, device := range cfg.Devices {
		wg.Go(func() {
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			err := s.ProvisionDevice(ctx, device, cfg.WiFi)
			results <- model.ProvisionResult{Device: device.Name, Err: err}
		})
	}

	// Wait for all to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	out := make([]model.ProvisionResult, 0, len(cfg.Devices))
	for r := range results {
		out = append(out, r)
	}
	return out
}

// ValidateBulkProvisionConfig validates all device names in the config.
// The isDeviceRegistered function checks if a device name is registered.
func ValidateBulkProvisionConfig(cfg *model.BulkProvisionConfig, isDeviceRegistered func(name string) bool) error {
	var errors []string

	for _, d := range cfg.Devices {
		// Validate device name format
		if err := config.ValidateDeviceName(d.Name); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", d.Name, err))
			continue
		}

		// If no address specified, device must be registered
		if d.Address == "" && !isDeviceRegistered(d.Name) {
			errors = append(errors, fmt.Sprintf("%s: not a registered device and no address specified", d.Name))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("invalid device configuration:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}

// ParseBulkProvisionFile reads and parses a bulk provision configuration file.
func ParseBulkProvisionFile(file string) (*model.BulkProvisionConfig, error) {
	data, err := afero.ReadFile(config.Fs(), file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg model.BulkProvisionConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

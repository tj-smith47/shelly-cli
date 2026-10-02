// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DeviceTemplate holds device information and configuration for templates.
type DeviceTemplate struct {
	Model      string         `json:"model"`
	App        string         `json:"app"`
	Generation int            `json:"generation"`
	Config     map[string]any `json:"config"`
}

// CaptureTemplate captures a device's configuration for use as a template.
func (s *Service) CaptureTemplate(ctx context.Context, identifier string, includeWiFi bool) (*DeviceTemplate, error) {
	var result *DeviceTemplate

	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		// Get device info
		info := conn.Info()

		// Get full device config
		cfg, err := conn.GetConfig(ctx)
		if err != nil {
			return fmt.Errorf("failed to get device config: %w", err)
		}

		// Sanitize config if WiFi not included
		if !includeWiFi {
			sanitizeConfig(cfg)
		}

		result = &DeviceTemplate{
			Model:      info.Model,
			App:        info.App,
			Generation: info.Generation,
			Config:     cfg,
		}
		return nil
	})

	return result, err
}

// ApplyTemplate applies a template configuration to a device and returns the
// changes made, plus a warning for each WiFi station left out of the write. A
// template's station address belongs to the device it was captured from and
// is never copied; its network is written only when the device is not already
// on it and this host has the network's password.
func (s *Service) ApplyTemplate(ctx context.Context, identifier string, cfg map[string]any, dryRun bool) (changes, warnings []string, err error) {
	return s.applyTemplate(ctx, identifier, cfg, dryRun, stationsCrossDevice)
}

func (s *Service) applyTemplate(ctx context.Context, identifier string, cfg map[string]any, dryRun bool, mode stationMode) (changes, warnings []string, err error) {
	err = s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		changes, warnings, err = s.applyConfig(ctx, conn, cfg, dryRun, mode)
		return err
	})
	return changes, warnings, err
}

// ImportConfig writes a full config file to a Gen2+ device, or on a dry run
// returns the changes it would make. The file's stations follow the template
// rule unless its sys.device.mac is the device's own, in which case they
// follow the same-device rule; the warnings name each station left out.
func (s *Service) ImportConfig(ctx context.Context, identifier string, cfg map[string]any, dryRun bool) (changes, warnings []string, err error) {
	err = s.withGenAwareAction(ctx, identifier,
		func(_ *client.Gen1Client) error {
			return fmt.Errorf("bulk config updates are not supported on Gen1 devices; use component-specific commands instead")
		},
		func(conn *client.Client) error {
			changes, warnings, err = s.applyConfig(ctx, conn, cfg, dryRun, stationsMatchMAC)
			return err
		})
	return changes, warnings, err
}

// applyConfig plans cfg's stations for mode and writes cfg with
// Shelly.SetConfig, or on a dry run returns the top-level changes and one line
// per planned station.
func (s *Service) applyConfig(ctx context.Context, conn *client.Client, cfg map[string]any, dryRun bool, mode stationMode) (changes, warnings []string, err error) {
	var current map[string]any
	if dryRun || (mode != stationsOmit && configHasStation(cfg)) {
		if current, err = conn.GetConfig(ctx); err != nil {
			return nil, nil, fmt.Errorf("failed to get current config: %w", err)
		}
	}
	planned, warnings := s.planConfigStations(ctx, cfg, current, mode)
	if dryRun {
		return append(compareForApply(current, planned), describeStations(planned)...), warnings, nil
	}
	if err := conn.SetConfig(ctx, planned); err != nil {
		return nil, warnings, fmt.Errorf("failed to apply config: %w", err)
	}
	return []string{"Configuration applied successfully"}, warnings, nil
}

// CompareTemplate compares a template configuration with a device's current config.
func (s *Service) CompareTemplate(ctx context.Context, identifier string, templateCfg map[string]any) ([]model.ConfigDiff, error) {
	var diffs []model.ConfigDiff

	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		current, err := conn.GetConfig(ctx)
		if err != nil {
			return fmt.Errorf("failed to get device config: %w", err)
		}

		diffs = backup.CompareConfigs(current, templateCfg)
		return nil
	})

	return diffs, err
}

// GetDeviceInfo returns basic device information.
func (s *Service) GetDeviceInfo(ctx context.Context, identifier string) (*DeviceTemplate, error) {
	var result *DeviceTemplate

	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		info := conn.Info()
		result = &DeviceTemplate{
			Model:      info.Model,
			App:        info.App,
			Generation: info.Generation,
		}
		return nil
	})

	return result, err
}

// sanitizeConfig removes sensitive data from config.
func sanitizeConfig(cfg map[string]any) {
	// Remove WiFi credentials
	if wifi, ok := cfg[componentWiFi].(map[string]any); ok {
		if sta, ok := wifi[fieldSTA].(map[string]any); ok {
			delete(sta, "pass")
		}
		if sta1, ok := wifi[fieldSTA1].(map[string]any); ok {
			delete(sta1, "pass")
		}
		if ap, ok := wifi["ap"].(map[string]any); ok {
			delete(ap, "pass")
		}
	}

	// Remove auth credentials
	if auth, ok := cfg["auth"].(map[string]any); ok {
		delete(auth, "pass")
	}

	// Remove cloud credentials
	if cloud, ok := cfg["cloud"].(map[string]any); ok {
		delete(cloud, fieldServer)
	}
}

// compareForApply returns a list of what would change when applying config.
func compareForApply(current, template map[string]any) []string {
	var changes []string

	for key, templateVal := range template {
		currentVal, exists := current[key]
		if !exists {
			changes = append(changes, fmt.Sprintf("+ %s: %v", key, summarizeValue(templateVal)))
			continue
		}

		if !utils.DeepEqualJSON(currentVal, templateVal) {
			changes = append(changes, fmt.Sprintf("~ %s: %v -> %v", key, summarizeValue(currentVal), summarizeValue(templateVal)))
		}
	}

	return changes
}

// summarizeValue returns a short string representation of a value.
func summarizeValue(v any) string {
	switch val := v.(type) {
	case map[string]any:
		return fmt.Sprintf("{...%d keys}", len(val))
	case []any:
		return fmt.Sprintf("[...%d items]", len(val))
	case string:
		if len(val) > 20 {
			return fmt.Sprintf("%q...", val[:17])
		}
		return fmt.Sprintf("%q", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

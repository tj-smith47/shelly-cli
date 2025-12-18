// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
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

// ApplyTemplate applies a template configuration to a device.
// Returns a list of changes made.
func (s *Service) ApplyTemplate(ctx context.Context, identifier string, cfg map[string]any, dryRun bool) ([]string, error) {
	var changes []string

	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		if dryRun {
			// In dry run mode, just compare and report what would change
			current, err := conn.GetConfig(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current config: %w", err)
			}
			changes = compareForApply(current, cfg)
			return nil
		}

		// Apply the configuration
		if err := conn.SetConfig(ctx, cfg); err != nil {
			return fmt.Errorf("failed to apply config: %w", err)
		}

		changes = []string{"Configuration applied successfully"}
		return nil
	})

	return changes, err
}

// CompareTemplate compares a template configuration with a device's current config.
func (s *Service) CompareTemplate(ctx context.Context, identifier string, templateCfg map[string]any) ([]model.ConfigDiff, error) {
	var diffs []model.ConfigDiff

	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		current, err := conn.GetConfig(ctx)
		if err != nil {
			return fmt.Errorf("failed to get device config: %w", err)
		}

		diffs = compareConfigs(current, templateCfg)
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
	if wifi, ok := cfg["wifi"].(map[string]any); ok {
		if sta, ok := wifi["sta"].(map[string]any); ok {
			delete(sta, "pass")
		}
		if sta1, ok := wifi["sta1"].(map[string]any); ok {
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
		delete(cloud, "server")
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

		if !deepEqualJSON(currentVal, templateVal) {
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

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidateTemplateName checks if a template name is valid.
func ValidateTemplateName(name string) error {
	return ValidateName(name, "template")
}

// ParseDeviceTemplateFile parses a device template from file data, detecting format by extension.
func ParseDeviceTemplateFile(filename string, data []byte) (DeviceTemplate, error) {
	var tpl DeviceTemplate

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &tpl); err != nil {
			return tpl, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &tpl); err != nil {
			return tpl, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		// Try YAML first, then JSON
		yamlErr := yaml.Unmarshal(data, &tpl)
		if yamlErr == nil {
			break
		}
		jsonErr := json.Unmarshal(data, &tpl)
		if jsonErr != nil {
			return tpl, fmt.Errorf("failed to parse file: %w", errors.Join(yamlErr, jsonErr))
		}
	}

	// Validate required fields
	if tpl.Name == "" {
		return tpl, fmt.Errorf("template missing required field: name")
	}
	if tpl.Model == "" {
		return tpl, fmt.Errorf("template missing required field: model")
	}
	if tpl.Config == nil {
		return tpl, fmt.Errorf("template missing required field: config")
	}

	return tpl, nil
}

// IsCompatibleModel checks if a device template is compatible with a device model.
func IsCompatibleModel(template DeviceTemplate, model string) bool {
	return template.Model == model
}

// IsCompatibleGeneration checks if a device template is compatible with a device generation.
func IsCompatibleGeneration(template DeviceTemplate, generation int) bool {
	return template.Generation == generation
}

// Package-level functions delegate to the default manager.

// CreateDeviceTemplate creates a new device template.
func CreateDeviceTemplate(name, description, deviceModel, app string, generation int, cfg map[string]any, sourceDevice string) error {
	return getDefaultManager().CreateDeviceTemplate(name, description, deviceModel, app, generation, cfg, sourceDevice)
}

// DeleteDeviceTemplate removes a device template.
func DeleteDeviceTemplate(name string) error {
	return getDefaultManager().DeleteDeviceTemplate(name)
}

// GetDeviceTemplate returns a device template by name.
func GetDeviceTemplate(name string) (DeviceTemplate, bool) {
	return getDefaultManager().GetDeviceTemplate(name)
}

// ListDeviceTemplates returns all device templates.
func ListDeviceTemplates() map[string]DeviceTemplate {
	return getDefaultManager().ListDeviceTemplates()
}

// UpdateDeviceTemplate updates a device template's metadata.
func UpdateDeviceTemplate(name, description string) error {
	return getDefaultManager().UpdateDeviceTemplate(name, description)
}

// SaveDeviceTemplate saves or updates a device template.
func SaveDeviceTemplate(template DeviceTemplate) error {
	return getDefaultManager().SaveDeviceTemplate(template)
}

// =============================================================================
// Script Template Functions
// =============================================================================

// ParseScriptTemplateFile parses a script template from file data, detecting format by extension.
func ParseScriptTemplateFile(filename string, data []byte) (ScriptTemplate, error) {
	var tpl ScriptTemplate

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &tpl); err != nil {
			return tpl, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &tpl); err != nil {
			return tpl, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		// Try YAML first, then JSON
		yamlErr := yaml.Unmarshal(data, &tpl)
		if yamlErr == nil {
			break
		}
		jsonErr := json.Unmarshal(data, &tpl)
		if jsonErr != nil {
			return tpl, fmt.Errorf("failed to parse file: %w", errors.Join(yamlErr, jsonErr))
		}
	}

	// Validate required fields
	if tpl.Name == "" {
		return tpl, fmt.Errorf("script template missing required field: name")
	}
	if tpl.Code == "" {
		return tpl, fmt.Errorf("script template missing required field: code")
	}

	return tpl, nil
}

// SaveScriptTemplate saves or updates a script template.
func SaveScriptTemplate(template ScriptTemplate) error {
	return getDefaultManager().SaveScriptTemplate(template)
}

// DeleteScriptTemplate removes a script template.
func DeleteScriptTemplate(name string) error {
	return getDefaultManager().DeleteScriptTemplate(name)
}

// GetScriptTemplate returns a script template by name.
func GetScriptTemplate(name string) (ScriptTemplate, bool) {
	return getDefaultManager().GetScriptTemplate(name)
}

// ListScriptTemplates returns all user-defined script templates.
func ListScriptTemplates() map[string]ScriptTemplate {
	return getDefaultManager().ListScriptTemplates()
}

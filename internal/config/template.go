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

// ParseTemplateFile parses a template from file data, detecting format by extension.
func ParseTemplateFile(filename string, data []byte) (Template, error) {
	var tpl Template

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

// IsCompatibleModel checks if a template is compatible with a device model.
func IsCompatibleModel(template Template, model string) bool {
	return template.Model == model
}

// IsCompatibleGeneration checks if a template is compatible with a device generation.
func IsCompatibleGeneration(template Template, generation int) bool {
	return template.Generation == generation
}

// Package-level functions delegate to the default manager.

// CreateTemplate creates a new template.
func CreateTemplate(name, description, deviceModel, app string, generation int, cfg map[string]any, sourceDevice string) error {
	return getDefaultManager().CreateTemplate(name, description, deviceModel, app, generation, cfg, sourceDevice)
}

// DeleteTemplate removes a template.
func DeleteTemplate(name string) error {
	return getDefaultManager().DeleteTemplate(name)
}

// GetTemplate returns a template by name.
func GetTemplate(name string) (Template, bool) {
	return getDefaultManager().GetTemplate(name)
}

// ListTemplates returns all templates.
func ListTemplates() map[string]Template {
	return getDefaultManager().ListTemplates()
}

// UpdateTemplate updates a template's metadata.
func UpdateTemplate(name, description string) error {
	return getDefaultManager().UpdateTemplate(name, description)
}

// SaveTemplate saves or updates a template.
func SaveTemplate(template Template) error {
	return getDefaultManager().SaveTemplate(template)
}

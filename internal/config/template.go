package config

import (
	"fmt"
	"regexp"
)

// templateNameRegex validates template names (alphanumeric, hyphens, underscores).
var templateNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// ValidateTemplateName checks if a template name is valid.
func ValidateTemplateName(name string) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("template name too long (max 64 characters)")
	}
	if !templateNameRegex.MatchString(name) {
		return fmt.Errorf("template name contains invalid characters (use alphanumeric, hyphens, underscores)")
	}
	return nil
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

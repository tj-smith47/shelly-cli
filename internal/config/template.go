package config

import (
	"fmt"
	"regexp"
	"time"
)

// templateNameRegex validates template names (alphanumeric, hyphens, underscores).
var templateNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// CreateTemplate creates a new template.
func CreateTemplate(name, description, model, app string, generation int, config map[string]any, sourceDevice string) error {
	if err := ValidateTemplateName(name); err != nil {
		return err
	}

	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	if _, exists := c.Templates[name]; exists {
		return fmt.Errorf("template %q already exists", name)
	}

	c.Templates[name] = Template{
		Name:         name,
		Description:  description,
		Model:        model,
		App:          app,
		Generation:   generation,
		Config:       config,
		CreatedAt:    time.Now().Format(time.RFC3339),
		SourceDevice: sourceDevice,
	}

	return Save()
}

// DeleteTemplate removes a template.
func DeleteTemplate(name string) error {
	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	if _, exists := c.Templates[name]; !exists {
		return fmt.Errorf("template %q not found", name)
	}

	delete(c.Templates, name)
	return Save()
}

// GetTemplate returns a template by name.
func GetTemplate(name string) (Template, bool) {
	c := Get()

	cfgMu.RLock()
	defer cfgMu.RUnlock()

	template, exists := c.Templates[name]
	return template, exists
}

// ListTemplates returns all templates.
func ListTemplates() map[string]Template {
	c := Get()

	cfgMu.RLock()
	defer cfgMu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]Template, len(c.Templates))
	for k, v := range c.Templates {
		result[k] = v
	}
	return result
}

// UpdateTemplate updates a template's metadata.
func UpdateTemplate(name, description string) error {
	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	template, exists := c.Templates[name]
	if !exists {
		return fmt.Errorf("template %q not found", name)
	}

	if description != "" {
		template.Description = description
	}

	c.Templates[name] = template
	return Save()
}

// SaveTemplate saves or updates a template.
func SaveTemplate(template Template) error {
	if err := ValidateTemplateName(template.Name); err != nil {
		return err
	}

	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	c.Templates[template.Name] = template
	return Save()
}

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
// For now, we check exact model match for strict compatibility.
func IsCompatibleModel(template Template, model string) bool {
	return template.Model == model
}

// IsCompatibleGeneration checks if a template is compatible with a device generation.
func IsCompatibleGeneration(template Template, generation int) bool {
	return template.Generation == generation
}

// Package config provides configuration management for the Shelly CLI.
package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// GetSetting retrieves a CLI configuration value by key.
// Returns the value and whether it was set.
func GetSetting(key string) (any, bool) {
	if !viper.IsSet(key) {
		return nil, false
	}
	return viper.Get(key), true
}

// GetAllSettings returns all CLI configuration values.
func GetAllSettings() map[string]any {
	return viper.AllSettings()
}

// SetSetting sets a CLI configuration value.
func SetSetting(key string, value any) error {
	viper.Set(key, value)
	return viper.WriteConfig()
}

// DeleteSetting removes a CLI configuration value by key.
// The key is removed from the config file, not just set to nil.
// Returns an error if the key does not exist.
func DeleteSetting(key string) error {
	if !viper.IsSet(key) {
		return fmt.Errorf("key %q is not set", key)
	}

	// Get the config file path
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return fmt.Errorf("no config file found")
	}

	// Read the config file as a map
	fs := Fs()
	data, err := afero.ReadFile(fs, configFile)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var configMap map[string]any
	if err := yaml.Unmarshal(data, &configMap); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	// Delete the key from the map (handling dot notation)
	if !deleteNestedKey(configMap, key) {
		return fmt.Errorf("key %q not found in config file", key)
	}

	// Write the updated config back
	updatedData, err := yaml.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := afero.WriteFile(fs, configFile, updatedData, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	// Clear the key from viper's in-memory state
	viper.Set(key, nil)

	return nil
}

// deleteNestedKey removes a key from a nested map using dot notation.
// Returns true if the key was found and deleted.
func deleteNestedKey(m map[string]any, key string) bool {
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		if _, exists := m[key]; exists {
			delete(m, key)
			return true
		}
		return false
	}

	// Navigate to the parent map
	current := m
	for i := range len(parts) - 1 {
		next, ok := current[parts[i]]
		if !ok {
			return false
		}
		nextMap, ok := next.(map[string]any)
		if !ok {
			return false
		}
		current = nextMap
	}

	// Delete the final key
	finalKey := parts[len(parts)-1]
	if _, exists := current[finalKey]; exists {
		delete(current, finalKey)
		return true
	}
	return false
}

// IsParentSetting checks if a setting key has nested child values.
// Returns the list of child keys if any exist.
func IsParentSetting(key string) ([]string, bool) {
	prefix := key + "."
	var children []string
	for _, k := range viper.AllKeys() {
		if strings.HasPrefix(k, prefix) {
			children = append(children, k)
		}
	}
	return children, len(children) > 0
}

// ResetSettings resets CLI configuration to defaults.
func ResetSettings() error {
	// Clear all settings and write empty config
	for _, key := range viper.AllKeys() {
		viper.Set(key, nil)
	}
	return viper.WriteConfig()
}

// KnownSettingKeys returns the list of known CLI setting keys for completion.
func KnownSettingKeys() []string {
	return []string{
		"defaults.timeout",
		"defaults.output",
		"defaults.concurrent",
		"editor",
		"theme.name",
		"theme.colors",
		"theme.file",
		"log.json",
		"log.categories",
	}
}

// GetEditor returns the configured editor command.
// Returns empty string if not configured (caller should fall back to env vars).
func GetEditor() string {
	if cfg := Get(); cfg != nil && cfg.Editor != "" {
		return cfg.Editor
	}
	// Fallback to viper setting if not in Config struct
	if val, ok := GetSetting("editor"); ok {
		if cmd, isStr := val.(string); isStr && cmd != "" {
			return cmd
		}
	}
	return ""
}

// FilterSettingKeys filters known keys by prefix for completion.
func FilterSettingKeys(prefix string) []string {
	var result []string
	for _, k := range KnownSettingKeys() {
		if strings.HasPrefix(k, prefix) {
			result = append(result, k)
		}
	}
	return result
}

// FormatSettingValue formats a setting value for display.
func FormatSettingValue(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return "(not set)"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// SaveTheme saves the theme name to configuration file.
// This uses config.Fs() for proper test isolation.
func SaveTheme(themeName string) error {
	viper.Set("theme", themeName)

	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// Create default config path
		configDir, err := Dir()
		if err != nil {
			return err
		}
		fs := Fs()
		if err := fs.MkdirAll(configDir, 0o700); err != nil {
			return err
		}
		configFile = filepath.Join(configDir, "config.yaml")
	}

	// In test mode, write to the test filesystem
	if IsTestFs() {
		data := []byte("theme: " + themeName + "\n")
		return afero.WriteFile(Fs(), configFile, data, 0o600)
	}

	return viper.WriteConfigAs(configFile)
}

// Package config provides configuration management for the Shelly CLI.
package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
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

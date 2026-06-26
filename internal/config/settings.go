// Package config provides configuration management for the Shelly CLI.
package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Known CLI setting keys referenced by completion and tests.
const (
	settingKeyOutput              = "output"
	settingKeyEditor              = "editor"
	settingKeyTelemetry           = "telemetry"
	settingKeyDiscoveryTimeout    = "discovery.timeout"
	settingKeyRateLimitConcurrent = "ratelimit.global.max_concurrent"
	settingKeyThemeName           = "theme.name"
)

// durationType is the reflect.Type of time.Duration, which shares its Kind
// (Int64) with plain integers and must be matched by type, not Kind.
var durationType = reflect.TypeFor[time.Duration]()

// settingType resolves the Go type of the Config field addressed by a dotted
// setting key, walking nested structs by their mapstructure (then yaml) tags.
// Returns ok=false for keys that do not map to a known struct field.
func settingType(key string) (reflect.Type, bool) {
	t := reflect.TypeFor[Config]()
	for seg := range strings.SplitSeq(key, ".") {
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return nil, false
		}
		f, ok := fieldByConfigTag(t, seg)
		if !ok {
			return nil, false
		}
		t = f.Type
	}
	return t, true
}

// fieldByConfigTag finds the struct field whose mapstructure (or yaml) tag name
// matches the given key segment.
func fieldByConfigTag(t reflect.Type, name string) (reflect.StructField, bool) {
	for f := range t.Fields() {
		tag, _, _ := strings.Cut(f.Tag.Get("mapstructure"), ",")
		if tag == "" {
			tag, _, _ = strings.Cut(f.Tag.Get("yaml"), ",")
		}
		if tag == "" {
			tag = strings.ToLower(f.Name)
		}
		if tag == "-" {
			continue
		}
		if tag == name {
			return f, true
		}
	}
	return reflect.StructField{}, false
}

// CoerceSettingValue converts a raw string CLI value to the type expected by the
// Config field addressed by key (dot notation), so typed settings round-trip
// correctly through the YAML config file. For example "telemetry=true" is stored
// as a real bool — storing the string "true" would otherwise fail to unmarshal
// into the Telemetry bool field and break config loading. Keys that do not map to
// a known typed field are returned as a string unchanged.
func CoerceSettingValue(key, raw string) (any, error) {
	typ, ok := settingType(key)
	raw = unquote(raw)
	if !ok {
		return raw, nil
	}
	switch {
	case typ == durationType:
		d, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid duration %q for %s (e.g. 30s, 5m)", raw, key)
		}
		return d, nil
	case typ.Kind() == reflect.Bool:
		b, ok := parseBoolLenient(raw)
		if !ok {
			return nil, fmt.Errorf("invalid boolean %q for %s (use true/false)", raw, key)
		}
		return b, nil
	case isSignedInt(typ.Kind()):
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q for %s", raw, key)
		}
		return n, nil
	case isUnsignedInt(typ.Kind()):
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q for %s", raw, key)
		}
		return n, nil
	case typ.Kind() == reflect.Float32 || typ.Kind() == reflect.Float64:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q for %s", raw, key)
		}
		return f, nil
	default:
		return raw, nil
	}
}

func isSignedInt(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func isUnsignedInt(k reflect.Kind) bool {
	switch k {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

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
	return writeViperConfig()
}

// writeViperConfig persists viper's in-memory settings to disk. On a fresh
// install no config file exists yet, which makes viper.WriteConfig fail with
// "Config File Not Found"; this resolves the default path, creates the
// directory, and writes a new file so the very first `shelly config set` (and
// `config reset`) succeeds instead of erroring.
func writeViperConfig() error {
	err := viper.WriteConfig()
	if err == nil {
		return nil
	}
	var notFound viper.ConfigFileNotFoundError
	if !errors.As(err, &notFound) {
		return err
	}

	path := viper.ConfigFileUsed()
	if path == "" {
		dir, dirErr := Dir()
		if dirErr != nil {
			return dirErr
		}
		if mkErr := Fs().MkdirAll(dir, 0o700); mkErr != nil {
			return fmt.Errorf("create config dir: %w", mkErr)
		}
		path = filepath.Join(dir, "config.yaml")
	}

	// In test mode viper still uses the OS filesystem, so write through the
	// package filesystem (an in-memory FS under test) to keep isolation.
	if IsTestFs() {
		data, marshalErr := yaml.Marshal(viper.AllSettings())
		if marshalErr != nil {
			return fmt.Errorf("marshal config: %w", marshalErr)
		}
		return afero.WriteFile(Fs(), path, data, 0o600)
	}
	return viper.WriteConfigAs(path)
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
	return writeViperConfig()
}

// KnownSettingKeys returns the list of known CLI setting keys for completion.
func KnownSettingKeys() []string {
	return []string{
		settingKeyOutput,
		settingKeyEditor,
		settingKeyTelemetry,
		"color",
		"verbose",
		"quiet",
		settingKeyDiscoveryTimeout,
		"discovery.mdns",
		"discovery.network",
		settingKeyRateLimitConcurrent,
		settingKeyThemeName,
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
	if val, ok := GetSetting(settingKeyEditor); ok {
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
			return valTrue
		}
		return valFalse
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

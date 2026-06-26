package config

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func TestKnownSettingKeys(t *testing.T) {
	t.Parallel()

	keys := KnownSettingKeys()
	if len(keys) == 0 {
		t.Error("KnownSettingKeys() returned empty slice")
	}

	// Check for expected keys
	expectedKeys := []string{
		"discovery.timeout",
		"output",
		"editor",
		"theme.name",
	}

	for _, expected := range expectedKeys {
		found := false
		for _, k := range keys {
			if k == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected key %q not found in KnownSettingKeys()", expected)
		}
	}
}

func TestFilterSettingKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prefix   string
		wantMin  int // minimum expected matches
		contains string
	}{
		{"discovery prefix", "discovery.", 2, "discovery.timeout"},
		{"theme prefix", "theme.", 2, "theme.name"},
		{"log prefix", "log.", 1, "log.json"},
		{"no match", "nonexistent.", 0, ""},
		{"empty prefix", "", 5, ""}, // should return all
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := FilterSettingKeys(tt.prefix)
			if len(result) < tt.wantMin {
				t.Errorf("FilterSettingKeys(%q) returned %d items, want at least %d", tt.prefix, len(result), tt.wantMin)
			}
			if tt.contains != "" {
				found := false
				for _, k := range result {
					if k == tt.contains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("FilterSettingKeys(%q) should contain %q", tt.prefix, tt.contains)
				}
			}
		})
	}
}

func TestFormatSettingValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"string value", "hello", "hello"},
		{"empty string", "", ""},
		{"true bool", true, "true"},
		{"false bool", false, "false"},
		{"nil value", nil, "(not set)"},
		{"int value", 42, "42"},
		{"float value", 3.14, "3.14"},
		{"slice value", []string{"a", "b"}, "[a b]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatSettingValue(tt.input)
			if got != tt.want {
				t.Errorf("FormatSettingValue(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestGetSetting(t *testing.T) {
	setupViperTest(t)

	// Test unset key
	val, ok := GetSetting("nonexistent.key")
	if ok {
		t.Error("GetSetting() should return false for unset key")
	}
	if val != nil {
		t.Errorf("GetSetting() value should be nil for unset key, got %v", val)
	}

	// Set a value and test
	viper.Set("test.key", "test-value")
	val, ok = GetSetting("test.key")
	if !ok {
		t.Error("GetSetting() should return true for set key")
	}
	if val != "test-value" {
		t.Errorf("GetSetting() = %v, want 'test-value'", val)
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestGetAllSettings(t *testing.T) {
	setupViperTest(t)

	// Set some values
	viper.Set("setting1", "value1")
	viper.Set("setting2", 42)

	all := GetAllSettings()
	if all == nil {
		t.Fatal("GetAllSettings() returned nil")
	}

	// Check that our settings are in the result
	if val, ok := all["setting1"]; !ok || val != "value1" {
		t.Errorf("GetAllSettings() missing setting1 or wrong value, got %v", val)
	}
	if val, ok := all["setting2"]; !ok || val != 42 {
		t.Errorf("GetAllSettings() missing setting2 or wrong value, got %v", val)
	}
}

func TestSetSetting_FreshInstallCreatesFile(t *testing.T) {
	setupPackageTestSettings(t)
	t.Setenv("XDG_CONFIG_HOME", "/testconfig")

	// On a fresh install no config file exists yet. SetSetting must create one
	// rather than failing with "Config File Not Found" (the issue #3 bug).
	if err := SetSetting(settingKeyTelemetry, true); err != nil {
		t.Fatalf("SetSetting on fresh install should succeed, got: %v", err)
	}

	if v := viper.Get(settingKeyTelemetry); v != true {
		t.Errorf("telemetry not set in memory, got %v", v)
	}

	// The value must persist to disk as a real bool, not a quoted string.
	data, err := afero.ReadFile(Fs(), "/testconfig/shelly/config.yaml")
	if err != nil {
		t.Fatalf("config file was not created: %v", err)
	}
	if got := string(data); !strings.Contains(got, "telemetry: true") {
		t.Errorf("config file should contain bare bool, got:\n%s", got)
	}
}

func TestResetSettings(t *testing.T) {
	setupPackageTestSettings(t)
	t.Setenv("XDG_CONFIG_HOME", "/testconfig")

	viper.Set("reset.test1", "value1")
	viper.Set("reset.test2", "value2")

	// ResetSettings must succeed even when no config file exists yet.
	if err := ResetSettings(); err != nil {
		t.Fatalf("ResetSettings should succeed on fresh install, got: %v", err)
	}

	if viper.IsSet("reset.test1") {
		t.Error("ResetSettings() did not clear reset.test1")
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestGetEditor(t *testing.T) {
	setupPackageTestSettings(t)

	// Test with config editor set
	cfg := Get()
	if cfg.Editor != "" {
		originalEditor := cfg.Editor
		t.Cleanup(func() { cfg.Editor = originalEditor })
	}

	// Initially should be empty if not configured
	// Editor might be set or not depending on config, just verify it doesn't panic
	_ = GetEditor()

	// Set via viper and test fallback
	viper.Set(settingKeyEditor, "nano")
	editor := GetEditor()
	// Should return the editor from config or viper
	if editor == "" {
		// Might still be empty if config overrides
		t.Log("GetEditor() returned empty string")
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestGetEditor_FromConfig(t *testing.T) {
	ResetDefaultManagerForTesting()
	t.Cleanup(func() { ResetDefaultManagerForTesting() })

	// Set up a config with editor
	testCfg := &Config{
		Editor: "vim",
	}
	testMgr := NewTestManager(testCfg)
	SetDefaultManager(testMgr)

	editor := GetEditor()
	if editor != "vim" {
		t.Errorf("GetEditor() = %q, want %q", editor, "vim")
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestGetEditor_FromViperSetting(t *testing.T) {
	ResetDefaultManagerForTesting()
	viper.Reset()
	t.Cleanup(func() {
		ResetDefaultManagerForTesting()
		viper.Reset()
	})

	// Set up a config WITHOUT editor
	testCfg := &Config{}
	testMgr := NewTestManager(testCfg)
	SetDefaultManager(testMgr)

	// Set via viper
	viper.Set(settingKeyEditor, "nano")

	editor := GetEditor()
	if editor != "nano" {
		t.Errorf("GetEditor() = %q, want %q", editor, "nano")
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestGetEditor_NonStringViperValue(t *testing.T) {
	ResetDefaultManagerForTesting()
	viper.Reset()
	t.Cleanup(func() {
		ResetDefaultManagerForTesting()
		viper.Reset()
	})

	// Set up a config WITHOUT editor
	testCfg := &Config{}
	testMgr := NewTestManager(testCfg)
	SetDefaultManager(testMgr)

	// Set a non-string value via viper
	viper.Set(settingKeyEditor, 123)

	// Should return empty since value is not a string
	editor := GetEditor()
	if editor != "" {
		t.Errorf("GetEditor() = %q, want empty for non-string viper value", editor)
	}
}

func TestCoerceSettingValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		raw     string
		want    any
		wantErr bool
	}{
		// Typed struct fields must coerce to their real Go type so they
		// round-trip through the YAML config instead of being stored as the
		// strings that broke config loading (the issue #3 telemetry bug).
		{"bool true", settingKeyTelemetry, "true", true, false},
		{"bool lenient on", settingKeyTelemetry, "on", true, false},
		{"bool lenient no", settingKeyTelemetry, "no", false, false},
		{"bool invalid", settingKeyTelemetry, "maybe", nil, true},
		{"top-level string", settingKeyOutput, "json", "json", false},
		{"nested duration", settingKeyDiscoveryTimeout, "30s", 30 * time.Second, false},
		{"nested duration invalid", settingKeyDiscoveryTimeout, "xyz", nil, true},
		{"nested int", settingKeyRateLimitConcurrent, "12", int64(12), false},
		{"nested int invalid", settingKeyRateLimitConcurrent, "lots", nil, true},
		{"nested bool", "discovery.mdns", "yes", true, false},
		// Unknown keys are not struct-backed: keep them as free-form strings.
		{"unknown key stays string", settingKeyEditor, "nano", "nano", false},
		{"unknown dotted key stays string", "some.unknown.key", "0", "0", false},
		// Surrounding quotes are stripped before coercion.
		{"quoted bool", settingKeyTelemetry, `"true"`, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := CoerceSettingValue(tt.key, tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CoerceSettingValue(%q, %q) err = %v, wantErr %v", tt.key, tt.raw, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("CoerceSettingValue(%q, %q) = %#v (%T), want %#v (%T)", tt.key, tt.raw, got, got, tt.want, tt.want)
			}
		})
	}
}

// setupViperTest resets viper state for isolated testing.
func setupViperTest(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(func() { viper.Reset() })
}

// setupPackageTestSettings sets up an isolated environment for package-level settings tests.
func setupPackageTestSettings(t *testing.T) {
	t.Helper()
	setupViperTest(t)
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })
	ResetDefaultManagerForTesting()
}

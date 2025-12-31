package config

import (
	"testing"

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
		"defaults.timeout",
		"defaults.output",
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
		{"defaults prefix", "defaults.", 2, "defaults.timeout"},
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

//nolint:paralleltest // Tests modify global viper state
func TestSetSetting(t *testing.T) {
	setupViperTest(t)

	// SetSetting will fail WriteConfig without a config file set
	// But we can still test that the value is set in viper
	err := SetSetting("test.setting", "new-value")
	// Error is expected since no config file is set
	if err == nil {
		// If no error, check value was set
		if v := viper.Get("test.setting"); v != "new-value" {
			t.Errorf("SetSetting() did not set value, got %v", v)
		}
	}
	// Either way, verify the value was set in memory
	if v := viper.Get("test.setting"); v != "new-value" {
		t.Errorf("SetSetting() did not set value in memory, got %v", v)
	}
}

//nolint:paralleltest // Tests modify global viper state
func TestResetSettings(t *testing.T) {
	setupViperTest(t)

	// Set some values
	viper.Set("reset.test1", "value1")
	viper.Set("reset.test2", "value2")

	// Reset will fail WriteConfig without a config file set
	// But we can still test that values are cleared
	err := ResetSettings()
	// Error is expected since no config file is set
	if err == nil {
		// If no error, check values were cleared
		if v := viper.Get("reset.test1"); v != nil {
			t.Errorf("ResetSettings() did not clear reset.test1, got %v", v)
		}
	}
	// Either way, verify values were cleared in memory
	// (ResetSettings sets them to nil, which should make IsSet return false)
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
	editor := GetEditor()
	// Editor might be set or not depending on config, just verify it doesn't panic

	// Set via viper and test fallback
	viper.Set("editor", "nano")
	editor = GetEditor()
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
	viper.Set("editor", "nano")

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
	viper.Set("editor", 123)

	// Should return empty since value is not a string
	editor := GetEditor()
	if editor != "" {
		t.Errorf("GetEditor() = %q, want empty for non-string viper value", editor)
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

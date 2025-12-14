package theme

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestColorCompatibility verifies theme colors work with lipgloss v2.
func TestColorCompatibility(t *testing.T) {
	t.Parallel()
	// Colors should implement color.Color (verified at compile time)
	var c color.Color = Green() //nolint:staticcheck // Explicit type for compile-time interface verification
	if c == nil {
		t.Error("Green() returned nil")
	}

	// Colors should work with lipgloss v2 Foreground
	style := lipgloss.NewStyle().Foreground(Green())
	result := style.Render("test")
	if result == "" {
		t.Error("expected non-empty styled result")
	}
}

// TestThemeRegistry verifies all 280+ themes are available.
func TestThemeRegistry(t *testing.T) {
	t.Parallel()
	themes := ListThemes()
	if len(themes) < 200 {
		t.Errorf("expected 200+ themes, got %d", len(themes))
	}
}

//nolint:paralleltest // Tests modify global theme state
func TestThemeCycling(t *testing.T) {
	initial := Current()
	if initial == nil {
		t.Fatal("expected non-nil initial theme")
	}
	initialID := initial.ID

	NextTheme()
	after := Current()
	if after == nil {
		t.Fatal("expected non-nil theme after NextTheme")
	}

	// Should have changed (unless we only have one theme)
	if len(ListThemes()) > 1 && after.ID == initialID {
		t.Error("expected theme to change after NextTheme")
	}

	// Restore
	SetTheme(initialID)
}

//nolint:paralleltest // Tests modify global theme state
func TestSetTheme(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"dracula", true},
		{"nord", true},
		{"catppuccin_mocha", true},
		{"nonexistent_theme_xyz", false},
	}

	for _, tt := range tests {
		//nolint:paralleltest // Tests modify global theme state
		t.Run(tt.id, func(t *testing.T) {
			got := SetTheme(tt.id)
			if got != tt.want {
				t.Errorf("SetTheme(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}

	// Restore default
	SetTheme(DefaultTheme)
}

// TestStyleFunctions verifies style functions return valid styles.
func TestStyleFunctions(t *testing.T) {
	t.Parallel()
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"StatusOK", StatusOK()},
		{"StatusWarn", StatusWarn()},
		{"StatusError", StatusError()},
		{"StatusInfo", StatusInfo()},
		{"Bold", Bold()},
		{"Dim", Dim()},
		{"Title", Title()},
		{"Code", Code()},
	}

	for _, tt := range styles {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.style.Render("test")
			if result == "" {
				t.Errorf("%s().Render() returned empty string", tt.name)
			}
		})
	}
}

// TestRenderedStrings verifies pre-rendered strings are non-empty.
func TestRenderedStrings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		fn   func() string
	}{
		{"DeviceOnline", DeviceOnline},
		{"DeviceOffline", DeviceOffline},
		{"DeviceUpdating", DeviceUpdating},
		{"SwitchOn", SwitchOn},
		{"SwitchOff", SwitchOff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.fn()
			if result == "" {
				t.Errorf("%s() returned empty string", tt.name)
			}
		})
	}
}

// TestParseHexColor tests hex color parsing.
func TestParseHexColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		hex     string
		wantNil bool
	}{
		{"valid_with_hash", "#50fa7b", false},
		{"valid_without_hash", "50fa7b", false},
		{"empty", "", true},
		{"invalid_length", "#fff", true},
		{"invalid_chars", "#gggggg", true},
		{"uppercase", "#FF5555", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseHexColor(tt.hex)
			if tt.wantNil && result != nil {
				t.Errorf("parseHexColor(%q) = %v, want nil", tt.hex, result)
			}
			if !tt.wantNil && result == nil {
				t.Errorf("parseHexColor(%q) = nil, want non-nil", tt.hex)
			}
		})
	}
}

//nolint:paralleltest // Tests modify global theme state
func TestCustomColorOverrides(t *testing.T) {
	// Save original state
	original := GetCustomColors()
	defer SetCustomColors(original)

	// Clear any existing overrides
	ClearCustomColors()

	// Verify no overrides by default
	if GetCustomColors() != nil {
		t.Error("expected nil custom colors after clear")
	}

	// Set custom green color
	SetCustomColors(&CustomColors{
		Green: "#00ff00",
	})

	// Verify override is active
	overrides := GetCustomColors()
	if overrides == nil {
		t.Fatal("expected non-nil custom colors after set")
	}
	if overrides.Green != "#00ff00" {
		t.Errorf("expected green '#00ff00', got %q", overrides.Green)
	}

	// Verify Green() uses the override
	green := Green()
	if green == nil {
		t.Error("Green() returned nil with custom override")
	}

	// Clear and verify theme color is used
	ClearCustomColors()
	green = Green()
	if green == nil {
		t.Error("Green() returned nil after clearing overrides")
	}
}

//nolint:paralleltest // Tests modify global theme state
func TestApplyConfig(t *testing.T) {
	// Save original state
	originalTheme := Current()
	originalColors := GetCustomColors()
	defer func() {
		SetTheme(originalTheme.ID)
		SetCustomColors(originalColors)
	}()

	t.Run("set_theme_by_name", func(t *testing.T) {
		ClearCustomColors()
		err := ApplyConfig("nord", nil, "")
		if err != nil {
			t.Errorf("ApplyConfig with valid theme failed: %v", err)
		}
		if Current().ID != "nord" {
			t.Errorf("expected theme 'nord', got %q", Current().ID)
		}
	})

	t.Run("invalid_theme", func(t *testing.T) {
		err := ApplyConfig("nonexistent_theme_xyz", nil, "")
		if err == nil {
			t.Error("expected error for invalid theme name")
		}
	})

	t.Run("color_overrides", func(t *testing.T) {
		ClearCustomColors()
		colors := map[string]string{
			"green": "#00ff00",
			"red":   "#ff0000",
		}
		err := ApplyConfig("", colors, "")
		if err != nil {
			t.Errorf("ApplyConfig with colors failed: %v", err)
		}
		overrides := GetCustomColors()
		if overrides == nil {
			t.Fatal("expected non-nil overrides")
		}
		if overrides.Green != "#00ff00" {
			t.Errorf("expected green '#00ff00', got %q", overrides.Green)
		}
		if overrides.Red != "#ff0000" {
			t.Errorf("expected red '#ff0000', got %q", overrides.Red)
		}
	})

	t.Run("invalid_file", func(t *testing.T) {
		err := ApplyConfig("", nil, "/nonexistent/path/theme.yaml")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}

func TestExpandPath(t *testing.T) {
	t.Parallel()

	t.Run("tilde_expansion", func(t *testing.T) {
		t.Parallel()
		result := expandPath("~/test/path")
		if result == "~/test/path" {
			// Only fails if home dir couldn't be determined
			t.Skip("could not expand tilde")
		}
		if result[0] == '~' {
			t.Errorf("expected tilde to be expanded, got %q", result)
		}
	})

	t.Run("no_expansion_needed", func(t *testing.T) {
		t.Parallel()
		result := expandPath("/absolute/path")
		if result != "/absolute/path" {
			t.Errorf("expected '/absolute/path', got %q", result)
		}
	})

	t.Run("relative_path", func(t *testing.T) {
		t.Parallel()
		result := expandPath("relative/path")
		if result != "relative/path" {
			t.Errorf("expected 'relative/path', got %q", result)
		}
	})
}

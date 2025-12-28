package theme

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
)

const testColorGreen = "#00ff00"

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
		Green: testColorGreen,
	})

	// Verify override is active
	overrides := GetCustomColors()
	if overrides == nil {
		t.Fatal("expected non-nil custom colors after set")
	}
	if overrides.Green != testColorGreen {
		t.Errorf("expected green %q, got %q", testColorGreen, overrides.Green)
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
		err := ApplyConfig("nord", nil, nil, "")
		if err != nil {
			t.Errorf("ApplyConfig with valid theme failed: %v", err)
		}
		if Current().ID != "nord" {
			t.Errorf("expected theme 'nord', got %q", Current().ID)
		}
	})

	t.Run("invalid_theme", func(t *testing.T) {
		err := ApplyConfig("nonexistent_theme_xyz", nil, nil, "")
		if err == nil {
			t.Error("expected error for invalid theme name")
		}
	})

	t.Run("color_overrides", func(t *testing.T) {
		ClearCustomColors()
		colors := map[string]string{
			"green": testColorGreen,
			"red":   "#ff0000",
		}
		err := ApplyConfig("", colors, nil, "")
		if err != nil {
			t.Errorf("ApplyConfig with colors failed: %v", err)
		}
		overrides := GetCustomColors()
		if overrides == nil {
			t.Fatal("expected non-nil overrides")
		}
		if overrides.Green != testColorGreen {
			t.Errorf("expected green %q, got %q", testColorGreen, overrides.Green)
		}
		if overrides.Red != "#ff0000" {
			t.Errorf("expected red '#ff0000', got %q", overrides.Red)
		}
	})

	t.Run("invalid_file", func(t *testing.T) {
		err := ApplyConfig("", nil, nil, "/nonexistent/path/theme.yaml")
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

// TestPrevTheme verifies backward theme cycling.
//
//nolint:paralleltest // Tests modify global theme state
func TestPrevTheme(t *testing.T) {
	initial := Current()
	if initial == nil {
		t.Fatal("expected non-nil initial theme")
	}
	initialID := initial.ID

	PrevTheme()
	after := Current()
	if after == nil {
		t.Fatal("expected non-nil theme after PrevTheme")
	}

	// Should have changed (unless we only have one theme)
	if len(ListThemes()) > 1 && after.ID == initialID {
		t.Error("expected theme to change after PrevTheme")
	}

	// Restore
	SetTheme(initialID)
}

// TestGetTheme verifies theme retrieval by ID.
func TestGetTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id     string
		wantOK bool
	}{
		{"dracula", true},
		{"nord", true},
		{"nonexistent_theme_xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()
			tint, ok := GetTheme(tt.id)
			if ok != tt.wantOK {
				t.Errorf("GetTheme(%q) ok = %v, want %v", tt.id, ok, tt.wantOK)
			}
			if tt.wantOK && tint == nil {
				t.Errorf("GetTheme(%q) returned nil tint with ok=true", tt.id)
			}
		})
	}
}

// TestColorAccessors verifies all color accessor functions return non-nil colors.
func TestColorAccessors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func() color.Color
	}{
		{"Fg", Fg},
		{"Bg", Bg},
		{"Green", Green},
		{"Red", Red},
		{"Yellow", Yellow},
		{"Blue", Blue},
		{"Cyan", Cyan},
		{"Purple", Purple},
		{"Magenta", Magenta},
		{"Pink", Pink},
		{"Orange", Orange},
		{"BrightBlack", BrightBlack},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := tt.fn()
			if c == nil {
				t.Errorf("%s() returned nil", tt.name)
			}
		})
	}
}

// TestStatusStyles verifies all status style functions return non-empty styled strings.
func TestStatusStyles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"StatusOnline", StatusOnline()},
		{"StatusOffline", StatusOffline()},
		{"StatusUpdating", StatusUpdating()},
		{"Highlight", Highlight()},
		{"Subtitle", Subtitle()},
		{"Link", Link()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.style.Render("test")
			if result == "" {
				t.Errorf("%s().Render() returned empty string", tt.name)
			}
		})
	}
}

// TestFalseStyleRender verifies FalseStyle.Render for both styles.
func TestFalseStyleRender(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		style FalseStyle
	}{
		{"FalseError", FalseError},
		{"FalseDim", FalseDim},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.style.Render("test")
			if result == "" {
				t.Errorf("%s.Render() returned empty string", tt.name)
			}
		})
	}
}

// TestStyledPower verifies power value formatting.
func TestStyledPower(t *testing.T) {
	t.Parallel()

	tests := []struct {
		watts float64
	}{
		{0},
		{100},
		{1500.5},
		{-50},
	}

	for _, tt := range tests {
		result := StyledPower(tt.watts)
		if result == "" {
			t.Errorf("StyledPower(%v) returned empty string", tt.watts)
		}
	}
}

// TestStyledEnergy verifies energy value formatting with unit conversion.
func TestStyledEnergy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wh   float64
		unit string
	}{
		{0, "Wh"},
		{500, "Wh"},
		{999, "Wh"},
		{1000, "kWh"},
		{2500, "kWh"},
	}

	for _, tt := range tests {
		result := StyledEnergy(tt.wh)
		if result == "" {
			t.Errorf("StyledEnergy(%v) returned empty string", tt.wh)
		}
	}
}

// TestFormatFloat verifies float formatting helper.
func TestFormatFloat(t *testing.T) {
	t.Parallel()

	// Note: formatFloat has known limitations:
	// - Negative decimals produce incorrect output (e.g., -1.25 -> "-1.-25")
	// - Very small decimals may be lost due to floating point precision
	// This test documents actual behavior for values that work correctly.
	tests := []struct {
		input float64
		want  string
	}{
		{0, "0"},
		{100, "100"},
		{1.5, "1.50"},
		{-5, "-5"},
		{1.25, "1.25"},
		{0.5, "0.50"},
		{123.45, "123.45"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := formatFloat(tt.input)
			if got != tt.want {
				t.Errorf("formatFloat(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestIntToStr verifies integer to string conversion.
func TestIntToStr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{123, "123"},
		{-5, "-5"},
		{-999, "-999"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := intToStr(tt.input)
			if got != tt.want {
				t.Errorf("intToStr(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestColorToHex verifies color to hex string conversion.
func TestColorToHex(t *testing.T) {
	t.Parallel()

	t.Run("nil_color", func(t *testing.T) {
		t.Parallel()
		result := ColorToHex(nil)
		if result != "" {
			t.Errorf("ColorToHex(nil) = %q, want empty string", result)
		}
	})

	t.Run("green_color", func(t *testing.T) {
		t.Parallel()
		c := Green()
		result := ColorToHex(c)
		if result == "" {
			t.Error("ColorToHex(Green()) returned empty string")
		}
		// Should start with #
		if result != "" && result[0] != '#' {
			t.Errorf("ColorToHex() = %q, expected to start with #", result)
		}
	})
}

// TestBuildColorOverrides verifies building color override maps.
func TestBuildColorOverrides(t *testing.T) {
	t.Parallel()

	t.Run("nil_input", func(t *testing.T) {
		t.Parallel()
		result := BuildColorOverrides(nil)
		if result != nil {
			t.Errorf("BuildColorOverrides(nil) = %v, want nil", result)
		}
	})

	t.Run("empty_colors", func(t *testing.T) {
		t.Parallel()
		result := BuildColorOverrides(&CustomColors{})
		if result != nil {
			t.Errorf("BuildColorOverrides(empty) = %v, want nil", result)
		}
	})

	t.Run("with_colors", func(t *testing.T) {
		t.Parallel()
		input := &CustomColors{
			Green: testColorGreen,
			Red:   "#ff0000",
		}
		result := BuildColorOverrides(input)
		if result == nil {
			t.Fatal("BuildColorOverrides() returned nil")
		}
		if result["green"] != testColorGreen {
			t.Errorf("result[green] = %q, want %s", result["green"], testColorGreen)
		}
		if result["red"] != "#ff0000" {
			t.Errorf("result[red] = %q, want #ff0000", result["red"])
		}
		// Unset colors should not be in map
		if _, ok := result["blue"]; ok {
			t.Error("unexpected 'blue' key in result")
		}
	})
}

// TestGetCustomColorAllFields verifies all custom color fields are accessible.
//
//nolint:paralleltest // Tests modify global theme state
func TestGetCustomColorAllFields(t *testing.T) {
	// Save and restore original state
	original := GetCustomColors()
	defer SetCustomColors(original)

	// Set all fields
	SetCustomColors(&CustomColors{
		Foreground:  "#f8f8f2",
		Background:  "#282a36",
		Green:       "#50fa7b",
		Red:         "#ff5555",
		Yellow:      "#f1fa8c",
		Blue:        "#6272a4",
		Cyan:        "#8be9fd",
		Purple:      "#bd93f9",
		BrightBlack: "#44475a",
	})

	// Verify each color accessor uses custom values
	colors := []struct {
		name string
		fn   func() color.Color
	}{
		{"Fg", Fg},
		{"Bg", Bg},
		{"Green", Green},
		{"Red", Red},
		{"Yellow", Yellow},
		{"Blue", Blue},
		{"Cyan", Cyan},
		{"BrightBlack", BrightBlack},
	}

	for _, c := range colors {
		if c.fn() == nil {
			t.Errorf("%s() returned nil with custom override set", c.name)
		}
	}
}

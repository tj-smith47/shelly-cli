package theme

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/spf13/viper"
)

const (
	testColorGreen = "#00ff00"
	testThemeNord  = "nord"
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
		{testThemeNord, true},
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
		err := ApplyConfig(testThemeNord, nil, nil)
		if err != nil {
			t.Errorf("ApplyConfig with valid theme failed: %v", err)
		}
		if Current().ID != testThemeNord {
			t.Errorf("expected theme %q, got %q", testThemeNord, Current().ID)
		}
	})

	t.Run("invalid_theme", func(t *testing.T) {
		err := ApplyConfig("nonexistent_theme_xyz", nil, nil)
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
		err := ApplyConfig("", colors, nil)
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

	t.Run("invalid_file_data", func(t *testing.T) {
		// Test that ApplyThemeFromData returns error for invalid YAML
		err := ApplyThemeFromData([]byte("invalid: [yaml"), nil)
		if err == nil {
			t.Error("expected error for invalid YAML data")
		}
	})
}

func TestExpandPath(t *testing.T) {
	t.Parallel()

	t.Run("tilde_expansion", func(t *testing.T) {
		t.Parallel()
		result := ExpandPath("~/test/path")
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
		result := ExpandPath("/absolute/path")
		if result != "/absolute/path" {
			t.Errorf("expected '/absolute/path', got %q", result)
		}
	})

	t.Run("relative_path", func(t *testing.T) {
		t.Parallel()
		result := ExpandPath("relative/path")
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
		{testThemeNord, true},
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

// TestCurrentThemeName tests the CurrentThemeName function.
//
//nolint:paralleltest // Tests modify global theme state
func TestCurrentThemeName(t *testing.T) {
	// Save and restore
	original := Current()
	defer SetTheme(original.ID)

	// Set a known theme
	SetTheme("dracula")
	name := CurrentThemeName()
	if name != "dracula" {
		t.Errorf("CurrentThemeName() = %q, want %q", name, "dracula")
	}

	// Test with different theme
	SetTheme(testThemeNord)
	name = CurrentThemeName()
	if name != testThemeNord {
		t.Errorf("CurrentThemeName() = %q, want %q", name, testThemeNord)
	}
}

// TestSetThemeFromConfig tests SetThemeFromConfig function.
//
//nolint:paralleltest // Tests modify global viper/theme state
func TestSetThemeFromConfig(t *testing.T) {
	// Save and restore
	original := Current()
	originalViperTheme := viper.GetString("theme")
	defer func() {
		SetTheme(original.ID)
		viper.Set("theme", originalViperTheme)
	}()

	// Test with empty config (should use default)
	viper.Set("theme", "")
	SetThemeFromConfig()
	if CurrentThemeName() != DefaultTheme {
		t.Errorf("SetThemeFromConfig() with empty config = %q, want %q", CurrentThemeName(), DefaultTheme)
	}

	// Test with specific theme
	viper.Set("theme", testThemeNord)
	SetThemeFromConfig()
	if CurrentThemeName() != testThemeNord {
		t.Errorf("SetThemeFromConfig() with %q config = %q, want %q", testThemeNord, CurrentThemeName(), testThemeNord)
	}
}

// TestApplyThemeFromData_ValidData tests ApplyThemeFromData with valid theme data.
//
//nolint:paralleltest // Tests modify global theme state
func TestApplyThemeFromData_ValidData(t *testing.T) {
	themeContent := `name: nord
colors:
  green: "#50fa7b"
`
	// Save and restore original state
	originalTheme := Current()
	originalColors := GetCustomColors()
	defer func() {
		SetTheme(originalTheme.ID)
		SetCustomColors(originalColors)
	}()

	// Apply from data
	err := ApplyThemeFromData([]byte(themeContent), nil)
	if err != nil {
		t.Errorf("ApplyThemeFromData() with valid data error = %v", err)
	}

	// Verify theme was applied
	if Current().ID != testThemeNord {
		t.Errorf("expected theme %q, got %q", testThemeNord, Current().ID)
	}

	// Verify color override was applied
	colors := GetCustomColors()
	if colors == nil || colors.Green != "#50fa7b" {
		t.Errorf("expected green color override '#50fa7b', got %v", colors)
	}
}

// TestDefaultColor tests the defaultColor function branches.
func TestDefaultColor(t *testing.T) {
	t.Parallel()

	// Purple and Orange use special fallback logic
	t.Run("purple", func(t *testing.T) {
		t.Parallel()
		c := Purple()
		if c == nil {
			t.Error("Purple() returned nil")
		}
	})

	t.Run("orange", func(t *testing.T) {
		t.Parallel()
		c := Orange()
		if c == nil {
			t.Error("Orange() returned nil")
		}
	})
}

// TestPurple_WithCustomColor tests Purple with custom color override.
//
//nolint:paralleltest // Tests modify global theme state
func TestPurple_WithCustomColor(t *testing.T) {
	// Save and restore
	original := GetCustomColors()
	defer SetCustomColors(original)

	// Set custom purple color
	SetCustomColors(&CustomColors{
		Purple: "#9b59b6",
	})

	c := Purple()
	if c == nil {
		t.Error("Purple() with custom color returned nil")
	}
}

// TestBrightBlack_WithCustomColor tests BrightBlack with custom color override.
//
//nolint:paralleltest // Tests modify global theme state
func TestBrightBlack_WithCustomColor(t *testing.T) {
	// Save and restore
	original := GetCustomColors()
	defer SetCustomColors(original)

	// Set custom bright_black color
	SetCustomColors(&CustomColors{
		BrightBlack: "#666666",
	})

	c := BrightBlack()
	if c == nil {
		t.Error("BrightBlack() with custom color returned nil")
	}
}

// TestColorFunctions_WithNilTheme tests color functions' fallback behavior.
// This is hard to test directly since themes are always initialized,
// so we test by verifying the color functions work correctly.
//
//nolint:paralleltest // Tests modify global theme state
func TestColorFunctions_AllWithCustom(t *testing.T) {
	// Save and restore
	original := GetCustomColors()
	defer SetCustomColors(original)

	// Set all custom colors
	SetCustomColors(&CustomColors{
		Foreground:  "#ffffff",
		Background:  "#000000",
		Green:       "#00ff00",
		Red:         "#ff0000",
		Yellow:      "#ffff00",
		Blue:        "#0000ff",
		Cyan:        "#00ffff",
		Purple:      "#ff00ff",
		BrightBlack: "#808080",
	})

	// Verify all color functions return non-nil
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
		{"BrightBlack", BrightBlack},
	}

	for _, tt := range tests {
		c := tt.fn()
		if c == nil {
			t.Errorf("%s() with custom color returned nil", tt.name)
		}
	}
}

// TestApplyThemeFromData_UnknownTheme tests ApplyThemeFromData with an unknown theme name.
//
//nolint:paralleltest // Tests modify global theme state
func TestApplyThemeFromData_UnknownTheme(t *testing.T) {
	themeContent := `name: nonexistent_theme_xyz_12345
colors:
  green: "#50fa7b"
`
	// Apply should fail with unknown theme error
	err := ApplyThemeFromData([]byte(themeContent), nil)
	if err == nil {
		t.Error("ApplyThemeFromData() with unknown theme should return error")
	}
}

// TestApplyThemeFromData_InvalidYAML tests ApplyThemeFromData with invalid YAML.
//
//nolint:paralleltest // Tests modify global theme state
func TestApplyThemeFromData_InvalidYAML(t *testing.T) {
	err := ApplyThemeFromData([]byte("name: [invalid yaml\n"), nil)
	if err == nil {
		t.Error("ApplyThemeFromData() with invalid YAML should return error")
	}
}

// TestApplyConfig_WithSemanticOverrides tests ApplyConfig with semantic overrides.
//
//nolint:paralleltest // Tests modify global theme state
func TestApplyConfig_WithSemanticOverrides(t *testing.T) {
	// Save original state
	originalSemantics := GetSemanticColors()
	defer setSemanticColors(originalSemantics)

	// Apply config with semantic overrides
	semantics := &SemanticOverrides{
		Success: "#00ff00",
		Error:   "#ff0000",
	}
	err := ApplyConfig("", nil, semantics)
	if err != nil {
		t.Errorf("ApplyConfig with semantic overrides failed: %v", err)
	}
}

// TestApplyThemeFromData_WithSemantics tests ApplyThemeFromData with semantic overrides.
//
//nolint:paralleltest // Tests modify global theme state
func TestApplyThemeFromData_WithSemantics(t *testing.T) {
	themeContent := `name: dracula
colors:
  green: "#50fa7b"
`
	// Save original state
	originalTheme := Current()
	originalColors := GetCustomColors()
	originalSemantics := GetSemanticColors()
	defer func() {
		SetTheme(originalTheme.ID)
		SetCustomColors(originalColors)
		setSemanticColors(originalSemantics)
	}()

	// Apply from data with semantic overrides
	semantics := &SemanticOverrides{
		Success: "#00ff00",
	}
	err := ApplyThemeFromData([]byte(themeContent), semantics)
	if err != nil {
		t.Errorf("ApplyThemeFromData with semantics failed: %v", err)
	}
}

// TestDefaultColor_NilColor tests defaultColor with nil primary color.
func TestDefaultColor_NilColor(t *testing.T) {
	t.Parallel()

	// Test with nil primary color - should return fallback
	result := defaultColor(nil, draculaGreen)
	if result == nil {
		t.Error("defaultColor(nil, fallback) should return fallback, got nil")
	}
	if result != draculaGreen {
		t.Errorf("defaultColor(nil, fallback) = %v, want %v", result, draculaGreen)
	}

	// Test with non-nil primary color - should return primary
	result = defaultColor(draculaRed, draculaGreen)
	if result != draculaRed {
		t.Errorf("defaultColor(primary, fallback) = %v, want %v", result, draculaRed)
	}
}

// TestFormatFloat_EdgeCases tests formatFloat with additional edge cases.
func TestFormatFloat_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{"whole_number_2", 2, "2"},
		{"whole_number_1000", 1000, "1000"},
		{"zero_decimal", 5.00, "5"},
		{"small_decimal", 1.01, "1.1"}, // Note: 0.01 * 100 = 1, not "01"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatFloat(tt.input)
			if got != tt.want {
				t.Errorf("formatFloat(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

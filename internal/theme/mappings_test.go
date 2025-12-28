package theme

import (
	"testing"
)

// TestGetThemeMapping verifies theme mapping retrieval.
func TestGetThemeMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		themeName string
	}{
		{"dracula", "dracula"},
		{"nord", "nord"},
		{"tokyo-night", "tokyo-night"},
		{"tokyonight", "tokyonight"},
		{"tokyonight-night", "tokyonight-night"},
		{"gruvbox", "gruvbox"},
		{"gruvbox-dark", "gruvbox-dark"},
		{"catppuccin", "catppuccin"},
		{"catppuccin-mocha", "catppuccin-mocha"},
		{"unknown_theme", "unknown_theme"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			colors := GetThemeMapping(tt.themeName)

			// All mappings should return valid SemanticColors
			if colors.Primary == nil {
				t.Errorf("GetThemeMapping(%q).Primary is nil", tt.themeName)
			}
			if colors.Success == nil {
				t.Errorf("GetThemeMapping(%q).Success is nil", tt.themeName)
			}
			if colors.Error == nil {
				t.Errorf("GetThemeMapping(%q).Error is nil", tt.themeName)
			}
			if colors.Text == nil {
				t.Errorf("GetThemeMapping(%q).Text is nil", tt.themeName)
			}
		})
	}
}

// TestMappingFromTheme verifies fallback mapping generation.
func TestMappingFromTheme(t *testing.T) {
	t.Parallel()

	colors := MappingFromTheme()

	// Verify all fields are populated
	fields := []struct {
		name  string
		color interface{}
	}{
		{"Primary", colors.Primary},
		{"Secondary", colors.Secondary},
		{"Highlight", colors.Highlight},
		{"Muted", colors.Muted},
		{"Text", colors.Text},
		{"AltText", colors.AltText},
		{"Success", colors.Success},
		{"Warning", colors.Warning},
		{"Error", colors.Error},
		{"Info", colors.Info},
		{"Background", colors.Background},
		{"AltBackground", colors.AltBackground},
		{"Online", colors.Online},
		{"Offline", colors.Offline},
		{"Updating", colors.Updating},
		{"Idle", colors.Idle},
		{"TableHeader", colors.TableHeader},
		{"TableCell", colors.TableCell},
		{"TableAltCell", colors.TableAltCell},
		{"TableBorder", colors.TableBorder},
	}

	for _, f := range fields {
		if f.color == nil {
			t.Errorf("MappingFromTheme().%s is nil", f.name)
		}
	}
}

// TestDraculaSemanticMapping verifies Dracula theme mapping.
func TestDraculaSemanticMapping(t *testing.T) {
	t.Parallel()

	colors := DraculaSemanticMapping()

	// Verify Dracula-specific colors
	if colors.Primary == nil {
		t.Error("DraculaSemanticMapping().Primary is nil")
	}
	if colors.Success == nil {
		t.Error("DraculaSemanticMapping().Success is nil")
	}
	if colors.TableBorder == nil {
		t.Error("DraculaSemanticMapping().TableBorder is nil")
	}
}

// TestNordSemanticMapping verifies Nord theme mapping.
func TestNordSemanticMapping(t *testing.T) {
	t.Parallel()

	colors := NordSemanticMapping()

	// Verify Nord-specific colors
	if colors.Primary == nil {
		t.Error("NordSemanticMapping().Primary is nil")
	}
	if colors.Success == nil {
		t.Error("NordSemanticMapping().Success is nil")
	}
	if colors.TableBorder == nil {
		t.Error("NordSemanticMapping().TableBorder is nil")
	}
}

// TestTokyoNightSemanticMapping verifies Tokyo Night theme mapping.
func TestTokyoNightSemanticMapping(t *testing.T) {
	t.Parallel()

	colors := TokyoNightSemanticMapping()

	// Verify Tokyo Night-specific colors
	if colors.Primary == nil {
		t.Error("TokyoNightSemanticMapping().Primary is nil")
	}
	if colors.Success == nil {
		t.Error("TokyoNightSemanticMapping().Success is nil")
	}
	if colors.TableBorder == nil {
		t.Error("TokyoNightSemanticMapping().TableBorder is nil")
	}
}

// TestGruvboxSemanticMapping verifies Gruvbox theme mapping.
func TestGruvboxSemanticMapping(t *testing.T) {
	t.Parallel()

	colors := GruvboxSemanticMapping()

	// Verify Gruvbox-specific colors
	if colors.Primary == nil {
		t.Error("GruvboxSemanticMapping().Primary is nil")
	}
	if colors.Success == nil {
		t.Error("GruvboxSemanticMapping().Success is nil")
	}
	if colors.TableBorder == nil {
		t.Error("GruvboxSemanticMapping().TableBorder is nil")
	}
}

// TestCatppuccinSemanticMapping verifies Catppuccin theme mapping.
func TestCatppuccinSemanticMapping(t *testing.T) {
	t.Parallel()

	colors := CatppuccinSemanticMapping()

	// Verify Catppuccin-specific colors
	if colors.Primary == nil {
		t.Error("CatppuccinSemanticMapping().Primary is nil")
	}
	if colors.Success == nil {
		t.Error("CatppuccinSemanticMapping().Success is nil")
	}
	if colors.TableBorder == nil {
		t.Error("CatppuccinSemanticMapping().TableBorder is nil")
	}
}

// TestAllMappingsFull verifies all theme mappings populate all fields.
func TestAllMappingsFull(t *testing.T) {
	t.Parallel()

	mappings := []struct {
		name string
		fn   func() SemanticColors
	}{
		{"Dracula", DraculaSemanticMapping},
		{"Nord", NordSemanticMapping},
		{"TokyoNight", TokyoNightSemanticMapping},
		{"Gruvbox", GruvboxSemanticMapping},
		{"Catppuccin", CatppuccinSemanticMapping},
		{"FromTheme", MappingFromTheme},
	}

	for _, m := range mappings {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			colors := m.fn()

			// Check all 20 semantic color fields
			checks := []struct {
				field string
				color interface{}
			}{
				{"Primary", colors.Primary},
				{"Secondary", colors.Secondary},
				{"Highlight", colors.Highlight},
				{"Muted", colors.Muted},
				{"Text", colors.Text},
				{"AltText", colors.AltText},
				{"Success", colors.Success},
				{"Warning", colors.Warning},
				{"Error", colors.Error},
				{"Info", colors.Info},
				{"Background", colors.Background},
				{"AltBackground", colors.AltBackground},
				{"Online", colors.Online},
				{"Offline", colors.Offline},
				{"Updating", colors.Updating},
				{"Idle", colors.Idle},
				{"TableHeader", colors.TableHeader},
				{"TableCell", colors.TableCell},
				{"TableAltCell", colors.TableAltCell},
				{"TableBorder", colors.TableBorder},
			}

			for _, c := range checks {
				if c.color == nil {
					t.Errorf("%s().%s is nil", m.name, c.field)
				}
			}
		})
	}
}

// TestThemeMappingsRegistration verifies themeMappings var is populated.
func TestThemeMappingsRegistration(t *testing.T) {
	t.Parallel()

	// These themes should have custom mappings
	expectedThemes := []string{
		"dracula",
		"nord",
		"tokyo-night",
		"tokyonight",
		"tokyonight-night",
		"gruvbox",
		"gruvbox-dark",
		"catppuccin",
		"catppuccin-mocha",
	}

	for _, theme := range expectedThemes {
		t.Run(theme, func(t *testing.T) {
			t.Parallel()
			// GetThemeMapping should find these in themeMappings
			colors := GetThemeMapping(theme)
			if colors.Primary == nil {
				t.Errorf("theme %q should have custom mapping", theme)
			}
		})
	}
}

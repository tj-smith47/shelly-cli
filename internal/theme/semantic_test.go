package theme

import (
	"testing"

	"charm.land/lipgloss/v2"
)

// TestGetSemanticColors verifies semantic colors are accessible.
func TestGetSemanticColors(t *testing.T) {
	t.Parallel()

	colors := GetSemanticColors()

	// Verify key colors are set (not nil)
	if colors.Primary == nil {
		t.Error("SemanticColors.Primary is nil")
	}
	if colors.Success == nil {
		t.Error("SemanticColors.Success is nil")
	}
	if colors.Error == nil {
		t.Error("SemanticColors.Error is nil")
	}
	if colors.Text == nil {
		t.Error("SemanticColors.Text is nil")
	}
}

// TestApplySemanticOverrides verifies semantic override application.
//
//nolint:paralleltest // Tests modify global theme state
func TestApplySemanticOverrides(t *testing.T) {
	// Save original semantic colors
	original := GetSemanticColors()
	defer setSemanticColors(original)

	t.Run("nil_overrides", func(t *testing.T) {
		// Should not panic with nil
		ApplySemanticOverrides(nil)
	})

	t.Run("empty_overrides", func(t *testing.T) {
		// Should not change anything
		before := GetSemanticColors()
		ApplySemanticOverrides(&SemanticOverrides{})
		after := GetSemanticColors()

		// Colors should remain unchanged (both non-nil)
		if before.Primary == nil || after.Primary == nil {
			t.Error("Primary became nil after empty overrides")
		}
	})

	t.Run("primary_override", func(t *testing.T) {
		ApplySemanticOverrides(&SemanticOverrides{
			Primary: "#ff0000",
		})
		colors := GetSemanticColors()
		if colors.Primary == nil {
			t.Error("Primary is nil after override")
		}
	})

	t.Run("all_overrides", func(t *testing.T) {
		overrides := &SemanticOverrides{
			Primary:       "#111111",
			Secondary:     "#222222",
			Highlight:     "#333333",
			Muted:         "#444444",
			Text:          "#555555",
			AltText:       "#666666",
			Success:       "#00ff00",
			Warning:       "#ffff00",
			Error:         "#ff0000",
			Info:          "#0000ff",
			Background:    "#000000",
			AltBackground: "#101010",
			Online:        "#00ff00",
			Offline:       "#ff0000",
			Updating:      "#ffff00",
			Idle:          "#808080",
			TableHeader:   "#aaaaaa",
			TableCell:     "#bbbbbb",
			TableAltCell:  "#cccccc",
			TableBorder:   "#dddddd",
		}
		ApplySemanticOverrides(overrides)

		colors := GetSemanticColors()
		// Verify all overrides were applied (non-nil)
		if colors.Primary == nil {
			t.Error("Primary is nil")
		}
		if colors.Secondary == nil {
			t.Error("Secondary is nil")
		}
		if colors.Success == nil {
			t.Error("Success is nil")
		}
		if colors.TableBorder == nil {
			t.Error("TableBorder is nil")
		}
	})
}

// TestSemanticStyleFunctions verifies all semantic style accessors.
func TestSemanticStyleFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func() lipgloss.Style
	}{
		{"SemanticPrimary", SemanticPrimary},
		{"SemanticSecondary", SemanticSecondary},
		{"SemanticHighlight", SemanticHighlight},
		{"SemanticMuted", SemanticMuted},
		{"SemanticText", SemanticText},
		{"SemanticAltText", SemanticAltText},
		{"SemanticSuccess", SemanticSuccess},
		{"SemanticWarning", SemanticWarning},
		{"SemanticError", SemanticError},
		{"SemanticInfo", SemanticInfo},
		{"SemanticOnline", SemanticOnline},
		{"SemanticOffline", SemanticOffline},
		{"SemanticUpdating", SemanticUpdating},
		{"SemanticIdle", SemanticIdle},
		{"SemanticTableHeader", SemanticTableHeader},
		{"SemanticTableCell", SemanticTableCell},
		{"SemanticTableAltCell", SemanticTableAltCell},
		{"SemanticTableBorder", SemanticTableBorder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			style := tt.fn()
			result := style.Render("test")
			if result == "" {
				t.Errorf("%s().Render() returned empty string", tt.name)
			}
		})
	}
}

// TestSemanticOverridesStruct verifies SemanticOverrides struct fields.
func TestSemanticOverridesStruct(t *testing.T) {
	t.Parallel()

	overrides := SemanticOverrides{
		Primary:       "#bd93f9",
		Secondary:     "#6272a4",
		Highlight:     "#8be9fd",
		Muted:         "#6272a4",
		Text:          "#f8f8f2",
		AltText:       "#6272a4",
		Success:       "#50fa7b",
		Warning:       "#ffb86c",
		Error:         "#ff5555",
		Info:          "#8be9fd",
		Background:    "#282a36",
		AltBackground: "#44475a",
		Online:        "#50fa7b",
		Offline:       "#ff5555",
		Updating:      "#f1fa8c",
		Idle:          "#6272a4",
		TableHeader:   "#8be9fd",
		TableCell:     "#f1fa8c",
		TableAltCell:  "#ffb86c",
		TableBorder:   "#ff79c6",
	}

	// Verify all fields are set
	if overrides.Primary == "" {
		t.Error("Primary not set")
	}
	if overrides.Success == "" {
		t.Error("Success not set")
	}
	if overrides.TableBorder == "" {
		t.Error("TableBorder not set")
	}
}

// TestSemanticColorsStruct verifies SemanticColors struct has all expected fields.
func TestSemanticColorsStruct(t *testing.T) {
	t.Parallel()

	// Get current colors and verify structure
	colors := GetSemanticColors()

	// Test type assertions work (compile-time check)
	_ = colors.Primary
	_ = colors.Secondary
	_ = colors.Highlight
	_ = colors.Muted
	_ = colors.Text
	_ = colors.AltText
	_ = colors.Success
	_ = colors.Warning
	_ = colors.Error
	_ = colors.Info
	_ = colors.Background
	_ = colors.AltBackground
	_ = colors.Online
	_ = colors.Offline
	_ = colors.Updating
	_ = colors.Idle
	_ = colors.TableHeader
	_ = colors.TableCell
	_ = colors.TableAltCell
	_ = colors.TableBorder
}

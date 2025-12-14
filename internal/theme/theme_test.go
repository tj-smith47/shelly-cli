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

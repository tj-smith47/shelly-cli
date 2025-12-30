package syntax

import (
	"testing"

	"github.com/spf13/viper"
)

//nolint:paralleltest // Tests modify shared IsTTY and viper state
func TestShouldHighlight(t *testing.T) {
	// Save and restore state
	oldIsTTY := IsTTY
	defer func() { IsTTY = oldIsTTY }()

	t.Run("returns false when not TTY", func(t *testing.T) {
		IsTTY = func() bool { return false }
		if ShouldHighlight() {
			t.Error("ShouldHighlight() = true, want false")
		}
	})

	t.Run("returns false when plain flag set", func(t *testing.T) {
		IsTTY = func() bool { return true }
		viper.Set("plain", true)
		defer viper.Set("plain", false)
		if ShouldHighlight() {
			t.Error("ShouldHighlight() = true, want false")
		}
	})

	t.Run("returns false when no-color flag set", func(t *testing.T) {
		IsTTY = func() bool { return true }
		viper.Set("no-color", true)
		defer viper.Set("no-color", false)
		if ShouldHighlight() {
			t.Error("ShouldHighlight() = true, want false")
		}
	})

	t.Run("returns false when NO_COLOR env set", func(t *testing.T) {
		IsTTY = func() bool { return true }
		t.Setenv("NO_COLOR", "1")
		if ShouldHighlight() {
			t.Error("ShouldHighlight() = true, want false")
		}
	})

	t.Run("returns false when SHELLY_NO_COLOR env set", func(t *testing.T) {
		IsTTY = func() bool { return true }
		t.Setenv("SHELLY_NO_COLOR", "1")
		if ShouldHighlight() {
			t.Error("ShouldHighlight() = true, want false")
		}
	})

	t.Run("returns false when TERM=dumb", func(t *testing.T) {
		IsTTY = func() bool { return true }
		t.Setenv("TERM", "dumb")
		if ShouldHighlight() {
			t.Error("ShouldHighlight() = true, want false")
		}
	})

	t.Run("returns true when TTY and no disabling flags", func(t *testing.T) {
		IsTTY = func() bool { return true }
		viper.Set("plain", false)
		viper.Set("no-color", false)
		t.Setenv("TERM", "xterm-256color")
		if !ShouldHighlight() {
			t.Error("ShouldHighlight() = false, want true")
		}
	})
}

func TestHighlightCode(t *testing.T) {
	t.Parallel()

	t.Run("JSON code", func(t *testing.T) {
		t.Parallel()
		result := HighlightCode(`{"key": "value"}`, "json")
		if result == "" {
			t.Error("HighlightCode() returned empty string")
		}
	})

	t.Run("YAML code", func(t *testing.T) {
		t.Parallel()
		result := HighlightCode("key: value\n", "yaml")
		if result == "" {
			t.Error("HighlightCode() returned empty string")
		}
	})

	t.Run("unknown language returns original", func(t *testing.T) {
		t.Parallel()
		code := "some text"
		result := HighlightCode(code, "nonexistent")
		if result != code {
			t.Errorf("HighlightCode() = %q, want %q", result, code)
		}
	})
}

//nolint:paralleltest // Tests modify shared viper state
func TestGetChromaStyle(t *testing.T) {
	t.Run("dracula theme", func(t *testing.T) {
		viper.Set("theme.name", "dracula")
		defer viper.Set("theme.name", "")
		if GetChromaStyle() == nil {
			t.Error("GetChromaStyle() returned nil")
		}
	})

	t.Run("nord theme", func(t *testing.T) {
		viper.Set("theme.name", "nord")
		defer viper.Set("theme.name", "")
		if GetChromaStyle() == nil {
			t.Error("GetChromaStyle() returned nil")
		}
	})

	t.Run("gruvbox theme", func(t *testing.T) {
		viper.Set("theme.name", "gruvbox")
		defer viper.Set("theme.name", "")
		if GetChromaStyle() == nil {
			t.Error("GetChromaStyle() returned nil")
		}
	})

	t.Run("unknown theme uses fallback", func(t *testing.T) {
		viper.Set("theme.name", "unknown-theme")
		defer viper.Set("theme.name", "")
		if GetChromaStyle() == nil {
			t.Error("GetChromaStyle() returned nil")
		}
	})

	t.Run("empty theme uses default", func(t *testing.T) {
		viper.Set("theme.name", "")
		if GetChromaStyle() == nil {
			t.Error("GetChromaStyle() returned nil")
		}
	})
}

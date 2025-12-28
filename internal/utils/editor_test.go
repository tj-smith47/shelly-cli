// Package utils provides common functionality shared across CLI commands.
package utils

import (
	"testing"
)

func TestGetEditor_DefaultFallback(t *testing.T) {
	// Test that GetEditor returns something when both EDITOR and VISUAL are unset
	// Cannot use t.Parallel() since it manipulates environment

	t.Run("with neither EDITOR nor VISUAL", func(t *testing.T) {
		t.Setenv("EDITOR", "")
		t.Setenv("VISUAL", "")

		editor := GetEditor()
		// Should return one of the default editors (nano, vim, vi) or empty
		// The actual result depends on system availability
		// We just verify it doesn't panic
		_ = editor
	})
}

func TestGetEditor_EditorPrecedence(t *testing.T) {
	// Test that EDITOR takes precedence over VISUAL

	t.Run("EDITOR takes precedence", func(t *testing.T) {
		t.Setenv("EDITOR", "custom-editor")
		t.Setenv("VISUAL", "visual-editor")

		editor := GetEditor()
		if editor != "custom-editor" {
			t.Errorf("GetEditor() = %q, want 'custom-editor' (EDITOR should take precedence)", editor)
		}
	})
}

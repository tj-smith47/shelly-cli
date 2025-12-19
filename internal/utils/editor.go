package utils

import (
	"os"
	"os/exec"
)

// GetEditor returns the user's preferred editor.
// It checks EDITOR, then VISUAL environment variables,
// then falls back to common editors (nano, vim, vi).
// Returns empty string if no editor is found.
func GetEditor() string {
	// Check EDITOR first
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// Then VISUAL
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	// Platform defaults - try common editors in order
	editors := []string{"nano", "vim", "vi"}
	for _, e := range editors {
		if path, err := exec.LookPath(e); err == nil {
			return path
		}
	}

	return ""
}

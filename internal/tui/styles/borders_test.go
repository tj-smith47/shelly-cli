package styles

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestPanelBorder(t *testing.T) {
	t.Parallel()

	style := PanelBorder()

	// Verify it has a border style set
	if style.GetBorderStyle() == (lipgloss.Border{}) {
		t.Error("expected border style to be set")
	}
}

func TestPanelBorderActive(t *testing.T) {
	t.Parallel()

	style := PanelBorderActive()

	// Verify it has a border style set
	if style.GetBorderStyle() == (lipgloss.Border{}) {
		t.Error("expected border style to be set")
	}
}

func TestPanelBorderFocused(t *testing.T) {
	t.Parallel()

	t.Run("focused", func(t *testing.T) {
		t.Parallel()
		style := PanelBorderFocused(true)
		if style.GetBorderStyle() == (lipgloss.Border{}) {
			t.Error("expected border style to be set")
		}
	})

	t.Run("unfocused", func(t *testing.T) {
		t.Parallel()
		style := PanelBorderFocused(false)
		if style.GetBorderStyle() == (lipgloss.Border{}) {
			t.Error("expected border style to be set")
		}
	})
}

func TestModalBorder(t *testing.T) {
	t.Parallel()

	style := ModalBorder()

	// Verify it has a border style set
	if style.GetBorderStyle() == (lipgloss.Border{}) {
		t.Error("expected border style to be set")
	}
}

func TestErrorBorder(t *testing.T) {
	t.Parallel()

	style := ErrorBorder()

	// Verify it has a border style set
	if style.GetBorderStyle() == (lipgloss.Border{}) {
		t.Error("expected border style to be set")
	}
}

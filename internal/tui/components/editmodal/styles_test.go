package editmodal

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestDefaultStyles(t *testing.T) {
	t.Parallel()

	styles := DefaultStyles()

	// Verify all styles are initialized (not zero-value)
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"Modal", styles.Modal},
		{"Title", styles.Title},
		{"Label", styles.Label},
		{"LabelFocus", styles.LabelFocus},
		{"Error", styles.Error},
		{"Help", styles.Help},
		{"Selector", styles.Selector},
		{"Warning", styles.Warning},
		{"Info", styles.Info},
		{"Muted", styles.Muted},
		{"Overlay", styles.Overlay},
		{"Input", styles.Input},
		{"InputFocus", styles.InputFocus},
		{"Button", styles.Button},
		{"ButtonFocus", styles.ButtonFocus},
		{"ButtonDanger", styles.ButtonDanger},
		{"Value", styles.Value},
		{"StatusOn", styles.StatusOn},
		{"StatusOff", styles.StatusOff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Render something to verify the style works
			result := tt.style.Render("test")
			if result == "" {
				t.Error("expected non-empty rendered result")
			}
		})
	}
}

func TestDefaultLabelWidth(t *testing.T) {
	t.Parallel()

	if DefaultLabelWidth != 14 {
		t.Errorf("expected DefaultLabelWidth to be 14, got %d", DefaultLabelWidth)
	}
}

func TestWithLabelWidth(t *testing.T) {
	t.Parallel()

	styles := DefaultStyles()
	newStyles := styles.WithLabelWidth(20)

	// Original should be unchanged
	if styles.Label.GetWidth() != 14 {
		t.Errorf("expected original Label width to be 14, got %d", styles.Label.GetWidth())
	}

	// New should have the new width
	if newStyles.Label.GetWidth() != 20 {
		t.Errorf("expected new Label width to be 20, got %d", newStyles.Label.GetWidth())
	}
	if newStyles.LabelFocus.GetWidth() != 20 {
		t.Errorf("expected new LabelFocus width to be 20, got %d", newStyles.LabelFocus.GetWidth())
	}
}

func TestLabelStyle(t *testing.T) {
	t.Parallel()

	styles := DefaultStyles()

	t.Run("focused", func(t *testing.T) {
		t.Parallel()
		style := styles.LabelStyle(true)
		// LabelFocus should have bold
		if !style.GetBold() {
			t.Error("expected focused label to be bold")
		}
	})

	t.Run("unfocused", func(t *testing.T) {
		t.Parallel()
		style := styles.LabelStyle(false)
		// Label should not be bold
		if style.GetBold() {
			t.Error("expected unfocused label to not be bold")
		}
	})
}

func TestInputStyle(t *testing.T) {
	t.Parallel()

	styles := DefaultStyles()

	t.Run("focused", func(t *testing.T) {
		t.Parallel()
		style := styles.InputStyle(true)
		if style.GetBorderStyle() == (lipgloss.Border{}) {
			t.Error("expected input to have border")
		}
	})

	t.Run("unfocused", func(t *testing.T) {
		t.Parallel()
		style := styles.InputStyle(false)
		if style.GetBorderStyle() == (lipgloss.Border{}) {
			t.Error("expected input to have border")
		}
	})
}

func TestButtonStyle(t *testing.T) {
	t.Parallel()

	styles := DefaultStyles()

	t.Run("focused", func(t *testing.T) {
		t.Parallel()
		style := styles.ButtonStyle(true)
		if !style.GetBold() {
			t.Error("expected focused button to be bold")
		}
	})

	t.Run("unfocused", func(t *testing.T) {
		t.Parallel()
		style := styles.ButtonStyle(false)
		if style.GetBold() {
			t.Error("expected unfocused button to not be bold")
		}
	})
}

func TestStatusStyle(t *testing.T) {
	t.Parallel()

	styles := DefaultStyles()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()
		style := styles.StatusStyle(true)
		// Just verify it returns a valid style
		result := style.Render("ON")
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()
		style := styles.StatusStyle(false)
		result := style.Render("OFF")
		if result == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestRenderSelector(t *testing.T) {
	t.Parallel()

	styles := DefaultStyles()

	t.Run("selected", func(t *testing.T) {
		t.Parallel()
		result := styles.RenderSelector(true)
		if !strings.Contains(result, "▶") {
			t.Error("expected selector to contain ▶ when selected")
		}
	})

	t.Run("not selected", func(t *testing.T) {
		t.Parallel()
		result := styles.RenderSelector(false)
		if strings.Contains(result, "▶") {
			t.Error("expected selector to not contain ▶ when not selected")
		}
		if result != "  " {
			t.Errorf("expected two spaces, got %q", result)
		}
	})
}

func TestRenderLabel(t *testing.T) {
	t.Parallel()

	styles := DefaultStyles()

	t.Run("focused", func(t *testing.T) {
		t.Parallel()
		result := styles.RenderLabel("Name", true)
		if !strings.Contains(result, "Name") {
			t.Error("expected result to contain label text")
		}
	})

	t.Run("unfocused", func(t *testing.T) {
		t.Parallel()
		result := styles.RenderLabel("Name", false)
		if !strings.Contains(result, "Name") {
			t.Error("expected result to contain label text")
		}
	})
}

func TestRenderFieldRow(t *testing.T) {
	t.Parallel()

	styles := DefaultStyles()

	t.Run("selected", func(t *testing.T) {
		t.Parallel()
		result := styles.RenderFieldRow(true, "Name:", "test value")
		if !strings.Contains(result, "▶") {
			t.Error("expected selector in selected row")
		}
		if !strings.Contains(result, "Name:") {
			t.Error("expected label in row")
		}
		if !strings.Contains(result, "test value") {
			t.Error("expected value in row")
		}
	})

	t.Run("not selected", func(t *testing.T) {
		t.Parallel()
		result := styles.RenderFieldRow(false, "Name:", "test value")
		if strings.Contains(result, "▶") {
			t.Error("expected no selector in unselected row")
		}
	})
}

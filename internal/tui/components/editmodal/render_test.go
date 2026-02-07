package editmodal

import (
	"fmt"
	"strings"
	"testing"
)

func TestBase_RenderField_Focused(t *testing.T) {
	t.Parallel()
	b := Base{Cursor: 1, Styles: DefaultStyles()}
	result := b.RenderField(1, "Label:", "value")

	if !strings.Contains(result, "Label:") {
		t.Error("expected label text in output")
	}
	if !strings.Contains(result, "value") {
		t.Error("expected value text in output")
	}
	// When focused, should contain the selector character
	if !strings.Contains(result, "▶") {
		t.Error("expected ▶ selector for focused field")
	}
}

func TestBase_RenderField_Unfocused(t *testing.T) {
	t.Parallel()
	b := Base{Cursor: 0, Styles: DefaultStyles()}
	result := b.RenderField(1, "Label:", "value")

	if !strings.Contains(result, "Label:") {
		t.Error("expected label text in output")
	}
	if !strings.Contains(result, "value") {
		t.Error("expected value text in output")
	}
	// When not focused, should NOT contain the selector character
	if strings.Contains(result, "▶") {
		t.Error("expected no ▶ selector for unfocused field")
	}
}

func TestBase_RenderModal(t *testing.T) {
	t.Parallel()
	b := Base{Width: 80, Height: 30, Styles: DefaultStyles()}
	result := b.RenderModal("Test Title", "content here", "Esc: Close")

	if result == "" {
		t.Error("RenderModal returned empty string")
	}
	if !strings.Contains(result, "Test Title") {
		t.Error("expected title in modal output")
	}
	if !strings.Contains(result, "content here") {
		t.Error("expected content in modal output")
	}
}

func TestBase_RenderError_WithError(t *testing.T) {
	t.Parallel()
	b := Base{Err: fmt.Errorf("something went wrong"), Styles: DefaultStyles()}
	result := b.RenderError()

	if result == "" {
		t.Error("RenderError returned empty string when error is set")
	}
	if !strings.Contains(result, "something went wrong") {
		t.Errorf("expected error message in output, got %q", result)
	}
}

func TestBase_RenderError_NoError(t *testing.T) {
	t.Parallel()
	b := Base{Styles: DefaultStyles()}
	result := b.RenderError()

	if result != "" {
		t.Errorf("RenderError returned %q when no error is set, want empty", result)
	}
}

func TestBase_RenderSavingFooter_Saving(t *testing.T) {
	t.Parallel()
	b := Base{Saving: true}
	result := b.RenderSavingFooter("Normal footer")

	if result != "Saving..." {
		t.Errorf("RenderSavingFooter while saving = %q, want %q", result, "Saving...")
	}
}

func TestBase_RenderSavingFooter_NotSaving(t *testing.T) {
	t.Parallel()
	b := Base{Saving: false}
	result := b.RenderSavingFooter("Normal footer")

	if result != "Normal footer" {
		t.Errorf("RenderSavingFooter not saving = %q, want %q", result, "Normal footer")
	}
}

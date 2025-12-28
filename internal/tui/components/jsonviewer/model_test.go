package jsonviewer

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil)
	if m.visible {
		t.Error("visible should be false initially")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil).SetSize(100, 40)
	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestModel_Visible(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil)
	if m.Visible() {
		t.Error("Visible() should be false initially")
	}

	m.visible = true
	if !m.Visible() {
		t.Error("Visible() should be true")
	}
}

func TestModel_Close(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil)
	m.visible = true
	m.data = map[string]any{"test": true}

	m = m.Close()

	if m.Visible() {
		t.Error("Visible() should be false after Close()")
	}
	if m.data != nil {
		t.Error("data should be nil after Close()")
	}
}

func TestModel_View_NotVisible(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil).SetSize(80, 24)
	view := m.View()
	if view != "" {
		t.Errorf("View() = %q, want empty string when not visible", view)
	}
}

func TestModel_View_Loading(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil).SetSize(80, 24)
	m.visible = true
	m.isLoading = true

	view := m.View()
	if view == "" {
		t.Error("View() should not be empty when loading")
	}
}

func TestModel_View_WithData(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil).SetSize(80, 30)
	m.visible = true
	m.endpoint = "Test.GetStatus"
	m.data = map[string]any{
		"output": true,
		"power":  45.2,
	}
	m.viewport.SetContent(m.formatJSON())

	view := m.View()
	if view == "" {
		t.Error("View() should not be empty with data")
	}
}

func TestHighlightJSON(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil)

	// Simple JSON
	input := `{"key": "value", "num": 42, "bool": true, "nil": null}`
	result := m.highlightJSON(input)

	// Should have some output (with ANSI codes)
	if len(result) < len(input) {
		t.Errorf("highlighted output too short: %d vs %d", len(result), len(input))
	}
}

func TestGetChromaStyle(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil)

	// Should return a non-nil style
	style := m.getChromaStyle()
	if style == nil {
		t.Error("getChromaStyle() should return a non-nil style")
	}
}

func TestFormatJSON(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil)

	// nil data
	m.data = nil
	result := m.formatJSON()
	if result == "" {
		t.Error("formatJSON() should not be empty for nil data")
	}

	// valid data
	m.data = map[string]any{
		"test": true,
		"num":  123,
	}
	result = m.formatJSON()
	if result == "" {
		t.Error("formatJSON() should not be empty for valid data")
	}
}

func TestModel_renderNav(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil).SetSize(80, 24)
	m.endpoints = []string{"A", "B", "C"}
	m.endpointIdx = 1
	m.endpoint = "B"

	nav := m.renderNav()
	if nav == "" {
		t.Error("renderNav() should not be empty")
	}
}

func TestModel_renderFooter(t *testing.T) {
	t.Parallel()
	m := New(context.Background(), nil)
	footer := m.renderFooter()
	if footer == "" {
		t.Error("renderFooter() should not be empty")
	}
}

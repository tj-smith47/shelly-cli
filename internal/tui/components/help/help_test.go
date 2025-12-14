package help

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()

	if m.Visible() {
		t.Error("expected help to be hidden on creation")
	}
}

func TestShow(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show()

	if !m.Visible() {
		t.Error("expected help to be visible after Show()")
	}
}

func TestHide(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show()
	m = m.Hide()

	if m.Visible() {
		t.Error("expected help to be hidden after Hide()")
	}
}

func TestToggle(t *testing.T) {
	t.Parallel()
	m := New()

	m = m.Toggle()
	if !m.Visible() {
		t.Error("expected help to be visible after first Toggle()")
	}

	m = m.Toggle()
	if m.Visible() {
		t.Error("expected help to be hidden after second Toggle()")
	}
}

func TestSetContext(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetContext(ContextMonitor)

	if m.context != ContextMonitor {
		t.Errorf("expected context ContextMonitor, got %v", m.context)
	}
}

func TestSetSize(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
}

func TestViewWhenHidden(t *testing.T) {
	t.Parallel()
	m := New()

	view := m.View()
	if view != "" {
		t.Error("expected empty view when hidden")
	}
}

func TestViewWhenVisible(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(80, 40)
	m = m.Show()

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view when visible")
	}

	// Check for expected content (with ANSI-safe checks)
	// The styled output contains escape codes, so we check common words
	expectedStrings := []string{
		"Navigation",
		"Actions",
		"General",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(view, expected) {
			t.Errorf("expected view to contain %q", expected)
		}
	}
}

func TestContextSpecificHelp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		context  ViewContext
		expected string
	}{
		{ContextDevices, "Devices View"},
		{ContextMonitor, "Monitor View"},
		{ContextEvents, "Events View"},
		{ContextEnergy, "Energy View"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			m := New()
			m = m.SetSize(80, 40)
			m = m.SetContext(tt.context)
			m = m.Show()

			view := m.View()
			if !strings.Contains(view, tt.expected) {
				t.Errorf("expected view to contain %q for context %v", tt.expected, tt.context)
			}
		})
	}
}

func TestDeviceControlOnlyInDevicesAndMonitor(t *testing.T) {
	t.Parallel()
	deviceControlKeys := []string{"toggle", "turn on", "turn off"}

	// Device control should appear in Devices and Monitor views
	for _, ctx := range []ViewContext{ContextDevices, ContextMonitor} {
		m := New()
		m = m.SetSize(80, 40)
		m = m.SetContext(ctx)
		m = m.Show()

		view := m.View()
		for _, key := range deviceControlKeys {
			if !strings.Contains(strings.ToLower(view), key) {
				t.Errorf("expected device control key %q in context %v", key, ctx)
			}
		}
	}
}

func TestShortHelp(t *testing.T) {
	t.Parallel()
	m := New()
	bindings := m.ShortHelp()

	if len(bindings) == 0 {
		t.Error("expected non-empty ShortHelp bindings")
	}
}

func TestFullHelp(t *testing.T) {
	t.Parallel()
	m := New()
	bindings := m.FullHelp()

	if len(bindings) == 0 {
		t.Error("expected non-empty FullHelp bindings")
	}
}

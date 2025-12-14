package search

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()

	if m.IsActive() {
		t.Error("expected search to be inactive on creation")
	}

	if m.Value() != "" {
		t.Error("expected empty value on creation")
	}
}

func TestNewWithFilter(t *testing.T) {
	t.Parallel()
	m := NewWithFilter("test-filter")

	if m.Value() != "test-filter" {
		t.Errorf("expected value 'test-filter', got %q", m.Value())
	}
}

func TestActivate(t *testing.T) {
	t.Parallel()
	m := New()

	m, _ = m.Activate()

	if !m.IsActive() {
		t.Error("expected search to be active after Activate()")
	}
}

func TestDeactivate(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()
	m = m.Deactivate()

	if m.IsActive() {
		t.Error("expected search to be inactive after Deactivate()")
	}
}

func TestClear(t *testing.T) {
	t.Parallel()
	m := NewWithFilter("some-filter")
	m = m.Clear()

	if m.Value() != "" {
		t.Errorf("expected empty value after Clear(), got %q", m.Value())
	}
}

func TestMatchesFilter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filter   string
		input    string
		expected bool
	}{
		{"empty filter matches all", "", "anything", true},
		{"exact match", "test", "test", true},
		{"case insensitive", "TEST", "test", true},
		{"partial match", "foo", "foobar", true},
		{"no match", "xyz", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := NewWithFilter(tt.filter)
			if got := m.MatchesFilter(tt.input); got != tt.expected {
				t.Errorf("MatchesFilter(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSetWidth(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetWidth(100)

	// Width should be stored (we can't easily verify internal textInput width)
	// but the function should not panic
	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
}

func TestUpdateWhenInactive(t *testing.T) {
	t.Parallel()
	m := New()

	// Updates should be no-ops when inactive
	_, cmd := m.Update(tea.KeyPressMsg{})

	if cmd != nil {
		t.Error("expected nil command when inactive")
	}
}

func TestViewWhenInactive(t *testing.T) {
	t.Parallel()
	m := New()

	view := m.View()
	if view != "" {
		t.Error("expected empty view when inactive")
	}
}

func TestViewWhenActive(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetWidth(80)
	m, _ = m.Activate()

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view when active")
	}
}

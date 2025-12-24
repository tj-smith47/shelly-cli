// Package tabs provides the tab bar component for the TUI.
package tabs

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		tabs       []string
		active     int
		wantActive int
	}{
		{
			name:       "valid active index",
			tabs:       []string{"A", "B", "C"},
			active:     1,
			wantActive: 1,
		},
		{
			name:       "negative active index defaults to 0",
			tabs:       []string{"A", "B", "C"},
			active:     -1,
			wantActive: 0,
		},
		{
			name:       "out of bounds active index defaults to 0",
			tabs:       []string{"A", "B", "C"},
			active:     10,
			wantActive: 0,
		},
		{
			name:       "zero active index",
			tabs:       []string{"A", "B", "C"},
			active:     0,
			wantActive: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := New(tt.tabs, tt.active)
			if got := m.Active(); got != tt.wantActive {
				t.Errorf("New() active = %d, want %d", got, tt.wantActive)
			}
		})
	}
}

func TestModel_SetActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		initial   int
		setTo     int
		wantOK    bool
		wantFinal int
	}{
		{"valid index 1", 0, 1, true, 1},
		{"valid index 2", 0, 2, true, 2},
		{"valid index 0", 1, 0, true, 0},
		{"negative index ignored", 1, -1, false, 1},
		{"out of bounds ignored", 1, 5, false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := New([]string{"A", "B", "C"}, tt.initial)
			m = m.SetActive(tt.setTo)
			got := m.Active()

			if got != tt.wantFinal {
				t.Errorf("SetActive(%d) = %d, want %d", tt.setTo, got, tt.wantFinal)
			}
		})
	}
}

func TestModel_ActiveName(t *testing.T) {
	t.Parallel()

	m := New([]string{"Dashboard", "Config", "Fleet"}, 1)

	if got := m.ActiveName(); got != "Config" {
		t.Errorf("ActiveName() = %q, want %q", got, "Config")
	}

	// Test with first tab
	m = m.SetActive(0)
	if got := m.ActiveName(); got != "Dashboard" {
		t.Errorf("ActiveName() = %q, want %q", got, "Dashboard")
	}
}

func TestModel_Next(t *testing.T) {
	t.Parallel()

	m := New([]string{"A", "B", "C"}, 0)

	// Move forward
	m = m.Next()
	if got := m.Active(); got != 1 {
		t.Errorf("Next() active = %d, want 1", got)
	}

	m = m.Next()
	if got := m.Active(); got != 2 {
		t.Errorf("Next() active = %d, want 2", got)
	}

	// Wrap around
	m = m.Next()
	if got := m.Active(); got != 0 {
		t.Errorf("Next() should wrap to 0, got %d", got)
	}
}

func TestModel_Prev(t *testing.T) {
	t.Parallel()

	m := New([]string{"A", "B", "C"}, 0)

	// Wrap backwards
	m = m.Prev()
	if got := m.Active(); got != 2 {
		t.Errorf("Prev() should wrap to 2, got %d", got)
	}

	m = m.Prev()
	if got := m.Active(); got != 1 {
		t.Errorf("Prev() active = %d, want 1", got)
	}

	m = m.Prev()
	if got := m.Active(); got != 0 {
		t.Errorf("Prev() active = %d, want 0", got)
	}
}

func TestModel_Count(t *testing.T) {
	t.Parallel()

	tests := []struct {
		tabs []string
		want int
	}{
		{[]string{"A"}, 1},
		{[]string{"A", "B"}, 2},
		{[]string{"A", "B", "C", "D", "E"}, 5},
	}

	for _, tt := range tests {
		m := New(tt.tabs, 0)
		if got := m.Count(); got != tt.want {
			t.Errorf("Count() = %d, want %d", got, tt.want)
		}
	}
}

func TestModel_View(t *testing.T) {
	t.Parallel()

	m := New([]string{"Dashboard", "Config"}, 0)
	view := m.View()

	// Should contain tab names
	if !strings.Contains(view, "Dashboard") {
		t.Error("View() should contain 'Dashboard'")
	}
	if !strings.Contains(view, "Config") {
		t.Error("View() should contain 'Config'")
	}

	// Should contain number hints
	if !strings.Contains(view, "1") {
		t.Error("View() should contain number hint '1'")
	}
	if !strings.Contains(view, "2") {
		t.Error("View() should contain number hint '2'")
	}

	// Should contain separator
	if !strings.Contains(view, "|") {
		t.Error("View() should contain separator '|'")
	}
}

func TestModel_Init(t *testing.T) {
	t.Parallel()

	m := New([]string{"A", "B"}, 0)
	if cmd := m.Init(); cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_Update(t *testing.T) {
	t.Parallel()

	m := New([]string{"A", "B"}, 0)
	updated, cmd := m.Update(nil)

	if cmd != nil {
		t.Error("Update(nil) should return nil cmd")
	}
	if updated.Active() != m.Active() {
		t.Error("Update(nil) should not change state")
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()

	styles := DefaultStyles()

	// Just verify the styles are non-empty
	if styles.Container.GetBorderStyle() == (lipgloss.Border{}) {
		t.Error("Container style should have a border")
	}
}

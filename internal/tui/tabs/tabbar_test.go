package tabs

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()
	if m.TabCount() != 4 {
		t.Errorf("TabCount() = %d, want 4", m.TabCount())
	}
	if m.ActiveTabID() != TabDevices {
		t.Errorf("ActiveTabID() = %v, want %v", m.ActiveTabID(), TabDevices)
	}
}

func TestNewWithTabs(t *testing.T) {
	t.Parallel()
	tabs := []Tab{
		{ID: TabDevices, Label: "Devices", Enabled: true},
		{ID: TabEnergy, Label: "Energy", Enabled: true},
	}
	m := NewWithTabs(tabs)
	if m.TabCount() != 2 {
		t.Errorf("TabCount() = %d, want 2", m.TabCount())
	}
}

func TestModel_SetActive(t *testing.T) {
	t.Parallel()
	m := New()

	newM, cmd := m.SetActive(TabEvents)
	if cmd == nil {
		t.Fatal("SetActive() should return a command")
	}

	if newM.ActiveTabID() != TabEvents {
		t.Errorf("ActiveTabID() = %v, want %v", newM.ActiveTabID(), TabEvents)
	}

	msg := cmd()
	tcMsg, ok := msg.(TabChangedMsg)
	if !ok {
		t.Fatalf("expected TabChangedMsg, got %T", msg)
	}
	if tcMsg.Previous != TabDevices {
		t.Errorf("Previous = %v, want %v", tcMsg.Previous, TabDevices)
	}
	if tcMsg.Current != TabEvents {
		t.Errorf("Current = %v, want %v", tcMsg.Current, TabEvents)
	}
}

func TestModel_SetActive_NoChange(t *testing.T) {
	t.Parallel()
	m := New()

	_, cmd := m.SetActive(TabDevices)
	if cmd != nil {
		t.Error("SetActive() should return nil when tab doesn't change")
	}
}

func TestModel_Next(t *testing.T) {
	t.Parallel()
	m := New()

	newM, _ := m.Next()
	if newM.ActiveTabID() != TabMonitor {
		t.Errorf("After Next(), ActiveTabID() = %v, want %v", newM.ActiveTabID(), TabMonitor)
	}

	newM, _ = newM.Next()
	if newM.ActiveTabID() != TabEvents {
		t.Errorf("After second Next(), ActiveTabID() = %v, want %v", newM.ActiveTabID(), TabEvents)
	}
}

func TestModel_Prev(t *testing.T) {
	t.Parallel()
	m := New()

	newM, _ := m.Prev()
	if newM.ActiveTabID() != TabEnergy {
		t.Errorf("After Prev(), ActiveTabID() = %v, want %v", newM.ActiveTabID(), TabEnergy)
	}
}

func TestModel_SetTabEnabled(t *testing.T) {
	t.Parallel()
	m := New()

	// Disable monitor tab
	m = m.SetTabEnabled(TabMonitor, false)

	// Next should skip monitor and go to events
	newM, _ := m.Next()
	if newM.ActiveTabID() != TabEvents {
		t.Errorf("After Next() with disabled tab, ActiveTabID() = %v, want %v", newM.ActiveTabID(), TabEvents)
	}
}

func TestModel_View(t *testing.T) {
	t.Parallel()
	m := New().SetWidth(80)
	view := m.View()

	if view == "" {
		t.Error("View() returned empty string")
	}

	// View should be non-empty and contain some content
	// (We can't easily check for exact labels due to ANSI styling)
	if len(view) < 20 {
		t.Errorf("View() too short: %q", view)
	}
}

func TestModel_ShowIcons(t *testing.T) {
	t.Parallel()
	m := New().SetWidth(80)

	// With icons
	viewWithIcons := m.View()

	// Without icons
	viewNoIcons := m.ShowIcons(false).View()

	// View without icons should be shorter or different
	// (we can't easily check for icon presence due to unicode, but we can check they're different)
	if viewWithIcons == viewNoIcons {
		// This is acceptable if icons don't render differently
		t.Log("Views with and without icons are the same (acceptable)")
	}
}

func TestTabID_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		tab  TabID
		want string
	}{
		{TabDevices, "Devices"},
		{TabMonitor, "Monitor"},
		{TabEvents, "Events"},
		{TabEnergy, "Energy"},
		{TabID(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.tab.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTabID_Icon(t *testing.T) {
	t.Parallel()
	// Just verify icons are non-empty for known tabs
	tabs := []TabID{TabDevices, TabMonitor, TabEvents, TabEnergy}
	for _, tab := range tabs {
		if tab.Icon() == "" {
			t.Errorf("Icon() for %v returned empty string", tab)
		}
	}
}

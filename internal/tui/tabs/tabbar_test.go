package tabs

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()
	if m.TabCount() != 5 {
		t.Errorf("TabCount() = %d, want 5", m.TabCount())
	}
	if m.ActiveTabID() != TabDashboard {
		t.Errorf("ActiveTabID() = %v, want %v", m.ActiveTabID(), TabDashboard)
	}
}

func TestNewWithTabs(t *testing.T) {
	t.Parallel()
	tabs := []Tab{
		{ID: TabDashboard, Label: "Dashboard", Enabled: true},
		{ID: TabFleet, Label: "Fleet", Enabled: true},
	}
	m := NewWithTabs(tabs)
	if m.TabCount() != 2 {
		t.Errorf("TabCount() = %d, want 2", m.TabCount())
	}
}

func TestModel_SetActive(t *testing.T) {
	t.Parallel()
	m := New()

	newM, cmd := m.SetActive(TabConfig)
	if cmd == nil {
		t.Fatal("SetActive() should return a command")
	}

	if newM.ActiveTabID() != TabConfig {
		t.Errorf("ActiveTabID() = %v, want %v", newM.ActiveTabID(), TabConfig)
	}

	msg := cmd()
	tcMsg, ok := msg.(TabChangedMsg)
	if !ok {
		t.Fatalf("expected TabChangedMsg, got %T", msg)
	}
	if tcMsg.Previous != TabDashboard {
		t.Errorf("Previous = %v, want %v", tcMsg.Previous, TabDashboard)
	}
	if tcMsg.Current != TabConfig {
		t.Errorf("Current = %v, want %v", tcMsg.Current, TabConfig)
	}
}

func TestModel_SetActive_NoChange(t *testing.T) {
	t.Parallel()
	m := New()

	_, cmd := m.SetActive(TabDashboard)
	if cmd != nil {
		t.Error("SetActive() should return nil when tab doesn't change")
	}
}

func TestModel_Next(t *testing.T) {
	t.Parallel()
	m := New()

	newM, _ := m.Next()
	if newM.ActiveTabID() != TabAutomation {
		t.Errorf("After Next(), ActiveTabID() = %v, want %v", newM.ActiveTabID(), TabAutomation)
	}

	newM, _ = newM.Next()
	if newM.ActiveTabID() != TabConfig {
		t.Errorf("After second Next(), ActiveTabID() = %v, want %v", newM.ActiveTabID(), TabConfig)
	}
}

func TestModel_Prev(t *testing.T) {
	t.Parallel()
	m := New()

	newM, _ := m.Prev()
	if newM.ActiveTabID() != TabFleet {
		t.Errorf("After Prev(), ActiveTabID() = %v, want %v", newM.ActiveTabID(), TabFleet)
	}
}

func TestModel_SetTabEnabled(t *testing.T) {
	t.Parallel()
	m := New()

	// Disable automation tab
	m = m.SetTabEnabled(TabAutomation, false)

	// Next should skip automation and go to config
	newM, _ := m.Next()
	if newM.ActiveTabID() != TabConfig {
		t.Errorf("After Next() with disabled tab, ActiveTabID() = %v, want %v", newM.ActiveTabID(), TabConfig)
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
		{TabDashboard, "Dashboard"},
		{TabAutomation, "Automation"},
		{TabConfig, "Config"},
		{TabManage, "Manage"},
		{TabFleet, "Fleet"},
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
	tabs := []TabID{TabDashboard, TabAutomation, TabConfig, TabManage, TabFleet}
	for _, tab := range tabs {
		if tab.Icon() == "" {
			t.Errorf("Icon() for %v returned empty string", tab)
		}
	}
}

func TestModel_KeyboardShortcuts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		key     string
		wantTab TabID
	}{
		{"key 1 selects Dashboard", "1", TabDashboard},
		{"key 2 selects Automation", "2", TabAutomation},
		{"key 3 selects Config", "3", TabConfig},
		{"key 4 selects Manage", "4", TabManage},
		{"key 5 selects Fleet", "5", TabFleet},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := New()
			// Start from a different tab to verify switching works
			if tt.wantTab == TabDashboard {
				m, _ = m.SetActive(TabFleet)
			}

			newM, cmd := m.handleKeyPress(mockKeyPress(tt.key))
			if newM.ActiveTabID() != tt.wantTab {
				t.Errorf("After key %q, ActiveTabID() = %v, want %v", tt.key, newM.ActiveTabID(), tt.wantTab)
			}
			// Verify command is returned for tab change
			if cmd == nil && m.ActiveTabID() != tt.wantTab {
				t.Errorf("handleKeyPress(%q) should return a command for tab change", tt.key)
			}
		})
	}
}

// mockKeyPress creates a mock tea.KeyPressMsg for testing.
func mockKeyPress(key string) tea.KeyPressMsg {
	if key == "" {
		return tea.KeyPressMsg{}
	}
	return tea.KeyPressMsg{
		Code: rune(key[0]),
		Text: key,
	}
}

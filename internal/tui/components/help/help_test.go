package help

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()

	if m.Visible() {
		t.Error("expected help to be hidden on creation")
	}
	if m.keyMap == nil {
		t.Error("expected keyMap to be initialized")
	}
}

func TestShow(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show()

	if !m.Visible() {
		t.Error("expected help to be visible after Show()")
	}
	if m.scrollOffset != 0 {
		t.Error("expected scrollOffset to be reset on Show()")
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
	m = m.SetContext(keys.ContextDevices)

	if m.context != keys.ContextDevices {
		t.Errorf("expected context ContextDevices, got %v", m.context)
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

func TestSetKeyMap(t *testing.T) {
	t.Parallel()
	m := New()
	km := keys.NewContextMap()
	m = m.SetKeyMap(km)

	if m.keyMap != km {
		t.Error("expected keyMap to be set")
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
	m = m.SetSize(120, 60)
	m = m.SetContext(keys.ContextDevices)
	m = m.Show()

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view when visible")
	}

	// Check for expected content
	if !strings.Contains(view, "Help") {
		t.Error("expected view to contain 'Help'")
	}
}

func TestViewCompact(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(120, 40)

	view := m.ViewCompact()
	if view == "" {
		t.Error("expected non-empty compact view")
	}

	// Check for expected keybindings
	expectedStrings := []string{
		"j/k",
		"toggle",
		"quit",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(view, expected) {
			t.Errorf("expected compact view to contain %q", expected)
		}
	}
}

func TestViewHeight(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(80, 40)

	height := m.ViewHeight()
	if height != 2 {
		t.Errorf("expected ViewHeight 2, got %d", height)
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

func TestGetContextBindings(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetContext(keys.ContextDevices)

	sections := m.getContextBindings()
	if len(sections) == 0 {
		t.Error("expected non-empty bindings")
	}

	// Should have device-specific and global sections
	hasDevices := false
	hasGlobal := false
	for _, s := range sections {
		if s.Name == "Devices" {
			hasDevices = true
		}
		if s.Name == "Global" {
			hasGlobal = true
		}
	}

	if !hasDevices {
		t.Error("expected Devices section")
	}
	if !hasGlobal {
		t.Error("expected Global section")
	}
}

func TestFormatBindings(t *testing.T) {
	t.Parallel()
	m := New()

	bindings := []keys.KeyBinding{
		{Key: "q", Action: keys.ActionQuit, Desc: "Quit"},
		{Key: "?", Action: keys.ActionHelp, Desc: "Show help"},
	}

	formatted := m.formatBindings(bindings)
	if len(formatted) == 0 {
		t.Error("expected non-empty formatted bindings")
	}

	joined := strings.Join(formatted, "\n")
	if !strings.Contains(joined, "Quit") {
		t.Error("expected formatted output to contain 'Quit'")
	}
}

func TestSortBindings(t *testing.T) {
	t.Parallel()
	m := New()

	bindings := []keys.KeyBinding{
		{Key: "z", Action: keys.ActionQuit, Desc: "Quit"},
		{Key: "a", Action: keys.ActionHelp, Desc: "Help"},
		{Key: "m", Action: keys.ActionToggle, Desc: "Toggle"},
	}

	sorted := m.sortBindings(bindings)
	if sorted[0].Key != "a" {
		t.Errorf("expected first key to be 'a', got %q", sorted[0].Key)
	}
	if sorted[2].Key != "z" {
		t.Errorf("expected last key to be 'z', got %q", sorted[2].Key)
	}
}

func TestBindingSection(t *testing.T) {
	t.Parallel()

	section := BindingSection{
		Name: "Test",
		Bindings: []keys.KeyBinding{
			{Key: "q", Action: keys.ActionQuit, Desc: "Quit"},
		},
	}

	if section.Name != "Test" {
		t.Errorf("expected Name 'Test', got %q", section.Name)
	}
	if len(section.Bindings) != 1 {
		t.Errorf("expected 1 binding, got %d", len(section.Bindings))
	}
}

const testFilterToggle = "toggle"

func TestSearchActivation(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(120, 60)
	m = m.SetContext(keys.ContextDevices)
	m = m.Show()

	if m.Searching() {
		t.Error("should not be searching initially")
	}

	// Press / to activate search
	m, _ = m.Update(tea.KeyPressMsg{Code: '/'})

	if !m.Searching() {
		t.Error("should be searching after pressing /")
	}
}

func TestSearchEscapeExitsSearch(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(120, 60)
	m = m.SetContext(keys.ContextDevices)
	m = m.Show()

	// Activate search
	m, _ = m.Update(tea.KeyPressMsg{Code: '/'})
	if !m.Searching() {
		t.Fatal("should be searching")
	}

	// Press Escape to exit search
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if m.Searching() {
		t.Error("should not be searching after Escape")
	}
	if !m.Visible() {
		t.Error("help should still be visible after exiting search")
	}
	if m.searchFilter != "" {
		t.Error("search filter should be cleared after Escape")
	}
}

func TestSearchEnterConfirmsFilter(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(120, 60)
	m = m.SetContext(keys.ContextDevices)
	m = m.Show()

	// Activate search and type something
	m, _ = m.Update(tea.KeyPressMsg{Code: '/'})
	m.searchInput.SetValue(testFilterToggle)
	m.searchFilter = testFilterToggle

	// Press Enter to confirm
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if m.Searching() {
		t.Error("should not be in search mode after Enter")
	}
	if m.searchFilter != testFilterToggle {
		t.Errorf("search filter should be preserved after Enter, got %q", m.searchFilter)
	}
	if !m.Visible() {
		t.Error("help should still be visible")
	}
}

func TestFilterSections(t *testing.T) {
	t.Parallel()
	m := New()
	m.searchFilter = "quit"

	sections := []BindingSection{
		{
			Name: "Global",
			Bindings: []keys.KeyBinding{
				{Key: "q", Action: keys.ActionQuit, Desc: "Quit"},
				{Key: "?", Action: keys.ActionHelp, Desc: "Show help"},
				{Key: "/", Action: keys.ActionFilter, Desc: "Filter"},
			},
		},
	}

	filtered := m.filterSections(sections)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 section, got %d", len(filtered))
	}
	if len(filtered[0].Bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(filtered[0].Bindings))
	}
	if filtered[0].Bindings[0].Key != "q" {
		t.Errorf("expected key 'q', got %q", filtered[0].Bindings[0].Key)
	}
}

func TestFilterSectionsMatchesKey(t *testing.T) {
	t.Parallel()
	m := New()
	m.searchFilter = "ctrl"

	sections := []BindingSection{
		{
			Name: "Global",
			Bindings: []keys.KeyBinding{
				{Key: "ctrl+r", Action: keys.ActionRefreshAll, Desc: "Refresh all"},
				{Key: "q", Action: keys.ActionQuit, Desc: "Quit"},
			},
		},
	}

	filtered := m.filterSections(sections)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 section, got %d", len(filtered))
	}
	if len(filtered[0].Bindings) != 1 {
		t.Errorf("expected 1 binding matching 'ctrl', got %d", len(filtered[0].Bindings))
	}
}

func TestFilterSectionsNoMatch(t *testing.T) {
	t.Parallel()
	m := New()
	m.searchFilter = "xyznonexistent"

	sections := []BindingSection{
		{
			Name: "Global",
			Bindings: []keys.KeyBinding{
				{Key: "q", Action: keys.ActionQuit, Desc: "Quit"},
			},
		},
	}

	filtered := m.filterSections(sections)

	if len(filtered) != 0 {
		t.Errorf("expected 0 sections, got %d", len(filtered))
	}
}

func TestFilterSectionsCaseInsensitive(t *testing.T) {
	t.Parallel()
	m := New()
	m.searchFilter = "QUIT"

	sections := []BindingSection{
		{
			Name: "Global",
			Bindings: []keys.KeyBinding{
				{Key: "q", Action: keys.ActionQuit, Desc: "Quit"},
				{Key: "?", Action: keys.ActionHelp, Desc: "Show help"},
			},
		},
	}

	filtered := m.filterSections(sections)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 section, got %d", len(filtered))
	}
	if len(filtered[0].Bindings) != 1 {
		t.Errorf("expected 1 binding, got %d", len(filtered[0].Bindings))
	}
}

func TestViewWithSearchFilter(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(120, 60)
	m = m.SetContext(keys.ContextDevices)
	m = m.Show()

	// Set a filter
	m.searchFilter = testFilterToggle

	view := m.View()
	if !strings.Contains(view, testFilterToggle) {
		t.Error("view should contain filter text 'toggle'")
	}
}

func TestViewNoMatchesMessage(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(120, 60)
	m = m.SetContext(keys.ContextDevices)
	m = m.Show()

	m.searchFilter = "xyznonexistent"

	view := m.View()
	if !strings.Contains(view, "No matching") {
		t.Error("view should show 'No matching' message for empty results")
	}
}

func TestShowClearsSearch(t *testing.T) {
	t.Parallel()
	m := New()
	m.searchFilter = "old search"
	m.searching = true

	m = m.Show()

	if m.searchFilter != "" {
		t.Error("Show should clear search filter")
	}
	if m.Searching() {
		t.Error("Show should exit search mode")
	}
}

func TestHideClearsSearch(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show()
	m.searchFilter = "test"
	m.searching = true

	m = m.Hide()

	if m.searchFilter != "" {
		t.Error("Hide should clear search filter")
	}
	if m.Searching() {
		t.Error("Hide should exit search mode")
	}
}

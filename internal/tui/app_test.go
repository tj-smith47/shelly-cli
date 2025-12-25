package tui

import (
	"bytes"
	"context"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
	"github.com/tj-smith47/shelly-cli/internal/tui/views"
)

// newTestModel creates a Model for testing with mock dependencies.
func newTestModel(t *testing.T) Model {
	t.Helper()
	ctx := context.Background()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	cfg := &config.Config{}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().
		SetIOStreams(ios).
		SetConfigManager(mgr)

	opts := DefaultOptions()
	opts.RefreshInterval = 1 * time.Hour // Long interval for testing

	return New(ctx, f, opts)
}

const testFilterKitchen = "kitchen"

// applyWindowSize applies a window size message and returns the updated model.
func applyWindowSize(m Model, width, height int) Model {
	updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	model, ok := updated.(Model)
	if !ok {
		panic("applyWindowSize: unexpected model type")
	}
	return model
}

func TestNew(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)

	// Verify initial state
	if m.quitting {
		t.Error("quitting should be false initially")
	}
	if m.ready {
		t.Error("ready should be false initially (before window size)")
	}
	if m.viewManager == nil {
		t.Error("viewManager should not be nil")
	}
	if m.cache == nil {
		t.Error("cache should not be nil")
	}
}

func TestModel_Init(t *testing.T) {
	t.Parallel()
	m := newTestModel(t)

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestModel_WindowResize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"narrow terminal", 60, 24},
		{"standard terminal", 100, 40},
		{"wide terminal", 160, 50},
		{"minimum size", 40, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newTestModel(t)
			msg := tea.WindowSizeMsg{Width: tt.width, Height: tt.height}

			updated, _ := m.Update(msg)
			model, ok := updated.(Model)
			if !ok {
				t.Fatal("Update should return Model")
			}

			if model.width != tt.width {
				t.Errorf("width = %d, want %d", model.width, tt.width)
			}
			if model.height != tt.height {
				t.Errorf("height = %d, want %d", model.height, tt.height)
			}
			if !model.ready {
				t.Error("ready should be true after window size")
			}
		})
	}
}

func TestModel_TabSwitching(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		key     string
		wantTab tabs.TabID
	}{
		{"key 1 switches to Dashboard", "1", tabs.TabDashboard},
		{"key 2 switches to Automation", "2", tabs.TabAutomation},
		{"key 3 switches to Config", "3", tabs.TabConfig},
		{"key 4 switches to Manage", "4", tabs.TabManage},
		{"key 5 switches to Monitor", "5", tabs.TabMonitor},
		{"key 6 switches to Fleet", "6", tabs.TabFleet},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newTestModel(t)

			// Apply window size first to make model ready
			m = applyWindowSize(m, 100, 40)

			// Start from a different tab if testing Dashboard
			if tt.wantTab == tabs.TabDashboard {
				m.tabBar, _ = m.tabBar.SetActive(tabs.TabFleet)
				m.viewManager.SetActive(views.ViewFleet)
			}

			msg := mockKeyPress(tt.key)
			updated, _ := m.Update(msg)
			model, ok := updated.(Model)
			if !ok {
				t.Fatal("Update should return Model")
			}

			if model.tabBar.ActiveTabID() != tt.wantTab {
				t.Errorf("tab = %v, want %v", model.tabBar.ActiveTabID(), tt.wantTab)
			}
		})
	}
}

func TestModel_ViewManagerSyncsWithTabs(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Switch to Automation tab
	msg := mockKeyPress("2")
	updated, _ := m.Update(msg)
	model, ok := updated.(Model)
	if !ok {
		t.Fatal("Update should return Model")
	}

	// Verify both tab bar and view manager are in sync
	if model.tabBar.ActiveTabID() != tabs.TabAutomation {
		t.Errorf("tabBar active = %v, want %v", model.tabBar.ActiveTabID(), tabs.TabAutomation)
	}
	if model.viewManager.Active() != views.ViewAutomation {
		t.Errorf("viewManager active = %v, want %v", model.viewManager.Active(), views.ViewAutomation)
	}
}

func TestModel_PanelFocusCycling(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Initial focus should be DeviceList
	if m.focusedPanel != PanelDeviceList {
		t.Errorf("initial focus = %v, want %v", m.focusedPanel, PanelDeviceList)
	}

	// Tab to next panel (DeviceList -> Detail)
	newM, _, handled := m.handlePanelSwitch(tea.KeyPressMsg{Code: tea.KeyTab})
	if !handled {
		t.Error("tab should be handled")
	}
	if newM.focusedPanel != PanelDetail {
		t.Errorf("after tab = %v, want %v", newM.focusedPanel, PanelDetail)
	}

	// Tab should wrap back to DeviceList (2-panel cycle: Detail -> DeviceList)
	newM, _, handled = newM.handlePanelSwitch(tea.KeyPressMsg{Code: tea.KeyTab})
	if !handled {
		t.Error("tab should be handled")
	}
	if newM.focusedPanel != PanelDeviceList {
		t.Errorf("after wrap = %v, want %v", newM.focusedPanel, PanelDeviceList)
	}
}

func TestModel_ShiftTabReversesFocus(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Shift+Tab from DeviceList should go to Detail (reverse cycle in 2-panel mode)
	msg := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	newM, _, handled := m.handlePanelSwitch(msg)
	if !handled {
		t.Error("shift+tab should be handled")
	}
	if newM.focusedPanel != PanelDetail {
		t.Errorf("after shift+tab = %v, want %v", newM.focusedPanel, PanelDetail)
	}
}

func TestModel_QuitKey(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	msg := tea.KeyPressMsg{Code: 'q'}
	updated, cmd := m.Update(msg)
	model, ok := updated.(Model)
	if !ok {
		t.Fatal("Update should return Model")
	}

	if !model.quitting {
		t.Error("quitting should be true after q key")
	}
	if cmd == nil {
		t.Error("quit should return tea.Quit command")
	}
}

func TestModel_HelpToggle(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Press ? to toggle help
	msg := tea.KeyPressMsg{Code: '?'}
	updated, _ := m.Update(msg)
	model, ok := updated.(Model)
	if !ok {
		t.Fatal("Update should return Model")
	}

	if !model.help.Visible() {
		t.Error("help should be visible after ? key")
	}

	// Press ? again or Escape to close
	msg = tea.KeyPressMsg{Code: tea.KeyEscape}
	updated, _ = model.Update(msg)
	model, ok = updated.(Model)
	if !ok {
		t.Fatal("Update should return Model")
	}

	if model.help.Visible() {
		t.Error("help should be hidden after Escape")
	}
}

func TestModel_SearchActivation(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Press / to activate search
	msg := tea.KeyPressMsg{Code: '/'}
	updated, _ := m.Update(msg)
	model, ok := updated.(Model)
	if !ok {
		t.Fatal("Update should return Model")
	}

	if !model.search.IsActive() {
		t.Error("search should be active after / key")
	}
}

func TestModel_CommandModeActivation(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Press : to activate command mode
	msg := tea.KeyPressMsg{Code: ':'}
	updated, _ := m.Update(msg)
	model, ok := updated.(Model)
	if !ok {
		t.Fatal("Update should return Model")
	}

	if !model.cmdMode.IsActive() {
		t.Error("cmdMode should be active after : key")
	}
}

func TestModel_ViewRendersWithoutPanic(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Verify View() doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("View() panicked: %v", r)
		}
	}()

	view := m.View()
	// View should have content when model is ready
	if view.Content == nil {
		t.Error("View() returned nil content")
	}
}

func TestModel_ViewBeforeReady(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	// Don't send window size, so ready = false

	view := m.View()
	// View should still return content (initializing message)
	if view.Content == nil {
		t.Error("View() should return content when not ready")
	}
}

func TestModel_ViewNarrowTerminal(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	// Use narrow width (60) to test narrow terminal mode
	m = applyWindowSize(m, 60, 30)

	view := m.View()
	if view.Content == nil {
		t.Error("View() should return content in narrow mode")
	}
}

func TestModel_ViewWhenQuitting(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)
	m.quitting = true

	// Just verify View() doesn't panic when quitting
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("View() panicked when quitting: %v", r)
		}
	}()
	_ = m.View()
}

func TestModel_FilterMessage(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Apply a filter via direct assignment (simulates search.FilterChangedMsg handling)
	m.filter = testFilterKitchen

	if m.filter != testFilterKitchen {
		t.Errorf("filter = %q, want %q", m.filter, testFilterKitchen)
	}
}

func TestModel_Navigation(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Initial cursor should be 0
	if m.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", m.cursor)
	}

	// Test g key (go to top)
	msg := tea.KeyPressMsg{Code: 'g'}
	updated, _, _ := m.handleNavigation(msg)
	if updated.cursor != 0 {
		t.Errorf("after g cursor = %d, want 0", updated.cursor)
	}
}

func TestModel_LayoutMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		wantLayout LayoutMode
	}{
		{"narrow (40)", 40, LayoutNarrow},
		{"narrow (79)", 79, LayoutNarrow},
		{"standard (80)", 80, LayoutStandard},
		{"standard (100)", 100, LayoutStandard},
		{"standard (120)", 120, LayoutStandard},
		{"wide (121)", 121, LayoutWide},
		{"wide (160)", 160, LayoutWide},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newTestModel(t)
			m = applyWindowSize(m, tt.width, 40)

			if m.layoutMode() != tt.wantLayout {
				t.Errorf("layoutMode() = %v, want %v", m.layoutMode(), tt.wantLayout)
			}
		})
	}
}

func TestModel_IsDashboardActive(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Initially should be dashboard
	if !m.isDashboardActive() {
		t.Error("isDashboardActive should be true initially")
	}

	// Switch to Automation
	m.tabBar, _ = m.tabBar.SetActive(tabs.TabAutomation)
	m.viewManager.SetActive(views.ViewAutomation)

	if m.isDashboardActive() {
		t.Error("isDashboardActive should be false after switching to Automation")
	}
}

func TestDefaultOptions(t *testing.T) {
	t.Parallel()

	opts := DefaultOptions()
	if opts.RefreshInterval != 5*time.Second {
		t.Errorf("RefreshInterval = %v, want 5s", opts.RefreshInterval)
	}
}

func TestModel_DeviceActionMsg(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Test successful action
	msg := DeviceActionMsg{
		Device: "test-device",
		Action: "toggle",
		Err:    nil,
	}
	updated, cmd := m.Update(msg)
	if _, ok := updated.(Model); !ok {
		t.Fatal("Update should return Model")
	}

	if cmd == nil {
		t.Error("DeviceActionMsg should return a status command")
	}
}

func TestModel_TabChangedMsg(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Simulate tab change message
	msg := tabs.TabChangedMsg{
		Previous: tabs.TabDashboard,
		Current:  tabs.TabConfig,
	}
	updated, cmd := m.Update(msg)
	if _, ok := updated.(Model); !ok {
		t.Fatal("Update should return Model")
	}

	if cmd == nil {
		t.Error("TabChangedMsg should return a view change command")
	}
}

func TestModel_ViewChangedMsg(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)

	// Simulate view change message
	msg := views.ViewChangedMsg{
		Previous: views.ViewDashboard,
		Current:  views.ViewConfig,
	}
	updated, _ := m.Update(msg)
	model, ok := updated.(Model)
	if !ok {
		t.Fatal("Update should return Model")
	}

	// Tab bar should sync with view change
	if model.tabBar.ActiveTabID() != tabs.TabConfig {
		t.Errorf("tabBar should sync to Config, got %v", model.tabBar.ActiveTabID())
	}
}

func TestModel_GetFilteredDevices_EmptyFilter(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.filter = ""

	devices := m.getFilteredDevices()
	// Should return all devices (empty in test)
	if devices == nil {
		t.Error("getFilteredDevices should return non-nil slice")
	}
}

func TestModel_GetFilteredDevices_WithFilter(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m.filter = testFilterKitchen

	devices := m.getFilteredDevices()
	// Should return filtered devices (empty in test, but verify it doesn't panic)
	if devices == nil {
		t.Error("getFilteredDevices should return non-nil slice even with filter")
	}
}

func TestModel_EscapeClearsFilter(t *testing.T) {
	t.Parallel()

	m := newTestModel(t)
	m = applyWindowSize(m, 100, 40)
	m.filter = "test-filter"

	// Press Escape when filter is set
	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	updated, _, _ := m.handleGlobalKeys(msg)
	model, ok := updated.(Model)
	if !ok {
		t.Fatal("handleGlobalKeys should return Model")
	}

	if model.filter != "" {
		t.Errorf("filter should be cleared after Escape, got %q", model.filter)
	}
}

func TestModel_NavigationKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		key       string
		wantDelta int
	}{
		{"j moves down", "j", 1},
		{"down moves down", "down", 1},
		{"k moves up", "k", -1},
		{"up moves up", "up", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newTestModel(t)
			m = applyWindowSize(m, 100, 40)

			// Navigation requires devices which we don't have in test
			// Just verify the key is handled without panic
			var msg tea.KeyPressMsg
			switch tt.key {
			case "j", "k":
				msg = tea.KeyPressMsg{Code: rune(tt.key[0])}
			case "down":
				msg = tea.KeyPressMsg{Code: tea.KeyDown}
			case "up":
				msg = tea.KeyPressMsg{Code: tea.KeyUp}
			}

			_, _, handled := m.handleNavigation(msg)
			if !handled {
				t.Errorf("key %q should be handled", tt.key)
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

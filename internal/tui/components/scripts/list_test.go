package scripts

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestScript(t *testing.T) {
	t.Parallel()
	s := Script{
		ID:      1,
		Name:    "test_script",
		Enabled: true,
		Running: true,
	}

	if s.ID != 1 {
		t.Errorf("ID = %d, want 1", s.ID)
	}
	if s.Name != "test_script" {
		t.Errorf("Name = %q, want %q", s.Name, "test_script")
	}
	if !s.Enabled {
		t.Error("Enabled = false, want true")
	}
	if !s.Running {
		t.Error("Running = false, want true")
	}
}

func TestListDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    ListDeps
		wantErr bool
	}{
		{
			name:    "nil context",
			deps:    ListDeps{Ctx: nil, Svc: nil},
			wantErr: true,
		},
		{
			name:    "nil service",
			deps:    ListDeps{Ctx: context.Background(), Svc: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.deps.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListModel_SetSize(t *testing.T) {
	t.Parallel()
	m := ListModel{}
	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestListModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := ListModel{}

	m = m.SetFocused(true)
	if !m.focused {
		t.Error("focused = false, want true")
	}

	m = m.SetFocused(false)
	if m.focused {
		t.Error("focused = true, want false")
	}
}

func TestListModel_VisibleRows(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{"zero height", 0, 1},
		{"small height", 5, 1},
		{"normal height", 20, 16},
		{"large height", 100, 96},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := ListModel{height: tt.height}
			got := m.visibleRows()
			if got != tt.want {
				t.Errorf("visibleRows() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestListModel_CursorNavigation(t *testing.T) {
	t.Parallel()
	scripts := []Script{
		{ID: 1, Name: "script1"},
		{ID: 2, Name: "script2"},
		{ID: 3, Name: "script3"},
	}

	m := ListModel{
		scripts: scripts,
		cursor:  0,
		height:  20,
	}

	// Test cursor down
	m = m.cursorDown()
	if m.cursor != 1 {
		t.Errorf("after cursorDown: cursor = %d, want 1", m.cursor)
	}

	m = m.cursorDown()
	if m.cursor != 2 {
		t.Errorf("after 2nd cursorDown: cursor = %d, want 2", m.cursor)
	}

	// Test cursor doesn't go past end
	m = m.cursorDown()
	if m.cursor != 2 {
		t.Errorf("cursor at end: cursor = %d, want 2", m.cursor)
	}

	// Test cursor up
	m = m.cursorUp()
	if m.cursor != 1 {
		t.Errorf("after cursorUp: cursor = %d, want 1", m.cursor)
	}

	// Test cursor doesn't go before start
	m.cursor = 0
	m = m.cursorUp()
	if m.cursor != 0 {
		t.Errorf("cursor at start: cursor = %d, want 0", m.cursor)
	}

	// Test cursor to end
	m = m.cursorToEnd()
	if m.cursor != 2 {
		t.Errorf("after cursorToEnd: cursor = %d, want 2", m.cursor)
	}
}

func TestListModel_EnsureVisible(t *testing.T) {
	t.Parallel()
	scripts := make([]Script, 20)
	for i := range scripts {
		scripts[i] = Script{ID: i, Name: "script"}
	}

	m := ListModel{
		scripts: scripts,
		height:  10,
		cursor:  0,
		scroll:  0,
	}

	// Move cursor past visible area
	m.cursor = 10
	m = m.ensureVisible()
	if m.scroll < 5 {
		t.Errorf("scroll should adjust for cursor=10, got scroll=%d", m.scroll)
	}

	// Move cursor before scroll
	m.cursor = 2
	m = m.ensureVisible()
	if m.scroll != 2 {
		t.Errorf("scroll should adjust to cursor=2, got scroll=%d", m.scroll)
	}
}

func TestListModel_SelectedScript(t *testing.T) {
	t.Parallel()
	scripts := []Script{
		{ID: 1, Name: "script1"},
		{ID: 2, Name: "script2"},
	}

	m := ListModel{
		scripts: scripts,
		cursor:  1,
	}

	selected := m.SelectedScript()
	if selected == nil {
		t.Fatal("SelectedScript() = nil, want script")
	}
	if selected.ID != 2 {
		t.Errorf("SelectedScript().ID = %d, want 2", selected.ID)
	}

	// Test empty list
	m.scripts = nil
	selected = m.SelectedScript()
	if selected != nil {
		t.Error("SelectedScript() on empty list should return nil")
	}
}

func TestListModel_ScriptCount(t *testing.T) {
	t.Parallel()
	m := ListModel{
		scripts: []Script{{ID: 1}, {ID: 2}, {ID: 3}},
	}

	if got := m.ScriptCount(); got != 3 {
		t.Errorf("ScriptCount() = %d, want 3", got)
	}

	m.scripts = nil
	if got := m.ScriptCount(); got != 0 {
		t.Errorf("ScriptCount() on nil = %d, want 0", got)
	}
}

func TestListModel_Device(t *testing.T) {
	t.Parallel()
	m := ListModel{device: "192.168.1.100"}
	if got := m.Device(); got != "192.168.1.100" {
		t.Errorf("Device() = %q, want %q", got, "192.168.1.100")
	}
}

func TestListModel_Loading(t *testing.T) {
	t.Parallel()
	m := ListModel{loading: true}
	if !m.Loading() {
		t.Error("Loading() = false, want true")
	}

	m.loading = false
	if m.Loading() {
		t.Error("Loading() = true, want false")
	}
}

func TestListModel_Error(t *testing.T) {
	t.Parallel()
	m := ListModel{}
	if err := m.Error(); err != nil {
		t.Errorf("Error() = %v, want nil", err)
	}

	m.err = context.DeadlineExceeded
	if err := m.Error(); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Error() = %v, want %v", err, context.DeadlineExceeded)
	}
}

func TestListModel_View_NoDevice(t *testing.T) {
	t.Parallel()
	m := ListModel{
		device: "",
		width:  40,
		height: 10,
		styles: DefaultListStyles(),
	}

	view := m.View()
	if !strings.Contains(view, "No device selected") {
		t.Errorf("View() should show 'No device selected', got:\n%s", view)
	}
}

func TestListModel_View_Loading(t *testing.T) {
	t.Parallel()
	m := ListModel{
		device:  "192.168.1.100",
		loading: true,
		width:   40,
		height:  10,
		styles:  DefaultListStyles(),
	}

	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("View() should show 'Loading', got:\n%s", view)
	}
}

func TestListModel_View_NoScripts(t *testing.T) {
	t.Parallel()
	m := ListModel{
		device:  "192.168.1.100",
		loading: false,
		scripts: []Script{},
		width:   40,
		height:  10,
		styles:  DefaultListStyles(),
	}

	view := m.View()
	if !strings.Contains(view, "No scripts") {
		t.Errorf("View() should show 'No scripts', got:\n%s", view)
	}
}

func TestListModel_View_WithScripts(t *testing.T) {
	t.Parallel()
	m := ListModel{
		device:  "192.168.1.100",
		loading: false,
		scripts: []Script{
			{ID: 1, Name: "auto_lights", Enabled: true, Running: true},
			{ID: 2, Name: "morning_routine", Enabled: true, Running: false},
			{ID: 3, Name: "disabled_script", Enabled: false, Running: false},
		},
		cursor: 0,
		width:  50,
		height: 15,
		styles: DefaultListStyles(),
	}

	view := m.View()
	if !strings.Contains(view, "auto_lights") {
		t.Errorf("View() should show script name 'auto_lights', got:\n%s", view)
	}
	if !strings.Contains(view, "morning_routine") {
		t.Errorf("View() should show script name 'morning_routine', got:\n%s", view)
	}
}

func TestListModel_RenderScriptLine(t *testing.T) {
	t.Parallel()
	m := ListModel{styles: DefaultListStyles()}

	tests := []struct {
		name       string
		script     Script
		isSelected bool
		wantIcon   bool
		wantName   bool
	}{
		{
			name:       "running script selected",
			script:     Script{ID: 1, Name: "test", Enabled: true, Running: true},
			isSelected: true,
			wantIcon:   true,
			wantName:   true,
		},
		{
			name:       "stopped script",
			script:     Script{ID: 2, Name: "stopped", Enabled: true, Running: false},
			isSelected: false,
			wantIcon:   true,
			wantName:   true,
		},
		{
			name:       "disabled script",
			script:     Script{ID: 3, Name: "disabled", Enabled: false, Running: false},
			isSelected: false,
			wantIcon:   true,
			wantName:   true,
		},
		{
			name:       "unnamed script",
			script:     Script{ID: 4, Name: "", Enabled: true, Running: false},
			isSelected: false,
			wantIcon:   true,
			wantName:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			line := m.renderScriptLine(tt.script, tt.isSelected)
			if line == "" {
				t.Error("renderScriptLine() returned empty string")
			}
			// Check for selection indicator
			if tt.isSelected && !strings.Contains(line, "▶") {
				t.Errorf("selected line should contain ▶, got: %s", line)
			}
		})
	}
}

func TestListModel_Update_LoadedMsg(t *testing.T) {
	t.Parallel()
	m := ListModel{
		loading: true,
		styles:  DefaultListStyles(),
	}

	scripts := []Script{
		{ID: 1, Name: "script1"},
		{ID: 2, Name: "script2"},
	}

	m, _ = m.Update(LoadedMsg{Scripts: scripts})

	if m.loading {
		t.Error("loading should be false after LoadedMsg")
	}
	if len(m.scripts) != 2 {
		t.Errorf("scripts length = %d, want 2", len(m.scripts))
	}
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestListModel_Update_LoadedMsg_Error(t *testing.T) {
	t.Parallel()
	m := ListModel{
		loading: true,
		styles:  DefaultListStyles(),
	}

	m, _ = m.Update(LoadedMsg{Err: context.DeadlineExceeded})

	if m.loading {
		t.Error("loading should be false after error LoadedMsg")
	}
	if !errors.Is(m.err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want %v", m.err, context.DeadlineExceeded)
	}
}

func TestListModel_Update_KeyPress_NotFocused(t *testing.T) {
	t.Parallel()
	m := ListModel{
		focused: false,
		scripts: []Script{{ID: 1}},
		cursor:  0,
		styles:  DefaultListStyles(),
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 106}) // 'j'

	// Cursor should not change when not focused
	if m.cursor != 0 {
		t.Errorf("cursor changed when not focused: %d", m.cursor)
	}
}

func TestListModel_Init(t *testing.T) {
	t.Parallel()
	m := ListModel{}
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestDefaultListStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultListStyles()

	// Just verify styles are created without panic
	_ = styles.Running.Render("test")
	_ = styles.Stopped.Render("test")
	_ = styles.Disabled.Render("test")
	_ = styles.Name.Render("test")
	_ = styles.Selected.Render("test")
	_ = styles.Status.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
}

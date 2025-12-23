package schedules

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestSchedule(t *testing.T) {
	t.Parallel()
	s := Schedule{
		ID:       1,
		Enable:   true,
		Timespec: "0 0 9 * * MON-FRI",
		Calls: []shelly.ScheduleCall{
			{Method: "Switch.Set", Params: map[string]any{"on": true}},
		},
	}

	if s.ID != 1 {
		t.Errorf("ID = %d, want 1", s.ID)
	}
	if !s.Enable {
		t.Error("Enable = false, want true")
	}
	if len(s.Calls) != 1 {
		t.Errorf("Calls length = %d, want 1", len(s.Calls))
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
	schedules := []Schedule{
		{ID: 1, Timespec: "0 0 9 * * *"},
		{ID: 2, Timespec: "0 0 18 * * *"},
		{ID: 3, Timespec: "0 30 12 * * *"},
	}

	m := ListModel{
		schedules: schedules,
		cursor:    0,
		height:    20,
	}

	// Cursor down
	m = m.cursorDown()
	if m.cursor != 1 {
		t.Errorf("after cursorDown: cursor = %d, want 1", m.cursor)
	}

	m = m.cursorDown()
	if m.cursor != 2 {
		t.Errorf("after 2nd cursorDown: cursor = %d, want 2", m.cursor)
	}

	// Don't go past end
	m = m.cursorDown()
	if m.cursor != 2 {
		t.Errorf("cursor at end: cursor = %d, want 2", m.cursor)
	}

	// Cursor up
	m = m.cursorUp()
	if m.cursor != 1 {
		t.Errorf("after cursorUp: cursor = %d, want 1", m.cursor)
	}

	// Don't go before start
	m.cursor = 0
	m = m.cursorUp()
	if m.cursor != 0 {
		t.Errorf("cursor at start: cursor = %d, want 0", m.cursor)
	}

	// Cursor to end
	m = m.cursorToEnd()
	if m.cursor != 2 {
		t.Errorf("after cursorToEnd: cursor = %d, want 2", m.cursor)
	}
}

func TestListModel_EnsureVisible(t *testing.T) {
	t.Parallel()
	schedules := make([]Schedule, 20)
	for i := range schedules {
		schedules[i] = Schedule{ID: i, Timespec: "0 0 * * * *"}
	}

	m := ListModel{
		schedules: schedules,
		height:    10,
		cursor:    0,
		scroll:    0,
	}

	// Move cursor past visible
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

func TestListModel_SelectedSchedule(t *testing.T) {
	t.Parallel()
	schedules := []Schedule{
		{ID: 1, Timespec: "0 0 9 * * *"},
		{ID: 2, Timespec: "0 0 18 * * *"},
	}

	m := ListModel{
		schedules: schedules,
		cursor:    1,
	}

	selected := m.SelectedSchedule()
	if selected == nil {
		t.Fatal("SelectedSchedule() = nil, want schedule")
	}
	if selected.ID != 2 {
		t.Errorf("SelectedSchedule().ID = %d, want 2", selected.ID)
	}

	// Empty list
	m.schedules = nil
	selected = m.SelectedSchedule()
	if selected != nil {
		t.Error("SelectedSchedule() on empty list should return nil")
	}
}

func TestListModel_ScheduleCount(t *testing.T) {
	t.Parallel()
	m := ListModel{
		schedules: []Schedule{{ID: 1}, {ID: 2}, {ID: 3}},
	}

	if got := m.ScheduleCount(); got != 3 {
		t.Errorf("ScheduleCount() = %d, want 3", got)
	}

	m.schedules = nil
	if got := m.ScheduleCount(); got != 0 {
		t.Errorf("ScheduleCount() on nil = %d, want 0", got)
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

func TestListModel_View_NoSchedules(t *testing.T) {
	t.Parallel()
	m := ListModel{
		device:    "192.168.1.100",
		loading:   false,
		schedules: []Schedule{},
		width:     40,
		height:    10,
		styles:    DefaultListStyles(),
	}

	view := m.View()
	if !strings.Contains(view, "No schedules") {
		t.Errorf("View() should show 'No schedules', got:\n%s", view)
	}
}

func TestListModel_View_WithSchedules(t *testing.T) {
	t.Parallel()
	m := ListModel{
		device:  "192.168.1.100",
		loading: false,
		schedules: []Schedule{
			{
				ID:       1,
				Enable:   true,
				Timespec: "0 0 9 * * MON-FRI",
				Calls:    []shelly.ScheduleCall{{Method: "Switch.Set"}},
			},
			{
				ID:       2,
				Enable:   false,
				Timespec: "0 0 22 * * *",
				Calls:    []shelly.ScheduleCall{{Method: "Switch.Set"}},
			},
		},
		cursor: 0,
		width:  60,
		height: 15,
		styles: DefaultListStyles(),
	}

	view := m.View()
	if !strings.Contains(view, "Switch.Set") {
		t.Errorf("View() should show method name, got:\n%s", view)
	}
}

func TestListModel_RenderScheduleLine(t *testing.T) {
	t.Parallel()
	m := ListModel{styles: DefaultListStyles()}

	tests := []struct {
		name       string
		schedule   Schedule
		isSelected bool
	}{
		{
			name: "enabled schedule selected",
			schedule: Schedule{
				ID:       1,
				Enable:   true,
				Timespec: "0 0 9 * * *",
				Calls:    []shelly.ScheduleCall{{Method: "Switch.Set"}},
			},
			isSelected: true,
		},
		{
			name: "disabled schedule",
			schedule: Schedule{
				ID:       2,
				Enable:   false,
				Timespec: "0 0 18 * * *",
				Calls:    []shelly.ScheduleCall{{Method: "Light.Set"}},
			},
			isSelected: false,
		},
		{
			name: "multiple calls",
			schedule: Schedule{
				ID:       3,
				Enable:   true,
				Timespec: "@sunrise",
				Calls: []shelly.ScheduleCall{
					{Method: "Switch.Set"},
					{Method: "Light.Set"},
				},
			},
			isSelected: false,
		},
		{
			name: "no calls",
			schedule: Schedule{
				ID:       4,
				Enable:   true,
				Timespec: "0 30 12 * * *",
				Calls:    []shelly.ScheduleCall{},
			},
			isSelected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			line := m.renderScheduleLine(tt.schedule, tt.isSelected)
			if line == "" {
				t.Error("renderScheduleLine() returned empty string")
			}
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

	schedules := []Schedule{
		{ID: 1, Timespec: "0 0 9 * * *"},
		{ID: 2, Timespec: "0 0 18 * * *"},
	}

	m, _ = m.Update(LoadedMsg{Schedules: schedules})

	if m.loading {
		t.Error("loading should be false after LoadedMsg")
	}
	if len(m.schedules) != 2 {
		t.Errorf("schedules length = %d, want 2", len(m.schedules))
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
		t.Error("loading should be false after error")
	}
	if !errors.Is(m.err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want %v", m.err, context.DeadlineExceeded)
	}
}

func TestListModel_Update_KeyPress_NotFocused(t *testing.T) {
	t.Parallel()
	m := ListModel{
		focused:   false,
		schedules: []Schedule{{ID: 1}},
		cursor:    0,
		styles:    DefaultListStyles(),
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 106}) // 'j'

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

	_ = styles.Enabled.Render("test")
	_ = styles.Disabled.Render("test")
	_ = styles.Method.Render("test")
	_ = styles.Timespec.Render("test")
	_ = styles.Selected.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
}

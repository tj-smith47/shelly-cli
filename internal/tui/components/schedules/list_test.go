package schedules

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
)

func TestSchedule(t *testing.T) {
	t.Parallel()
	s := Schedule{
		ID:       1,
		Enable:   true,
		Timespec: "0 0 9 * * MON-FRI",
		Calls: []automation.ScheduleCall{
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
	m := newTestListModel()
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
	m := newTestListModel()

	m = m.SetFocused(true)
	if !m.focused {
		t.Error("focused = false, want true")
	}

	m = m.SetFocused(false)
	if m.focused {
		t.Error("focused = true, want false")
	}
}

func TestListModel_ScrollerVisibleRows(t *testing.T) {
	t.Parallel()
	m := newTestListModel()

	// SetSize configures visible rows (height - 4 overhead)
	m = m.SetSize(80, 20)
	if m.scroller.VisibleRows() != 16 {
		t.Errorf("visibleRows = %d, want 16", m.scroller.VisibleRows())
	}

	m = m.SetSize(80, 5)
	if m.scroller.VisibleRows() < 1 {
		t.Errorf("visibleRows with small height = %d, want >= 1", m.scroller.VisibleRows())
	}
}

func TestListModel_ScrollerCursorNavigation(t *testing.T) {
	t.Parallel()
	m := newTestListModel()
	m.schedules = []Schedule{
		{ID: 1, Timespec: "0 0 9 * * *"},
		{ID: 2, Timespec: "0 0 18 * * *"},
		{ID: 3, Timespec: "0 30 12 * * *"},
	}
	m.scroller.SetItemCount(len(m.schedules))
	m = m.SetSize(80, 20)

	// Cursor down
	m.scroller.CursorDown()
	if m.Cursor() != 1 {
		t.Errorf("after CursorDown: cursor = %d, want 1", m.Cursor())
	}

	m.scroller.CursorDown()
	if m.Cursor() != 2 {
		t.Errorf("after 2nd CursorDown: cursor = %d, want 2", m.Cursor())
	}

	// Don't go past end
	m.scroller.CursorDown()
	if m.Cursor() != 2 {
		t.Errorf("cursor at end: cursor = %d, want 2", m.Cursor())
	}

	// Cursor up
	m.scroller.CursorUp()
	if m.Cursor() != 1 {
		t.Errorf("after CursorUp: cursor = %d, want 1", m.Cursor())
	}

	// Don't go before start
	m.scroller.CursorToStart()
	m.scroller.CursorUp()
	if m.Cursor() != 0 {
		t.Errorf("cursor at start: cursor = %d, want 0", m.Cursor())
	}

	// Cursor to end
	m.scroller.CursorToEnd()
	if m.Cursor() != 2 {
		t.Errorf("after CursorToEnd: cursor = %d, want 2", m.Cursor())
	}
}

func TestListModel_ScrollerEnsureVisible(t *testing.T) {
	t.Parallel()
	m := newTestListModel()
	schedules := make([]Schedule, 20)
	for i := range schedules {
		schedules[i] = Schedule{ID: i, Timespec: "0 0 * * * *"}
	}
	m.schedules = schedules
	m.scroller.SetItemCount(20)
	m = m.SetSize(80, 15) // Sets visibleRows = 15 - 4 = 11

	// Cursor at end should scroll
	m.scroller.CursorToEnd()
	start, _ := m.scroller.VisibleRange()
	if start == 0 {
		t.Error("scroll should increase when cursor at end of long list")
	}

	// Cursor back to start
	m.scroller.CursorToStart()
	start, _ = m.scroller.VisibleRange()
	if start != 0 {
		t.Errorf("scroll = %d, want 0 when cursor at beginning", start)
	}
}

func TestListModel_SelectedSchedule(t *testing.T) {
	t.Parallel()
	m := newTestListModel()
	m.schedules = []Schedule{
		{ID: 1, Timespec: "0 0 9 * * *"},
		{ID: 2, Timespec: "0 0 18 * * *"},
	}
	m.scroller.SetItemCount(len(m.schedules))
	m.scroller.SetCursor(1)

	selected := m.SelectedSchedule()
	if selected == nil {
		t.Fatal("SelectedSchedule() = nil, want schedule")
	}
	if selected.ID != 2 {
		t.Errorf("SelectedSchedule().ID = %d, want 2", selected.ID)
	}

	// Empty list
	m.schedules = nil
	m.scroller.SetItemCount(0)
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
	m := newTestListModel()
	m.device = "192.168.1.100"
	m.loading = false
	m.schedules = []Schedule{
		{
			ID:       1,
			Enable:   true,
			Timespec: "0 0 9 * * MON-FRI",
			Calls:    []automation.ScheduleCall{{Method: "Switch.Set"}},
		},
		{
			ID:       2,
			Enable:   false,
			Timespec: "0 0 22 * * *",
			Calls:    []automation.ScheduleCall{{Method: "Switch.Set"}},
		},
	}
	m.scroller.SetItemCount(len(m.schedules))
	m = m.SetSize(60, 15)

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
				Calls:    []automation.ScheduleCall{{Method: "Switch.Set"}},
			},
			isSelected: true,
		},
		{
			name: "disabled schedule",
			schedule: Schedule{
				ID:       2,
				Enable:   false,
				Timespec: "0 0 18 * * *",
				Calls:    []automation.ScheduleCall{{Method: "Light.Set"}},
			},
			isSelected: false,
		},
		{
			name: "multiple calls",
			schedule: Schedule{
				ID:       3,
				Enable:   true,
				Timespec: "@sunrise",
				Calls: []automation.ScheduleCall{
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
				Calls:    []automation.ScheduleCall{},
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
	m := newTestListModel()
	m.loading = true

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
	m := newTestListModel()
	m.loading = true

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
	m := newTestListModel()
	m.focused = false
	m.schedules = []Schedule{{ID: 1}}
	m.scroller.SetItemCount(len(m.schedules))

	m, _ = m.Update(tea.KeyPressMsg{Code: 106}) // 'j'

	if m.Cursor() != 0 {
		t.Errorf("cursor changed when not focused: %d", m.Cursor())
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

func newTestListModel() ListModel {
	ctx := context.Background()
	svc := &automation.Service{}
	deps := ListDeps{Ctx: ctx, Svc: svc}
	return NewList(deps)
}

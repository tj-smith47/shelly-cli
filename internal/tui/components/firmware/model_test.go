package firmware

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}

	m := New(deps)

	if m.ctx != ctx {
		t.Error("ctx not set")
	}
	if m.svc != svc {
		t.Error("svc not set")
	}
	if m.checking {
		t.Error("should not be checking initially")
	}
	if m.updating {
		t.Error("should not be updating initially")
	}
}

func TestNew_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	deps := Deps{Ctx: nil, Svc: &shelly.Service{}}
	New(deps)
}

func TestNew_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	deps := Deps{Ctx: context.Background(), Svc: nil}
	New(deps)
}

func TestDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    Deps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    Deps{Ctx: context.Background(), Svc: &shelly.Service{}},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    Deps{Ctx: nil, Svc: &shelly.Service{}},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    Deps{Ctx: context.Background(), Svc: nil},
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

func TestModel_Init(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	cmd := m.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated := m.SetSize(100, 50)

	if updated.Width != 100 {
		t.Errorf("width = %d, want 100", updated.Width)
	}
	if updated.Height != 50 {
		t.Errorf("height = %d, want 50", updated.Height)
	}
}

func TestModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	if m.focused {
		t.Error("should not be focused initially")
	}

	updated := m.SetFocused(true)

	if !updated.focused {
		t.Error("should be focused after SetFocused(true)")
	}
}

func TestModel_Update_CheckCompleteMsg(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.checking = true
	m.devices = []DeviceFirmware{
		{Name: "device1", Address: "192.168.1.100"},
		{Name: "device2", Address: "192.168.1.101"},
	}
	results := []DeviceFirmware{
		{Name: "device1", Address: "192.168.1.100", Current: "1.0.0", Available: "1.1.0", HasUpdate: true, Checked: true},
		{Name: "device2", Address: "192.168.1.101", Current: "2.0.0", HasUpdate: false, Checked: true},
	}
	msg := CheckCompleteMsg{Results: results}

	updated, _ := m.Update(msg)

	if updated.checking {
		t.Error("should not be checking after CheckCompleteMsg")
	}
	if !updated.devices[0].HasUpdate {
		t.Error("device1 should have update")
	}
	if updated.devices[1].HasUpdate {
		t.Error("device2 should not have update")
	}
}

func TestModel_Update_UpdateCompleteMsg(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceFirmware{
		{Name: "device1", Updating: true, HasUpdate: true},
	}
	msg := UpdateCompleteMsg{Name: "device1", Success: true}

	updated, _ := m.Update(msg)

	if updated.devices[0].Updating {
		t.Error("device should not be updating after success")
	}
	if updated.devices[0].HasUpdate {
		t.Error("device should not have update after success")
	}
}

func TestModel_Update_UpdateCompleteMsg_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceFirmware{
		{Name: "device1", Updating: true, HasUpdate: true},
	}
	testErr := errors.New("update failed")
	msg := UpdateCompleteMsg{Name: "device1", Success: false, Err: testErr}

	updated, _ := m.Update(msg)

	if updated.devices[0].Updating {
		t.Error("device should not be updating after error")
	}
	if updated.devices[0].Err == nil {
		t.Error("device should have error")
	}
}

func TestModel_HandleKey_Navigation(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceFirmware{
		{Name: "device0"},
		{Name: "device1"},
		{Name: "device2"},
	}
	m.Scroller.SetItemCount(len(m.devices))

	// Move down
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.Cursor() != 1 {
		t.Errorf("cursor after j = %d, want 1", updated.Cursor())
	}

	// Move up
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'k'})
	if updated.Cursor() != 0 {
		t.Errorf("cursor after k = %d, want 0", updated.Cursor())
	}
}

func TestModel_HandleKey_Selection(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceFirmware{
		{Name: "device0", Selected: false},
		{Name: "device1", Selected: false},
	}

	// Toggle selection with space
	updated, _ := m.Update(tea.KeyPressMsg{Code: ' '})
	if !updated.devices[0].Selected {
		t.Error("device0 should be selected after space")
	}

	// Toggle again
	updated, _ = updated.Update(tea.KeyPressMsg{Code: ' '})
	if updated.devices[0].Selected {
		t.Error("device0 should be unselected after second space")
	}
}

func TestModel_HandleKey_SelectAllWithUpdates(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceFirmware{
		{Name: "device0", HasUpdate: true, Selected: false},
		{Name: "device1", HasUpdate: false, Selected: false},
		{Name: "device2", HasUpdate: true, Selected: false},
	}

	// Select all with updates via 'a'
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'a'})
	if !updated.devices[0].Selected {
		t.Error("device0 with update should be selected after 'a'")
	}
	if updated.devices[1].Selected {
		t.Error("device1 without update should not be selected after 'a'")
	}
	if !updated.devices[2].Selected {
		t.Error("device2 with update should be selected after 'a'")
	}
}

func TestModel_HandleKey_SelectNone(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceFirmware{
		{Name: "device0", Selected: true},
		{Name: "device1", Selected: true},
	}

	// Deselect all with 'n'
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	if updated.devices[0].Selected || updated.devices[1].Selected {
		t.Error("all devices should be deselected after 'n'")
	}
}

func TestModel_HandleKey_Check(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceFirmware{{Name: "device0"}}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'c'})

	if !updated.checking {
		t.Error("should be checking after 'c' key")
	}
	if cmd == nil {
		t.Error("should return check command")
	}
}

func TestModel_HandleKey_Update_NoSelection(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceFirmware{
		{Name: "device0", HasUpdate: true, Selected: false},
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'u'})

	if cmd != nil {
		t.Error("should not return command when no devices selected")
	}
	if updated.err == nil {
		t.Error("should set error when no devices selected")
	}
}

func TestModel_HandleKey_Update_WithSelection(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceFirmware{
		{Name: "device0", HasUpdate: true, Selected: true},
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'u'})

	if !updated.updating {
		t.Error("should be updating after 'u'")
	}
	if cmd == nil {
		t.Error("should return command when devices selected")
	}
	if !updated.devices[0].Updating {
		t.Error("selected device should have Updating=true")
	}
}

func TestModel_HandleKey_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.devices = []DeviceFirmware{{Name: "device0"}}
	m.Scroller.SetItemCount(len(m.devices))

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})

	if updated.Cursor() != 0 {
		t.Error("cursor should not change when not focused")
	}
}

func TestModel_ScrollerCursorBounds(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceFirmware{
		{Name: "device0"},
		{Name: "device1"},
	}
	m.Scroller.SetItemCount(len(m.devices))

	// Can't go below 0
	m.Scroller.CursorUp()
	if m.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 (can't go below)", m.Cursor())
	}

	// Can't exceed list length
	m.Scroller.SetCursor(1)
	m.Scroller.CursorDown()
	if m.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 (can't exceed list)", m.Cursor())
	}
}

func TestModel_ScrollerVisibleRows(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = make([]DeviceFirmware, 20)
	m.Scroller.SetItemCount(20)

	// SetSize configures visible rows (height - 10 overhead)
	m = m.SetSize(80, 20)
	if m.Scroller.VisibleRows() != 10 {
		t.Errorf("visibleRows = %d, want 10", m.Scroller.VisibleRows())
	}

	m = m.SetSize(80, 5)
	if m.Scroller.VisibleRows() < 1 {
		t.Errorf("visibleRows with small height = %d, want >= 1", m.Scroller.VisibleRows())
	}
}

func TestModel_SelectedDevices(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceFirmware{
		{Name: "device0", Selected: true},
		{Name: "device1", Selected: false},
		{Name: "device2", Selected: true},
	}

	selected := m.selectedDevices()

	if len(selected) != 2 {
		t.Errorf("selectedDevices() len = %d, want 2", len(selected))
	}
}

func TestModel_View_NoDevices(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_WithDevices(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceFirmware{
		{Name: "device0", Current: "1.0.0", Available: "1.1.0", HasUpdate: true, Checked: true},
		{Name: "device1", Current: "2.0.0", HasUpdate: false, Checked: true},
	}
	m = m.SetSize(80, 30)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Checking(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceFirmware{{Name: "device0"}}
	m.checking = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Updating(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceFirmware{{Name: "device0", Updating: true}}
	m.updating = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceFirmware{
		{Name: "device0", Selected: true, HasUpdate: true},
		{Name: "device1", Selected: false, HasUpdate: true},
		{Name: "device2", Selected: true, HasUpdate: false},
	}
	m.Scroller.SetItemCount(len(m.devices))
	m.checking = true
	m.updating = true
	m.err = errors.New("test error")
	m.Scroller.SetCursor(2)

	if len(m.Devices()) != 3 {
		t.Errorf("Devices() len = %d, want 3", len(m.Devices()))
	}
	if !m.Checking() {
		t.Error("Checking() should be true")
	}
	if !m.Updating() {
		t.Error("Updating() should be true")
	}
	if m.Error() == nil {
		t.Error("Error() should not be nil")
	}
	if m.Cursor() != 2 {
		t.Errorf("Cursor() = %d, want 2", m.Cursor())
	}
	if m.SelectedCount() != 2 {
		t.Errorf("SelectedCount() = %d, want 2", m.SelectedCount())
	}
	if m.UpdateCount() != 2 {
		t.Errorf("UpdateCount() = %d, want 2", m.UpdateCount())
	}
}

func TestModel_ScrollerEnsureVisible(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = make([]DeviceFirmware, 20)
	for i := range m.devices {
		m.devices[i] = DeviceFirmware{Name: string(rune('a' + i))}
	}
	m.Scroller.SetItemCount(20)
	m = m.SetSize(80, 15) // Sets visibleRows = 15 - 10 = 5

	// Cursor at end should scroll
	m.Scroller.CursorToEnd()
	start, _ := m.Scroller.VisibleRange()
	if start == 0 {
		t.Error("scroll should increase when cursor at end of long list")
	}

	// Cursor back to start
	m.Scroller.CursorToStart()
	start, _ = m.Scroller.VisibleRange()
	if start != 0 {
		t.Errorf("scroll = %d, want 0 when cursor at beginning", start)
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Verify styles are created without panic
	_ = styles.HasUpdate.Render("test")
	_ = styles.UpToDate.Render("test")
	_ = styles.Unknown.Render("test")
	_ = styles.Updating.Render("test")
	_ = styles.Selected.Render("test")
	_ = styles.Cursor.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.Button.Render("test")
	_ = styles.Version.Render("test")
}

func TestModel_CheckAll_NoDevices(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated, cmd := m.CheckAll()

	if cmd != nil {
		t.Error("CheckAll with no devices should not return command")
	}
	if updated.checking {
		t.Error("should not be checking with no devices")
	}
}

func TestModel_CheckAll_AlreadyChecking(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.checking = true
	m.devices = []DeviceFirmware{{Name: "device0"}}

	updated, cmd := m.CheckAll()

	if cmd != nil {
		t.Error("CheckAll when already checking should not return command")
	}
	if !updated.checking {
		t.Error("should still be checking")
	}
}

func TestModel_UpdateSelected_NoSelection(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceFirmware{
		{Name: "device0", HasUpdate: true, Selected: false},
	}

	updated, cmd := m.UpdateSelected()

	if cmd != nil {
		t.Error("should not return command when no devices selected")
	}
	if updated.err == nil {
		t.Error("should set error when no devices selected")
	}
}

func TestModel_UpdateSelected_AlreadyUpdating(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.updating = true
	m.devices = []DeviceFirmware{
		{Name: "device0", HasUpdate: true, Selected: true},
	}

	updated, cmd := m.UpdateSelected()

	if cmd != nil {
		t.Error("should not return command when already updating")
	}
	if !updated.updating {
		t.Error("should still be updating")
	}
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}

package batch

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
	if m.executing {
		t.Error("should not be executing initially")
	}
	if m.operation != OpToggle {
		t.Errorf("operation = %v, want OpToggle", m.operation)
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

func TestModel_SetOperation(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated := m.SetOperation(OpOn)

	if updated.operation != OpOn {
		t.Errorf("operation = %v, want OpOn", updated.operation)
	}
}

func TestOperation_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		op   Operation
		want string
	}{
		{OpToggle, "Toggle"},
		{OpOn, "On"},
		{OpOff, "Off"},
		{OpReboot, "Reboot"},
		{OpCheckFirmware, "Check Firmware"},
		{Operation(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.op.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestModel_Update_CompleteMsg(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.executing = true
	results := []OperationResult{
		{Name: "device1", Success: true},
		{Name: "device2", Success: false, Err: errors.New("failed")},
	}
	msg := CompleteMsg{Results: results}

	updated, _ := m.Update(msg)

	if updated.executing {
		t.Error("should not be executing after CompleteMsg")
	}
	if len(updated.results) != 2 {
		t.Errorf("results len = %d, want 2", len(updated.results))
	}
	if !updated.results[0].Success {
		t.Error("results[0] should be success")
	}
	if updated.results[1].Success {
		t.Error("results[1] should be failure")
	}
}

func TestModel_HandleKey_Navigation(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceSelection{
		{Name: "device0", Address: "192.168.1.100"},
		{Name: "device1", Address: "192.168.1.101"},
		{Name: "device2", Address: "192.168.1.102"},
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
	m.devices = []DeviceSelection{
		{Name: "device0", Address: "192.168.1.100", Selected: false},
		{Name: "device1", Address: "192.168.1.101", Selected: false},
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

func TestModel_HandleKey_SelectAll(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceSelection{
		{Name: "device0", Selected: false},
		{Name: "device1", Selected: false},
	}

	// Select all with 'a'
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'a'})
	if !updated.devices[0].Selected || !updated.devices[1].Selected {
		t.Error("all devices should be selected after 'a'")
	}
}

func TestModel_HandleKey_SelectNone(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceSelection{
		{Name: "device0", Selected: true},
		{Name: "device1", Selected: true},
	}

	// Deselect all with 'n'
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	if updated.devices[0].Selected || updated.devices[1].Selected {
		t.Error("all devices should be deselected after 'n'")
	}
}

func TestModel_HandleKey_OperationSwitch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		key  string
		want Operation
	}{
		{"1", OpToggle},
		{"2", OpOn},
		{"3", OpOff},
		{"4", OpReboot},
		{"5", OpCheckFirmware},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			m := newTestModel()
			m.focused = true

			var msg tea.KeyPressMsg
			switch tt.key {
			case "1":
				msg = tea.KeyPressMsg{Code: '1'}
			case "2":
				msg = tea.KeyPressMsg{Code: '2'}
			case "3":
				msg = tea.KeyPressMsg{Code: '3'}
			case "4":
				msg = tea.KeyPressMsg{Code: '4'}
			case "5":
				msg = tea.KeyPressMsg{Code: '5'}
			}

			updated, _ := m.Update(msg)

			if updated.operation != tt.want {
				t.Errorf("operation = %v, want %v", updated.operation, tt.want)
			}
		})
	}
}

func TestModel_HandleKey_Execute_NoSelection(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceSelection{
		{Name: "device0", Selected: false},
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'x'})

	if cmd != nil {
		t.Error("should not return command when no devices selected")
	}
	if updated.err == nil {
		t.Error("should set error when no devices selected")
	}
}

func TestModel_HandleKey_Execute_WithSelection(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceSelection{
		{Name: "device0", Address: "192.168.1.100", Selected: true},
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'x'})

	if !updated.executing {
		t.Error("should be executing after 'x'")
	}
	if cmd == nil {
		t.Error("should return command when devices selected")
	}
}

func TestModel_HandleKey_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.devices = []DeviceSelection{{Name: "device0"}}
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
	m.devices = []DeviceSelection{
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
	m.devices = make([]DeviceSelection, 20)
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
	m.devices = []DeviceSelection{
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
	m.devices = []DeviceSelection{
		{Name: "device0", Address: "192.168.1.100", Selected: true},
		{Name: "device1", Address: "192.168.1.101", Selected: false},
	}
	m = m.SetSize(80, 30)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_WithResults(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceSelection{{Name: "device0"}}
	m.results = []OperationResult{
		{Name: "device0", Success: true},
		{Name: "device1", Success: false, Err: errors.New("failed")},
	}
	m = m.SetSize(80, 30)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Executing(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceSelection{{Name: "device0"}}
	m.executing = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceSelection{
		{Name: "device0", Selected: true},
		{Name: "device1", Selected: false},
		{Name: "device2", Selected: false},
	}
	m.Scroller.SetItemCount(len(m.devices))
	m.operation = OpReboot
	m.executing = true
	m.results = []OperationResult{{Name: "device0", Success: true}}
	m.err = errors.New("test error")
	m.Scroller.SetCursor(2)

	if len(m.Devices()) != 3 {
		t.Errorf("Devices() len = %d, want 3", len(m.Devices()))
	}
	if m.Operation() != OpReboot {
		t.Errorf("Operation() = %v, want OpReboot", m.Operation())
	}
	if !m.Executing() {
		t.Error("Executing() should be true")
	}
	if len(m.Results()) != 1 {
		t.Errorf("Results() len = %d, want 1", len(m.Results()))
	}
	if m.Error() == nil {
		t.Error("Error() should not be nil")
	}
	if m.Cursor() != 2 {
		t.Errorf("Cursor() = %d, want 2", m.Cursor())
	}
	if m.SelectedCount() != 1 {
		t.Errorf("SelectedCount() = %d, want 1", m.SelectedCount())
	}
}

func TestModel_Refresh(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.results = []OperationResult{{Name: "device0"}}
	m.err = errors.New("old error")

	updated, _ := m.Refresh()

	if len(updated.results) != 0 {
		t.Error("results should be cleared")
	}
	if updated.err != nil {
		t.Error("err should be cleared")
	}
}

func TestModel_ScrollerEnsureVisible(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = make([]DeviceSelection, 20)
	for i := range m.devices {
		m.devices[i] = DeviceSelection{Name: string(rune('a' + i))}
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
	_ = styles.Selected.Render("test")
	_ = styles.Unselected.Render("test")
	_ = styles.Cursor.Render("test")
	_ = styles.Operation.Render("test")
	_ = styles.Success.Render("test")
	_ = styles.Failure.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}

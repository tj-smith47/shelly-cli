package inputs

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

const testDevice = "192.168.1.100"

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
	if m.loading {
		t.Error("should not be loading initially")
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

func TestModel_SetDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated, cmd := m.SetDevice(testDevice)

	if updated.device != testDevice {
		t.Errorf("device = %q, want %q", updated.device, testDevice)
	}
	if cmd == nil {
		t.Error("SetDevice should return a command")
	}
	if !updated.loading {
		t.Error("should be loading after SetDevice")
	}
}

func TestModel_SetDevice_Empty(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice

	updated, cmd := m.SetDevice("")

	if updated.device != "" {
		t.Errorf("device = %q, want empty", updated.device)
	}
	if cmd != nil {
		t.Error("SetDevice with empty should return nil")
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

	updated = updated.SetFocused(false)

	if updated.focused {
		t.Error("should not be focused after SetFocused(false)")
	}
}

func TestModel_Update_InputsLoaded(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true
	inputs := []shelly.InputInfo{
		{ID: 0, Name: "Button 1", Type: "button", State: true},
		{ID: 1, Name: "Switch 1", Type: "switch", State: false},
	}
	msg := LoadedMsg{Inputs: inputs}

	updated, _ := m.Update(msg)

	if updated.loading {
		t.Error("should not be loading after LoadedMsg")
	}
	if len(updated.inputs) != 2 {
		t.Errorf("inputs len = %d, want 2", len(updated.inputs))
	}
	if updated.inputs[0].Name != "Button 1" {
		t.Errorf("inputs[0].Name = %q, want Button 1", updated.inputs[0].Name)
	}
}

func TestModel_Update_InputsLoadedError(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true
	testErr := errors.New("connection failed")
	msg := LoadedMsg{Err: testErr}

	updated, _ := m.Update(msg)

	if updated.loading {
		t.Error("should not be loading after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
	if !errors.Is(updated.err, testErr) {
		t.Errorf("err = %v, want %v", updated.err, testErr)
	}
}

func TestModel_HandleKey_Navigation(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.inputs = []shelly.InputInfo{
		{ID: 0, Name: "Input 0"},
		{ID: 1, Name: "Input 1"},
		{ID: 2, Name: "Input 2"},
	}
	m.Scroller.SetItemCount(len(m.inputs))

	// Move down
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.Cursor() != 1 {
		t.Errorf("cursor after j = %d, want 1", updated.Cursor())
	}

	// Move down again
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.Cursor() != 2 {
		t.Errorf("cursor after second j = %d, want 2", updated.Cursor())
	}

	// Move up
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'k'})
	if updated.Cursor() != 1 {
		t.Errorf("cursor after k = %d, want 1", updated.Cursor())
	}
}

func TestModel_HandleKey_Refresh(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'r'})

	if !updated.loading {
		t.Error("should be loading after 'r' key")
	}
	if cmd == nil {
		t.Error("should return refresh command")
	}
}

func TestModel_HandleKey_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.inputs = []shelly.InputInfo{{ID: 0}}
	m.Scroller.SetItemCount(len(m.inputs))

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})

	if updated.Cursor() != 0 {
		t.Error("cursor should not change when not focused")
	}
}

func TestModel_ScrollerCursorBounds(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.inputs = []shelly.InputInfo{
		{ID: 0},
		{ID: 1},
	}
	m.Scroller.SetItemCount(len(m.inputs))

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
	m.inputs = make([]shelly.InputInfo, 20)
	m.Scroller.SetItemCount(20)

	// SetSize configures visible rows (height - 6 overhead)
	m = m.SetSize(80, 20)
	if m.Scroller.VisibleRows() != 14 {
		t.Errorf("visibleRows = %d, want 14", m.Scroller.VisibleRows())
	}

	m = m.SetSize(80, 5)
	if m.Scroller.VisibleRows() < 1 {
		t.Errorf("visibleRows with small height = %d, want >= 1", m.Scroller.VisibleRows())
	}
}

func TestModel_View_NoDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Loading(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.loading = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.err = errors.New("connection failed")
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_NoInputs(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.inputs = []shelly.InputInfo{}
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_WithInputs(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.inputs = []shelly.InputInfo{
		{ID: 0, Name: "Button 1", Type: "button", State: true},
		{ID: 1, Name: "Switch 1", Type: "switch", State: false},
		{ID: 2, Name: "", Type: "analog", State: true}, // Unnamed input
	}
	m = m.SetSize(80, 30)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.inputs = []shelly.InputInfo{{ID: 0}, {ID: 1}, {ID: 2}}
	m.Scroller.SetItemCount(len(m.inputs))
	m.loading = true
	m.err = errors.New("test error")
	m.Scroller.SetCursor(2)

	if m.Device() != testDevice {
		t.Errorf("Device() = %q, want %q", m.Device(), testDevice)
	}
	if len(m.Inputs()) != 3 {
		t.Errorf("Inputs() len = %d, want 3", len(m.Inputs()))
	}
	if !m.Loading() {
		t.Error("Loading() should be true")
	}
	if m.Error() == nil {
		t.Error("Error() should not be nil")
	}
	if m.Cursor() != 2 {
		t.Errorf("Cursor() = %d, want 2", m.Cursor())
	}
}

func TestModel_Refresh(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice

	updated, cmd := m.Refresh()

	if !updated.loading {
		t.Error("should be loading after Refresh")
	}
	if cmd == nil {
		t.Error("Refresh should return a command")
	}
}

func TestModel_Refresh_NoDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated, cmd := m.Refresh()

	if updated.loading {
		t.Error("should not be loading without device")
	}
	if cmd != nil {
		t.Error("Refresh without device should return nil")
	}
}

func TestModel_ScrollerEnsureVisible(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.inputs = make([]shelly.InputInfo, 20)
	for i := range m.inputs {
		m.inputs[i] = shelly.InputInfo{ID: i}
	}
	m.Scroller.SetItemCount(20)
	m = m.SetSize(80, 15) // Sets visibleRows = 15 - 6 = 9

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
	_ = styles.StateOn.Render("test")
	_ = styles.StateOff.Render("test")
	_ = styles.Type.Render("test")
	_ = styles.Name.Render("test")
	_ = styles.ID.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Selected.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}

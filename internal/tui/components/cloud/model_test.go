package cloud

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

func TestModel_Update_StatusLoaded(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true
	msg := StatusLoadedMsg{
		Status:  &shelly.CloudStatus{Connected: true},
		Enabled: true,
		Server:  "shelly-cloud.com",
	}

	updated, _ := m.Update(msg)

	if updated.loading {
		t.Error("should not be loading after StatusLoadedMsg")
	}
	if !updated.connected {
		t.Error("connected should be true")
	}
	if !updated.enabled {
		t.Error("enabled should be true")
	}
	if updated.server != "shelly-cloud.com" {
		t.Errorf("server = %q, want shelly-cloud.com", updated.server)
	}
}

func TestModel_Update_StatusLoadedError(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.loading = true
	testErr := errors.New("connection failed")
	msg := StatusLoadedMsg{Err: testErr}

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

func TestModel_Update_ToggleResult(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.toggling = true
	m.enabled = false
	msg := ToggleResultMsg{Enabled: true}

	updated, cmd := m.Update(msg)

	if updated.toggling {
		t.Error("should not be toggling after ToggleResultMsg")
	}
	if !updated.enabled {
		t.Error("enabled should be true after toggle")
	}
	if cmd == nil {
		t.Error("should return command to refresh status")
	}
}

func TestModel_Update_ToggleResultError(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.toggling = true
	testErr := errors.New("toggle failed")
	msg := ToggleResultMsg{Err: testErr}

	updated, _ := m.Update(msg)

	if updated.toggling {
		t.Error("should not be toggling after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModel_HandleKey_Toggle(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if !updated.toggling {
		t.Error("should be toggling after 't' key")
	}
	if cmd == nil {
		t.Error("should return toggle command")
	}
}

func TestModel_HandleKey_ToggleEnter(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !updated.toggling {
		t.Error("should be toggling after enter key")
	}
	if cmd == nil {
		t.Error("should return toggle command")
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
	m.device = testDevice

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if updated.toggling {
		t.Error("should not toggle when not focused")
	}
	if cmd != nil {
		t.Error("should not return command when not focused")
	}
}

func TestModel_HandleKey_NoDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if updated.toggling {
		t.Error("should not toggle without device")
	}
	if cmd != nil {
		t.Error("should not return command without device")
	}
}

func TestModel_HandleKey_AlreadyToggling(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.toggling = true

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if cmd != nil {
		t.Error("should not return command while toggling")
	}
	if !updated.toggling {
		t.Error("toggling should remain true")
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

func TestModel_View_Connected(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.connected = true
	m.enabled = true
	m.server = "shelly-cloud.com"
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Disconnected(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.connected = false
	m.enabled = false
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Toggling(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.toggling = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.connected = true
	m.enabled = true
	m.server = "test.server.com"
	m.loading = true
	m.toggling = true
	m.err = errors.New("test error")

	if m.Device() != testDevice {
		t.Errorf("Device() = %q, want %q", m.Device(), testDevice)
	}
	if !m.Connected() {
		t.Error("Connected() should be true")
	}
	if !m.Enabled() {
		t.Error("Enabled() should be true")
	}
	if m.Server() != "test.server.com" {
		t.Errorf("Server() = %q, want test.server.com", m.Server())
	}
	if !m.Loading() {
		t.Error("Loading() should be true")
	}
	if !m.Toggling() {
		t.Error("Toggling() should be true")
	}
	if m.Error() == nil {
		t.Error("Error() should not be nil")
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

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Verify styles are created without panic
	_ = styles.Connected.Render("test")
	_ = styles.Disconnected.Render("test")
	_ = styles.Enabled.Render("test")
	_ = styles.Disabled.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Value.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.Title.Render("test")
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}

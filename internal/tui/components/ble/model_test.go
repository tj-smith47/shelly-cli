package ble

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

	if updated.width != 100 {
		t.Errorf("width = %d, want 100", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height = %d, want 50", updated.height)
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
		BLE: &shelly.BLEConfig{
			Enable:       true,
			RPCEnabled:   true,
			ObserverMode: false,
		},
		Discovery: &shelly.BTHomeDiscovery{
			Active:   false,
			Duration: 30,
		},
	}

	updated, _ := m.Update(msg)

	if updated.loading {
		t.Error("should not be loading after StatusLoadedMsg")
	}
	if updated.ble == nil {
		t.Error("ble should be set")
	}
	if !updated.ble.Enable {
		t.Error("ble.Enable should be true")
	}
	if updated.discovery == nil {
		t.Error("discovery should be set")
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

func TestModel_Update_DiscoveryStarted(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.starting = true
	msg := DiscoveryStartedMsg{Err: nil}

	updated, cmd := m.Update(msg)

	if updated.starting {
		t.Error("should not be starting after DiscoveryStartedMsg")
	}
	if !updated.loading {
		t.Error("should be loading to refresh status")
	}
	if cmd == nil {
		t.Error("should return refresh command")
	}
}

func TestModel_Update_DiscoveryStartedError(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.starting = true
	testErr := errors.New("discovery failed")
	msg := DiscoveryStartedMsg{Err: testErr}

	updated, _ := m.Update(msg)

	if updated.starting {
		t.Error("should not be starting after error")
	}
	if updated.err == nil {
		t.Error("err should be set")
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

func TestModel_HandleKey_Discovery(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.ble = &shelly.BLEConfig{Enable: true}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'd'})

	if !updated.starting {
		t.Error("should be starting after 'd' key")
	}
	if cmd == nil {
		t.Error("should return discovery command")
	}
}

func TestModel_HandleKey_Discovery_BLEDisabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.ble = &shelly.BLEConfig{Enable: false}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'd'})

	if updated.starting {
		t.Error("should not start discovery when BLE disabled")
	}
	if cmd != nil {
		t.Error("should not return command when BLE disabled")
	}
}

func TestModel_HandleKey_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.device = testDevice

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'r'})

	if updated.loading {
		t.Error("should not respond to keys when not focused")
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

func TestModel_View_BLEEnabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.ble = &shelly.BLEConfig{
		Enable:       true,
		RPCEnabled:   true,
		ObserverMode: true,
	}
	m.discovery = &shelly.BTHomeDiscovery{
		Active: false,
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_BLEDisabled(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.ble = &shelly.BLEConfig{
		Enable: false,
	}
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_DiscoveryActive(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.ble = &shelly.BLEConfig{Enable: true}
	m.discovery = &shelly.BTHomeDiscovery{
		Active:   true,
		Duration: 30,
	}
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Starting(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.ble = &shelly.BLEConfig{Enable: true}
	m.starting = true
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_NilBLE(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	// ble is nil = not supported
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
	m.ble = &shelly.BLEConfig{Enable: true}
	m.discovery = &shelly.BTHomeDiscovery{Active: true}
	m.loading = true
	m.starting = true
	m.err = errors.New("test error")

	if m.Device() != testDevice {
		t.Errorf("Device() = %q, want %q", m.Device(), testDevice)
	}
	if m.BLE() == nil || !m.BLE().Enable {
		t.Error("BLE() incorrect")
	}
	if m.Discovery() == nil || !m.Discovery().Active {
		t.Error("Discovery() incorrect")
	}
	if !m.Loading() {
		t.Error("Loading() should be true")
	}
	if !m.Starting() {
		t.Error("Starting() should be true")
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
	_ = styles.Enabled.Render("test")
	_ = styles.Disabled.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Value.Render("test")
	_ = styles.Highlight.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.Section.Render("test")
	_ = styles.Warning.Render("test")
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}

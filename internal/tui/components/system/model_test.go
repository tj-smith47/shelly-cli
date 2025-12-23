package system

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
		Status: &shelly.SysStatus{
			MAC:     "AA:BB:CC:DD:EE:FF",
			Uptime:  3600,
			RAMFree: 50000,
			RAMSize: 100000,
			FSFree:  20000,
			FSSize:  50000,
		},
		Config: &shelly.SysConfig{
			Name:         "My Device",
			Timezone:     "America/New_York",
			EcoMode:      true,
			Discoverable: true,
		},
	}

	updated, _ := m.Update(msg)

	if updated.loading {
		t.Error("should not be loading after StatusLoadedMsg")
	}
	if updated.status == nil {
		t.Error("status should be set")
	}
	if updated.status.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("status.MAC = %q, want AA:BB:CC:DD:EE:FF", updated.status.MAC)
	}
	if updated.config == nil {
		t.Error("config should be set")
	}
	if updated.config.Name != "My Device" {
		t.Errorf("config.Name = %q, want My Device", updated.config.Name)
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

func TestModel_HandleKey_Navigation(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true

	// Move down
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.cursor != FieldTimezone {
		t.Errorf("cursor after j = %v, want FieldTimezone", updated.cursor)
	}

	// Move down again
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.cursor != FieldEcoMode {
		t.Errorf("cursor after second j = %v, want FieldEcoMode", updated.cursor)
	}

	// Move up
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'k'})
	if updated.cursor != FieldTimezone {
		t.Errorf("cursor after k = %v, want FieldTimezone", updated.cursor)
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

func TestModel_HandleKey_Toggle(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.device = testDevice
	m.config = &shelly.SysConfig{
		EcoMode:      false,
		Discoverable: true,
	}
	m.cursor = FieldEcoMode

	_, cmd := m.Update(tea.KeyPressMsg{Code: 't'})

	if cmd == nil {
		t.Error("toggle should return a command")
	}
}

func TestModel_HandleKey_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})

	if updated.cursor != FieldName {
		t.Error("cursor should not change when not focused")
	}
}

func TestModel_CursorBounds(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true

	// Can't go below 0
	updated := m.cursorUp()
	if updated.cursor != FieldName {
		t.Errorf("cursor = %v, want FieldName (can't go below)", updated.cursor)
	}

	// Can't exceed FieldCount-1
	updated.cursor = FieldDiscoverable
	updated = updated.cursorDown()
	if updated.cursor != FieldDiscoverable {
		t.Errorf("cursor = %v, want FieldDiscoverable (can't exceed)", updated.cursor)
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

func TestModel_View_WithStatus(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.status = &shelly.SysStatus{
		MAC:             "AA:BB:CC:DD:EE:FF",
		Uptime:          90061, // 1 day, 1 hour, 1 minute
		Time:            "12:34:56",
		RAMFree:         50000,
		RAMSize:         100000,
		FSFree:          20000,
		FSSize:          50000,
		RestartRequired: true,
		UpdateAvailable: "1.2.3",
	}
	m.config = &shelly.SysConfig{
		Name:         "My Device",
		Timezone:     "America/New_York",
		EcoMode:      true,
		Discoverable: false,
	}
	m = m.SetSize(80, 30)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_EmptyConfig(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.status = &shelly.SysStatus{
		Uptime:  3600,
		RAMFree: 50000,
		RAMSize: 100000,
		FSFree:  20000,
		FSSize:  50000,
	}
	m.config = &shelly.SysConfig{
		// Empty timezone
		Name: "Device",
	}
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
	m.status = &shelly.SysStatus{MAC: "AA:BB:CC:DD:EE:FF"}
	m.config = &shelly.SysConfig{Name: "Test"}
	m.loading = true
	m.err = errors.New("test error")

	if m.Device() != testDevice {
		t.Errorf("Device() = %q, want %q", m.Device(), testDevice)
	}
	if m.Status() == nil || m.Status().MAC != "AA:BB:CC:DD:EE:FF" {
		t.Error("Status() incorrect")
	}
	if m.Config() == nil || m.Config().Name != "Test" {
		t.Error("Config() incorrect")
	}
	if !m.Loading() {
		t.Error("Loading() should be true")
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

func TestModel_ToggleCurrentField_NoConfig(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.cursor = FieldEcoMode

	_, cmd := m.toggleCurrentField()

	if cmd != nil {
		t.Error("toggle without config should return nil")
	}
}

func TestModel_ToggleCurrentField_NoDevice(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.config = &shelly.SysConfig{}
	m.cursor = FieldEcoMode

	_, cmd := m.toggleCurrentField()

	if cmd != nil {
		t.Error("toggle without device should return nil")
	}
}

func TestModel_ToggleCurrentField_NonToggleable(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.device = testDevice
	m.config = &shelly.SysConfig{}
	m.cursor = FieldName

	_, cmd := m.toggleCurrentField()

	if cmd != nil {
		t.Error("toggle on non-toggleable field should return nil")
	}
}

func TestFormatUptime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		seconds int
		want    string
	}{
		{60, "1m"},
		{3600, "1h 0m"},
		{3661, "1h 1m"},
		{86400, "1d 0h 0m"},
		{90061, "1d 1h 1m"},
	}

	for _, tt := range tests {
		result := formatUptime(tt.seconds)
		if result != tt.want {
			t.Errorf("formatUptime(%d) = %q, want %q", tt.seconds, result, tt.want)
		}
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Verify styles are created without panic
	_ = styles.Label.Render("test")
	_ = styles.Value.Render("test")
	_ = styles.ValueEnabled.Render("test")
	_ = styles.ValueMuted.Render("test")
	_ = styles.Selected.Render("test")
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

package fleet

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNewDevices(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	deps := DevicesDeps{Ctx: ctx}

	m := NewDevices(deps)

	if m.ctx != ctx {
		t.Error("ctx not set")
	}
}

func TestNewDevices_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	deps := DevicesDeps{Ctx: nil}
	NewDevices(deps)
}

func TestDevicesDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    DevicesDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    DevicesDeps{Ctx: context.Background()},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    DevicesDeps{Ctx: nil},
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

func TestDevicesModel_Init(t *testing.T) {
	t.Parallel()
	m := newTestDevices()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestDevicesModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestDevices()

	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestDevicesModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := newTestDevices()

	m = m.SetFocused(true)
	if !m.focused {
		t.Error("focused should be true")
	}

	m = m.SetFocused(false)
	if m.focused {
		t.Error("focused should be false")
	}
}

func TestDevicesModel_ScrollerNavigation(t *testing.T) {
	t.Parallel()
	m := newTestDevices()
	m.focused = true

	// Test down navigation (no devices, should stay at 0)
	m, _ = m.handleKey(tea.KeyPressMsg{Code: 'j'})
	if m.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", m.Cursor())
	}

	// Test up navigation (should stay at 0)
	m, _ = m.handleKey(tea.KeyPressMsg{Code: 'k'})
	if m.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", m.Cursor())
	}
}

func TestDevicesModel_View_NoFleet(t *testing.T) {
	t.Parallel()
	m := newTestDevices()
	m = m.SetSize(80, 24)

	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestDevicesModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestDevices()

	if m.SelectedDevice() != nil {
		t.Error("SelectedDevice() should return nil with no devices")
	}

	if len(m.Devices()) != 0 {
		t.Error("Devices() should return empty slice")
	}

	if m.DeviceCount() != 0 {
		t.Error("DeviceCount() should return 0")
	}

	if m.OnlineCount() != 0 {
		t.Error("OnlineCount() should return 0")
	}

	if m.Loading() {
		t.Error("Loading() should return false initially")
	}

	if m.Error() != nil {
		t.Error("Error() should return nil initially")
	}
}

func TestDefaultDevicesStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultDevicesStyles()

	// Verify styles are created without panic
	_ = styles.Online.Render("test")
	_ = styles.Offline.Render("test")
	_ = styles.Name.Render("test")
	_ = styles.Type.Render("test")
	_ = styles.Cursor.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.Error.Render("test")
}

func newTestDevices() DevicesModel {
	ctx := context.Background()
	deps := DevicesDeps{Ctx: ctx}
	return NewDevices(deps)
}

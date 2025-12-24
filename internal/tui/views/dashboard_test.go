package views

import (
	"context"
	"testing"
)

const testDashboardDevice = "kitchen-plug"

func TestNewDashboard(t *testing.T) {
	t.Parallel()

	d := NewDashboard(DashboardDeps{
		Ctx: context.Background(),
	})

	if d == nil {
		t.Fatal("NewDashboard() returned nil")
	}
	if d.ID() != ViewDashboard {
		t.Errorf("ID() = %v, want %v", d.ID(), ViewDashboard)
	}
	if d.FocusedPanel() != DashboardPanelDevices {
		t.Errorf("FocusedPanel() = %v, want %v", d.FocusedPanel(), DashboardPanelDevices)
	}
}

func TestNewDashboard_NilCtx(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewDashboard with nil ctx should panic")
		}
	}()

	NewDashboard(DashboardDeps{Ctx: nil})
}

func TestDashboardDeps_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		deps    DashboardDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    DashboardDeps{Ctx: context.Background()},
			wantErr: false,
		},
		{
			name:    "nil_ctx",
			deps:    DashboardDeps{Ctx: nil},
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

func TestDashboard_ID(t *testing.T) {
	t.Parallel()

	d := NewDashboard(DashboardDeps{Ctx: context.Background()})
	if d.ID() != ViewDashboard {
		t.Errorf("ID() = %v, want %v", d.ID(), ViewDashboard)
	}
}

func TestDashboard_Init(t *testing.T) {
	t.Parallel()

	d := NewDashboard(DashboardDeps{Ctx: context.Background()})
	cmd := d.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestDashboard_View(t *testing.T) {
	t.Parallel()

	d := NewDashboard(DashboardDeps{Ctx: context.Background()})
	view := d.View()
	// Dashboard.View() returns empty because app.go handles rendering
	if view != "" {
		t.Errorf("View() = %q, want empty string", view)
	}
}

func TestDashboard_SetSize(t *testing.T) {
	t.Parallel()

	d := NewDashboard(DashboardDeps{Ctx: context.Background()})
	result := d.SetSize(100, 50)

	if result == nil {
		t.Fatal("SetSize() returned nil")
	}
	if d.width != 100 {
		t.Errorf("width = %d, want 100", d.width)
	}
	if d.height != 50 {
		t.Errorf("height = %d, want 50", d.height)
	}
}

func TestDashboard_SetFocusedPanel(t *testing.T) {
	t.Parallel()

	d := NewDashboard(DashboardDeps{Ctx: context.Background()})

	d = d.SetFocusedPanel(DashboardPanelInfo)
	if d.FocusedPanel() != DashboardPanelInfo {
		t.Errorf("FocusedPanel() = %v, want %v", d.FocusedPanel(), DashboardPanelInfo)
	}

	d = d.SetFocusedPanel(DashboardPanelJSON)
	if d.FocusedPanel() != DashboardPanelJSON {
		t.Errorf("FocusedPanel() = %v, want %v", d.FocusedPanel(), DashboardPanelJSON)
	}
}

func TestDashboard_Update_DeviceSelected(t *testing.T) {
	t.Parallel()

	d := NewDashboard(DashboardDeps{Ctx: context.Background()})

	msg := DashboardDeviceSelectedMsg{
		Device:  testDashboardDevice,
		Address: "192.168.1.100",
	}

	newView, cmd := d.Update(msg)
	if newView == nil {
		t.Fatal("Update() returned nil view")
	}
	if cmd == nil {
		t.Fatal("Update() should return a command for device selection")
	}

	// Execute the command and check the message
	resultMsg := cmd()
	dsMsg, ok := resultMsg.(DeviceSelectedMsg)
	if !ok {
		t.Fatalf("expected DeviceSelectedMsg, got %T", resultMsg)
	}
	if dsMsg.Device != testDashboardDevice {
		t.Errorf("Device = %q, want %q", dsMsg.Device, testDashboardDevice)
	}
	if dsMsg.Address != "192.168.1.100" {
		t.Errorf("Address = %q, want %q", dsMsg.Address, "192.168.1.100")
	}
}

func TestDashboard_Update_SameDevice(t *testing.T) {
	t.Parallel()

	d := NewDashboard(DashboardDeps{Ctx: context.Background()})
	d.selectedDevice = testDashboardDevice

	msg := DashboardDeviceSelectedMsg{
		Device:  testDashboardDevice,
		Address: "192.168.1.100",
	}

	_, cmd := d.Update(msg)
	if cmd != nil {
		t.Error("Update() should return nil when device hasn't changed")
	}
}

func TestDashboard_Update_PanelFocus(t *testing.T) {
	t.Parallel()

	d := NewDashboard(DashboardDeps{Ctx: context.Background()})

	msg := DashboardPanelFocusMsg{Panel: DashboardPanelJSON}

	newView, _ := d.Update(msg)
	dashboard, ok := newView.(*Dashboard)
	if !ok {
		t.Fatalf("expected *Dashboard, got %T", newView)
	}

	if dashboard.FocusedPanel() != DashboardPanelJSON {
		t.Errorf("FocusedPanel() = %v, want %v", dashboard.FocusedPanel(), DashboardPanelJSON)
	}
}

func TestDashboard_SelectedDevice(t *testing.T) {
	t.Parallel()

	d := NewDashboard(DashboardDeps{Ctx: context.Background()})

	if d.SelectedDevice() != "" {
		t.Errorf("SelectedDevice() = %q, want empty", d.SelectedDevice())
	}

	d.selectedDevice = "test-device"
	if d.SelectedDevice() != "test-device" {
		t.Errorf("SelectedDevice() = %q, want %q", d.SelectedDevice(), "test-device")
	}
}

func TestDashboard_IsDashboardView(t *testing.T) {
	t.Parallel()

	d := NewDashboard(DashboardDeps{Ctx: context.Background()})
	if !d.IsDashboardView() {
		t.Error("IsDashboardView() should return true")
	}
}

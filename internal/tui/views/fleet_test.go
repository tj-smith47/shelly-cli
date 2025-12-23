package views

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestNewFleet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := FleetDeps{Ctx: ctx, Svc: svc}

	f := NewFleet(deps)

	if f.ctx != ctx {
		t.Error("ctx not set")
	}
	if f.svc != svc {
		t.Error("svc not set")
	}
	if f.focusedPanel != FleetPanelDevices {
		t.Errorf("focusedPanel = %v, want FleetPanelDevices", f.focusedPanel)
	}
	if f.id != TabFleet {
		t.Errorf("id = %v, want TabFleet", f.id)
	}
}

func TestNewFleet_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	deps := FleetDeps{Ctx: nil, Svc: &shelly.Service{}}
	NewFleet(deps)
}

func TestNewFleet_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	deps := FleetDeps{Ctx: context.Background(), Svc: nil}
	NewFleet(deps)
}

func TestFleetDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    FleetDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    FleetDeps{Ctx: context.Background(), Svc: &shelly.Service{}},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    FleetDeps{Ctx: nil, Svc: &shelly.Service{}},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    FleetDeps{Ctx: context.Background(), Svc: nil},
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

func TestFleet_Init(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	// Init may return nil or a batch command
	_ = f.Init()
}

func TestFleet_ID(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	if f.ID() != TabFleet {
		t.Errorf("ID() = %v, want TabFleet", f.ID())
	}
}

func TestFleet_SetSize(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	result := f.SetSize(100, 50)
	updated, ok := result.(*Fleet)
	if !ok {
		t.Fatal("SetSize should return *Fleet")
	}

	if updated.width != 100 {
		t.Errorf("width = %d, want 100", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height = %d, want 50", updated.height)
	}
}

func TestFleet_Update_FocusNext(t *testing.T) {
	t.Parallel()
	f := newTestFleet()
	f.focusedPanel = FleetPanelDevices

	msg := tea.KeyPressMsg{Code: tea.KeyTab}
	updated, _ := f.Update(msg)
	fleet, ok := updated.(*Fleet)
	if !ok {
		t.Fatal("Update should return *Fleet")
	}

	if fleet.focusedPanel != FleetPanelGroups {
		t.Errorf("focusedPanel after tab = %v, want FleetPanelGroups", fleet.focusedPanel)
	}
}

func TestFleet_Update_FocusPrev(t *testing.T) {
	t.Parallel()
	f := newTestFleet()
	f.focusedPanel = FleetPanelGroups

	msg := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	updated, _ := f.Update(msg)
	fleet, ok := updated.(*Fleet)
	if !ok {
		t.Fatal("Update should return *Fleet")
	}

	if fleet.focusedPanel != FleetPanelDevices {
		t.Errorf("focusedPanel after shift+tab = %v, want FleetPanelDevices", fleet.focusedPanel)
	}
}

func TestFleet_Update_NumberKeyFocus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		key      rune
		expected FleetPanel
	}{
		{'1', FleetPanelDevices},
		{'2', FleetPanelGroups},
		{'3', FleetPanelHealth},
		{'4', FleetPanelOperations},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			t.Parallel()
			f := newTestFleet()

			msg := tea.KeyPressMsg{Code: tt.key}
			updated, _ := f.Update(msg)
			fleet, ok := updated.(*Fleet)
			if !ok {
				t.Fatal("Update should return *Fleet")
			}

			if fleet.focusedPanel != tt.expected {
				t.Errorf("focusedPanel after '%c' = %v, want %v", tt.key, fleet.focusedPanel, tt.expected)
			}
		})
	}
}

func TestFleet_FocusCycle(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	// Start at devices
	if f.focusedPanel != FleetPanelDevices {
		t.Fatal("should start at FleetPanelDevices")
	}

	// Cycle through all panels
	f.focusNext()
	if f.focusedPanel != FleetPanelGroups {
		t.Errorf("after 1 focusNext = %v, want FleetPanelGroups", f.focusedPanel)
	}

	f.focusNext()
	if f.focusedPanel != FleetPanelHealth {
		t.Errorf("after 2 focusNext = %v, want FleetPanelHealth", f.focusedPanel)
	}

	f.focusNext()
	if f.focusedPanel != FleetPanelOperations {
		t.Errorf("after 3 focusNext = %v, want FleetPanelOperations", f.focusedPanel)
	}

	f.focusNext()
	if f.focusedPanel != FleetPanelDevices {
		t.Errorf("after 4 focusNext = %v, want FleetPanelDevices (wrap)", f.focusedPanel)
	}
}

func TestFleet_FocusPrevCycle(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	// Start at devices
	f.focusedPanel = FleetPanelDevices

	// Go backwards
	f.focusPrev()
	if f.focusedPanel != FleetPanelOperations {
		t.Errorf("after 1 focusPrev = %v, want FleetPanelOperations (wrap)", f.focusedPanel)
	}

	f.focusPrev()
	if f.focusedPanel != FleetPanelHealth {
		t.Errorf("after 2 focusPrev = %v, want FleetPanelHealth", f.focusedPanel)
	}
}

func TestFleet_View_Empty(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	// Without SetSize, should return empty
	view := f.View()
	if view != "" {
		t.Error("View() without SetSize should return empty string")
	}
}

func TestFleet_View_NotConnected(t *testing.T) {
	t.Parallel()
	f := newTestFleet()
	result, ok := f.SetSize(100, 50).(*Fleet)
	if !ok {
		t.Fatal("SetSize should return *Fleet")
	}
	f = result

	view := f.View()

	// Should show connection prompt
	if view == "" {
		t.Error("View() with SetSize should not return empty string")
	}
}

func TestFleet_Accessors(t *testing.T) {
	t.Parallel()
	f := newTestFleet()
	f.focusedPanel = FleetPanelHealth

	if f.FocusedPanel() != FleetPanelHealth {
		t.Errorf("FocusedPanel() = %v, want FleetPanelHealth", f.FocusedPanel())
	}

	// Verify components are accessible
	_ = f.Devices()
	_ = f.Groups()
	_ = f.Health()
	_ = f.Operations()
}

func TestFleet_Connected(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	if f.Connected() {
		t.Error("Connected() should return false initially")
	}

	if f.Connecting() {
		t.Error("Connecting() should return false initially")
	}

	if f.ConnectionError() != nil {
		t.Error("ConnectionError() should return nil initially")
	}
}

func TestFleet_StatusSummary(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	summary := f.StatusSummary()

	if summary == "" {
		t.Error("StatusSummary() should not return empty string")
	}
}

func TestFleet_StatusSummary_Connecting(t *testing.T) {
	t.Parallel()
	f := newTestFleet()
	f.connecting = true

	summary := f.StatusSummary()

	if summary == "" {
		t.Error("StatusSummary() should not return empty string when connecting")
	}
}

func TestFleet_StatusSummary_Error(t *testing.T) {
	t.Parallel()
	f := newTestFleet()
	f.connErr = errors.New("test error")

	summary := f.StatusSummary()

	if summary == "" {
		t.Error("StatusSummary() should not return empty string with error")
	}
}

func TestFleet_UpdateFocusStates(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	// Set focus to health
	f.focusedPanel = FleetPanelHealth
	f.updateFocusStates()

	// Verify focus states are updated (method doesn't panic)
	_ = f.devices.Loading()
}

func TestDefaultFleetStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultFleetStyles()

	// Verify styles are created without panic
	_ = styles.Panel.Render("test")
	_ = styles.PanelActive.Render("test")
	_ = styles.Title.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.Connected.Render("test")
	_ = styles.Error.Render("test")
}

func TestFleet_Close(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	// Close should not panic even when not connected
	f.Close()
}

func TestFleetConnectMsg(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	// Test error message
	errMsg := FleetConnectMsg{Err: errors.New("test error")}
	updated, _ := f.Update(errMsg)
	fleet, ok := updated.(*Fleet)
	if !ok {
		t.Fatal("Update should return *Fleet")
	}

	if fleet.connErr == nil {
		t.Error("connErr should be set after error message")
	}

	if fleet.connecting {
		t.Error("connecting should be false after error message")
	}
}

func newTestFleet() *Fleet {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := FleetDeps{Ctx: ctx, Svc: svc}
	return NewFleet(deps)
}

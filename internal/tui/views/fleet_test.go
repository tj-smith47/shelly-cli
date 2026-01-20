package views

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

func TestNewFleet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	focusState := focus.NewState()
	focusState.SetActiveTab(tabs.TabFleet)
	deps := FleetDeps{Ctx: ctx, Svc: svc, FocusState: focusState}

	f := NewFleet(deps)

	if f.ctx != ctx {
		t.Error("ctx not set")
	}
	if f.svc != svc {
		t.Error("svc not set")
	}
	if f.id != tabs.TabFleet {
		t.Errorf("id = %v, want tabs.TabFleet", f.id)
	}
}

func TestNewFleet_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	focusState := focus.NewState()
	deps := FleetDeps{Ctx: nil, Svc: &shelly.Service{}, FocusState: focusState}
	NewFleet(deps)
}

func TestNewFleet_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	focusState := focus.NewState()
	deps := FleetDeps{Ctx: context.Background(), Svc: nil, FocusState: focusState}
	NewFleet(deps)
}

func TestFleetDeps_Validate(t *testing.T) {
	t.Parallel()
	focusState := focus.NewState()
	tests := []struct {
		name    string
		deps    FleetDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    FleetDeps{Ctx: context.Background(), Svc: &shelly.Service{}, FocusState: focusState},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    FleetDeps{Ctx: nil, Svc: &shelly.Service{}, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    FleetDeps{Ctx: context.Background(), Svc: nil, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil focus state",
			deps:    FleetDeps{Ctx: context.Background(), Svc: &shelly.Service{}, FocusState: nil},
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

	if f.ID() != tabs.TabFleet {
		t.Errorf("ID() = %v, want tabs.TabFleet", f.ID())
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
	initialPanel := f.focusState.ActivePanel()

	msg := tea.KeyPressMsg{Code: tea.KeyTab}
	updated, _ := f.Update(msg)
	fleet, ok := updated.(*Fleet)
	if !ok {
		t.Fatal("Update should return *Fleet")
	}

	newPanel := fleet.focusState.ActivePanel()
	if newPanel == initialPanel {
		t.Error("Tab should change focused panel")
	}
}

func TestFleet_Update_FocusPrev(t *testing.T) {
	t.Parallel()
	f := newTestFleet()
	// Move to second panel first
	f.focusState.NextPanel()
	panelAfterNext := f.focusState.ActivePanel()

	msg := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	updated, _ := f.Update(msg)
	fleet, ok := updated.(*Fleet)
	if !ok {
		t.Fatal("Update should return *Fleet")
	}

	panelAfterPrev := fleet.focusState.ActivePanel()
	if panelAfterPrev == panelAfterNext {
		t.Error("Shift+Tab should change focused panel")
	}
}

func TestFleet_FocusCycle(t *testing.T) {
	t.Parallel()
	f := newTestFleet()

	// Start at first panel
	initialPanel := f.focusState.ActivePanel()

	// Cycle through all panels (Fleet has 4 panels)
	f.focusState.NextPanel()
	p1 := f.focusState.ActivePanel()
	if p1 == initialPanel {
		t.Error("NextPanel should change panel")
	}

	f.focusState.NextPanel()
	p2 := f.focusState.ActivePanel()
	if p2 == p1 {
		t.Error("NextPanel should change panel again")
	}

	f.focusState.NextPanel()
	p3 := f.focusState.ActivePanel()
	if p3 == p2 {
		t.Error("NextPanel should change panel again")
	}

	// After 4 more NextPanel, should wrap back
	f.focusState.NextPanel()
	p4 := f.focusState.ActivePanel()
	if p4 != initialPanel {
		t.Errorf("After cycling through all 4 panels, should wrap back to initial, got %v", p4)
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

	// Set focus to a different panel
	f.focusState.NextPanel()
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
	focusState := focus.NewState()
	// Set to fleet tab so panel cycling works correctly
	focusState.SetActiveTab(tabs.TabFleet)
	deps := FleetDeps{Ctx: ctx, Svc: svc, FocusState: focusState}
	return NewFleet(deps)
}

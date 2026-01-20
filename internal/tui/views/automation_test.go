package views

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

const testAutomationDevice = "192.168.1.100"

func TestNewAutomation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	autoSvc := automation.New(svc, nil, nil)
	kvsSvc := kvs.NewService(svc.WithConnection)
	focusState := focus.NewState()
	focusState.SetActiveTab(tabs.TabAutomation)
	deps := AutomationDeps{Ctx: ctx, Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc, FocusState: focusState}

	a := NewAutomation(deps)

	if a.ctx != ctx {
		t.Error("ctx not set")
	}
	if a.svc != svc {
		t.Error("svc not set")
	}
	if a.ID() != tabs.TabAutomation {
		t.Errorf("ID() = %v, want tabs.TabAutomation", a.ID())
	}
}

func TestNewAutomation_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	svc := &shelly.Service{}
	autoSvc := automation.New(svc, nil, nil)
	kvsSvc := kvs.NewService(svc.WithConnection)
	focusState := focus.NewState()
	deps := AutomationDeps{Ctx: nil, Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc, FocusState: focusState}
	NewAutomation(deps)
}

func TestNewAutomation_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	svc := &shelly.Service{}
	autoSvc := automation.New(svc, nil, nil)
	kvsSvc := kvs.NewService(svc.WithConnection)
	focusState := focus.NewState()
	deps := AutomationDeps{Ctx: context.Background(), Svc: nil, AutoSvc: autoSvc, KVSSvc: kvsSvc, FocusState: focusState}
	NewAutomation(deps)
}

func TestAutomationDeps_Validate(t *testing.T) {
	t.Parallel()
	svc := &shelly.Service{}
	autoSvc := automation.New(svc, nil, nil)
	kvsSvc := kvs.NewService(svc.WithConnection)
	focusState := focus.NewState()
	tests := []struct {
		name    string
		deps    AutomationDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    AutomationDeps{Ctx: context.Background(), Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc, FocusState: focusState},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    AutomationDeps{Ctx: nil, Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    AutomationDeps{Ctx: context.Background(), Svc: nil, AutoSvc: autoSvc, KVSSvc: kvsSvc, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil auto svc",
			deps:    AutomationDeps{Ctx: context.Background(), Svc: svc, AutoSvc: nil, KVSSvc: kvsSvc, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil kvs svc",
			deps:    AutomationDeps{Ctx: context.Background(), Svc: svc, AutoSvc: autoSvc, KVSSvc: nil, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil focus state",
			deps:    AutomationDeps{Ctx: context.Background(), Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc, FocusState: nil},
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

func TestAutomation_Init(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()

	// Init returns a batch of sub-component init commands.
	// The batch may resolve to nil if all components return nil.
	_ = a.Init()

	// Just verify it doesn't panic.
}

func TestAutomation_SetDevice(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()

	cmd := a.SetDevice(testAutomationDevice)

	if a.device != testAutomationDevice {
		t.Errorf("device = %q, want %q", a.device, testAutomationDevice)
	}
	if cmd == nil {
		t.Error("SetDevice should return a command")
	}
}

func TestAutomation_SetDevice_Same(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()
	a.device = testAutomationDevice

	cmd := a.SetDevice(testAutomationDevice)

	if cmd != nil {
		t.Error("SetDevice with same device should return nil")
	}
}

func TestAutomation_SetSize(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()

	result := a.SetSize(120, 40)

	// SetSize returns View interface, cast back
	updated, ok := result.(*Automation)
	if !ok {
		t.Fatal("SetSize should return *Automation")
	}
	if updated.width != 120 {
		t.Errorf("width = %d, want 120", updated.width)
	}
	if updated.height != 40 {
		t.Errorf("height = %d, want 40", updated.height)
	}
}

func TestAutomation_FocusCycling(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()

	// Initial panel should be first in automation tab (DeviceList)
	// But since automation doesn't have device list as first, it's PanelAutoScripts
	// Actually, let's check what the focusState says
	initialPanel := a.focusState.ActivePanel()

	// Send Tab key to cycle focus
	msg := tea.KeyPressMsg{Code: tea.KeyTab}
	a.handleKeyPress(msg)

	newPanel := a.focusState.ActivePanel()
	if newPanel == initialPanel {
		t.Error("Tab should change focused panel")
	}
}

func TestAutomation_HandleKeyPress_ShiftTab(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()

	// Move to second panel first
	a.focusState.NextPanel()
	panelAfterNext := a.focusState.ActivePanel()

	// Send Shift+Tab key to go back
	msg := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	a.handleKeyPress(msg)

	panelAfterPrev := a.focusState.ActivePanel()
	if panelAfterPrev == panelAfterNext {
		t.Error("Shift+Tab should change focused panel")
	}
}

func TestAutomation_View_NoDevice(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()
	updated, ok := a.SetSize(80, 24).(*Automation)
	if !ok {
		t.Fatal("SetSize should return *Automation")
	}

	view := updated.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestAutomation_View_WithDevice(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()
	updated, ok := a.SetSize(120, 40).(*Automation)
	if !ok {
		t.Fatal("SetSize should return *Automation")
	}
	updated.device = testAutomationDevice

	view := updated.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestAutomation_Device(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()

	if a.Device() != "" {
		t.Errorf("Device() = %q, want empty", a.Device())
	}

	a.device = testAutomationDevice

	if a.Device() != testAutomationDevice {
		t.Errorf("Device() = %q, want %q", a.Device(), testAutomationDevice)
	}
}

func TestDefaultAutomationStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultAutomationStyles()

	// Just verify styles are created without panic
	_ = styles.Panel.Render("test")
	_ = styles.PanelActive.Render("test")
	_ = styles.Title.Render("test")
	_ = styles.Muted.Render("test")
}

func TestAutomationDeps_Errors(t *testing.T) {
	t.Parallel()
	svc := &shelly.Service{}
	autoSvc := automation.New(svc, nil, nil)
	kvsSvc := kvs.NewService(svc.WithConnection)
	focusState := focus.NewState()

	t.Run("nil context error", func(t *testing.T) {
		t.Parallel()
		deps := AutomationDeps{Ctx: nil, Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc, FocusState: focusState}
		err := deps.Validate()
		if !errors.Is(err, errNilContext) {
			t.Errorf("Validate() error = %v, want errNilContext", err)
		}
	})

	t.Run("nil service error", func(t *testing.T) {
		t.Parallel()
		deps := AutomationDeps{Ctx: context.Background(), Svc: nil, AutoSvc: autoSvc, KVSSvc: kvsSvc, FocusState: focusState}
		err := deps.Validate()
		if !errors.Is(err, errNilService) {
			t.Errorf("Validate() error = %v, want errNilService", err)
		}
	})

	t.Run("nil focus state error", func(t *testing.T) {
		t.Parallel()
		deps := AutomationDeps{Ctx: context.Background(), Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc, FocusState: nil}
		err := deps.Validate()
		if !errors.Is(err, errNilFocusState) {
			t.Errorf("Validate() error = %v, want errNilFocusState", err)
		}
	})
}

// newTestAutomation creates a test automation view.
func newTestAutomation() *Automation {
	ctx := context.Background()
	svc := &shelly.Service{}
	autoSvc := automation.New(svc, nil, nil)
	kvsSvc := kvs.NewService(svc.WithConnection)
	focusState := focus.NewState()
	// Set to automation tab so panel cycling works correctly
	focusState.SetActiveTab(tabs.TabAutomation)
	deps := AutomationDeps{Ctx: ctx, Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc, FocusState: focusState}
	return NewAutomation(deps)
}

package views

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

const testAutomationDevice = "192.168.1.100"

func TestNewAutomation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	autoSvc := automation.New(svc, nil, nil)
	kvsSvc := kvs.NewService(svc.WithConnection)
	deps := AutomationDeps{Ctx: ctx, Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc}

	a := NewAutomation(deps)

	if a.ctx != ctx {
		t.Error("ctx not set")
	}
	if a.svc != svc {
		t.Error("svc not set")
	}
	if a.focusedPanel != PanelScripts {
		t.Errorf("focusedPanel = %v, want PanelScripts", a.focusedPanel)
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
	deps := AutomationDeps{Ctx: nil, Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc}
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
	deps := AutomationDeps{Ctx: context.Background(), Svc: nil, AutoSvc: autoSvc, KVSSvc: kvsSvc}
	NewAutomation(deps)
}

func TestAutomationDeps_Validate(t *testing.T) {
	t.Parallel()
	svc := &shelly.Service{}
	autoSvc := automation.New(svc, nil, nil)
	kvsSvc := kvs.NewService(svc.WithConnection)
	tests := []struct {
		name    string
		deps    AutomationDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    AutomationDeps{Ctx: context.Background(), Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    AutomationDeps{Ctx: nil, Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    AutomationDeps{Ctx: context.Background(), Svc: nil, AutoSvc: autoSvc, KVSSvc: kvsSvc},
			wantErr: true,
		},
		{
			name:    "nil auto svc",
			deps:    AutomationDeps{Ctx: context.Background(), Svc: svc, AutoSvc: nil, KVSSvc: kvsSvc},
			wantErr: true,
		},
		{
			name:    "nil kvs svc",
			deps:    AutomationDeps{Ctx: context.Background(), Svc: svc, AutoSvc: autoSvc, KVSSvc: nil},
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

func TestAutomation_FocusNext(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()

	// Start at PanelScripts
	if a.focusedPanel != PanelScripts {
		t.Fatalf("initial panel = %v, want PanelScripts", a.focusedPanel)
	}

	// Panel order: Scripts -> Schedules -> Webhooks -> Virtuals -> KVS -> Alerts
	a.focusNext()
	if a.focusedPanel != PanelSchedules {
		t.Errorf("after focusNext = %v, want PanelSchedules", a.focusedPanel)
	}

	a.focusNext()
	if a.focusedPanel != PanelWebhooks {
		t.Errorf("after focusNext = %v, want PanelWebhooks", a.focusedPanel)
	}
}

func TestAutomation_FocusPrev(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()
	a.focusedPanel = PanelSchedules

	// Panel order: Scripts -> Schedules -> Webhooks -> Virtuals -> KVS -> Alerts
	// Prev from Schedules is Scripts
	a.focusPrev()
	if a.focusedPanel != PanelScripts {
		t.Errorf("after focusPrev = %v, want PanelScripts", a.focusedPanel)
	}

	// Prev from Scripts wraps to Alerts (last panel)
	a.focusPrev()
	if a.focusedPanel != PanelAlerts {
		t.Errorf("after focusPrev = %v, want PanelAlerts (wrap)", a.focusedPanel)
	}
}

func TestAutomation_FocusWrap(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()
	a.focusedPanel = PanelAlerts // Last panel

	// Focus next should wrap to PanelScripts
	a.focusNext()
	if a.focusedPanel != PanelScripts {
		t.Errorf("after wrap = %v, want PanelScripts", a.focusedPanel)
	}
}

func TestAutomation_HandleKeyPress_Tab(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()
	msg := tea.KeyPressMsg{Code: tea.KeyTab}

	a.handleKeyPress(msg)

	// Panel order: Scripts -> Schedules -> Webhooks -> Virtuals -> KVS -> Alerts
	// Tab from Scripts goes to Schedules
	if a.focusedPanel != PanelSchedules {
		t.Errorf("focusedPanel = %v, want PanelSchedules", a.focusedPanel)
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

func TestAutomation_FocusedPanel(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()

	if a.FocusedPanel() != PanelScripts {
		t.Errorf("FocusedPanel() = %v, want PanelScripts", a.FocusedPanel())
	}

	a = a.SetFocusedPanel(PanelWebhooks)

	if a.FocusedPanel() != PanelWebhooks {
		t.Errorf("FocusedPanel() = %v, want PanelWebhooks", a.FocusedPanel())
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
	kvsSvc := kvs.NewService(svc.WithConnection)

	t.Run("nil context error", func(t *testing.T) {
		t.Parallel()
		deps := AutomationDeps{Ctx: nil, Svc: svc, KVSSvc: kvsSvc}
		err := deps.Validate()
		if !errors.Is(err, errNilContext) {
			t.Errorf("Validate() error = %v, want errNilContext", err)
		}
	})

	t.Run("nil service error", func(t *testing.T) {
		t.Parallel()
		deps := AutomationDeps{Ctx: context.Background(), Svc: nil, KVSSvc: kvsSvc}
		err := deps.Validate()
		if !errors.Is(err, errNilService) {
			t.Errorf("Validate() error = %v, want errNilService", err)
		}
	})
}

// newTestAutomation creates a test automation view.
func newTestAutomation() *Automation {
	ctx := context.Background()
	svc := &shelly.Service{}
	autoSvc := automation.New(svc, nil, nil)
	kvsSvc := kvs.NewService(svc.WithConnection)
	deps := AutomationDeps{Ctx: ctx, Svc: svc, AutoSvc: autoSvc, KVSSvc: kvsSvc}
	return NewAutomation(deps)
}

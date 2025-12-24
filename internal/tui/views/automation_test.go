package views

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

const testAutomationDevice = "192.168.1.100"

func TestNewAutomation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := AutomationDeps{Ctx: ctx, Svc: svc}

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

	deps := AutomationDeps{Ctx: nil, Svc: &shelly.Service{}}
	NewAutomation(deps)
}

func TestNewAutomation_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	deps := AutomationDeps{Ctx: context.Background(), Svc: nil}
	NewAutomation(deps)
}

func TestAutomationDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    AutomationDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    AutomationDeps{Ctx: context.Background(), Svc: &shelly.Service{}},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    AutomationDeps{Ctx: nil, Svc: &shelly.Service{}},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    AutomationDeps{Ctx: context.Background(), Svc: nil},
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

	a.focusNext()
	if a.focusedPanel != PanelScriptEditor {
		t.Errorf("after focusNext = %v, want PanelScriptEditor", a.focusedPanel)
	}

	a.focusNext()
	if a.focusedPanel != PanelSchedules {
		t.Errorf("after focusNext = %v, want PanelSchedules", a.focusedPanel)
	}
}

func TestAutomation_FocusPrev(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()
	a.focusedPanel = PanelSchedules

	a.focusPrev()
	if a.focusedPanel != PanelScriptEditor {
		t.Errorf("after focusPrev = %v, want PanelScriptEditor", a.focusedPanel)
	}

	a.focusPrev()
	if a.focusedPanel != PanelScripts {
		t.Errorf("after focusPrev = %v, want PanelScripts", a.focusedPanel)
	}
}

func TestAutomation_FocusWrap(t *testing.T) {
	t.Parallel()
	a := newTestAutomation()
	a.focusedPanel = PanelKVS

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

	if a.focusedPanel != PanelScriptEditor {
		t.Errorf("focusedPanel = %v, want PanelScriptEditor", a.focusedPanel)
	}
}

func TestAutomation_HandleKeyPress_Numbers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		key  string
		want AutomationPanel
	}{
		{"1", PanelScripts},
		{"2", PanelSchedules},
		{"3", PanelWebhooks},
		{"4", PanelVirtuals},
		{"5", PanelKVS},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			a := newTestAutomation()

			// Create key press message based on key
			var msg tea.KeyPressMsg
			switch tt.key {
			case "1":
				msg = tea.KeyPressMsg{Code: 49} // '1'
			case "2":
				msg = tea.KeyPressMsg{Code: 50} // '2'
			case "3":
				msg = tea.KeyPressMsg{Code: 51} // '3'
			case "4":
				msg = tea.KeyPressMsg{Code: 52} // '4'
			case "5":
				msg = tea.KeyPressMsg{Code: 53} // '5'
			}

			a.handleKeyPress(msg)

			if a.focusedPanel != tt.want {
				t.Errorf("focusedPanel = %v, want %v", a.focusedPanel, tt.want)
			}
		})
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

	t.Run("nil context error", func(t *testing.T) {
		t.Parallel()
		deps := AutomationDeps{Ctx: nil, Svc: &shelly.Service{}}
		err := deps.Validate()
		if !errors.Is(err, errNilContext) {
			t.Errorf("Validate() error = %v, want errNilContext", err)
		}
	})

	t.Run("nil service error", func(t *testing.T) {
		t.Parallel()
		deps := AutomationDeps{Ctx: context.Background(), Svc: nil}
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
	deps := AutomationDeps{Ctx: ctx, Svc: svc}
	return NewAutomation(deps)
}

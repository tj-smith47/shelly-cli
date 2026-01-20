package views

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

const testConfigDevice = "192.168.1.100"

func TestNewConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	focusState := focus.NewState()
	focusState.SetActiveTab(tabs.TabConfig)
	deps := ConfigDeps{Ctx: ctx, Svc: svc, FocusState: focusState}

	c := NewConfig(deps)

	if c.ctx != ctx {
		t.Error("ctx not set")
	}
	if c.svc != svc {
		t.Error("svc not set")
	}
	if c.ID() != tabs.TabConfig {
		t.Errorf("ID() = %v, want tabs.TabConfig", c.ID())
	}
}

func TestNewConfig_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	focusState := focus.NewState()
	deps := ConfigDeps{Ctx: nil, Svc: &shelly.Service{}, FocusState: focusState}
	NewConfig(deps)
}

func TestNewConfig_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	focusState := focus.NewState()
	deps := ConfigDeps{Ctx: context.Background(), Svc: nil, FocusState: focusState}
	NewConfig(deps)
}

func TestConfigDeps_Validate(t *testing.T) {
	t.Parallel()
	focusState := focus.NewState()
	tests := []struct {
		name    string
		deps    ConfigDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    ConfigDeps{Ctx: context.Background(), Svc: &shelly.Service{}, FocusState: focusState},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    ConfigDeps{Ctx: nil, Svc: &shelly.Service{}, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    ConfigDeps{Ctx: context.Background(), Svc: nil, FocusState: focusState},
			wantErr: true,
		},
		{
			name:    "nil focus state",
			deps:    ConfigDeps{Ctx: context.Background(), Svc: &shelly.Service{}, FocusState: nil},
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

func TestConfig_Init(t *testing.T) {
	t.Parallel()
	c := newTestConfig()

	// Init returns a batch of sub-component init commands.
	_ = c.Init()

	// Just verify it doesn't panic.
}

func TestConfig_SetDevice(t *testing.T) {
	t.Parallel()
	c := newTestConfig()

	cmd := c.SetDevice(testConfigDevice)

	if c.device != testConfigDevice {
		t.Errorf("device = %q, want %q", c.device, testConfigDevice)
	}
	if cmd == nil {
		t.Error("SetDevice should return a command")
	}
}

func TestConfig_SetDevice_Same(t *testing.T) {
	t.Parallel()
	c := newTestConfig()
	c.device = testConfigDevice

	cmd := c.SetDevice(testConfigDevice)

	if cmd != nil {
		t.Error("SetDevice with same device should return nil")
	}
}

func TestConfig_SetSize(t *testing.T) {
	t.Parallel()
	c := newTestConfig()

	result := c.SetSize(120, 40)

	updated, ok := result.(*Config)
	if !ok {
		t.Fatal("SetSize should return *Config")
	}
	if updated.width != 120 {
		t.Errorf("width = %d, want 120", updated.width)
	}
	if updated.height != 40 {
		t.Errorf("height = %d, want 40", updated.height)
	}
}

func TestConfig_FocusCycling(t *testing.T) {
	t.Parallel()
	c := newTestConfig()

	// Get initial panel
	initialPanel := c.focusState.ActivePanel()

	// Send Tab key to cycle focus
	msg := tea.KeyPressMsg{Code: tea.KeyTab}
	c.handleKeyPress(msg)

	newPanel := c.focusState.ActivePanel()
	if newPanel == initialPanel {
		t.Error("Tab should change focused panel")
	}
}

func TestConfig_HandleKeyPress_ShiftTab(t *testing.T) {
	t.Parallel()
	c := newTestConfig()

	// Move to second panel first
	c.focusState.NextPanel()
	panelAfterNext := c.focusState.ActivePanel()

	// Send Shift+Tab key to go back
	msg := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	c.handleKeyPress(msg)

	panelAfterPrev := c.focusState.ActivePanel()
	if panelAfterPrev == panelAfterNext {
		t.Error("Shift+Tab should change focused panel")
	}
}

func TestConfig_View_NoDevice(t *testing.T) {
	t.Parallel()
	c := newTestConfig()
	updated, ok := c.SetSize(80, 24).(*Config)
	if !ok {
		t.Fatal("SetSize should return *Config")
	}

	view := updated.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestConfig_View_WithDevice(t *testing.T) {
	t.Parallel()
	c := newTestConfig()
	updated, ok := c.SetSize(120, 40).(*Config)
	if !ok {
		t.Fatal("SetSize should return *Config")
	}
	updated.device = testConfigDevice

	view := updated.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestConfig_Device(t *testing.T) {
	t.Parallel()
	c := newTestConfig()

	if c.Device() != "" {
		t.Errorf("Device() = %q, want empty", c.Device())
	}

	c.device = testConfigDevice

	if c.Device() != testConfigDevice {
		t.Errorf("Device() = %q, want %q", c.Device(), testConfigDevice)
	}
}

func TestDefaultConfigStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultConfigStyles()

	// Just verify styles are created without panic
	_ = styles.Panel.Render("test")
	_ = styles.PanelActive.Render("test")
	_ = styles.Title.Render("test")
	_ = styles.Muted.Render("test")
}

func newTestConfig() *Config {
	ctx := context.Background()
	svc := &shelly.Service{}
	focusState := focus.NewState()
	// Set to config tab so panel cycling works correctly
	focusState.SetActiveTab(tabs.TabConfig)
	deps := ConfigDeps{Ctx: ctx, Svc: svc, FocusState: focusState}
	return NewConfig(deps)
}

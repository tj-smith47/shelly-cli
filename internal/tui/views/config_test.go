package views

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

const testConfigDevice = "192.168.1.100"

func TestNewConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := ConfigDeps{Ctx: ctx, Svc: svc}

	c := NewConfig(deps)

	if c.ctx != ctx {
		t.Error("ctx not set")
	}
	if c.svc != svc {
		t.Error("svc not set")
	}
	if c.focusedPanel != PanelWiFi {
		t.Errorf("focusedPanel = %v, want PanelWiFi", c.focusedPanel)
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

	deps := ConfigDeps{Ctx: nil, Svc: &shelly.Service{}}
	NewConfig(deps)
}

func TestNewConfig_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	deps := ConfigDeps{Ctx: context.Background(), Svc: nil}
	NewConfig(deps)
}

func TestConfigDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    ConfigDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    ConfigDeps{Ctx: context.Background(), Svc: &shelly.Service{}},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    ConfigDeps{Ctx: nil, Svc: &shelly.Service{}},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    ConfigDeps{Ctx: context.Background(), Svc: nil},
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

func TestConfig_FocusNext(t *testing.T) {
	t.Parallel()
	c := newTestConfig()

	// Start at PanelWiFi
	if c.focusedPanel != PanelWiFi {
		t.Fatalf("initial panel = %v, want PanelWiFi", c.focusedPanel)
	}

	c.focusNext()
	if c.focusedPanel != PanelSystem {
		t.Errorf("after focusNext = %v, want PanelSystem", c.focusedPanel)
	}

	c.focusNext()
	if c.focusedPanel != PanelCloud {
		t.Errorf("after focusNext = %v, want PanelCloud", c.focusedPanel)
	}

	c.focusNext()
	if c.focusedPanel != PanelInputs {
		t.Errorf("after focusNext = %v, want PanelInputs", c.focusedPanel)
	}
}

func TestConfig_FocusPrev(t *testing.T) {
	t.Parallel()
	c := newTestConfig()
	c.focusedPanel = PanelInputs

	c.focusPrev()
	if c.focusedPanel != PanelCloud {
		t.Errorf("after focusPrev = %v, want PanelCloud", c.focusedPanel)
	}

	c.focusPrev()
	if c.focusedPanel != PanelSystem {
		t.Errorf("after focusPrev = %v, want PanelSystem", c.focusedPanel)
	}
}

func TestConfig_FocusWrap(t *testing.T) {
	t.Parallel()
	c := newTestConfig()
	c.focusedPanel = PanelInputs

	// Focus next should wrap to PanelWiFi
	c.focusNext()
	if c.focusedPanel != PanelWiFi {
		t.Errorf("after wrap = %v, want PanelWiFi", c.focusedPanel)
	}
}

func TestConfig_HandleKeyPress_Tab(t *testing.T) {
	t.Parallel()
	c := newTestConfig()
	msg := tea.KeyPressMsg{Code: tea.KeyTab}

	c.handleKeyPress(msg)

	if c.focusedPanel != PanelSystem {
		t.Errorf("focusedPanel = %v, want PanelSystem", c.focusedPanel)
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

func TestConfig_FocusedPanel(t *testing.T) {
	t.Parallel()
	c := newTestConfig()

	if c.FocusedPanel() != PanelWiFi {
		t.Errorf("FocusedPanel() = %v, want PanelWiFi", c.FocusedPanel())
	}

	c = c.SetFocusedPanel(PanelCloud)

	if c.FocusedPanel() != PanelCloud {
		t.Errorf("FocusedPanel() = %v, want PanelCloud", c.FocusedPanel())
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
	deps := ConfigDeps{Ctx: ctx, Svc: svc}
	return NewConfig(deps)
}

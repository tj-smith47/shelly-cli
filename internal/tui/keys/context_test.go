package keys

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
)

func TestNewContextMap(t *testing.T) {
	t.Parallel()
	m := NewContextMap()
	if m == nil {
		t.Fatal("NewContextMap() returned nil")
	}
	if m.bindings == nil {
		t.Error("bindings map should be initialized")
	}
}

func TestContextMap_Match_Global(t *testing.T) {
	t.Parallel()
	m := NewContextMap()

	tests := []struct {
		key  string
		want Action
	}{
		{"q", ActionQuit},
		{"?", ActionHelp},
		{"/", ActionFilter},
		{":", ActionCommand},
		{"1", ActionTab1},
		{"2", ActionTab2},
		{"3", ActionTab3},
		{"4", ActionTab4},
		{"5", ActionTab5},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			// Verify binding exists
			got := m.bindings[ContextGlobal][tt.key]
			if got != tt.want {
				t.Errorf("bindings[ContextGlobal][%q] = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestContextMap_Match_ContextSpecific(t *testing.T) {
	t.Parallel()
	m := NewContextMap()

	// Test that 'j' in Events context returns ActionDown
	eventsBindings := m.bindings[ContextEvents]
	if eventsBindings["j"] != ActionDown {
		t.Errorf("Events context 'j' = %v, want ActionDown", eventsBindings["j"])
	}

	// Test that 't' in Devices context returns ActionToggle
	devicesBindings := m.bindings[ContextDevices]
	if devicesBindings["t"] != ActionToggle {
		t.Errorf("Devices context 't' = %v, want ActionToggle", devicesBindings["t"])
	}
}

func TestContextMap_GetBindings(t *testing.T) {
	t.Parallel()
	m := NewContextMap()

	bindings := m.GetBindings(ContextGlobal)
	if len(bindings) == 0 {
		t.Error("GetBindings(ContextGlobal) returned empty")
	}

	// Check that quit binding exists
	found := false
	for _, b := range bindings {
		if b.Action == ActionQuit {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetBindings should include ActionQuit in global context")
	}
}

func TestContextMap_GetGlobalBindings(t *testing.T) {
	t.Parallel()
	m := NewContextMap()

	bindings := m.GetGlobalBindings()
	if len(bindings) == 0 {
		t.Error("GetGlobalBindings() returned empty")
	}
}

func TestContextFromPanel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		panel focus.PanelID
		want  Context
	}{
		{focus.PanelEvents, ContextEvents},
		{focus.PanelDeviceList, ContextDevices},
		{focus.PanelDeviceInfo, ContextInfo},
		{focus.PanelJSON, ContextJSON},
		{focus.PanelEnergyBars, ContextEnergy},
		{focus.PanelEnergyHistory, ContextEnergy},
		{focus.PanelNone, ContextGlobal},
	}

	for _, tt := range tests {
		t.Run(ContextName(tt.want), func(t *testing.T) {
			t.Parallel()
			got := ContextFromPanel(tt.panel)
			if got != tt.want {
				t.Errorf("ContextFromPanel(%v) = %v, want %v", tt.panel, got, tt.want)
			}
		})
	}
}

func TestActionDesc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		action Action
		want   string
	}{
		{ActionQuit, "Quit"},
		{ActionHelp, "Show help"},
		{ActionUp, "Move up"},
		{ActionDown, "Move down"},
		{ActionToggle, "Toggle"},
		{ActionNone, ""},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := ActionDesc(tt.action)
			if got != tt.want {
				t.Errorf("ActionDesc(%v) = %q, want %q", tt.action, got, tt.want)
			}
		})
	}
}

func TestContextName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ctx  Context
		want string
	}{
		{ContextGlobal, "Global"},
		{ContextEvents, "Events"},
		{ContextDevices, "Devices"},
		{ContextInfo, "Device Info"},
		{ContextEnergy, "Energy"},
		{ContextJSON, "JSON Viewer"},
		{ContextAutomation, "Automation"},
		{ContextConfig, "Config"},
		{ContextManage, "Manage"},
		{ContextFleet, "Fleet"},
		{ContextHelp, "Help"},
		{Context(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := ContextName(tt.ctx)
			if got != tt.want {
				t.Errorf("ContextName(%v) = %q, want %q", tt.ctx, got, tt.want)
			}
		})
	}
}

func TestKeyBinding(t *testing.T) {
	t.Parallel()
	kb := KeyBinding{
		Key:    "q",
		Action: ActionQuit,
		Desc:   "Quit",
	}

	if kb.Key != "q" {
		t.Errorf("Key = %q, want %q", kb.Key, "q")
	}
	if kb.Action != ActionQuit {
		t.Errorf("Action = %v, want ActionQuit", kb.Action)
	}
	if kb.Desc != "Quit" {
		t.Errorf("Desc = %q, want %q", kb.Desc, "Quit")
	}
}

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
		{"esc", ActionEscape},
		{"ctrl+[", ActionEscape},
		{"tab", ActionNextPanel},
		{"shift+tab", ActionPrevPanel},
		{"alt+]", ActionNextPanel},
		{"alt+[", ActionPrevPanel},
		{"1", ActionTab1},
		{"2", ActionTab2},
		{"3", ActionTab3},
		{"4", ActionTab4},
		{"5", ActionTab5},
		{"6", ActionTab6},
		{"!", ActionPanel1},
		{"@", ActionPanel2},
		{"#", ActionPanel3},
		{"D", ActionDebug},
		{"ctrl+c", ActionQuit},
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

	tests := []struct {
		ctx  Context
		key  string
		want Action
	}{
		// Events context
		{ContextEvents, "j", ActionDown},
		{ContextEvents, "k", ActionUp},
		{ContextEvents, "space", ActionPause},
		{ContextEvents, "c", ActionClear},
		// Devices context
		{ContextDevices, "t", ActionToggle},
		{ContextDevices, "o", ActionOn},
		{ContextDevices, "O", ActionOff},
		{ContextDevices, "R", ActionReboot},
		{ContextDevices, "b", ActionBrowser},
		{ContextDevices, "ctrl+u", ActionPageUp},
		{ContextDevices, "ctrl+d", ActionPageDown},
		// Monitor context
		{ContextMonitor, "t", ActionToggle},
		{ContextMonitor, "o", ActionOn},
		{ContextMonitor, "O", ActionOff},
		{ContextMonitor, "R", ActionReboot},
		{ContextMonitor, "b", ActionBrowser},
		{ContextMonitor, "space", ActionPause},
		// Automation context
		{ContextAutomation, "e", ActionEdit},
		{ContextAutomation, "n", ActionNew},
		{ContextAutomation, "d", ActionDelete},
		// Config context
		{ContextConfig, "e", ActionEdit},
		// JSON context
		{ContextJSON, "y", ActionCopy},
		{ContextJSON, "ctrl+[", ActionEscape},
	}

	for _, tt := range tests {
		t.Run(ContextName(tt.ctx)+"/"+tt.key, func(t *testing.T) {
			t.Parallel()
			contextBindings := m.bindings[tt.ctx]
			got := contextBindings[tt.key]
			if got != tt.want {
				t.Errorf("%s context %q = %v, want %v", ContextName(tt.ctx), tt.key, got, tt.want)
			}
		})
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
		{focus.PanelMonitor, ContextMonitor},
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
		{ActionEdit, "Edit"},
		{ActionNew, "Create new"},
		{ActionDelete, "Delete"},
		{ActionBrowser, "Open in browser"},
		{ActionDebug, "Toggle debug"},
		{ActionTab6, "Fleet tab"},
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
		{ContextMonitor, "Monitor"},
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

func TestContextActionDesc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ctx    Context
		action Action
		want   string
	}{
		// Context-specific overrides
		{ContextAutomation, ActionEnter, "View script"},
		{ContextAutomation, ActionEdit, "Edit script/schedule"},
		{ContextAutomation, ActionNew, "Create new"},
		{ContextAutomation, ActionDelete, "Delete item"},
		{ContextEvents, ActionPause, "Pause events"},
		{ContextEvents, ActionClear, "Clear events"},
		{ContextDevices, ActionBrowser, "Open web UI"},
		{ContextMonitor, ActionPause, "Pause monitoring"},
		{ContextMonitor, ActionBrowser, "Open web UI"},
		{ContextConfig, ActionEdit, "Edit configuration"},
		// Falls back to generic description
		{ContextGlobal, ActionToggle, "Toggle"},
		{ContextFleet, ActionRefresh, "Refresh fleet"},
	}

	for _, tt := range tests {
		t.Run(ContextName(tt.ctx)+"/"+ActionDesc(tt.action), func(t *testing.T) {
			t.Parallel()
			got := ContextActionDesc(tt.ctx, tt.action)
			if got != tt.want {
				t.Errorf("ContextActionDesc(%v, %v) = %q, want %q", tt.ctx, tt.action, got, tt.want)
			}
		})
	}
}

func TestAllContextsHaveBindings(t *testing.T) {
	t.Parallel()
	m := NewContextMap()

	// Ensure all view contexts have at least basic navigation bindings
	contexts := []Context{
		ContextEvents,
		ContextDevices,
		ContextInfo,
		ContextEnergy,
		ContextJSON,
		ContextAutomation,
		ContextConfig,
		ContextManage,
		ContextMonitor,
		ContextFleet,
		ContextHelp,
	}

	for _, ctx := range contexts {
		t.Run(ContextName(ctx), func(t *testing.T) {
			t.Parallel()
			bindings := m.bindings[ctx]
			if len(bindings) == 0 {
				t.Errorf("Context %s has no bindings", ContextName(ctx))
			}
			// Check j/k navigation exists (common to all contexts)
			if bindings["j"] == ActionNone && bindings["down"] == ActionNone {
				t.Errorf("Context %s missing down navigation", ContextName(ctx))
			}
		})
	}
}

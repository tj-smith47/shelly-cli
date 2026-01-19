// Package keys provides context-sensitive key handling for the TUI.
package keys

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
)

// Context represents the current keybinding context.
type Context int

// Context values for different TUI views.
const (
	ContextGlobal Context = iota
	ContextEvents
	ContextDevices
	ContextInfo
	ContextEnergy
	ContextJSON
	ContextAutomation
	ContextConfig
	ContextManage
	ContextMonitor
	ContextFleet
	ContextHelp
)

// Action represents a user action.
type Action int

// Action values for user interactions.
const (
	ActionNone Action = iota
	ActionQuit
	ActionHelp
	ActionFilter
	ActionCommand
	ActionNextPanel
	ActionPrevPanel
	ActionUp
	ActionDown
	ActionLeft
	ActionRight
	ActionToggle
	ActionOn
	ActionOff
	ActionReboot
	ActionEnter
	ActionEscape
	ActionPageUp
	ActionPageDown
	ActionHome
	ActionEnd
	ActionCopy
	ActionPause
	ActionClear
	ActionSort
	ActionExpand
	ActionRefresh
	ActionRefreshAll
	ActionFilterToggle
	ActionEdit
	ActionNew
	ActionDelete
	ActionBrowser
	ActionDebug
	ActionTab1
	ActionTab2
	ActionTab3
	ActionTab4
	ActionTab5
	ActionTab6
	ActionPanel1
	ActionPanel2
	ActionPanel3
	ActionPanel4
	ActionPanel5
	ActionPanel6
	ActionPanel7
	ActionPanel8
	ActionPanel9
	ActionControl
	ActionDetail
)

// KeyBinding represents a key and its description.
type KeyBinding struct {
	Key    string
	Action Action
	Desc   string
}

// ContextMap holds keybindings for each context.
type ContextMap struct {
	bindings map[Context]map[string]Action
}

// NewContextMap creates a map with default bindings.
func NewContextMap() *ContextMap {
	m := &ContextMap{
		bindings: make(map[Context]map[string]Action),
	}
	m.initDefaults()
	return m
}

func (m *ContextMap) initDefaults() {
	// Global bindings
	m.bindings[ContextGlobal] = map[string]Action{
		"q":         ActionQuit,
		"?":         ActionHelp,
		"/":         ActionFilter,
		":":         ActionCommand,
		"esc":       ActionEscape,
		"ctrl+[":    ActionEscape,
		"tab":       ActionNextPanel,
		"shift+tab": ActionPrevPanel,
		"alt+]":     ActionNextPanel,
		"alt+[":     ActionPrevPanel,
		"1":         ActionTab1,
		"2":         ActionTab2,
		"3":         ActionTab3,
		"4":         ActionTab4,
		"5":         ActionTab5,
		"6":         ActionTab6,
		"!":         ActionPanel1, // Shift+1
		"@":         ActionPanel2, // Shift+2
		"#":         ActionPanel3, // Shift+3
		"$":         ActionPanel4, // Shift+4
		"%":         ActionPanel5, // Shift+5
		"^":         ActionPanel6, // Shift+6
		"&":         ActionPanel7, // Shift+7
		"*":         ActionPanel8, // Shift+8
		"(":         ActionPanel9, // Shift+9
		"D":         ActionDebug,
		"ctrl+c":    ActionQuit,
	}

	// Events context - dual column layout with h/l for column switching
	m.bindings[ContextEvents] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"h":      ActionLeft,  // Switch to user (left) column
		"l":      ActionRight, // Switch to system (right) column
		"left":   ActionLeft,
		"right":  ActionRight,
		"space":  ActionPause,
		"c":      ActionClear,
		"f":      ActionFilterToggle,
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
		"ctrl+d": ActionPageDown,
		"ctrl+u": ActionPageUp,
		"g":      ActionHome,
		"G":      ActionEnd,
	}

	// Devices context
	m.bindings[ContextDevices] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"h":      ActionLeft,
		"l":      ActionRight,
		"left":   ActionLeft,
		"right":  ActionRight,
		"t":      ActionToggle,
		"o":      ActionOn,
		"O":      ActionOff,
		"R":      ActionReboot,
		"c":      ActionControl, // Open control panel
		"d":      ActionDetail,  // Device detail overlay
		"enter":  ActionEnter,
		"r":      ActionRefresh,
		"ctrl+r": ActionRefreshAll,
		"b":      ActionBrowser, // Open in browser
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
		"ctrl+d": ActionPageDown, // Half-page down
		"ctrl+u": ActionPageUp,   // Half-page up
		"g":      ActionHome,
		"G":      ActionEnd,
		"home":   ActionHome,
		"end":    ActionEnd,
	}

	// Info context
	m.bindings[ContextInfo] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"h":      ActionLeft,
		"l":      ActionRight,
		"left":   ActionLeft,
		"right":  ActionRight,
		"enter":  ActionEnter,
		"a":      ActionExpand, // Toggle all/single view
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
		"ctrl+d": ActionPageDown,
		"ctrl+u": ActionPageUp,
		"g":      ActionHome,
		"G":      ActionEnd,
		"home":   ActionHome,
		"end":    ActionEnd,
	}

	// Energy context
	m.bindings[ContextEnergy] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"h":      ActionLeft,
		"l":      ActionRight,
		"left":   ActionLeft,
		"right":  ActionRight,
		"s":      ActionSort,
		"e":      ActionExpand,
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
		"ctrl+d": ActionPageDown,
		"ctrl+u": ActionPageUp,
		"g":      ActionHome,
		"G":      ActionEnd,
	}

	// JSON context
	m.bindings[ContextJSON] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"h":      ActionLeft,
		"l":      ActionRight,
		"left":   ActionLeft,
		"right":  ActionRight,
		"q":      ActionEscape,
		"esc":    ActionEscape,
		"ctrl+[": ActionEscape,
		"y":      ActionCopy,
		"r":      ActionRefresh,
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
		"ctrl+d": ActionPageDown,
		"ctrl+u": ActionPageUp,
		"g":      ActionHome,
		"G":      ActionEnd,
	}

	// Automation context - scripts, schedules, webhooks, KVS
	m.bindings[ContextAutomation] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"h":      ActionLeft,
		"l":      ActionRight,
		"left":   ActionLeft,
		"right":  ActionRight,
		"enter":  ActionEnter,
		"e":      ActionEdit,   // Edit script/schedule/webhook
		"n":      ActionNew,    // Create new
		"d":      ActionDelete, // Delete
		"r":      ActionRefresh,
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
		"ctrl+d": ActionPageDown,
		"ctrl+u": ActionPageUp,
		"g":      ActionHome,
		"G":      ActionEnd,
	}

	// Config context - device configuration panels
	m.bindings[ContextConfig] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"h":      ActionLeft,
		"l":      ActionRight,
		"left":   ActionLeft,
		"right":  ActionRight,
		"enter":  ActionEnter,
		"e":      ActionEdit, // Edit configuration
		"r":      ActionRefresh,
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
		"ctrl+d": ActionPageDown,
		"ctrl+u": ActionPageUp,
		"g":      ActionHome,
		"G":      ActionEnd,
	}

	// Manage context - discovery, firmware, backup, batch operations
	m.bindings[ContextManage] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"h":      ActionLeft,
		"l":      ActionRight,
		"left":   ActionLeft,
		"right":  ActionRight,
		"enter":  ActionEnter,
		"r":      ActionRefresh, // Refresh device list
		"space":  ActionToggle,  // Select device
		"a":      ActionExpand,  // Select all
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
		"ctrl+d": ActionPageDown,
		"ctrl+u": ActionPageUp,
		"g":      ActionHome,
		"G":      ActionEnd,
	}

	// Monitor context - real-time monitoring with device actions
	m.bindings[ContextMonitor] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"h":      ActionLeft,
		"l":      ActionRight,
		"left":   ActionLeft,
		"right":  ActionRight,
		"t":      ActionToggle,
		"o":      ActionOn,
		"O":      ActionOff,
		"R":      ActionReboot,
		"c":      ActionControl, // Open control panel
		"d":      ActionDetail,  // Device detail overlay
		"enter":  ActionEnter,
		"r":      ActionRefresh,
		"ctrl+r": ActionRefreshAll,
		"b":      ActionBrowser, // Open in browser
		"space":  ActionPause,   // Pause/resume monitoring
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
		"ctrl+d": ActionPageDown, // Half-page down
		"ctrl+u": ActionPageUp,   // Half-page up
		"g":      ActionHome,
		"G":      ActionEnd,
		"home":   ActionHome,
		"end":    ActionEnd,
	}

	// Fleet context - cloud fleet management
	m.bindings[ContextFleet] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"enter":  ActionEnter,
		"r":      ActionRefresh,
		"space":  ActionToggle, // Select device
		"a":      ActionExpand, // Select all
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
		"ctrl+d": ActionPageDown,
		"ctrl+u": ActionPageUp,
		"g":      ActionHome,
		"G":      ActionEnd,
	}

	// Help context
	m.bindings[ContextHelp] = map[string]Action{
		"?":      ActionEscape,
		"q":      ActionEscape,
		"esc":    ActionEscape,
		"ctrl+[": ActionEscape,
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
		"ctrl+d": ActionPageDown,
		"ctrl+u": ActionPageUp,
		"g":      ActionHome,
		"G":      ActionEnd,
	}
}

// Match returns the action for a key in the given context.
// Falls back to ContextGlobal if not found in specific context.
func (m *ContextMap) Match(ctx Context, msg tea.KeyPressMsg) Action {
	keyStr := keyString(msg)

	// Check context-specific bindings first
	if contextBindings, ok := m.bindings[ctx]; ok {
		if action, ok := contextBindings[keyStr]; ok {
			return action
		}
	}

	// Fall back to global bindings
	if globalBindings, ok := m.bindings[ContextGlobal]; ok {
		if action, ok := globalBindings[keyStr]; ok {
			return action
		}
	}

	return ActionNone
}

// GetBindings returns all bindings for a context (for help display).
func (m *ContextMap) GetBindings(ctx Context) []KeyBinding {
	var bindings []KeyBinding

	if contextBindings, ok := m.bindings[ctx]; ok {
		for key, action := range contextBindings {
			bindings = append(bindings, KeyBinding{
				Key:    key,
				Action: action,
				Desc:   ContextActionDesc(ctx, action),
			})
		}
	}

	return bindings
}

// GetGlobalBindings returns global bindings.
func (m *ContextMap) GetGlobalBindings() []KeyBinding {
	return m.GetBindings(ContextGlobal)
}

// ContextFromPanel converts a focus Panel to a keys Context.
func ContextFromPanel(p focus.PanelID) Context {
	switch p {
	case focus.PanelEvents:
		return ContextEvents
	case focus.PanelDeviceList:
		return ContextDevices
	case focus.PanelDeviceInfo:
		return ContextInfo
	case focus.PanelJSON:
		return ContextJSON
	case focus.PanelEnergyBars, focus.PanelEnergyHistory:
		return ContextEnergy
	case focus.PanelMonitor:
		return ContextMonitor
	default:
		return ContextGlobal
	}
}

// keyString converts a tea.KeyPressMsg to a canonical string.
func keyString(msg tea.KeyPressMsg) string {
	return msg.String()
}

// actionDescriptions maps actions to their human-readable descriptions.
var actionDescriptions = map[Action]string{
	ActionQuit:         "Quit",
	ActionHelp:         "Show help",
	ActionFilter:       "Filter",
	ActionCommand:      "Command mode",
	ActionNextPanel:    "Next panel",
	ActionPrevPanel:    "Previous panel",
	ActionUp:           "Move up",
	ActionDown:         "Move down",
	ActionLeft:         "Move left",
	ActionRight:        "Move right",
	ActionToggle:       "Toggle",
	ActionOn:           "Turn on",
	ActionOff:          "Turn off",
	ActionReboot:       "Reboot device",
	ActionEnter:        "Select/Enter",
	ActionEscape:       "Close/Cancel",
	ActionPageUp:       "Page up",
	ActionPageDown:     "Page down",
	ActionHome:         "Go to top",
	ActionEnd:          "Go to bottom",
	ActionCopy:         "Copy",
	ActionPause:        "Pause",
	ActionClear:        "Clear",
	ActionSort:         "Sort",
	ActionExpand:       "Expand/Toggle all",
	ActionRefresh:      "Refresh",
	ActionRefreshAll:   "Refresh all",
	ActionFilterToggle: "Toggle filter",
	ActionEdit:         "Edit",
	ActionNew:          "Create new",
	ActionDelete:       "Delete",
	ActionBrowser:      "Open in browser",
	ActionDebug:        "Toggle debug",
	ActionTab1:         "Dashboard tab",
	ActionTab2:         "Automation tab",
	ActionTab3:         "Config tab",
	ActionTab4:         "Manage tab",
	ActionTab5:         "Monitor tab",
	ActionTab6:         "Fleet tab",
	ActionPanel1:       "Jump to panel 1",
	ActionPanel2:       "Jump to panel 2",
	ActionPanel3:       "Jump to panel 3",
	ActionPanel4:       "Jump to panel 4",
	ActionPanel5:       "Jump to panel 5",
	ActionPanel6:       "Jump to panel 6",
	ActionPanel7:       "Jump to panel 7",
	ActionPanel8:       "Jump to panel 8",
	ActionPanel9:       "Jump to panel 9",
	ActionControl:      "Control panel",
	ActionDetail:       "Device detail",
}

// contextActionDescriptions overrides action descriptions for specific contexts.
var contextActionDescriptions = map[Context]map[Action]string{
	ContextAutomation: {
		ActionEnter:  "View script",
		ActionEdit:   "Edit script/schedule",
		ActionNew:    "Create new",
		ActionDelete: "Delete item",
	},
	ContextEvents: {
		ActionPause:        "Pause events",
		ActionClear:        "Clear events",
		ActionFilterToggle: "Filter by device",
	},
	ContextDevices: {
		ActionEnter:   "Select/Enter",
		ActionDetail:  "Device detail overlay",
		ActionRefresh: "Refresh device",
		ActionBrowser: "Open web UI",
		ActionControl: "Open control panel",
	},
	ContextInfo: {
		ActionEnter:  "View JSON",
		ActionExpand: "Toggle view",
	},
	ContextEnergy: {
		ActionSort:   "Sort by power",
		ActionExpand: "Expand details",
	},
	ContextConfig: {
		ActionEdit:    "Edit configuration",
		ActionRefresh: "Reload settings",
	},
	ContextManage: {
		ActionToggle:  "Select device",
		ActionExpand:  "Select all",
		ActionRefresh: "Refresh list",
	},
	ContextMonitor: {
		ActionEnter:   "Select/Enter",
		ActionDetail:  "Device detail overlay",
		ActionPause:   "Pause monitoring",
		ActionRefresh: "Refresh data",
		ActionBrowser: "Open web UI",
		ActionControl: "Open control panel",
	},
	ContextFleet: {
		ActionToggle:  "Select device",
		ActionExpand:  "Select all",
		ActionRefresh: "Refresh fleet",
	},
}

// ActionDesc returns a human-readable description for an action.
func ActionDesc(a Action) string {
	return actionDescriptions[a]
}

// ContextActionDesc returns a context-aware description for an action.
// Falls back to generic description if no context-specific one exists.
func ContextActionDesc(ctx Context, a Action) string {
	if ctxDescs, ok := contextActionDescriptions[ctx]; ok {
		if desc, ok := ctxDescs[a]; ok {
			return desc
		}
	}
	return actionDescriptions[a]
}

// ContextName returns the name of a context.
func ContextName(ctx Context) string {
	switch ctx {
	case ContextGlobal:
		return "Global"
	case ContextEvents:
		return "Events"
	case ContextDevices:
		return "Devices"
	case ContextInfo:
		return "Device Info"
	case ContextEnergy:
		return "Energy"
	case ContextJSON:
		return "JSON Viewer"
	case ContextAutomation:
		return "Automation"
	case ContextConfig:
		return "Config"
	case ContextManage:
		return "Manage"
	case ContextMonitor:
		return "Monitor"
	case ContextFleet:
		return "Fleet"
	case ContextHelp:
		return "Help"
	default:
		return "Unknown"
	}
}

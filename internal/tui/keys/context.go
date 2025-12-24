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
	ActionTab1
	ActionTab2
	ActionTab3
	ActionTab4
	ActionTab5
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
		"tab":       ActionNextPanel,
		"shift+tab": ActionPrevPanel,
		"1":         ActionTab1,
		"2":         ActionTab2,
		"3":         ActionTab3,
		"4":         ActionTab4,
		"5":         ActionTab5,
		"ctrl+c":    ActionQuit,
	}

	// Events context
	m.bindings[ContextEvents] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"space":  ActionPause,
		"c":      ActionClear,
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
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
		"enter":  ActionEnter,
		"r":      ActionRefresh,
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
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
	}

	// Energy context
	m.bindings[ContextEnergy] = map[string]Action{
		"j":      ActionDown,
		"k":      ActionUp,
		"down":   ActionDown,
		"up":     ActionUp,
		"s":      ActionSort,
		"e":      ActionExpand,
		"pgdown": ActionPageDown,
		"pgup":   ActionPageUp,
	}

	// JSON context
	m.bindings[ContextJSON] = map[string]Action{
		"j":     ActionDown,
		"k":     ActionUp,
		"down":  ActionDown,
		"up":    ActionUp,
		"h":     ActionLeft,
		"l":     ActionRight,
		"left":  ActionLeft,
		"right": ActionRight,
		"q":     ActionEscape,
		"esc":   ActionEscape,
		"y":     ActionCopy,
		"r":     ActionRefresh,
		"g":     ActionHome,
		"G":     ActionEnd,
	}

	// Automation context
	m.bindings[ContextAutomation] = map[string]Action{
		"j":     ActionDown,
		"k":     ActionUp,
		"down":  ActionDown,
		"up":    ActionUp,
		"enter": ActionEnter,
		"e":     ActionEnter,  // Edit
		"d":     ActionClear,  // Delete
		"n":     ActionExpand, // New
	}

	// Config context
	m.bindings[ContextConfig] = map[string]Action{
		"j":     ActionDown,
		"k":     ActionUp,
		"down":  ActionDown,
		"up":    ActionUp,
		"enter": ActionEnter,
		"e":     ActionEnter,
	}

	// Manage context
	m.bindings[ContextManage] = map[string]Action{
		"j":     ActionDown,
		"k":     ActionUp,
		"down":  ActionDown,
		"up":    ActionUp,
		"enter": ActionEnter,
		"r":     ActionRefresh, // Refresh device list
		"space": ActionToggle,  // Select device
		"a":     ActionExpand,  // Select all
	}

	// Fleet context
	m.bindings[ContextFleet] = map[string]Action{
		"j":     ActionDown,
		"k":     ActionUp,
		"down":  ActionDown,
		"up":    ActionUp,
		"enter": ActionEnter,
		"space": ActionToggle, // Select device
		"a":     ActionExpand, // Select all
	}

	// Help context
	m.bindings[ContextHelp] = map[string]Action{
		"?":   ActionEscape,
		"q":   ActionEscape,
		"esc": ActionEscape,
		"j":   ActionDown,
		"k":   ActionUp,
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
				Desc:   ActionDesc(action),
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
	case focus.PanelEnergy:
		return ContextEnergy
	case focus.PanelMonitor:
		return ContextDevices // Monitor uses device-like navigation
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
	ActionQuit:      "Quit",
	ActionHelp:      "Show help",
	ActionFilter:    "Filter",
	ActionCommand:   "Command mode",
	ActionNextPanel: "Next panel",
	ActionPrevPanel: "Previous panel",
	ActionUp:        "Move up",
	ActionDown:      "Move down",
	ActionLeft:      "Move left",
	ActionRight:     "Move right",
	ActionToggle:    "Toggle",
	ActionOn:        "Turn on",
	ActionOff:       "Turn off",
	ActionReboot:    "Reboot device",
	ActionEnter:     "Select/Enter",
	ActionEscape:    "Close/Cancel",
	ActionPageUp:    "Page up",
	ActionPageDown:  "Page down",
	ActionHome:      "Go to top",
	ActionEnd:       "Go to bottom",
	ActionCopy:      "Copy",
	ActionPause:     "Pause",
	ActionClear:     "Clear",
	ActionSort:      "Sort",
	ActionExpand:    "Expand/Toggle all",
	ActionRefresh:   "Refresh",
	ActionTab1:      "Dashboard tab",
	ActionTab2:      "Automation tab",
	ActionTab3:      "Config tab",
	ActionTab4:      "Manage tab",
	ActionTab5:      "Fleet tab",
}

// ActionDesc returns a human-readable description for an action.
func ActionDesc(a Action) string {
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
	case ContextFleet:
		return "Fleet"
	case ContextHelp:
		return "Help"
	default:
		return "Unknown"
	}
}

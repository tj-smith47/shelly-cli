// Package tui provides the TUI dashboard for the Shelly CLI.
package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// KeyMap defines keyboard bindings for the TUI.
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	// Actions
	Enter     key.Binding
	Escape    key.Binding
	Refresh   key.Binding
	Filter    key.Binding
	Command   key.Binding
	Help      key.Binding
	Quit      key.Binding
	ForceQuit key.Binding

	// Device actions
	Toggle  key.Binding
	TurnOn  key.Binding
	TurnOff key.Binding
	Reboot  key.Binding

	// View switching
	Tab      key.Binding
	ShiftTab key.Binding
	View1    key.Binding
	View2    key.Binding
	View3    key.Binding
	View4    key.Binding
}

// DefaultKeyMap returns vim-style keyboard bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation (vim-style)
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/up", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/down", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h/left", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/right", "right"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("ctrl+u", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("ctrl+d", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G", "go to bottom"),
		),

		// Actions
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc", "ctrl+["),
			key.WithHelp("esc", "back"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "ctrl+r"),
			key.WithHelp("r", "refresh"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Command: key.NewBinding(
			key.WithKeys(":"),
			key.WithHelp(":", "command"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),

		// Device actions
		Toggle: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle"),
		),
		TurnOn: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "turn on"),
		),
		TurnOff: key.NewBinding(
			key.WithKeys("O"),
			key.WithHelp("O", "turn off"),
		),
		Reboot: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "reboot"),
		),

		// View switching
		Tab: key.NewBinding(
			key.WithKeys("tab", "alt+]"),
			key.WithHelp("tab", "next view"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab", "alt+["),
			key.WithHelp("shift+tab", "prev view"),
		),
		View1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "devices"),
		),
		View2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "monitor"),
		),
		View3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "events"),
		),
		View4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "energy"),
		),
	}
}

// ShortHelp returns keybindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Quit, k.Help}
}

// FullHelp returns keybindings for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.PageUp, k.PageDown, k.Home, k.End},
		{k.Enter, k.Escape, k.Refresh, k.Filter},
		{k.Toggle, k.TurnOn, k.TurnOff, k.Reboot},
		{k.Tab, k.View1, k.View2, k.View3},
		{k.Command, k.Help, k.Quit},
	}
}

// KeyMapFromConfig creates a KeyMap with overrides from config.
// Any keybinding not specified in config uses the default.
func KeyMapFromConfig(cfg *config.Config) KeyMap {
	km := DefaultKeyMap()
	if cfg == nil {
		return km
	}

	kb := cfg.TUI.Keybindings

	// Apply navigation overrides
	applyNavigationBindings(&km, &kb)

	// Apply action overrides
	applyActionBindings(&km, &kb)

	// Apply device action overrides
	applyDeviceActionBindings(&km, &kb)

	// Apply view switching overrides
	applyViewBindings(&km, &kb)

	return km
}

// applyNavigationBindings applies navigation key overrides from config.
func applyNavigationBindings(km *KeyMap, kb *config.KeybindingsConfig) {
	applyBinding(&km.Up, kb.Up, "up")
	applyBinding(&km.Down, kb.Down, "down")
	applyBinding(&km.Left, kb.Left, "left")
	applyBinding(&km.Right, kb.Right, "right")
	applyBinding(&km.PageUp, kb.PageUp, "page up")
	applyBinding(&km.PageDown, kb.PageDown, "page down")
	applyBinding(&km.Home, kb.Home, "go to top")
	applyBinding(&km.End, kb.End, "go to bottom")
}

// applyActionBindings applies action key overrides from config.
func applyActionBindings(km *KeyMap, kb *config.KeybindingsConfig) {
	applyBinding(&km.Enter, kb.Enter, "select")
	applyBinding(&km.Escape, kb.Escape, "back")
	applyBinding(&km.Refresh, kb.Refresh, "refresh")
	applyBinding(&km.Filter, kb.Filter, "filter")
	applyBinding(&km.Command, kb.Command, "command")
	applyBinding(&km.Help, kb.Help, "help")
	applyBinding(&km.Quit, kb.Quit, "quit")
}

// applyDeviceActionBindings applies device action key overrides from config.
func applyDeviceActionBindings(km *KeyMap, kb *config.KeybindingsConfig) {
	applyBinding(&km.Toggle, kb.Toggle, "toggle")
	applyBinding(&km.TurnOn, kb.TurnOn, "turn on")
	applyBinding(&km.TurnOff, kb.TurnOff, "turn off")
	applyBinding(&km.Reboot, kb.Reboot, "reboot")
}

// applyViewBindings applies view switching key overrides from config.
func applyViewBindings(km *KeyMap, kb *config.KeybindingsConfig) {
	applyBinding(&km.Tab, kb.Tab, "next view")
	applyBinding(&km.ShiftTab, kb.ShiftTab, "prev view")
	applyBinding(&km.View1, kb.View1, "devices")
	applyBinding(&km.View2, kb.View2, "monitor")
	applyBinding(&km.View3, kb.View3, "events")
	applyBinding(&km.View4, kb.View4, "energy")
}

// applyBinding applies a keybinding override if keys are provided.
func applyBinding(binding *key.Binding, keys []string, helpText string) {
	if len(keys) > 0 {
		*binding = bindingWithHelp(keys, helpText)
	}
}

// bindingWithHelp creates a key.Binding with keys and help text.
func bindingWithHelp(keys []string, helpText string) key.Binding {
	helpKey := strings.Join(keys, "/")
	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(helpKey, helpText),
	)
}

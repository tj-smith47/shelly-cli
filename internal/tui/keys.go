// Package tui provides the TUI dashboard for the Shelly CLI.
package tui

import (
	"charm.land/bubbles/v2/key"
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
			key.WithKeys("esc"),
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
			key.WithKeys("tab"),
			key.WithHelp("tab", "next view"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
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
		{k.Help, k.Quit},
	}
}

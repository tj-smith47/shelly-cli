// Package focus provides a focus management system for TUI panels.
// It tracks which panel has focus and provides methods for cycling focus.
package focus

import (
	"image/color"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// PanelID identifies a focusable panel.
type PanelID int

const (
	// PanelNone indicates no panel is focused.
	PanelNone PanelID = iota
	// PanelEvents is the left events/sidebar panel.
	PanelEvents
	// PanelDeviceList is the device list panel.
	PanelDeviceList
	// PanelDeviceInfo is the device info/detail panel.
	PanelDeviceInfo
	// PanelJSON is the JSON viewer overlay.
	PanelJSON
	// PanelEnergyBars is the power consumption panel.
	PanelEnergyBars
	// PanelEnergyHistory is the energy history sparklines panel.
	PanelEnergyHistory
	// PanelMonitor is the monitor panel.
	PanelMonitor
)

// String returns the panel name for display.
func (p PanelID) String() string {
	switch p {
	case PanelNone:
		return "none"
	case PanelEvents:
		return "events"
	case PanelDeviceList:
		return "devices"
	case PanelDeviceInfo:
		return "info"
	case PanelJSON:
		return "json"
	case PanelEnergyBars:
		return "power"
	case PanelEnergyHistory:
		return "history"
	case PanelMonitor:
		return "monitor"
	default:
		return "unknown"
	}
}

// ChangedMsg is sent when focus changes between panels.
type ChangedMsg struct {
	Previous PanelID
	Current  PanelID
}

// Manager tracks focus state and provides cycling behavior.
type Manager struct {
	current    PanelID
	panels     []PanelID // Ordered list of focusable panels
	focusColor color.Color
	blurColor  color.Color
}

// New creates a new focus manager with the given panel order.
// The first panel in the list will be focused by default.
func New(panels ...PanelID) *Manager {
	colors := theme.GetSemanticColors()
	m := &Manager{
		panels:     panels,
		focusColor: colors.Highlight,
		blurColor:  colors.TableBorder,
	}
	if len(panels) > 0 {
		m.current = panels[0]
	}
	return m
}

// Current returns the currently focused panel.
func (m *Manager) Current() PanelID {
	return m.current
}

// IsFocused returns true if the given panel is currently focused.
func (m *Manager) IsFocused(panel PanelID) bool {
	return m.current == panel
}

// SetFocus sets focus to the specified panel.
// Returns a command that emits a ChangedMsg if focus changed.
func (m *Manager) SetFocus(panel PanelID) tea.Cmd {
	if m.current == panel {
		return nil
	}
	prev := m.current
	m.current = panel
	return func() tea.Msg {
		return ChangedMsg{Previous: prev, Current: panel}
	}
}

// Next moves focus to the next panel in the cycle.
// Returns a command that emits a ChangedMsg.
func (m *Manager) Next() tea.Cmd {
	if len(m.panels) == 0 {
		return nil
	}

	prev := m.current
	idx := m.findIndex(m.current)
	nextIdx := (idx + 1) % len(m.panels)
	m.current = m.panels[nextIdx]

	if m.current == prev {
		return nil
	}
	return func() tea.Msg {
		return ChangedMsg{Previous: prev, Current: m.current}
	}
}

// Prev moves focus to the previous panel in the cycle.
// Returns a command that emits a ChangedMsg.
func (m *Manager) Prev() tea.Cmd {
	if len(m.panels) == 0 {
		return nil
	}

	prev := m.current
	idx := m.findIndex(m.current)
	prevIdx := (idx - 1 + len(m.panels)) % len(m.panels)
	m.current = m.panels[prevIdx]

	if m.current == prev {
		return nil
	}
	return func() tea.Msg {
		return ChangedMsg{Previous: prev, Current: m.current}
	}
}

// findIndex returns the index of the given panel, or 0 if not found.
func (m *Manager) findIndex(panel PanelID) int {
	for i, p := range m.panels {
		if p == panel {
			return i
		}
	}
	return 0
}

// BorderColor returns the appropriate border color for the given panel.
// Returns the focus color if focused, blur color otherwise.
func (m *Manager) BorderColor(panel PanelID) color.Color {
	if m.current == panel {
		return m.focusColor
	}
	return m.blurColor
}

// SetFocusColor sets the color used for focused panel borders.
func (m *Manager) SetFocusColor(c color.Color) *Manager {
	m.focusColor = c
	return m
}

// SetBlurColor sets the color used for unfocused panel borders.
func (m *Manager) SetBlurColor(c color.Color) *Manager {
	m.blurColor = c
	return m
}

// SetPanels updates the ordered list of focusable panels.
// If the current panel is not in the new list, focus moves to the first panel.
func (m *Manager) SetPanels(panels ...PanelID) *Manager {
	m.panels = panels
	if len(panels) > 0 && m.findIndex(m.current) == 0 && m.current != panels[0] {
		// Current panel not found in new list, reset to first
		m.current = panels[0]
	}
	return m
}

// Reset sets focus to the first panel.
func (m *Manager) Reset() tea.Cmd {
	if len(m.panels) == 0 {
		return nil
	}
	return m.SetFocus(m.panels[0])
}

// PanelCount returns the number of focusable panels.
func (m *Manager) PanelCount() int {
	return len(m.panels)
}

// PanelIndex returns the 1-based index of the given panel for Shift+N navigation hints.
// Returns 0 if the panel is not in the list.
func (m *Manager) PanelIndex(panel PanelID) int {
	for i, p := range m.panels {
		if p == panel {
			return i + 1 // 1-based for display
		}
	}
	return 0
}

// Panels returns the ordered list of focusable panels.
func (m *Manager) Panels() []PanelID {
	return m.panels
}

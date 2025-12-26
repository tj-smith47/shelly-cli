// Package tabs provides a tab bar component for the TUI.
package tabs

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// TabID identifies a tab.
type TabID int

const (
	// TabDashboard is the main dashboard tab showing devices, events, and status.
	TabDashboard TabID = iota
	// TabAutomation is the automation tab for scripts, schedules, and webhooks.
	TabAutomation
	// TabConfig is the configuration tab for WiFi, system, cloud, and inputs.
	TabConfig
	// TabManage is the local device management tab for discovery, batch, firmware, and backup.
	TabManage
	// TabMonitor is the real-time monitoring tab.
	TabMonitor
	// TabFleet is the Shelly Cloud Fleet management tab.
	TabFleet
)

// String returns the tab name for display.
func (t TabID) String() string {
	switch t {
	case TabDashboard:
		return "Dashboard"
	case TabAutomation:
		return "Automation"
	case TabConfig:
		return "Config"
	case TabManage:
		return "Manage"
	case TabMonitor:
		return "Monitor"
	case TabFleet:
		return "Fleet"
	default:
		return "Unknown"
	}
}

// Icon returns the icon for the tab.
func (t TabID) Icon() string {
	switch t {
	case TabDashboard:
		return "󰋊" // dashboard/home icon
	case TabAutomation:
		return "󰃭" // automation/robot icon
	case TabConfig:
		return "󰒓" // config/settings icon
	case TabManage:
		return "󰑣" // manage/tools icon
	case TabMonitor:
		return "󰄪" // monitor/gauge icon
	case TabFleet:
		return "󰒍" // fleet/cloud icon
	default:
		return "?"
	}
}

// Tab represents a single tab.
type Tab struct {
	ID      TabID
	Label   string
	Icon    string
	Enabled bool
}

// TabChangedMsg is sent when the active tab changes.
type TabChangedMsg struct {
	Previous TabID
	Current  TabID
}

// Model represents the tab bar state.
type Model struct {
	tabs      []Tab
	active    int // Index of active tab
	width     int
	styles    Styles
	showIcons bool
}

// Styles for the tab bar.
type Styles struct {
	Container lipgloss.Style
	Tab       lipgloss.Style
	Active    lipgloss.Style
	Inactive  lipgloss.Style
	Disabled  lipgloss.Style
	Icon      lipgloss.Style
	Divider   lipgloss.Style
}

// DefaultStyles returns default styles for the tab bar.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.TableBorder).
			BorderBottom(true).
			Padding(0, 1),
		Tab: lipgloss.NewStyle(),
		Active: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Inactive: lipgloss.NewStyle().
			Foreground(colors.Text),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Icon: lipgloss.NewStyle().
			MarginRight(1),
		Divider: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// New creates a new tab bar with default tabs.
func New() Model {
	return Model{
		tabs: []Tab{
			{ID: TabDashboard, Label: TabDashboard.String(), Icon: TabDashboard.Icon(), Enabled: true},
			{ID: TabAutomation, Label: TabAutomation.String(), Icon: TabAutomation.Icon(), Enabled: true},
			{ID: TabConfig, Label: TabConfig.String(), Icon: TabConfig.Icon(), Enabled: true},
			{ID: TabManage, Label: TabManage.String(), Icon: TabManage.Icon(), Enabled: true},
			{ID: TabMonitor, Label: TabMonitor.String(), Icon: TabMonitor.Icon(), Enabled: true},
			{ID: TabFleet, Label: TabFleet.String(), Icon: TabFleet.Icon(), Enabled: true},
		},
		active:    0,
		styles:    DefaultStyles(),
		showIcons: true,
	}
}

// NewWithTabs creates a tab bar with custom tabs.
func NewWithTabs(tabs []Tab) Model {
	return Model{
		tabs:      tabs,
		active:    0,
		styles:    DefaultStyles(),
		showIcons: true,
	}
}

// Init initializes the tab bar.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the tab bar.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		return m.handleKeyPress(keyMsg)
	}
	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	keyStr := msg.String()
	switch keyStr {
	case "1":
		return m.setActive(0)
	case "2":
		return m.setActive(1)
	case "3":
		return m.setActive(2)
	case "4":
		return m.setActive(3)
	case "5":
		return m.setActive(4)
	case "6":
		return m.setActive(5)
	}
	return m, nil
}

// setActive sets the active tab and returns a command.
func (m Model) setActive(idx int) (Model, tea.Cmd) {
	if idx < 0 || idx >= len(m.tabs) || !m.tabs[idx].Enabled {
		return m, nil
	}
	if m.active == idx {
		return m, nil
	}

	prev := m.tabs[m.active].ID
	m.active = idx
	curr := m.tabs[m.active].ID

	return m, func() tea.Msg {
		return TabChangedMsg{Previous: prev, Current: curr}
	}
}

// View renders the tab bar.
func (m Model) View() string {
	if len(m.tabs) == 0 {
		return ""
	}

	tabs := make([]string, 0, len(m.tabs))
	divider := m.styles.Divider.Render(" │ ")

	for i, tab := range m.tabs {
		var style lipgloss.Style
		switch {
		case !tab.Enabled:
			style = m.styles.Disabled
		case i == m.active:
			style = m.styles.Active
		default:
			style = m.styles.Inactive
		}

		// Build tab content
		var tabContent string
		if m.showIcons && tab.Icon != "" {
			tabContent = tab.Icon + " " + tab.Label
		} else {
			tabContent = tab.Label
		}

		// Add number prefix for keyboard shortcut
		numPrefix := lipgloss.NewStyle().Foreground(theme.Purple()).Render(intToStr(i+1) + ":")
		content := numPrefix + style.Render(tabContent)

		tabs = append(tabs, content)
	}

	content := strings.Join(tabs, divider)
	// Use full width for the separator line and ensure full clearing
	if m.width > 0 {
		// Pad content to full width to clear any leftover characters from previous renders
		style := m.styles.Container.Width(m.width)
		return style.Render(content)
	}
	return m.styles.Container.Render(content)
}

// intToStr converts an int to a string without using strconv.
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToStr(-n)
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// SetWidth sets the tab bar width.
func (m Model) SetWidth(width int) Model {
	m.width = width
	return m
}

// SetActive sets the active tab by ID.
// Returns a command that emits TabChangedMsg.
func (m Model) SetActive(id TabID) (Model, tea.Cmd) {
	for i, tab := range m.tabs {
		if tab.ID == id {
			return m.setActive(i)
		}
	}
	return m, nil
}

// ActiveTab returns the currently active tab.
func (m Model) ActiveTab() Tab {
	if m.active < 0 || m.active >= len(m.tabs) {
		return Tab{}
	}
	return m.tabs[m.active]
}

// ActiveTabID returns the ID of the currently active tab.
func (m Model) ActiveTabID() TabID {
	return m.ActiveTab().ID
}

// Next moves to the next enabled tab.
func (m Model) Next() (Model, tea.Cmd) {
	for i := 1; i <= len(m.tabs); i++ {
		nextIdx := (m.active + i) % len(m.tabs)
		if m.tabs[nextIdx].Enabled {
			return m.setActive(nextIdx)
		}
	}
	return m, nil
}

// Prev moves to the previous enabled tab.
func (m Model) Prev() (Model, tea.Cmd) {
	for i := 1; i <= len(m.tabs); i++ {
		prevIdx := (m.active - i + len(m.tabs)) % len(m.tabs)
		if m.tabs[prevIdx].Enabled {
			return m.setActive(prevIdx)
		}
	}
	return m, nil
}

// SetTabEnabled enables or disables a tab.
func (m Model) SetTabEnabled(id TabID, enabled bool) Model {
	for i := range m.tabs {
		if m.tabs[i].ID == id {
			m.tabs[i].Enabled = enabled
			break
		}
	}
	return m
}

// ShowIcons toggles icon visibility.
func (m Model) ShowIcons(show bool) Model {
	m.showIcons = show
	return m
}

// TabCount returns the number of tabs.
func (m Model) TabCount() int {
	return len(m.tabs)
}

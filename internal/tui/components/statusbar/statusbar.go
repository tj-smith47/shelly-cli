// Package statusbar provides the status bar component for the TUI.
package statusbar

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// IconType determines the style of icons used in the status bar.
type IconType int

const (
	// IconTypeColor uses full-color emojis (⚡️ 💡 🪟).
	IconTypeColor IconType = iota
	// IconTypeOutline uses simple outline symbols (⏻ ☀ ↕).
	IconTypeOutline
)

// getIconType returns the icon type based on ICON_TYPE environment variable.
// Returns IconTypeColor for "color" (default), IconTypeOutline for "outline".
func getIconType() IconType {
	switch strings.ToLower(os.Getenv("ICON_TYPE")) {
	case "outline":
		return IconTypeOutline
	default:
		return IconTypeColor
	}
}

// componentIcons returns the icons for switches, lights, and covers based on icon type.
func componentIcons() (switchIcon, lightIcon, coverIcon string) {
	if getIconType() == IconTypeOutline {
		return "⏻", "☀", "↕"
	}
	// Color emojis (with variation selector for ⚡️)
	return "⚡️", "💡", "🪟"
}

// tickMsg is sent on each tick for time updates.
type tickMsg time.Time

// ComponentCounts holds counts for various component types.
type ComponentCounts struct {
	SwitchesOn   int
	SwitchesOff  int
	LightsOn     int
	LightsOff    int
	CoversOpen   int
	CoversClosed int
	CoversMoving int
}

// ContextType identifies which view's context to display.
type ContextType int

const (
	// ContextNone shows no context.
	ContextNone ContextType = iota
	// ContextDevice shows device info (Dashboard, Automation, Config views).
	ContextDevice
	// ContextMonitor shows WebSocket/refresh info (Monitor view).
	ContextMonitor
	// ContextManage shows firmware update info (Manage view).
	ContextManage
	// ContextFleet shows group info (Fleet view).
	ContextFleet
)

// Model holds the status bar state.
type Model struct {
	width       int
	message     string
	messageType MessageType
	lastUpdate  time.Time
	styles      Styles
	debugActive bool
	counts      ComponentCounts

	// Context display
	contextType ContextType
	panelName   string // Active panel/view name (used by all context types)

	// Device context (Dashboard, Automation, Config)
	deviceName string
	deviceIP   string

	// Monitor context
	wsConnected     int
	wsTotal         int
	refreshInterval time.Duration

	// Manage context
	firmwareUpdates int

	// Fleet context
	groupName string
}

// MessageType indicates the type of status message.
type MessageType int

const (
	// MessageNormal is a normal informational message.
	MessageNormal MessageType = iota
	// MessageSuccess indicates a successful operation.
	MessageSuccess
	// MessageError indicates an error.
	MessageError
	// MessageWarning indicates a warning.
	MessageWarning
)

// Styles for the status bar.
type Styles struct {
	Bar        lipgloss.Style
	Left       lipgloss.Style
	Right      lipgloss.Style
	Normal     lipgloss.Style
	Success    lipgloss.Style
	Error      lipgloss.Style
	Warning    lipgloss.Style
	Version    lipgloss.Style
	Time       lipgloss.Style
	Debug      lipgloss.Style
	CountOn    lipgloss.Style // For "on" state counts
	CountOff   lipgloss.Style // For "off" state counts
	CountLabel lipgloss.Style // For component type labels
}

// DefaultStyles returns default styles for the status bar.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Bar: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Padding(0, 1),
		Left: lipgloss.NewStyle().
			Align(lipgloss.Left),
		Right: lipgloss.NewStyle().
			Align(lipgloss.Right),
		Normal: lipgloss.NewStyle().
			Foreground(colors.Text),
		Success: lipgloss.NewStyle().
			Foreground(colors.Success),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Version: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Time: lipgloss.NewStyle().
			Foreground(colors.Info),
		Debug: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
		CountOn: lipgloss.NewStyle().
			Foreground(colors.Online),
		CountOff: lipgloss.NewStyle().
			Foreground(colors.Muted),
		CountLabel: lipgloss.NewStyle().
			Foreground(colors.Text),
	}
}

// New creates a new status bar model.
func New() Model {
	return Model{
		message:    "Ready",
		lastUpdate: time.Now(),
		styles:     DefaultStyles(),
	}
}

// Init returns the initial command for the status bar.
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// tickCmd returns a command that ticks every second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages for the status bar.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.lastUpdate = time.Time(msg)
		return m, tickCmd()

	case StatusMsg:
		m.message = msg.Message
		m.messageType = msg.Type
		return m, nil
	}

	return m, nil
}

// StatusMsg is a message to update the status bar.
type StatusMsg struct {
	Message string
	Type    MessageType
}

// SetMessage sets the status bar message.
func SetMessage(msg string, t MessageType) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{Message: msg, Type: t}
	}
}

// SetWidth sets the status bar width.
func (m Model) SetWidth(width int) Model {
	m.width = width
	return m
}

// SetDeviceContext sets device context for Dashboard, Automation, Config views.
// Full: "Device: Name (IP) │ Panel", Compact: "Name (IP)", Minimal: none.
func (m Model) SetDeviceContext(name, ip, panel string) Model {
	m.contextType = ContextDevice
	m.deviceName = name
	m.deviceIP = ip
	m.panelName = panel
	return m
}

// SetMonitorContext sets monitor context for the Monitor view.
// Full: "WS: X/Y Connected (⟳ Xs) │ Panel", Compact: "X/Y Connected (⟳ Xs)", Minimal: none.
func (m Model) SetMonitorContext(connected, total int, refreshInterval time.Duration, panel string) Model {
	m.contextType = ContextMonitor
	m.wsConnected = connected
	m.wsTotal = total
	m.refreshInterval = refreshInterval
	m.panelName = panel
	return m
}

// SetManageContext sets manage context for the Manage view.
// Full: "Firmware Updates: X │ Panel", Compact: "X Updates Available", Minimal: none.
func (m Model) SetManageContext(firmwareUpdates int, panel string) Model {
	m.contextType = ContextManage
	m.firmwareUpdates = firmwareUpdates
	m.panelName = panel
	return m
}

// SetFleetContext sets fleet context for the Fleet view.
// Full: "Group: Name │ Panel", Compact: "Name", Minimal: none.
// Pass empty groupName when no group is selected or fleet not configured.
func (m Model) SetFleetContext(groupName, panel string) Model {
	m.contextType = ContextFleet
	m.groupName = groupName
	m.panelName = panel
	return m
}

// ClearContext clears all context (shows nothing in center).
func (m Model) ClearContext() Model {
	m.contextType = ContextNone
	m.panelName = ""
	m.deviceName = ""
	m.deviceIP = ""
	m.wsConnected = 0
	m.wsTotal = 0
	m.refreshInterval = 0
	m.firmwareUpdates = 0
	m.groupName = ""
	return m
}

// SetDebugActive sets whether a debug session is active.
func (m Model) SetDebugActive(active bool) Model {
	m.debugActive = active
	return m
}

// IsDebugActive returns whether a debug session is active.
func (m Model) IsDebugActive() bool {
	return m.debugActive
}

// SetComponentCounts sets the component state counts for display.
func (m Model) SetComponentCounts(counts ComponentCounts) Model {
	m.counts = counts
	return m
}

// Tier represents the status bar display tier.
type Tier int

const (
	// TierMinimal is for narrow terminals (<80 columns).
	TierMinimal Tier = iota
	// TierCompact is for medium terminals (80-120 columns).
	TierCompact
	// TierFull is for wide terminals (>120 columns).
	TierFull
)

// GetTier returns the appropriate tier based on width.
func (m Model) GetTier() Tier {
	if m.width >= 120 {
		return TierFull
	}
	if m.width >= 80 {
		return TierCompact
	}
	return TierMinimal
}

// View renders the status bar.
func (m Model) View() string {
	tier := m.GetTier()

	// Left side: status message first
	var msgStyle lipgloss.Style
	switch m.messageType {
	case MessageSuccess:
		msgStyle = m.styles.Success
	case MessageError:
		msgStyle = m.styles.Error
	case MessageWarning:
		msgStyle = m.styles.Warning
	default:
		msgStyle = m.styles.Normal
	}
	left := msgStyle.Render(m.message)

	// Debug indicator (recording dot + text)
	if m.debugActive {
		debugText := "REC"
		if tier == TierFull {
			debugText = "Debug active"
		}
		left += "  " + m.styles.Debug.Render("● "+debugText)
	}

	// Component counts section (icon-prefixed)
	componentSection := m.renderComponentCounts(tier)
	if componentSection != "" {
		left += "  │ " + componentSection
	}

	// Right side: version and time (tier-dependent)
	timeStr := m.lastUpdate.Format("15:04:05")
	var right string
	switch tier {
	case TierFull:
		right = fmt.Sprintf("%s  %s",
			m.styles.Version.Render("v"+version.Version),
			m.styles.Time.Render(timeStr),
		)
	case TierCompact:
		right = fmt.Sprintf("%s  %s",
			m.styles.Version.Render(version.Version),
			m.styles.Time.Render(timeStr[:5]), // HH:MM only
		)
	default: // TierMinimal
		right = m.styles.Time.Render(timeStr[:5])
	}

	// Center section: Active device and panel context
	center := m.renderContext(tier)

	// Calculate widths
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	centerWidth := lipgloss.Width(center)

	// Available space for center (minus padding)
	availableWidth := m.width - 4
	sideSpace := availableWidth - leftWidth - rightWidth - centerWidth

	if sideSpace < 2 {
		// Not enough space for center - just left and right with minimal gap
		spacing := availableWidth - leftWidth - rightWidth
		if spacing < 1 {
			spacing = 1
		}
		content := left + strings.Repeat(" ", spacing) + right
		return m.styles.Bar.Width(m.width).Render(content)
	}

	// Center the context between left and right
	leftPad := sideSpace / 2
	rightPad := sideSpace - leftPad

	content := left + strings.Repeat(" ", leftPad) + center + strings.Repeat(" ", rightPad) + right

	return m.styles.Bar.Width(m.width).Render(content)
}

// renderContext renders the centered context based on context type and tier.
func (m Model) renderContext(tier Tier) string {
	if tier == TierMinimal {
		return "" // All context types show nothing in minimal tier
	}

	switch m.contextType {
	case ContextDevice:
		return m.renderDeviceContext(tier)
	case ContextMonitor:
		return m.renderMonitorContext(tier)
	case ContextManage:
		return m.renderManageContext(tier)
	case ContextFleet:
		return m.renderFleetContext(tier)
	default:
		return ""
	}
}

// renderDeviceContext renders device context.
// Full: "Device: Name (IP) │ Panel", Compact: "Name (IP)".
func (m Model) renderDeviceContext(tier Tier) string {
	if m.deviceName == "" {
		return ""
	}

	deviceInfo := m.deviceName
	if m.deviceIP != "" {
		deviceInfo += " (" + m.deviceIP + ")"
	}

	switch tier {
	case TierFull:
		result := m.styles.Normal.Render("Device: " + deviceInfo)
		if m.panelName != "" {
			result += " │ " + m.styles.Normal.Render(m.panelName)
		}
		return result
	case TierCompact:
		return m.styles.Normal.Render(deviceInfo)
	default:
		return ""
	}
}

// renderMonitorContext renders monitor context.
// Full: "WS: X/Y Connected (⟳ Xs) │ Panel", Compact: "X/Y Connected (⟳ Xs)".
func (m Model) renderMonitorContext(tier Tier) string {
	if m.wsTotal == 0 && m.refreshInterval == 0 {
		return ""
	}

	// Format refresh interval
	refreshStr := ""
	if m.refreshInterval > 0 {
		secs := int(m.refreshInterval.Seconds())
		refreshStr = fmt.Sprintf(" (⟳ %ds)", secs)
	}

	connStr := fmt.Sprintf("%d/%d Connected%s", m.wsConnected, m.wsTotal, refreshStr)

	switch tier {
	case TierFull:
		result := m.styles.Normal.Render("WS: " + connStr)
		if m.panelName != "" {
			result += " │ " + m.styles.Normal.Render(m.panelName)
		}
		return result
	case TierCompact:
		return m.styles.Normal.Render(connStr)
	default:
		return ""
	}
}

// renderManageContext renders manage context.
// Full: "Firmware Updates: X │ Panel", Compact: "X Updates Available".
func (m Model) renderManageContext(tier Tier) string {
	if m.firmwareUpdates == 0 {
		// No updates available - show nothing special
		if tier == TierFull && m.panelName != "" {
			return m.styles.Normal.Render(m.panelName)
		}
		return ""
	}

	switch tier {
	case TierFull:
		result := m.styles.Normal.Render(fmt.Sprintf("Firmware Updates: %d", m.firmwareUpdates))
		if m.panelName != "" {
			result += " │ " + m.styles.Normal.Render(m.panelName)
		}
		return result
	case TierCompact:
		return m.styles.Normal.Render(fmt.Sprintf("%d Updates Available", m.firmwareUpdates))
	default:
		return ""
	}
}

// renderFleetContext renders fleet context.
// Full: "Group: Name │ Panel", Compact: "Name".
func (m Model) renderFleetContext(tier Tier) string {
	if m.groupName == "" {
		// No group selected - show nothing
		return ""
	}

	switch tier {
	case TierFull:
		result := m.styles.Normal.Render("Group: " + m.groupName)
		if m.panelName != "" {
			result += " │ " + m.styles.Normal.Render(m.panelName)
		}
		return result
	case TierCompact:
		return m.styles.Normal.Render(m.groupName)
	default:
		return ""
	}
}

// renderComponentCounts renders the component state counts based on tier.
// Icons are determined by the ICON_TYPE environment variable.
func (m Model) renderComponentCounts(tier Tier) string {
	c := m.counts
	hasSwitches := c.SwitchesOn > 0 || c.SwitchesOff > 0
	hasLights := c.LightsOn > 0 || c.LightsOff > 0
	hasCovers := c.CoversOpen > 0 || c.CoversClosed > 0 || c.CoversMoving > 0

	if !hasSwitches && !hasLights && !hasCovers {
		return ""
	}

	switchIcon, lightIcon, coverIcon := componentIcons()
	var parts []string

	if hasSwitches {
		parts = append(parts, m.formatComponentCount(tier, switchIcon, "Switches", "Sw", c.SwitchesOn, c.SwitchesOff))
	}
	if hasLights {
		parts = append(parts, m.formatComponentCount(tier, lightIcon, "Lights", "Lt", c.LightsOn, c.LightsOff))
	}
	if hasCovers {
		parts = append(parts, m.formatCoverCount(tier, coverIcon, c.CoversOpen, c.CoversClosed, c.CoversMoving))
	}

	sep := " │ "
	if tier == TierMinimal {
		sep = " "
	}

	return joinStrings(parts, sep)
}

// formatComponentCount formats a component count for display (switches/lights).
func (m Model) formatComponentCount(tier Tier, icon, fullLabel, shortLabel string, on, off int) string {
	total := on + off
	onStr := m.styles.CountOn.Render(fmt.Sprintf("%d", on))
	offStr := m.styles.CountOff.Render(fmt.Sprintf("%d", off))

	switch tier {
	case TierFull:
		return icon + " " + m.styles.CountLabel.Render(fullLabel+": ") + onStr + "/" + offStr
	case TierCompact:
		return icon + " " + m.styles.CountLabel.Render(shortLabel+": ") + onStr + "/" + fmt.Sprintf("%d", total)
	default: // TierMinimal
		return icon + fmt.Sprintf("%d", on)
	}
}

// formatCoverCount formats cover counts for display.
func (m Model) formatCoverCount(tier Tier, icon string, open, closed, moving int) string {
	openStr := m.styles.CountOn.Render(fmt.Sprintf("%d", open))
	closedStr := m.styles.CountOff.Render(fmt.Sprintf("%d", closed))

	switch tier {
	case TierFull:
		if moving > 0 {
			return icon + " " + m.styles.CountLabel.Render("Covers: ") + openStr + "↑ " + closedStr + "↓ " +
				m.styles.Warning.Render(fmt.Sprintf("%d", moving)) + "~"
		}
		return icon + " " + m.styles.CountLabel.Render("Covers: ") + openStr + "↑ " + closedStr + "↓"
	case TierCompact:
		total := open + closed + moving
		return icon + " " + m.styles.CountLabel.Render("Cv: ") + openStr + "/" + fmt.Sprintf("%d", total)
	default: // TierMinimal
		return icon + fmt.Sprintf("%d", open)
	}
}

// joinStrings joins strings with a separator.
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}

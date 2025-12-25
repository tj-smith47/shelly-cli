// Package devicelist provides the device list component for the TUI.
// This component displays a split-pane view with a device list and detail panel.
// It uses the shared cache for device data rather than doing its own fetching.
package devicelist

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
)

// DeviceSelectedMsg is sent when a device is selected.
type DeviceSelectedMsg struct {
	Name    string
	Address string
}

// Model holds the device list state.
type Model struct {
	cache        *cache.Cache // Shared device cache
	filter       string       // Current filter string
	cursor       int          // Currently selected index in filtered list
	scrollOffset int          // Scroll offset for list
	focused      bool         // Whether this component has focus
	width        int
	height       int
	styles       Styles
}

// Styles for the device list component.
type Styles struct {
	Container     lipgloss.Style
	ListPanel     lipgloss.Style
	DetailPanel   lipgloss.Style
	ListHeader    lipgloss.Style
	DetailHeader  lipgloss.Style
	Row           lipgloss.Style
	SelectedRow   lipgloss.Style
	Online        lipgloss.Style
	Offline       lipgloss.Style
	Checking      lipgloss.Style
	DeviceName    lipgloss.Style
	DeviceAddress lipgloss.Style
	Power         lipgloss.Style
	Label         lipgloss.Style
	Value         lipgloss.Style
	Separator     lipgloss.Style
	StatusOK      lipgloss.Style
	StatusError   lipgloss.Style
	Table         lipgloss.Style // For compatibility
}

// DefaultStyles returns default styles for the device list.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			Padding(0),
		ListPanel: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Padding(0, 1),
		DetailPanel: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Padding(1, 2),
		ListHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.TableBorder).
			BorderBottom(true).
			MarginBottom(1),
		DetailHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight).
			MarginBottom(1),
		Row: lipgloss.NewStyle(),
		SelectedRow: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Online: lipgloss.NewStyle().
			Foreground(colors.Online),
		Offline: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Checking: lipgloss.NewStyle().
			Foreground(colors.Updating),
		DeviceName: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Text),
		DeviceAddress: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Power: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Width(12),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Separator: lipgloss.NewStyle().
			Foreground(colors.Muted),
		StatusOK: lipgloss.NewStyle().
			Foreground(colors.Success),
		StatusError: lipgloss.NewStyle().
			Foreground(colors.Error),
		Table: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder),
	}
}

// New creates a new device list model using the shared cache.
func New(c *cache.Cache) Model {
	return Model{
		cache:  c,
		styles: DefaultStyles(),
	}
}

// Init initializes the device list component.
// The cache handles device loading and refresh, so no commands needed here.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the device list.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		oldCursor := m.cursor
		m = m.handleKeyPress(keyMsg)
		// Emit selection change if cursor moved
		if m.cursor != oldCursor {
			return m, m.emitSelection()
		}
	}
	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) Model {
	devices := m.getFilteredDevices()
	switch msg.String() {
	case "j", "down":
		m = m.cursorDown(devices)
	case "k", "up":
		m = m.cursorUp()
	case "g":
		m.cursor = 0
		m.scrollOffset = 0
	case "G":
		m = m.cursorToEnd(devices)
	case "pgdown", "ctrl+d":
		m = m.pageDown(devices)
	case "pgup", "ctrl+u":
		m = m.pageUp()
	}
	return m
}

func (m Model) cursorDown(devices []*cache.DeviceData) Model {
	if m.cursor < len(devices)-1 {
		m.cursor++
		if m.cursor >= m.scrollOffset+m.visibleRows() {
			m.scrollOffset = m.cursor - m.visibleRows() + 1
		}
	}
	return m
}

func (m Model) cursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.scrollOffset {
			m.scrollOffset = m.cursor
		}
	}
	return m
}

func (m Model) cursorToEnd(devices []*cache.DeviceData) Model {
	if len(devices) > 0 {
		m.cursor = len(devices) - 1
		maxOffset := len(devices) - m.visibleRows()
		if maxOffset < 0 {
			maxOffset = 0
		}
		m.scrollOffset = maxOffset
	}
	return m
}

func (m Model) pageDown(devices []*cache.DeviceData) Model {
	if len(devices) == 0 {
		return m
	}
	m.cursor += m.visibleRows()
	if m.cursor >= len(devices) {
		m.cursor = len(devices) - 1
	}
	if m.cursor >= m.scrollOffset+m.visibleRows() {
		m.scrollOffset = m.cursor - m.visibleRows() + 1
	}
	return m
}

func (m Model) pageUp() Model {
	m.cursor -= m.visibleRows()
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	return m
}

// emitSelection returns a command that emits a DeviceSelectedMsg for the current selection.
func (m Model) emitSelection() tea.Cmd {
	devices := m.getFilteredDevices()
	if m.cursor < 0 || m.cursor >= len(devices) {
		return nil
	}
	d := devices[m.cursor]
	return func() tea.Msg {
		return DeviceSelectedMsg{
			Name:    d.Device.Name,
			Address: d.Device.Address,
		}
	}
}

// getFilteredDevices returns devices from cache filtered by the current filter.
func (m Model) getFilteredDevices() []*cache.DeviceData {
	if m.cache == nil {
		return nil
	}

	all := m.cache.GetAllDevices()
	if m.filter == "" {
		return all
	}

	filterLower := strings.ToLower(m.filter)
	filtered := make([]*cache.DeviceData, 0, len(all))
	for _, d := range all {
		if strings.Contains(strings.ToLower(d.Device.Name), filterLower) ||
			strings.Contains(strings.ToLower(d.Device.Address), filterLower) ||
			strings.Contains(strings.ToLower(d.Device.Type), filterLower) ||
			strings.Contains(strings.ToLower(d.Device.Model), filterLower) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

// SetFilter sets the filter string.
func (m Model) SetFilter(filter string) Model {
	m.filter = filter
	// Reset cursor if it's now out of bounds
	devices := m.getFilteredDevices()
	if m.cursor >= len(devices) {
		m.cursor = len(devices) - 1
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
	return m
}

// SetFocused sets whether this component has focus.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// Focused returns whether this component has focus.
func (m Model) Focused() bool {
	return m.focused
}

// visibleRows calculates how many rows can be displayed in the list panel.
func (m Model) visibleRows() int {
	// Account for borders, header, padding
	available := m.height - 4
	if available < 1 {
		return 1
	}
	return available
}

// listPanelWidth returns the width of the list panel (40% of total).
func (m Model) listPanelWidth() int {
	return (m.width * 40) / 100
}

// detailPanelWidth returns the width of the detail panel (60% of total).
func (m Model) detailPanelWidth() int {
	return m.width - m.listPanelWidth() - 1 // -1 for gap
}

// SetSize sets the component size.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// View renders the device list with split pane.
func (m Model) View() string {
	if m.cache == nil {
		return m.styles.Table.
			Width(m.width-4).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No cache available")
	}

	devices := m.getFilteredDevices()

	if m.cache.IsLoading() && len(devices) == 0 {
		return m.styles.Table.
			Width(m.width-4).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Loading devices...")
	}

	if m.cache.DeviceCount() == 0 {
		return m.styles.Table.
			Width(m.width-4).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No devices registered.\nUse 'shelly device add' to add devices.")
	}

	if len(devices) == 0 && m.filter != "" {
		return m.styles.Table.
			Width(m.width-4).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("No devices match filter %q.\nPress / to clear or modify filter.", m.filter))
	}

	// Split pane layout
	listWidth := m.listPanelWidth()
	detailWidth := m.detailPanelWidth()

	listPanel := m.renderListPanel(devices, listWidth)
	detailPanel := m.renderDetailPanel(devices, detailWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, listPanel, " ", detailPanel)
}

// renderListPanel renders the left panel with the device list.
func (m Model) renderListPanel(devices []*cache.DeviceData, width int) string {
	colors := theme.GetSemanticColors()
	borderColor := colors.TableBorder
	if m.focused {
		borderColor = colors.Highlight
	}

	panelStyle := m.styles.ListPanel.BorderForeground(borderColor)

	// Header
	header := m.styles.ListHeader.Width(width - 4).Render("Devices")

	// Rows
	visible := m.visibleRows()
	startIdx := m.scrollOffset
	endIdx := startIdx + visible
	if endIdx > len(devices) {
		endIdx = len(devices)
	}

	var rows strings.Builder
	for i := startIdx; i < endIdx; i++ {
		d := devices[i]
		isSelected := i == m.cursor
		row := m.renderListRow(d, isSelected, width-4)
		rows.WriteString(row + "\n")
	}

	// Scroll indicator
	scrollInfo := ""
	if len(devices) > visible {
		scrollInfo = m.styles.Separator.Render(
			fmt.Sprintf(" [%d/%d]", m.cursor+1, len(devices)),
		)
	}

	content := header + "\n" + rows.String() + scrollInfo

	return panelStyle.
		Width(width).
		Height(m.height).
		Render(content)
}

// renderListRow renders a single row in the device list.
func (m Model) renderListRow(d *cache.DeviceData, isSelected bool, width int) string {
	// Status icon
	var icon string
	switch {
	case !d.Fetched:
		icon = m.styles.Checking.Render("◐")
	case d.Online:
		icon = m.styles.Online.Render("●")
	default:
		icon = m.styles.Offline.Render("○")
	}

	// Selection indicator
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Name (truncate if needed)
	maxNameWidth := width - 6 // icon, selector, padding
	name := d.Device.Name
	if len(name) > maxNameWidth && maxNameWidth > 3 {
		name = name[:maxNameWidth-1] + "…"
	}

	row := fmt.Sprintf("%s%s %s", selector, icon, name)

	if isSelected {
		return m.styles.SelectedRow.Width(width).Render(row)
	}
	return m.styles.Row.Width(width).Render(row)
}

// renderDetailPanel renders the right panel with device details.
func (m Model) renderDetailPanel(devices []*cache.DeviceData, width int) string {
	colors := theme.GetSemanticColors()
	// Detail panel always uses standard border - only list panel highlights on focus
	borderColor := colors.TableBorder

	panelStyle := m.styles.DetailPanel.BorderForeground(borderColor)

	// Get selected device
	if len(devices) == 0 || m.cursor < 0 || m.cursor >= len(devices) {
		return panelStyle.
			Width(width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No device selected")
	}

	d := devices[m.cursor]

	var content strings.Builder

	// Header with device name
	content.WriteString(m.styles.DetailHeader.Render(d.Device.Name) + "\n\n")

	// Status
	content.WriteString(m.renderDeviceStatus(d) + "\n\n")

	// Basic info from config
	m.renderBasicInfo(&content, d)

	// Device info from API
	m.renderDeviceInfo(&content, d)

	// Power data
	m.renderPowerMetrics(&content, d)

	return panelStyle.
		Width(width).
		Height(m.height).
		Render(content.String())
}

// renderDeviceStatus renders the device status line.
func (m Model) renderDeviceStatus(d *cache.DeviceData) string {
	switch {
	case !d.Fetched:
		return m.styles.Checking.Render("◐ Checking...")
	case d.Online:
		return m.styles.Online.Render("● Online")
	default:
		status := m.styles.Offline.Render("○ Offline")
		if d.Error != nil {
			errMsg := d.Error.Error()
			if len(errMsg) > 40 {
				errMsg = errMsg[:40] + "..."
			}
			status += " - " + m.styles.StatusError.Render(errMsg)
		}
		return status
	}
}

// renderBasicInfo renders basic device info from config.
func (m Model) renderBasicInfo(content *strings.Builder, d *cache.DeviceData) {
	content.WriteString(m.renderDetailRow("Address", d.Device.Address))
	content.WriteString(m.renderDetailRow("Type", d.Device.Type))
	content.WriteString(m.renderDetailRow("Model", d.Device.Model))
	content.WriteString(m.renderDetailRow("Generation", fmt.Sprintf("Gen%d", d.Device.Generation)))
}

// renderDeviceInfo renders device info from API.
func (m Model) renderDeviceInfo(content *strings.Builder, d *cache.DeviceData) {
	if d.Info == nil {
		return
	}

	content.WriteString("\n")
	content.WriteString(m.styles.Separator.Render("─────────────────────────────") + "\n\n")
	content.WriteString(m.renderDetailRow("Firmware", d.Info.Firmware))
	content.WriteString(m.renderDetailRow("MAC", d.Info.MAC))
	if d.Info.App != "" {
		content.WriteString(m.renderDetailRow("App", d.Info.App))
	}
	authStr := "No"
	if d.Info.AuthEn {
		authStr = "Yes"
	}
	content.WriteString(m.renderDetailRow("Auth", authStr))
}

// renderPowerMetrics renders power metrics section.
func (m Model) renderPowerMetrics(content *strings.Builder, d *cache.DeviceData) {
	if !d.Online || (d.Power == 0 && d.Voltage == 0 && d.Current == 0) {
		return
	}

	content.WriteString("\n")
	content.WriteString(m.styles.Separator.Render("─────────────────────────────") + "\n\n")
	content.WriteString(m.styles.DetailHeader.Render("Power Metrics") + "\n\n")

	if d.Power != 0 {
		powerStr := fmt.Sprintf("%.1f W", d.Power)
		if d.Power >= 1000 || d.Power <= -1000 {
			powerStr = fmt.Sprintf("%.2f kW", d.Power/1000)
		}
		content.WriteString(m.renderDetailRow("Power", m.styles.Power.Render(powerStr)))
	}
	if d.Voltage != 0 {
		content.WriteString(m.renderDetailRow("Voltage", fmt.Sprintf("%.1f V", d.Voltage)))
	}
	if d.Current != 0 {
		content.WriteString(m.renderDetailRow("Current", fmt.Sprintf("%.2f A", d.Current)))
	}
}

// renderDetailRow renders a label-value row for the detail panel.
func (m Model) renderDetailRow(label, value string) string {
	return m.styles.Label.Render(label+":") + " " + m.styles.Value.Render(value) + "\n"
}

// SelectedDevice returns the currently selected device, if any.
func (m Model) SelectedDevice() *cache.DeviceData {
	devices := m.getFilteredDevices()
	if m.cursor < 0 || m.cursor >= len(devices) {
		return nil
	}
	return devices[m.cursor]
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.cursor
}

// SetCursor sets the cursor position.
func (m Model) SetCursor(cursor int) Model {
	devices := m.getFilteredDevices()
	if cursor >= 0 && cursor < len(devices) {
		m.cursor = cursor
		// Adjust scroll offset if needed
		if m.cursor < m.scrollOffset {
			m.scrollOffset = m.cursor
		} else if m.cursor >= m.scrollOffset+m.visibleRows() {
			m.scrollOffset = m.cursor - m.visibleRows() + 1
		}
	}
	return m
}

// DeviceCount returns the number of filtered devices.
func (m Model) DeviceCount() int {
	return len(m.getFilteredDevices())
}

// TotalDeviceCount returns the total number of devices (before filtering).
func (m Model) TotalDeviceCount() int {
	if m.cache == nil {
		return 0
	}
	return m.cache.DeviceCount()
}

// Filter returns the current filter string.
func (m Model) Filter() string {
	return m.filter
}

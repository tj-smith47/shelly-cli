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
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

// DeviceSelectedMsg is sent when a device is selected.
type DeviceSelectedMsg struct {
	Name    string
	Address string
}

// OpenBrowserMsg is sent when the user wants to open a device's web UI in browser.
type OpenBrowserMsg struct {
	Address string
}

// Model holds the device list state.
type Model struct {
	cache      *cache.Cache    // Shared device cache
	filter     string          // Current filter string
	scroller   *panel.Scroller // Shared scroll/cursor management
	focused    bool            // Whether this component has focus
	listOnly   bool            // When true, only render list panel (detail rendered separately)
	gPressed   bool            // Tracks if 'g' was just pressed (for vim-style gx, gg commands)
	panelIndex int
	width      int
	height     int
	styles     Styles
	loader     loading.Model // Loading spinner for initial load
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
		cache:    c,
		scroller: panel.NewScroller(0, 10),
		styles:   DefaultStyles(),
		loader: loading.New(
			loading.WithMessage("Loading devices..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
	}
}

// Init initializes the device list component.
// The cache handles device loading and refresh, but we start the spinner if loading.
func (m Model) Init() tea.Cmd {
	if m.cache != nil && m.cache.IsLoading() {
		return m.loader.Init()
	}
	return nil
}

// Update handles messages for the device list.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Update loader for spinner animation when loading
	if m.cache != nil && m.cache.IsLoading() && len(m.getFilteredDevices()) == 0 {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		return m, cmd
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		// Sync item count from cache before handling key
		devices := m.getFilteredDevices()
		m.scroller.SetItemCount(len(devices))

		oldCursor := m.scroller.Cursor()
		var cmd tea.Cmd
		m, cmd = m.handleKeyPress(keyMsg)
		// Emit selection change if cursor moved (unless a command is already being returned)
		if m.scroller.Cursor() != oldCursor && cmd == nil {
			return m, m.emitSelection()
		}
		return m, cmd
	}
	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	key := msg.String()

	// Handle g-prefix commands (gx for browser, gg for top)
	if m.gPressed {
		m.gPressed = false
		switch key {
		case "x":
			// gx: open device web UI in browser
			if d := m.SelectedDevice(); d != nil {
				return m, func() tea.Msg {
					return OpenBrowserMsg{Address: d.Device.Address}
				}
			}
			return m, nil
		case "g":
			// gg: go to top
			m.scroller.CursorToStart()
			return m, nil
		}
		// Other keys after g: just process normally
	}

	switch key {
	case "j", "down":
		m.scroller.CursorDown()
	case "k", "up":
		m.scroller.CursorUp()
	case "g":
		// Start g-prefix mode (for gg, gx commands)
		m.gPressed = true
		return m, nil
	case "G":
		m.scroller.CursorToEnd()
	case "pgdown", "ctrl+d":
		m.scroller.PageDown()
	case "pgup", "ctrl+u":
		m.scroller.PageUp()
	}
	return m, nil
}

// emitSelection returns a command that emits a DeviceSelectedMsg for the current selection.
func (m Model) emitSelection() tea.Cmd {
	devices := m.getFilteredDevices()
	cursor := m.scroller.Cursor()
	if cursor < 0 || cursor >= len(devices) {
		return nil
	}
	d := devices[cursor]
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
	// Update item count for new filter
	devices := m.getFilteredDevices()
	m.scroller.SetItemCount(len(devices))
	return m
}

// SetFocused sets whether this component has focus.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return m
}

// SetListOnly sets whether to render only the list panel (detail rendered separately).
func (m Model) SetListOnly(listOnly bool) Model {
	m.listOnly = listOnly
	// Update visible rows as overhead differs between modes
	m.scroller.SetVisibleRows(m.visibleRows())
	return m
}

// ListOnly returns whether list-only mode is enabled.
func (m Model) ListOnly() bool {
	return m.listOnly
}

// Focused returns whether this component has focus.
func (m Model) Focused() bool {
	return m.focused
}

// visibleRows calculates how many rows can be displayed in the list panel.
func (m Model) visibleRows() int {
	var overhead int
	if m.listOnly {
		// In listOnly mode, renderer handles borders (2) + vertical padding (2) = 4 lines
		overhead = 4
	} else {
		// In split-pane mode, account for panel borders (2), header (~3), padding
		overhead = 5
	}

	available := m.height - overhead
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
	m.scroller.SetVisibleRows(m.visibleRows())
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
			Width(m.width - 4).
			Height(m.height).
			Render(m.loader.SetSize(m.width-4, m.height).View())
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

	// List-only mode: render just the list panel at full width
	if m.listOnly {
		return m.renderListPanel(devices, m.width)
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

	// Update scroller with current device count
	m.scroller.SetItemCount(len(devices))

	// Get visible range from scroller
	startIdx, endIdx := m.scroller.VisibleRange()

	// In list-only mode, renderer adds borders (2) + padding (2) = 4 chars
	// So content rows should be width - 4 to fit properly
	rowWidth := width
	if m.listOnly {
		rowWidth = width - 4
	}

	var rows strings.Builder
	for i := startIdx; i < endIdx; i++ {
		d := devices[i]
		isSelected := m.scroller.IsCursorAt(i)
		row := m.renderListRow(d, isSelected, rowWidth)
		rows.WriteString(row + "\n")
	}

	content := strings.TrimSuffix(rows.String(), "\n")

	// In listOnly mode, just return content (border/padding handled by renderer)
	if m.listOnly {
		return content
	}

	// For split-pane mode, use panel styling with header
	borderColor := colors.TableBorder
	if m.focused {
		borderColor = colors.Highlight
	}
	panelStyle := m.styles.ListPanel.BorderForeground(borderColor)
	header := m.styles.ListHeader.Width(width - 4).Render("Devices")

	return panelStyle.
		Width(width).
		Height(m.height).
		Render(header + "\n" + content)
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
	cursor := m.scroller.Cursor()
	if len(devices) == 0 || cursor < 0 || cursor >= len(devices) {
		return panelStyle.
			Width(width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No device selected")
	}

	d := devices[cursor]

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
	cursor := m.scroller.Cursor()
	if cursor < 0 || cursor >= len(devices) {
		return nil
	}
	return devices[cursor]
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.scroller.Cursor()
}

// SetCursor sets the cursor position.
func (m Model) SetCursor(cursor int) Model {
	devices := m.getFilteredDevices()
	m.scroller.SetItemCount(len(devices))
	if cursor < 0 || cursor >= len(devices) {
		return m
	}
	m.scroller.SetCursor(cursor)
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

// OptimalWidth calculates the minimum width needed to display all device names
// without truncation, plus the status icon and selector (6 chars overhead).
func (m Model) OptimalWidth() int {
	if m.cache == nil {
		return 20 // Minimum default
	}

	devices := m.getFilteredDevices()
	if len(devices) == 0 {
		return 20
	}

	maxLen := 0
	for _, d := range devices {
		if len(d.Device.Name) > maxLen {
			maxLen = len(d.Device.Name)
		}
	}

	// Add overhead: selector (2) + icon (1) + space (1) + padding (4) + border (2)
	optimal := maxLen + 10

	// Apply min/max constraints
	if optimal < 20 {
		optimal = 20
	}
	if optimal > 50 {
		optimal = 50 // Don't let device list get too wide
	}

	return optimal
}

// MaxDeviceNameLen returns the length of the longest device name in the list.
func (m Model) MaxDeviceNameLen() int {
	if m.cache == nil {
		return 0
	}

	devices := m.getFilteredDevices()
	maxLen := 0
	for _, d := range devices {
		if len(d.Device.Name) > maxLen {
			maxLen = len(d.Device.Name)
		}
	}
	return maxLen
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll g/G:top/bottom /:filter"
}

// Package devicelist provides the device list component for the TUI.
package devicelist

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Deps holds the dependencies for the device list component.
type Deps struct {
	Ctx             context.Context
	Svc             *shelly.Service
	IOS             *iostreams.IOStreams
	RefreshInterval time.Duration
}

// validate ensures all required dependencies are set.
func (d Deps) validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	if d.IOS == nil {
		return fmt.Errorf("iostreams is required")
	}
	return nil
}

// DeviceStatus represents a device with its live status.
type DeviceStatus struct {
	Device   model.Device
	Online   bool
	Checking bool // True while status is being fetched
	Power    float64
	Voltage  float64
	Current  float64
	Info     *shelly.DeviceInfo // Full device info for detail panel
	Error    error
}

// DevicesLoadedMsg signals that devices were loaded.
type DevicesLoadedMsg struct {
	Devices []DeviceStatus
	Err     error
}

// DeviceStatusUpdateMsg signals a single device status update (for streaming).
type DeviceStatusUpdateMsg struct {
	Name   string
	Status DeviceStatus
}

// RefreshTickMsg triggers periodic refresh.
type RefreshTickMsg struct{}

// Model holds the device list state.
type Model struct {
	ctx             context.Context
	svc             *shelly.Service
	ios             *iostreams.IOStreams
	devices         map[string]*DeviceStatus // Keyed by name for fast updates
	deviceOrder     []string                 // Sorted order of device names
	filtered        []string                 // Filtered device names for display
	filter          string                   // Current filter string
	cursor          int                      // Currently selected index in filtered list
	scrollOffset    int                      // Scroll offset for list
	loading         bool                     // Initial load in progress
	err             error
	width           int
	height          int
	styles          Styles
	refreshInterval time.Duration
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

// New creates a new device list model.
func New(deps Deps) Model {
	if err := deps.validate(); err != nil {
		panic(fmt.Sprintf("devicelist: invalid deps: %v", err))
	}

	refreshInterval := deps.RefreshInterval
	if refreshInterval == 0 {
		refreshInterval = 5 * time.Second
	}

	return Model{
		ctx:             deps.Ctx,
		svc:             deps.Svc,
		ios:             deps.IOS,
		devices:         make(map[string]*DeviceStatus),
		deviceOrder:     []string{},
		filtered:        []string{},
		loading:         true,
		styles:          DefaultStyles(),
		refreshInterval: refreshInterval,
	}
}

// Init returns the initial command to fetch devices.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadDevicesFromConfig(),
		m.scheduleRefresh(),
	)
}

// scheduleRefresh schedules the next refresh tick.
func (m Model) scheduleRefresh() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}

// loadDevicesFromConfig immediately loads devices from config and starts fetching status.
func (m Model) loadDevicesFromConfig() tea.Cmd {
	return func() tea.Msg {
		deviceMap := config.ListDevices()
		if len(deviceMap) == 0 {
			return DevicesLoadedMsg{Devices: nil}
		}

		// Create initial device list with "checking" status
		devices := make([]DeviceStatus, 0, len(deviceMap))
		for _, d := range deviceMap {
			devices = append(devices, DeviceStatus{
				Device:   d,
				Checking: true,
			})
		}

		// Sort by name
		sort.Slice(devices, func(i, j int) bool {
			return devices[i].Device.Name < devices[j].Device.Name
		})

		return DevicesLoadedMsg{Devices: devices}
	}
}

// fetchAllDeviceStatuses fetches status for all devices concurrently, sending individual updates.
func (m Model) fetchAllDeviceStatuses() tea.Cmd {
	return func() tea.Msg {
		// Collect all device names to fetch
		names := make([]string, 0, len(m.devices))
		for name := range m.devices {
			names = append(names, name)
		}

		if len(names) == 0 {
			return nil
		}

		// Return a batch of commands, one per device
		return tea.BatchMsg(m.createStatusFetchCommands(names))
	}
}

// createStatusFetchCommands creates individual fetch commands for each device.
func (m Model) createStatusFetchCommands(names []string) []tea.Cmd {
	cmds := make([]tea.Cmd, len(names))
	for i, name := range names {
		deviceName := name
		device := m.devices[name]
		if device == nil {
			continue
		}
		cmds[i] = m.fetchSingleDeviceStatus(deviceName, device.Device)
	}
	return cmds
}

// fetchSingleDeviceStatus fetches status for a single device.
func (m Model) fetchSingleDeviceStatus(name string, device model.Device) tea.Cmd {
	return func() tea.Msg {
		status := DeviceStatus{
			Device:   device,
			Checking: false,
		}

		// Per-device timeout
		ctx, cancel := context.WithTimeout(m.ctx, 3*time.Second)
		defer cancel()

		// Get device info (includes firmware, model, etc.)
		info, err := m.svc.DeviceInfo(ctx, device.Address)
		if err != nil {
			status.Error = err
			return DeviceStatusUpdateMsg{Name: name, Status: status}
		}
		status.Info = info
		status.Online = true

		// Get monitoring snapshot for power data
		snapshot, err := m.svc.GetMonitoringSnapshot(ctx, device.Address)
		if err == nil {
			aggregatePowerData(&status, snapshot)
		}

		return DeviceStatusUpdateMsg{Name: name, Status: status}
	}
}

// aggregatePowerData aggregates power data from a monitoring snapshot.
func aggregatePowerData(status *DeviceStatus, snapshot *shelly.MonitoringSnapshot) {
	for _, pm := range snapshot.PM {
		status.Power += pm.APower
		if status.Voltage == 0 && pm.Voltage > 0 {
			status.Voltage = pm.Voltage
		}
		if status.Current == 0 && pm.Current > 0 {
			status.Current = pm.Current
		}
	}
	for _, em := range snapshot.EM {
		status.Power += em.TotalActivePower
		if status.Voltage == 0 && em.AVoltage > 0 {
			status.Voltage = em.AVoltage
		}
	}
	for _, em1 := range snapshot.EM1 {
		status.Power += em1.ActPower
		if status.Voltage == 0 && em1.Voltage > 0 {
			status.Voltage = em1.Voltage
		}
	}
}

// Refresh returns a command to refresh the device list.
func (m Model) Refresh() tea.Cmd {
	// Mark all devices as checking
	for _, d := range m.devices {
		d.Checking = true
	}
	return m.fetchAllDeviceStatuses()
}

// Update handles messages for the device list.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DevicesLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}

		// Populate device map and order
		m.devices = make(map[string]*DeviceStatus, len(msg.Devices))
		m.deviceOrder = make([]string, 0, len(msg.Devices))
		for i := range msg.Devices {
			d := msg.Devices[i]
			m.devices[d.Device.Name] = &d
			m.deviceOrder = append(m.deviceOrder, d.Device.Name)
		}
		m = m.applyFilter()

		// Start fetching statuses
		return m, m.fetchAllDeviceStatuses()

	case DeviceStatusUpdateMsg:
		// Update single device status
		if d, ok := m.devices[msg.Name]; ok {
			*d = msg.Status
		}
		return m, nil

	case RefreshTickMsg:
		return m, tea.Batch(
			m.Refresh(),
			m.scheduleRefresh(),
		)

	case tea.KeyPressMsg:
		m = m.handleKeyPress(msg)
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) Model {
	switch msg.String() {
	case "j", "down":
		m = m.cursorDown()
	case "k", "up":
		m = m.cursorUp()
	case "g":
		m.cursor = 0
		m.scrollOffset = 0
	case "G":
		m = m.cursorToEnd()
	case "pgdown", "ctrl+d":
		m = m.pageDown()
	case "pgup", "ctrl+u":
		m = m.pageUp()
	}
	return m
}

func (m Model) cursorDown() Model {
	if m.cursor < len(m.filtered)-1 {
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

func (m Model) cursorToEnd() Model {
	if len(m.filtered) > 0 {
		m.cursor = len(m.filtered) - 1
		maxOffset := len(m.filtered) - m.visibleRows()
		if maxOffset < 0 {
			maxOffset = 0
		}
		m.scrollOffset = maxOffset
	}
	return m
}

func (m Model) pageDown() Model {
	if len(m.filtered) == 0 {
		return m
	}
	m.cursor += m.visibleRows()
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor >= m.scrollOffset+m.visibleRows() {
		m.scrollOffset = m.cursor - m.visibleRows() + 1
	}
	return m
}

func (m Model) pageUp() Model {
	if len(m.filtered) == 0 {
		return m
	}
	m.cursor -= m.visibleRows()
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	return m
}

// SetFilter sets the filter string and re-applies it to the device list.
func (m Model) SetFilter(filter string) Model {
	m.filter = filter
	return m.applyFilter()
}

// applyFilter filters the device list based on the current filter string.
func (m Model) applyFilter() Model {
	if m.filter == "" {
		m.filtered = m.deviceOrder
	} else {
		filterLower := strings.ToLower(m.filter)
		m.filtered = make([]string, 0, len(m.deviceOrder))
		for _, name := range m.deviceOrder {
			d := m.devices[name]
			if d == nil {
				continue
			}
			if strings.Contains(strings.ToLower(d.Device.Name), filterLower) ||
				strings.Contains(strings.ToLower(d.Device.Address), filterLower) ||
				strings.Contains(strings.ToLower(d.Device.Type), filterLower) ||
				strings.Contains(strings.ToLower(d.Device.Model), filterLower) {
				m.filtered = append(m.filtered, name)
			}
		}
	}

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
		if m.cursor < 0 {
			m.cursor = 0
		}
	}

	return m
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
	if m.loading && len(m.devices) == 0 {
		return m.styles.Table.
			Width(m.width-4).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Loading devices...")
	}

	if m.err != nil {
		return m.styles.Table.
			Width(m.width - 4).
			Render(m.styles.StatusError.Render("Error: " + m.err.Error()))
	}

	if len(m.devices) == 0 {
		return m.styles.Table.
			Width(m.width-4).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No devices registered.\nUse 'shelly device add' to add devices.")
	}

	if len(m.filtered) == 0 && m.filter != "" {
		return m.styles.Table.
			Width(m.width-4).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("No devices match filter %q.\nPress / to clear or modify filter.", m.filter))
	}

	// Split pane layout
	listWidth := m.listPanelWidth()
	detailWidth := m.detailPanelWidth()

	listPanel := m.renderListPanel(listWidth)
	detailPanel := m.renderDetailPanel(detailWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, listPanel, " ", detailPanel)
}

// renderListPanel renders the left panel with the device list.
func (m Model) renderListPanel(width int) string {
	// Header
	header := m.styles.ListHeader.Width(width - 4).Render("Devices")

	// Rows
	visible := m.visibleRows()
	startIdx := m.scrollOffset
	endIdx := startIdx + visible
	if endIdx > len(m.filtered) {
		endIdx = len(m.filtered)
	}

	var rows strings.Builder
	for i := startIdx; i < endIdx; i++ {
		name := m.filtered[i]
		d := m.devices[name]
		if d == nil {
			continue
		}

		isSelected := i == m.cursor
		row := m.renderListRow(d, isSelected, width-4)
		rows.WriteString(row + "\n")
	}

	// Scroll indicator
	scrollInfo := ""
	if len(m.filtered) > visible {
		scrollInfo = m.styles.Separator.Render(
			fmt.Sprintf(" [%d/%d]", m.cursor+1, len(m.filtered)),
		)
	}

	content := header + "\n" + rows.String() + scrollInfo

	return m.styles.ListPanel.
		Width(width).
		Height(m.height).
		Render(content)
}

// renderListRow renders a single row in the device list.
func (m Model) renderListRow(d *DeviceStatus, isSelected bool, width int) string {
	// Status icon
	var icon string
	switch {
	case d.Checking:
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
	if len(name) > maxNameWidth {
		name = name[:maxNameWidth-1] + "…"
	}

	row := fmt.Sprintf("%s%s %s", selector, icon, name)

	if isSelected {
		return m.styles.SelectedRow.Width(width).Render(row)
	}
	return m.styles.Row.Width(width).Render(row)
}

// renderDetailPanel renders the right panel with device details.
func (m Model) renderDetailPanel(width int) string {
	// Get selected device
	if len(m.filtered) == 0 || m.cursor < 0 || m.cursor >= len(m.filtered) {
		return m.styles.DetailPanel.
			Width(width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No device selected")
	}

	name := m.filtered[m.cursor]
	d := m.devices[name]
	if d == nil {
		return m.styles.DetailPanel.
			Width(width).
			Height(m.height).
			Render("Device not found")
	}

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

	return m.styles.DetailPanel.
		Width(width).
		Height(m.height).
		Render(content.String())
}

// renderDeviceStatus renders the device status line.
func (m Model) renderDeviceStatus(d *DeviceStatus) string {
	switch {
	case d.Checking:
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
func (m Model) renderBasicInfo(content *strings.Builder, d *DeviceStatus) {
	content.WriteString(m.renderDetailRow("Address", d.Device.Address))
	content.WriteString(m.renderDetailRow("Type", d.Device.Type))
	content.WriteString(m.renderDetailRow("Model", d.Device.Model))
	content.WriteString(m.renderDetailRow("Generation", fmt.Sprintf("Gen%d", d.Device.Generation)))
}

// renderDeviceInfo renders device info from API.
func (m Model) renderDeviceInfo(content *strings.Builder, d *DeviceStatus) {
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
func (m Model) renderPowerMetrics(content *strings.Builder, d *DeviceStatus) {
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
func (m Model) SelectedDevice() *DeviceStatus {
	if len(m.filtered) == 0 || m.cursor < 0 || m.cursor >= len(m.filtered) {
		return nil
	}

	name := m.filtered[m.cursor]
	return m.devices[name]
}

// DeviceCount returns the number of filtered devices.
func (m Model) DeviceCount() int {
	return len(m.filtered)
}

// TotalDeviceCount returns the total number of devices (before filtering).
func (m Model) TotalDeviceCount() int {
	return len(m.devices)
}

// Filter returns the current filter string.
func (m Model) Filter() string {
	return m.filter
}

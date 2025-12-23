// Package fleet provides TUI components for Shelly Cloud Fleet management.
package fleet

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// DevicesDeps holds the dependencies for the Devices component.
type DevicesDeps struct {
	Ctx context.Context
}

// Validate ensures all required dependencies are set.
func (d DevicesDeps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	return nil
}

// DevicesLoadedMsg signals that devices were loaded from the fleet.
type DevicesLoadedMsg struct {
	Devices []integrator.AccountDevice
	Err     error
}

// DevicesModel displays cloud devices from the fleet manager.
type DevicesModel struct {
	ctx       context.Context
	fleet     *integrator.FleetManager
	devices   []integrator.AccountDevice
	cursor    int
	loading   bool
	err       error
	width     int
	height    int
	focused   bool
	styles    DevicesStyles
	lastFetch time.Time
}

// DevicesStyles holds styles for the Devices component.
type DevicesStyles struct {
	Online   lipgloss.Style
	Offline  lipgloss.Style
	Name     lipgloss.Style
	Type     lipgloss.Style
	Host     lipgloss.Style
	Cursor   lipgloss.Style
	Muted    lipgloss.Style
	Error    lipgloss.Style
	Title    lipgloss.Style
	Selected lipgloss.Style
}

// DefaultDevicesStyles returns the default styles for the Devices component.
func DefaultDevicesStyles() DevicesStyles {
	colors := theme.GetSemanticColors()
	return DevicesStyles{
		Online: lipgloss.NewStyle().
			Foreground(colors.Online),
		Offline: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Name: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		Type: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Host: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Cursor: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Selected: lipgloss.NewStyle().
			Background(colors.Highlight).
			Foreground(colors.Background),
	}
}

// NewDevices creates a new Devices model.
func NewDevices(deps DevicesDeps) DevicesModel {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("fleet/devices: invalid deps: %v", err))
	}

	return DevicesModel{
		ctx:    deps.Ctx,
		styles: DefaultDevicesStyles(),
	}
}

// Init returns the initial command.
func (m DevicesModel) Init() tea.Cmd {
	return nil
}

// SetFleetManager sets the fleet manager and triggers a device load.
func (m DevicesModel) SetFleetManager(fm *integrator.FleetManager) (DevicesModel, tea.Cmd) {
	m.fleet = fm
	if fm == nil {
		m.devices = nil
		return m, nil
	}
	m.loading = true
	return m, m.loadDevices()
}

func (m DevicesModel) loadDevices() tea.Cmd {
	return func() tea.Msg {
		if m.fleet == nil {
			return DevicesLoadedMsg{Err: fmt.Errorf("not connected to fleet")}
		}
		devices := m.fleet.AccountManager().ListDevices()
		return DevicesLoadedMsg{Devices: devices}
	}
}

// SetSize sets the component dimensions.
func (m DevicesModel) SetSize(width, height int) DevicesModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focus state.
func (m DevicesModel) SetFocused(focused bool) DevicesModel {
	m.focused = focused
	return m
}

// Update handles messages.
func (m DevicesModel) Update(msg tea.Msg) (DevicesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case DevicesLoadedMsg:
		m.loading = false
		m.lastFetch = time.Now()
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.devices = msg.Devices
		m.err = nil
		// Reset cursor if needed
		if m.cursor >= len(m.devices) {
			m.cursor = max(0, len(m.devices)-1)
		}
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m DevicesModel) handleKey(msg tea.KeyPressMsg) (DevicesModel, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(m.devices)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g":
		m.cursor = 0
	case "G":
		m.cursor = max(0, len(m.devices)-1)
	case "r":
		if !m.loading {
			m.loading = true
			return m, m.loadDevices()
		}
	}

	return m, nil
}

// View renders the Devices component.
func (m DevicesModel) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Cloud Devices").
		SetFocused(m.focused)

	// Handle early return cases
	if msg := m.getStatusMessage(); msg != "" {
		r.SetContent(msg)
		return r.Render()
	}

	r.SetContent(m.renderDeviceList())
	return r.Render()
}

func (m DevicesModel) getStatusMessage() string {
	switch {
	case m.fleet == nil:
		return m.styles.Muted.Render("Not connected to Shelly Cloud")
	case m.loading:
		return m.styles.Muted.Render("Loading devices...")
	case m.err != nil:
		return m.styles.Error.Render("Error: " + m.err.Error())
	case len(m.devices) == 0:
		return m.styles.Muted.Render("No devices in fleet")
	default:
		return ""
	}
}

func (m DevicesModel) renderDeviceList() string {
	var content strings.Builder

	// Stats line
	online := m.OnlineCount()
	statsLine := fmt.Sprintf("%d devices, %d online", len(m.devices), online)
	content.WriteString(m.styles.Muted.Render(statsLine))
	content.WriteString("\n\n")

	// Device list with scroll
	startIdx, endIdx := m.calculateScrollRange()
	for i := startIdx; i < endIdx; i++ {
		content.WriteString(m.renderDeviceLine(i))
		content.WriteString("\n")
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("j/k: navigate | r: refresh"))

	return content.String()
}

func (m DevicesModel) calculateScrollRange() (start, end int) {
	visibleHeight := m.height - 6
	if visibleHeight < 1 {
		visibleHeight = 5
	}

	start = 0
	if m.cursor >= visibleHeight {
		start = m.cursor - visibleHeight + 1
	}
	end = start + visibleHeight
	if end > len(m.devices) {
		end = len(m.devices)
	}
	return start, end
}

func (m DevicesModel) renderDeviceLine(idx int) string {
	device := m.devices[idx]

	// Online indicator
	statusIcon := m.styles.Offline.Render("○")
	if device.Online {
		statusIcon = m.styles.Online.Render("●")
	}

	// Cursor
	cursor := "  "
	if idx == m.cursor && m.focused {
		cursor = m.styles.Cursor.Render("> ")
	}

	// Device name (truncate if needed)
	name := device.Name
	if name == "" {
		name = device.DeviceID
	}
	if len(name) > 20 {
		name = name[:17] + "..."
	}

	// Device type (truncate if needed)
	deviceType := device.DeviceType
	if len(deviceType) > 12 {
		deviceType = deviceType[:12]
	}

	line := fmt.Sprintf("%s%s %s %s",
		cursor,
		statusIcon,
		m.styles.Name.Render(name),
		m.styles.Type.Render("("+deviceType+")"),
	)

	if idx == m.cursor && m.focused {
		line = m.styles.Selected.Render(line)
	}

	return line
}

// SelectedDevice returns the currently selected device.
func (m DevicesModel) SelectedDevice() *integrator.AccountDevice {
	if m.cursor >= 0 && m.cursor < len(m.devices) {
		return &m.devices[m.cursor]
	}
	return nil
}

// Devices returns all devices.
func (m DevicesModel) Devices() []integrator.AccountDevice {
	return m.devices
}

// DeviceCount returns the number of devices.
func (m DevicesModel) DeviceCount() int {
	return len(m.devices)
}

// OnlineCount returns the number of online devices.
func (m DevicesModel) OnlineCount() int {
	count := 0
	for _, d := range m.devices {
		if d.Online {
			count++
		}
	}
	return count
}

// Loading returns whether devices are being loaded.
func (m DevicesModel) Loading() bool {
	return m.loading
}

// Error returns any error that occurred.
func (m DevicesModel) Error() error {
	return m.err
}

// Refresh triggers a device reload.
func (m DevicesModel) Refresh() (DevicesModel, tea.Cmd) {
	if m.fleet == nil {
		return m, nil
	}
	m.loading = true
	return m, m.loadDevices()
}

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

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
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
	ctx        context.Context
	fleet      *integrator.FleetManager
	devices    []integrator.AccountDevice
	scroller   *panel.Scroller
	loading    bool
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int
	styles     DevicesStyles
	lastFetch  time.Time
	loader     loading.Model
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
			Foreground(colors.Primary),
	}
}

// NewDevices creates a new Devices model.
func NewDevices(deps DevicesDeps) DevicesModel {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("fleet/devices component init", err)
		panic(fmt.Sprintf("fleet/devices: invalid deps: %v", err))
	}

	return DevicesModel{
		ctx:      deps.Ctx,
		scroller: panel.NewScroller(0, 10),
		styles:   DefaultDevicesStyles(),
		loader: loading.New(
			loading.WithMessage("Loading devices..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
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
	return m, tea.Batch(m.loader.Tick(), m.loadDevices())
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
	visibleRows := height - 6 // Account for borders, title, stats line
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
	// Update loader size for proper centering
	m.loader = m.loader.SetSize(width-4, height-4)
	return m
}

// SetFocused sets the focus state.
func (m DevicesModel) SetFocused(focused bool) DevicesModel {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m DevicesModel) SetPanelIndex(index int) DevicesModel {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m DevicesModel) Update(msg tea.Msg) (DevicesModel, tea.Cmd) {
	// Forward tick messages to loader when loading
	if m.loading {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		// Continue processing DevicesLoadedMsg even during loading
		if _, ok := msg.(DevicesLoadedMsg); !ok {
			if cmd != nil {
				return m, cmd
			}
		}
	}

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
		m.scroller.SetItemCount(len(m.devices))
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
	case "j", keyconst.KeyDown:
		m.scroller.CursorDown()
	case "k", keyconst.KeyUp:
		m.scroller.CursorUp()
	case "g":
		m.scroller.CursorToStart()
	case "G":
		m.scroller.CursorToEnd()
	case "ctrl+d", keyconst.KeyPgDown:
		m.scroller.PageDown()
	case "ctrl+u", keyconst.KeyPgUp:
		m.scroller.PageUp()
	case "r":
		if !m.loading {
			m.loading = true
			return m, tea.Batch(m.loader.Tick(), m.loadDevices())
		}
	}

	return m, nil
}

// View renders the Devices component.
func (m DevicesModel) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Cloud Devices").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused {
		r.SetFooter("j/k:nav g/G:top/btm r:refresh")
	}

	// Handle early return cases
	if msg := m.getStatusMessage(); msg != "" {
		r.SetContent(msg)
		return r.Render()
	}

	r.SetContent(m.renderDeviceList())
	return r.Render()
}

func (m DevicesModel) getStatusMessage() string {
	// Calculate content area for centering (accounting for panel borders)
	contentWidth := m.width - 4
	contentHeight := m.height - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	switch {
	case m.fleet == nil:
		msg := m.styles.Muted.Render("Not connected to Shelly Cloud")
		return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, msg)
	case m.loading:
		return m.loader.View()
	case m.err != nil:
		msg, hint := tuierrors.FormatError(m.err)
		errMsg := m.styles.Error.Render(msg) + "\n" +
			m.styles.Muted.Render("  "+hint) + "\n" +
			m.styles.Muted.Render("  Press 'r' to retry")
		return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, errMsg)
	case len(m.devices) == 0:
		msg := m.styles.Muted.Render("No devices in fleet")
		return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, msg)
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
	start, end := m.scroller.VisibleRange()
	for i := start; i < end; i++ {
		content.WriteString(m.renderDeviceLine(i))
		content.WriteString("\n")
	}

	// Scroll indicator
	content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))

	return content.String()
}

func (m DevicesModel) renderDeviceLine(idx int) string {
	device := m.devices[idx]
	isSelected := m.scroller.IsCursorAt(idx)

	// Online indicator
	statusIcon := m.styles.Offline.Render("○")
	if device.Online {
		statusIcon = m.styles.Online.Render("●")
	}

	// Cursor
	cursor := "  "
	if isSelected && m.focused {
		cursor = m.styles.Cursor.Render("> ")
	}

	// Calculate available width for name and type
	// Fixed: cursor(2) + status(2) + spaces(2) + parens(2) = 8
	available := output.ContentWidth(m.width, 4+8) // panel border + fixed elements
	nameWidth, typeWidth := output.SplitWidth(available, 60, 10, 8)

	// Device name (truncate if needed)
	name := device.Name
	if name == "" {
		name = device.DeviceID
	}
	name = output.Truncate(name, nameWidth)

	// Device type (truncate if needed)
	deviceType := output.Truncate(device.DeviceType, typeWidth)

	line := fmt.Sprintf("%s%s %s %s",
		cursor,
		statusIcon,
		m.styles.Name.Render(name),
		m.styles.Type.Render("("+deviceType+")"),
	)

	if isSelected && m.focused {
		line = m.styles.Selected.Render(line)
	}

	return line
}

// SelectedDevice returns the currently selected device.
func (m DevicesModel) SelectedDevice() *integrator.AccountDevice {
	cursor := m.scroller.Cursor()
	if cursor >= 0 && cursor < len(m.devices) {
		return &m.devices[cursor]
	}
	return nil
}

// Cursor returns the current cursor position.
func (m DevicesModel) Cursor() int {
	return m.scroller.Cursor()
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
	return m, tea.Batch(m.loader.Tick(), m.loadDevices())
}

// FooterText returns keybinding hints for the footer.
func (m DevicesModel) FooterText() string {
	return "j/k:scroll g/G:top/bottom enter:details"
}

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
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
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
	panel.Sizable // Embeds Width, Height, Loader, Scroller
	ctx           context.Context
	fleet         *integrator.FleetManager
	devices       []integrator.AccountDevice
	loading       bool
	err           error
	focused       bool
	panelIndex    int
	styles        DevicesStyles
	lastFetch     time.Time
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

	m := DevicesModel{
		Sizable: panel.NewSizable(6, panel.NewScroller(0, 10)), // 6 accounts for borders, title, stats line
		ctx:     deps.Ctx,
		styles:  DefaultDevicesStyles(),
	}
	m.Loader = m.Loader.SetMessage("Loading devices...")
	return m
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
	return m, tea.Batch(m.Loader.Tick(), m.loadDevices())
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
	m.ApplySize(width, height)
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
		result := generics.UpdateLoader(m.Loader, msg, func(msg tea.Msg) bool {
			_, ok := msg.(DevicesLoadedMsg)
			return ok
		})
		m.Loader = result.Loader
		if result.Consumed {
			return m, result.Cmd
		}
	}

	return m.handleMessage(msg)
}

func (m DevicesModel) handleMessage(msg tea.Msg) (DevicesModel, tea.Cmd) {
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
		m.Scroller.SetItemCount(len(m.devices))
		return m, nil

	// Action messages from context system
	case messages.NavigationMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleNavigation(msg)
	case messages.RefreshRequestMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleRefresh()
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m DevicesModel) handleNavigation(msg messages.NavigationMsg) (DevicesModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		m.Scroller.CursorUp()
	case messages.NavDown:
		m.Scroller.CursorDown()
	case messages.NavPageUp:
		m.Scroller.PageUp()
	case messages.NavPageDown:
		m.Scroller.PageDown()
	case messages.NavHome:
		m.Scroller.CursorToStart()
	case messages.NavEnd:
		m.Scroller.CursorToEnd()
	case messages.NavLeft, messages.NavRight:
		// Not applicable for this component
	}
	return m, nil
}

func (m DevicesModel) handleRefresh() (DevicesModel, tea.Cmd) {
	if m.loading {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.loadDevices())
}

func (m DevicesModel) handleKey(msg tea.KeyPressMsg) (DevicesModel, tea.Cmd) {
	// Component-specific keys not covered by action messages
	// (none currently - all keys migrated to action messages)
	_ = msg
	return m, nil
}

// View renders the Devices component.
func (m DevicesModel) View() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Cloud Devices").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused {
		r.SetFooter(theme.StyledKeybindings(keys.FormatHints([]keys.Hint{{Key: "j/k", Desc: "nav"}, {Key: "g/G", Desc: "top/btm"}, {Key: "r", Desc: "refresh"}}, keys.FooterHintWidth(m.Width))))
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
	contentWidth := m.Width - 4
	contentHeight := m.Height - 4
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
		return m.Loader.View()
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

	// Device list with scroll using generic helper
	content.WriteString(generics.RenderScrollableList(generics.ListRenderConfig[integrator.AccountDevice]{
		Items:    m.devices,
		Scroller: m.Scroller,
		RenderItem: func(device integrator.AccountDevice, _ int, isCursor bool) string {
			return m.renderDeviceLine(device, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoAlways,
	}))

	return content.String()
}

func (m DevicesModel) renderDeviceLine(device integrator.AccountDevice, isSelected bool) string {
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
	available := output.ContentWidth(m.Width, 4+8) // panel border + fixed elements
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
	cursor := m.Scroller.Cursor()
	if cursor >= 0 && cursor < len(m.devices) {
		return &m.devices[cursor]
	}
	return nil
}

// Cursor returns the current cursor position.
func (m DevicesModel) Cursor() int {
	return m.Scroller.Cursor()
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
	return m, tea.Batch(m.Loader.Tick(), m.loadDevices())
}

// FooterText returns keybinding hints for the footer.
func (m DevicesModel) FooterText() string {
	return "j/k:scroll g/G:top/bottom enter:details"
}

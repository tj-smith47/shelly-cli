// Package discovery provides TUI components for device discovery.
package discovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// Deps holds the dependencies for the Discovery component.
type Deps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d Deps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	return nil
}

// ScanCompleteMsg signals that a discovery scan completed.
type ScanCompleteMsg struct {
	Devices []shelly.DiscoveredDevice
	Err     error
}

// DeviceAddedMsg signals that a device was added to the registry.
type DeviceAddedMsg struct {
	Address string
	Err     error
}

// Model displays device discovery.
type Model struct {
	ctx        context.Context
	svc        *shelly.Service
	devices    []shelly.DiscoveredDevice
	scroller   *panel.Scroller
	scanning   bool
	method     shelly.DiscoveryMethod
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int
	styles     Styles
	loader     loading.Model
}

// Styles holds styles for the Discovery component.
type Styles struct {
	Added      lipgloss.Style
	NotAdded   lipgloss.Style
	Model      lipgloss.Style
	Address    lipgloss.Style
	Generation lipgloss.Style
	Selected   lipgloss.Style
	Label      lipgloss.Style
	Error      lipgloss.Style
	Muted      lipgloss.Style
	ScanButton lipgloss.Style
}

// DefaultStyles returns the default styles for the Discovery component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Added: lipgloss.NewStyle().
			Foreground(colors.Online),
		NotAdded: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Model: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Address: lipgloss.NewStyle().
			Foreground(colors.Text),
		Generation: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		ScanButton: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// New creates a new Discovery model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("discovery component init", err)
		panic(fmt.Sprintf("discovery: invalid deps: %v", err))
	}

	return Model{
		ctx:      deps.Ctx,
		svc:      deps.Svc,
		scroller: panel.NewScroller(0, 10),
		scanning: false,
		method:   shelly.DiscoveryMDNS,
		styles:   DefaultStyles(),
		loader: loading.New(
			loading.WithMessage("Scanning for devices..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	// Reserve space for header, method selector, and footer
	visibleRows := height - 8
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
	// Update loader size for proper centering
	m.loader = m.loader.SetSize(width-4, height-4)
	return m
}

// SetFocused sets the focus state.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return m
}

// StartScan triggers a new discovery scan.
func (m Model) StartScan() (Model, tea.Cmd) {
	if m.scanning {
		return m, nil
	}
	m.scanning = true
	m.err = nil
	m.devices = nil
	m.scroller.SetItemCount(0)
	m.scroller.CursorToStart()
	return m, tea.Batch(m.loader.Tick(), m.scanDevices())
}

func (m Model) scanDevices() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		opts := shelly.DiscoveryOptions{
			Method:     m.method,
			Timeout:    15 * time.Second,
			AutoDetect: true,
		}

		devices, err := m.svc.DiscoverDevices(ctx, opts)
		return ScanCompleteMsg{Devices: devices, Err: err}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Forward tick messages to loader when scanning
	if m.scanning {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		// Continue processing ScanCompleteMsg even during scanning
		switch msg.(type) {
		case ScanCompleteMsg:
			// Pass through to main switch below
		default:
			if cmd != nil {
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case ScanCompleteMsg:
		m.scanning = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.devices = msg.Devices
		m.scroller.SetItemCount(len(m.devices))
		return m, nil

	case DeviceAddedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Mark device as added
		for i := range m.devices {
			if m.devices[i].Address == msg.Address {
				m.devices[i].Added = true
				break
			}
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

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.scroller.CursorDown()
	case "k", "up":
		m.scroller.CursorUp()
	case "g":
		m.scroller.CursorToStart()
	case "G":
		m.scroller.CursorToEnd()
	case "ctrl+d", "pgdown":
		m.scroller.PageDown()
	case "ctrl+u", "pgup":
		m.scroller.PageUp()
	case "s", "r":
		return m.StartScan()
	case "a", "enter":
		return m.addSelectedDevice()
	case "m":
		m.method = shelly.DiscoveryMDNS
	case "h":
		m.method = shelly.DiscoveryHTTP
	case "c":
		m.method = shelly.DiscoveryCoIoT
	case "b":
		m.method = shelly.DiscoveryBLE
	}

	return m, nil
}

func (m Model) addSelectedDevice() (Model, tea.Cmd) {
	cursor := m.scroller.Cursor()
	if len(m.devices) == 0 || cursor >= len(m.devices) {
		return m, nil
	}

	device := m.devices[cursor]
	if device.Added {
		return m, nil // Already added
	}

	return m, func() tea.Msg {
		err := shelly.RegisterDiscoveredDevice(m.ctx, device, m.svc)
		return DeviceAddedMsg{Address: device.Address, Err: err}
	}
}

// View renders the Discovery component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Discovery").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused {
		r.SetFooter("s:scan a:add p:provision m/h/c/b:method")
	}

	var content strings.Builder

	// Method selector
	content.WriteString(m.renderMethodSelector())
	content.WriteString("\n\n")

	// Scan button / status
	if m.scanning {
		content.WriteString(m.loader.View())
		content.WriteString("\n")
	} else {
		content.WriteString(m.styles.ScanButton.Render("[s] Scan"))
	}
	content.WriteString("\n\n")

	// Error display with categorized messaging and retry hint
	if m.err != nil {
		msg, hint := tuierrors.FormatError(m.err)
		content.WriteString(m.styles.Error.Render(msg))
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render("  " + hint))
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render("  Press 'r' to retry"))
		content.WriteString("\n\n")
	}

	// Device list
	content.WriteString(m.renderDeviceList())

	r.SetContent(content.String())
	return r.Render()
}

func (m Model) renderDeviceList() string {
	if len(m.devices) == 0 && !m.scanning && m.err == nil {
		return m.styles.Muted.Render("No devices found. Press [s] to scan.")
	}

	if len(m.devices) == 0 {
		return ""
	}

	var content strings.Builder
	content.WriteString(m.styles.Label.Render(fmt.Sprintf("Found %d device(s):", len(m.devices))))
	content.WriteString("\n\n")

	start, end := m.scroller.VisibleRange()
	for i := start; i < end; i++ {
		device := m.devices[i]
		isSelected := m.scroller.IsCursorAt(i)
		content.WriteString(m.renderDeviceLine(device, isSelected))
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))

	return content.String()
}

func (m Model) renderMethodSelector() string {
	methods := []struct {
		method shelly.DiscoveryMethod
		key    string
		name   string
	}{
		{shelly.DiscoveryMDNS, "m", "mDNS"},
		{shelly.DiscoveryHTTP, "h", "HTTP"},
		{shelly.DiscoveryCoIoT, "c", "CoIoT"},
		{shelly.DiscoveryBLE, "b", "BLE"},
	}

	parts := make([]string, 0, len(methods))
	for _, method := range methods {
		style := m.styles.Muted
		if method.method == m.method {
			style = m.styles.ScanButton
		}
		parts = append(parts, style.Render(fmt.Sprintf("[%s] %s", method.key, method.name)))
	}

	return m.styles.Label.Render("Method: ") + strings.Join(parts, " ")
}

func (m Model) renderDeviceLine(device shelly.DiscoveredDevice, isSelected bool) string {
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Status indicator
	var statusStr string
	if device.Added {
		statusStr = m.styles.Added.Render("✓")
	} else {
		statusStr = m.styles.NotAdded.Render("○")
	}

	// Address
	addressStr := m.styles.Address.Render(device.Address)

	// Model
	modelStr := ""
	if device.Model != "" {
		modelStr = m.styles.Model.Render(" " + device.Model)
	}

	// Name
	nameStr := ""
	if device.Name != "" && device.Name != device.Address {
		nameStr = m.styles.Muted.Render(fmt.Sprintf(" (%s)", device.Name))
	}

	// Generation
	genStr := ""
	if device.Generation > 0 {
		genStr = m.styles.Generation.Render(fmt.Sprintf(" Gen%d", device.Generation))
	}

	line := fmt.Sprintf("%s%s %s%s%s%s", selector, statusStr, addressStr, modelStr, nameStr, genStr)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

// Devices returns the discovered devices.
func (m Model) Devices() []shelly.DiscoveredDevice {
	return m.devices
}

// Scanning returns whether a scan is in progress.
func (m Model) Scanning() bool {
	return m.scanning
}

// Method returns the current discovery method.
func (m Model) Method() shelly.DiscoveryMethod {
	return m.method
}

// SetMethod sets the discovery method.
func (m Model) SetMethod(method shelly.DiscoveryMethod) Model {
	m.method = method
	return m
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.scroller.Cursor()
}

// Refresh triggers a new scan.
func (m Model) Refresh() (Model, tea.Cmd) {
	return m.StartScan()
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll g/G:top/bottom enter:add r:refresh"
}

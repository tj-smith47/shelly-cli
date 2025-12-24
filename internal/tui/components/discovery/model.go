// Package discovery provides TUI components for device discovery.
package discovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
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
	ctx      context.Context
	svc      *shelly.Service
	devices  []shelly.DiscoveredDevice
	cursor   int
	scroll   int
	scanning bool
	method   shelly.DiscoveryMethod
	err      error
	width    int
	height   int
	focused  bool
	styles   Styles
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
		panic(fmt.Sprintf("discovery: invalid deps: %v", err))
	}

	return Model{
		ctx:      deps.Ctx,
		svc:      deps.Svc,
		scanning: false,
		method:   shelly.DiscoveryMDNS,
		styles:   DefaultStyles(),
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
	return m
}

// SetFocused sets the focus state.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
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
	m.cursor = 0
	m.scroll = 0
	return m, m.scanDevices()
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
	switch msg := msg.(type) {
	case ScanCompleteMsg:
		m.scanning = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.devices = msg.Devices
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
		m = m.cursorDown()
	case "k", "up":
		m = m.cursorUp()
	case "s", "r":
		return m.StartScan()
	case "a", "enter":
		return m.addSelectedDevice()
	case "1":
		m.method = shelly.DiscoveryMDNS
	case "2":
		m.method = shelly.DiscoveryHTTP
	case "3":
		m.method = shelly.DiscoveryCoIoT
	}

	return m, nil
}

func (m Model) cursorDown() Model {
	if m.cursor < len(m.devices)-1 {
		m.cursor++
		m = m.ensureVisible()
	}
	return m
}

func (m Model) cursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
		m = m.ensureVisible()
	}
	return m
}

func (m Model) ensureVisible() Model {
	visible := m.visibleRows()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	} else if m.cursor >= m.scroll+visible {
		m.scroll = m.cursor - visible + 1
	}
	return m
}

func (m Model) visibleRows() int {
	rows := m.height - 8 // Reserve space for header, method selector, and footer
	if rows < 1 {
		return 1
	}
	return rows
}

func (m Model) addSelectedDevice() (Model, tea.Cmd) {
	if len(m.devices) == 0 || m.cursor >= len(m.devices) {
		return m, nil
	}

	device := m.devices[m.cursor]
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
		SetFocused(m.focused)

	var content strings.Builder

	// Method selector
	content.WriteString(m.renderMethodSelector())
	content.WriteString("\n\n")

	// Scan button / status
	if m.scanning {
		content.WriteString(m.styles.Muted.Render("Scanning..."))
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

	// Help text
	content.WriteString("\n\n")
	content.WriteString(m.styles.Muted.Render("j/k: navigate | a: add | s: scan"))

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
	content.WriteString(m.styles.Label.Render(fmt.Sprintf("Found %d device(s):\n\n", len(m.devices))))

	visible := m.visibleRows()
	endIdx := m.scroll + visible
	if endIdx > len(m.devices) {
		endIdx = len(m.devices)
	}

	for i := m.scroll; i < endIdx; i++ {
		device := m.devices[i]
		isSelected := i == m.cursor
		content.WriteString(m.renderDeviceLine(device, isSelected))
		if i < endIdx-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(m.devices) > visible {
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render(
			fmt.Sprintf("[%d/%d]", m.cursor+1, len(m.devices)),
		))
	}

	return content.String()
}

func (m Model) renderMethodSelector() string {
	methods := []struct {
		method shelly.DiscoveryMethod
		key    string
		name   string
	}{
		{shelly.DiscoveryMDNS, "1", "mDNS"},
		{shelly.DiscoveryHTTP, "2", "HTTP"},
		{shelly.DiscoveryCoIoT, "3", "CoIoT"},
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
	return m.cursor
}

// Refresh triggers a new scan.
func (m Model) Refresh() (Model, tea.Cmd) {
	return m.StartScan()
}

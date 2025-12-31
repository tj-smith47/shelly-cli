// Package deviceinfo provides a device information panel component.
package deviceinfo

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// RequestJSONMsg is sent when user requests JSON view for an endpoint.
type RequestJSONMsg struct {
	DeviceName string
	Endpoint   string
}

// Model represents the device info component state.
type Model struct {
	device          *cache.DeviceData
	focused         bool
	panelIndex      int
	componentCursor int // -1 = all, >=0 = specific component
	endpointCursor  int // Selected endpoint for JSON viewer
	scrollOffset    int
	width           int
	height          int
	styles          Styles
}

// ComponentInfo holds information about a device component.
type ComponentInfo struct {
	Name       string
	Type       string
	State      string
	StateColor color.Color
	Power      *float64 // nil if not applicable
	Endpoint   string   // API endpoint for JSON viewing
}

// Styles for the device info component.
type Styles struct {
	Container   lipgloss.Style
	Title       lipgloss.Style
	Section     lipgloss.Style
	Label       lipgloss.Style
	Value       lipgloss.Style
	Online      lipgloss.Style
	Offline     lipgloss.Style
	OnState     lipgloss.Style
	OffState    lipgloss.Style
	Selected    lipgloss.Style
	Muted       lipgloss.Style
	Power       lipgloss.Style
	Focused     lipgloss.Style
	FocusBorder lipgloss.Style
	BlurBorder  lipgloss.Style
}

// DefaultStyles returns default styles for the device info component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			Padding(0, 1),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Section: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text).
			Width(12),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Online: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		Offline: lipgloss.NewStyle().
			Foreground(colors.Offline).
			Bold(true),
		OnState: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		OffState: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Selected: lipgloss.NewStyle().
			Background(colors.Highlight).
			Foreground(colors.Primary),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Power: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),
		Focused: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight),
		FocusBorder: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		BlurBorder: lipgloss.NewStyle().
			Foreground(colors.TableBorder),
	}
}

// New creates a new device info model.
func New() Model {
	return Model{
		componentCursor: -1, // Show all by default
		endpointCursor:  0,
		styles:          DefaultStyles(),
	}
}

// Init initializes the device info component.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the device info component.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	return m.handleKeyPress(keyMsg)
}

func (m Model) handleKeyPress(keyMsg tea.KeyPressMsg) (Model, tea.Cmd) {
	components := m.getComponents()
	maxIdx := len(components) - 1

	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("h", "left"))) {
		return m.handleLeft(), nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("l", "right"))) {
		return m.handleRight(maxIdx), nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("j", "down"))) {
		return m.handleDown(maxIdx), nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("k", "up"))) {
		return m.handleUp(), nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("enter"))) {
		return m.handleEnter(components, maxIdx)
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("a"))) {
		return m.handleToggleAll(), nil
	}
	return m, nil
}

func (m Model) handleLeft() Model {
	if m.componentCursor > -1 {
		m.componentCursor--
	}
	return m
}

func (m Model) handleRight(maxIdx int) Model {
	if m.componentCursor < maxIdx {
		m.componentCursor++
	}
	return m
}

func (m Model) handleDown(maxIdx int) Model {
	if m.componentCursor >= 0 && m.componentCursor <= maxIdx {
		m.endpointCursor++
	}
	return m
}

func (m Model) handleUp() Model {
	if m.endpointCursor > 0 {
		m.endpointCursor--
	}
	return m
}

func (m Model) handleEnter(components []ComponentInfo, maxIdx int) (Model, tea.Cmd) {
	if m.device == nil || m.componentCursor < 0 || m.componentCursor > maxIdx {
		return m, nil
	}
	comp := components[m.componentCursor]
	if comp.Endpoint == "" {
		return m, nil
	}
	return m, func() tea.Msg {
		return RequestJSONMsg{
			DeviceName: m.device.Device.Name,
			Endpoint:   comp.Endpoint,
		}
	}
}

func (m Model) handleToggleAll() Model {
	if m.componentCursor == -1 {
		m.componentCursor = 0
	} else {
		m.componentCursor = -1
	}
	return m
}

// View renders the device info panel.
func (m Model) View() string {
	if m.device == nil {
		return m.renderEmpty()
	}

	colors := theme.GetSemanticColors()
	borderColor := colors.TableBorder
	if m.focused {
		borderColor = colors.Highlight
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Device Info").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex).
		SetFocusColor(borderColor).
		SetBlurColor(colors.TableBorder).
		SetFooter(m.FooterText())

	content := m.buildContent()

	return r.
		SetContent(content).
		Render()
}

func (m Model) buildContent() string {
	var content strings.Builder
	m.writeDeviceHeader(&content)
	m.writeDeviceInfo(&content)
	m.writeFirmwareInfo(&content)

	components := m.getComponents()
	m.writeComponentsSection(&content, components)
	m.writePowerSection(&content)
	m.writeMetricsSection(&content)
	m.writeNavigationHint(&content, components)

	return content.String()
}

func (m Model) writeDeviceHeader(content *strings.Builder) {
	statusStr := m.styles.Online.Render("Online")
	if !m.device.Online {
		statusStr = m.styles.Offline.Render("Offline")
	}
	content.WriteString(m.styles.Title.Render(m.device.Device.Name) + " " + statusStr + "\n")
}

func (m Model) writeDeviceInfo(content *strings.Builder) {
	if m.device.Device.Model != "" {
		content.WriteString(m.renderRow("Model", m.device.Device.Model) + "\n")
	}
	if m.device.Device.Type != "" {
		content.WriteString(m.renderRow("Type", m.device.Device.Type) + "\n")
	}
}

func (m Model) writeFirmwareInfo(content *strings.Builder) {
	if m.device.Info == nil {
		return
	}
	if m.device.Info.Firmware != "" {
		content.WriteString(m.renderRow("Firmware", m.device.Info.Firmware) + "\n")
	}
	// App field is redundant with Model - Type now shows model code/SKU
}

func (m Model) writeComponentsSection(content *strings.Builder, components []ComponentInfo) {
	if len(components) == 0 {
		return
	}
	content.WriteString("\n")
	content.WriteString(m.styles.Section.Render("Components") + "\n")
	content.WriteString(m.renderComponents(components))
}

func (m Model) writePowerSection(content *strings.Builder) {
	if m.device.Power == 0 {
		return
	}
	content.WriteString("\n")
	content.WriteString(m.styles.Section.Render("Power") + "\n")
	content.WriteString(m.styles.Power.Render(formatPower(m.device.Power)) + "\n")
}

func (m Model) writeMetricsSection(content *strings.Builder) {
	if m.device.Voltage == 0 && m.device.Current == 0 {
		return
	}
	content.WriteString("\n")
	content.WriteString(m.styles.Section.Render("Metrics") + "\n")
	if m.device.Voltage != 0 {
		content.WriteString(m.renderRow("Voltage", fmt.Sprintf("%.1f V", m.device.Voltage)) + "\n")
	}
	if m.device.Current != 0 {
		content.WriteString(m.renderRow("Current", fmt.Sprintf("%.3f A", m.device.Current)) + "\n")
	}
	if m.device.TotalEnergy != 0 {
		content.WriteString(m.renderRow("Energy", fmt.Sprintf("%.2f Wh", m.device.TotalEnergy)) + "\n")
	}
}

func (m Model) writeNavigationHint(content *strings.Builder, components []ComponentInfo) {
	// Footer is now shown via FooterText() - don't inline it
}

func (m Model) renderEmpty() string {
	colors := theme.GetSemanticColors()
	r := rendering.New(m.width, m.height).
		SetTitle("Device Info").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex).
		SetFocusColor(colors.Highlight).
		SetBlurColor(colors.TableBorder).
		SetContent(m.styles.Muted.Render("Select a device to view details"))

	return r.Render()
}

func (m Model) renderRow(label, value string) string {
	return m.styles.Label.Render(label+":") + " " + m.styles.Value.Render(value)
}

func (m Model) renderComponents(components []ComponentInfo) string {
	if m.componentCursor == -1 {
		// Show all components in compact format
		return m.renderAllComponents(components)
	}

	// Show single component with details
	if m.componentCursor >= 0 && m.componentCursor < len(components) {
		return m.renderSingleComponent(components[m.componentCursor])
	}

	return ""
}

func (m Model) renderAllComponents(components []ComponentInfo) string {
	lines := make([]string, 0, len(components))
	for i, comp := range components {
		prefix := "  "
		if i == m.componentCursor {
			prefix = "> "
		}

		stateStyle := m.styles.OffState
		if comp.State == "on" || comp.State == "On" {
			stateStyle = m.styles.OnState
		}

		line := prefix + comp.Name + " " + stateStyle.Render("["+comp.State+"]")
		if comp.Power != nil && *comp.Power != 0 {
			line += " " + m.styles.Power.Render(formatPower(*comp.Power))
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderSingleComponent(comp ComponentInfo) string {
	// Component header with selection indicator
	header := m.styles.Selected.Render(" " + comp.Name + " ")

	stateStyle := m.styles.OffState
	if comp.State == "on" || comp.State == "On" {
		stateStyle = m.styles.OnState
	}

	lines := []string{
		header,
		"",
		m.renderRow("Type", comp.Type),
		m.renderRow("State", stateStyle.Render(comp.State)),
	}

	if comp.Power != nil {
		lines = append(lines, m.renderRow("Power", m.styles.Power.Render(formatPower(*comp.Power))))
	}

	if comp.Endpoint != "" {
		lines = append(lines, "", m.styles.Muted.Render("Endpoint: "+comp.Endpoint))
	}

	return strings.Join(lines, "\n")
}

func (m Model) getComponents() []ComponentInfo {
	if m.device == nil || !m.device.Online {
		return nil
	}

	// Estimate capacity based on known components
	capacity := len(m.device.Switches) + len(m.device.Lights) + len(m.device.Covers)
	if m.device.Snapshot != nil {
		capacity += len(m.device.Snapshot.PM) + len(m.device.Snapshot.EM) + len(m.device.Snapshot.EM1)
	}
	components := make([]ComponentInfo, 0, capacity)

	// Add switch components from cache
	for _, sw := range m.device.Switches {
		state := "off"
		if sw.On {
			state = "on"
		}
		components = append(components, ComponentInfo{
			Name:     fmt.Sprintf("Switch:%d", sw.ID),
			Type:     "Switch",
			State:    state,
			Endpoint: fmt.Sprintf("Switch.GetStatus?id=%d", sw.ID),
		})
	}

	// Add light components from cache (includes dimmers)
	for _, lt := range m.device.Lights {
		state := "off"
		if lt.On {
			state = "on"
		}
		components = append(components, ComponentInfo{
			Name:     fmt.Sprintf("Light:%d", lt.ID),
			Type:     "Light",
			State:    state,
			Endpoint: fmt.Sprintf("Light.GetStatus?id=%d", lt.ID),
		})
	}

	// Add cover components from cache
	for _, cv := range m.device.Covers {
		state := cv.State
		if state == "" {
			state = "stopped"
		}
		components = append(components, ComponentInfo{
			Name:     fmt.Sprintf("Cover:%d", cv.ID),
			Type:     "Cover",
			State:    state,
			Endpoint: fmt.Sprintf("Cover.GetStatus?id=%d", cv.ID),
		})
	}

	// Check if there's power monitoring data
	if m.device.Snapshot != nil {
		// Add PM (Power Meter) components
		for i, pm := range m.device.Snapshot.PM {
			power := pm.APower
			components = append(components, ComponentInfo{
				Name:     fmt.Sprintf("PM:%d", i),
				Type:     "Power Meter",
				State:    "active",
				Power:    &power,
				Endpoint: fmt.Sprintf("PM.GetStatus?id=%d", i),
			})
		}

		// Add EM (Energy Meter) components
		for i, em := range m.device.Snapshot.EM {
			power := em.TotalActivePower
			components = append(components, ComponentInfo{
				Name:     fmt.Sprintf("EM:%d", i),
				Type:     "Energy Meter",
				State:    "active",
				Power:    &power,
				Endpoint: fmt.Sprintf("EM.GetStatus?id=%d", i),
			})
		}

		// Add EM1 (Single-phase Energy Meter) components
		for i, em1 := range m.device.Snapshot.EM1 {
			power := em1.ActPower
			components = append(components, ComponentInfo{
				Name:     fmt.Sprintf("EM1:%d", i),
				Type:     "Energy Meter 1",
				State:    "active",
				Power:    &power,
				Endpoint: fmt.Sprintf("EM1.GetStatus?id=%d", i),
			})
		}
	}

	return components
}

// SetDevice sets the device to display.
func (m Model) SetDevice(d *cache.DeviceData) Model {
	m.device = d
	m.componentCursor = -1 // Reset to show all
	m.endpointCursor = 0
	m.scrollOffset = 0
	return m
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets whether the component is focused.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// Focused returns whether the component is focused.
func (m Model) Focused() bool {
	return m.focused
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return m
}

// SelectedComponent returns the selected component index (-1 = all).
func (m Model) SelectedComponent() int {
	return m.componentCursor
}

// SelectedEndpoint returns the selected endpoint for JSON viewing.
func (m Model) SelectedEndpoint() string {
	components := m.getComponents()
	if m.componentCursor >= 0 && m.componentCursor < len(components) {
		return components[m.componentCursor].Endpoint
	}
	return ""
}

// Device returns the current device data.
func (m Model) Device() *cache.DeviceData {
	return m.device
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	if m.device == nil {
		return ""
	}
	components := m.getComponents()
	if len(components) == 0 {
		return ""
	}
	if len(components) == 1 {
		return "space:toggle enter:json"
	}
	return "h/l:select a:all space:toggle enter:json"
}

// formatPower formats a power value with appropriate units.
func formatPower(value float64) string {
	absVal := value
	if absVal < 0 {
		absVal = -absVal
	}

	if absVal >= 1000 {
		return fmt.Sprintf("%.2f kW", value/1000)
	}
	return fmt.Sprintf("%.1f W", value)
}

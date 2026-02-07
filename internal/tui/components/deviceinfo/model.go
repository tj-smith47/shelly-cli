// Package deviceinfo provides a device information panel component.
package deviceinfo

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
)

// RequestJSONMsg is sent when user requests JSON view for an endpoint.
type RequestJSONMsg struct {
	DeviceName string
	Endpoint   string
}

// RequestToggleMsg is sent when user requests to toggle a component.
type RequestToggleMsg struct {
	DeviceName    string
	ComponentType string // "switch", "light", "cover"
	ComponentID   int
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
	ID         int
	Name       string
	Type       string
	State      string
	StateColor color.Color
	Power      *float64
	Endpoint   string
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
	Warning     lipgloss.Style
	Good        lipgloss.Style
	Fair        lipgloss.Style
	Weak        lipgloss.Style
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
			Foreground(colors.Highlight),
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
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Good: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		Fair: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Weak: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Focused: styles.PanelBorderActive(),
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

	switch msg := msg.(type) {
	case messages.NavigationMsg:
		return m.handleNavigation(msg), nil
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}
	return m, nil
}

// handleNavigation handles NavigationMsg for cursor movement.
func (m Model) handleNavigation(msg messages.NavigationMsg) Model {
	components := m.getComponents()
	maxIdx := len(components) - 1

	switch msg.Direction {
	case messages.NavDown:
		return m.handleDown(maxIdx)
	case messages.NavUp:
		return m.handleUp()
	case messages.NavLeft:
		return m.handleLeft()
	case messages.NavRight:
		return m.handleRight(maxIdx)
	case messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Page navigation not applicable for device info panel
	}
	return m
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
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys(keyconst.KeyEnter))) {
		return m.handleEnter(components, maxIdx)
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys("a"))) {
		return m.handleToggleAll(), nil
	}
	if key.Matches(keyMsg, key.NewBinding(key.WithKeys(" ", "t"))) {
		return m.handleToggle(components)
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

func (m Model) handleToggle(components []ComponentInfo) (Model, tea.Cmd) {
	if m.device == nil {
		return m, nil
	}

	idx := m.componentCursor
	if idx < 0 {
		idx = 0
	}
	if idx >= len(components) {
		return m, nil
	}

	comp := components[idx]

	var compType string
	switch comp.Type {
	case "Switch":
		compType = "switch"
	case "Light":
		compType = "light"
	case "Cover":
		compType = "cover"
	default:
		return m, nil
	}

	return m, func() tea.Msg {
		return RequestToggleMsg{
			DeviceName:    m.device.Device.Name,
			ComponentType: compType,
			ComponentID:   comp.ID,
		}
	}
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

// buildContent builds a compact horizontal layout for the device info.
func (m Model) buildContent() string {
	// Line 1: Device header with status
	lines := []string{m.buildHeaderLine()}

	// Power section (right after device name)
	if powerLine := m.buildPowerLine(); powerLine != "" {
		lines = append(lines, powerLine)
	}

	// Identity section - inline header.
	if identLine := m.buildIdentityLine(); identLine != "" {
		lines = append(lines, m.inlineSection("ID", identLine))
	}

	// Network section - inline header.
	if netLine := m.buildNetworkLine(); netLine != "" {
		lines = append(lines, m.inlineSection("Net", netLine))
	}

	// Runtime section - inline header.
	if runtimeLine := m.buildRuntimeLine(); runtimeLine != "" {
		lines = append(lines, m.inlineSection("Sys", runtimeLine))
	}

	// Components - horizontal compact format.
	if compLine := m.buildComponentsHorizontal(); compLine != "" {
		lines = append(lines, compLine)
	}

	return strings.Join(lines, "\n")
}

// buildHeaderLine builds the device header with name and status.
func (m Model) buildHeaderLine() string {
	statusStr := m.styles.Online.Render("● Online")
	if !m.device.Online {
		statusStr = m.styles.Offline.Render("○ Offline")
	}
	return m.styles.Title.Render(m.device.Device.Name) + " " + statusStr
}

// buildIdentityLine builds a compact line with model and generation info.
func (m Model) buildIdentityLine() string {
	var parts []string

	// Model
	if m.device.Device.Model != "" {
		parts = append(parts, m.kv("Model", m.device.Device.Model))
	}

	// Generation with chip
	if m.device.Info != nil && m.device.Info.Generation > 0 {
		gen := m.device.Info.Generation
		chip := chipTypeForGeneration(gen)
		parts = append(parts, m.kv("Gen", fmt.Sprintf("%d (%s)", gen, chip)))
	}

	// Type
	if m.device.Device.Type != "" {
		parts = append(parts, m.kv("Type", m.device.Device.Type))
	}

	return strings.Join(parts, " │ ")
}

// buildNetworkLine builds a compact line with network info.
func (m Model) buildNetworkLine() string {
	var parts []string

	// IP Address
	if m.device.Device.Address != "" {
		parts = append(parts, m.kv("IP", m.device.Device.Address))
	}

	// MAC (full address)
	if mac := m.getMAC(); mac != "" {
		parts = append(parts, m.kv("MAC", mac))
	}

	// WiFi SSID
	if m.device.WiFi != nil && m.device.WiFi.SSID != "" {
		parts = append(parts, m.kv("SSID", m.device.WiFi.SSID))
	}

	// WiFi signal
	if m.device.WiFi != nil && m.device.WiFi.RSSI != 0 {
		rssi := m.device.WiFi.RSSI
		quality := m.signalQualityShort(rssi)
		parts = append(parts, m.kv("Signal", fmt.Sprintf("%d dBm %s", rssi, quality)))
	}

	// AP Mode client count
	if m.device.WiFi != nil && m.device.WiFi.APCount > 0 {
		parts = append(parts, m.kv("AP Clients", fmt.Sprintf("%d", m.device.WiFi.APCount)))
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " │ ")
}

// buildRuntimeLine builds a compact line with runtime and firmware info.
func (m Model) buildRuntimeLine() string {
	var parts []string

	// Uptime
	if m.device.Sys != nil && m.device.Sys.Uptime > 0 {
		parts = append(parts, m.kv("Up", formatUptimeShort(m.device.Sys.Uptime)))
	}

	// RAM usage
	if m.device.Sys != nil && m.device.Sys.RAMSize > 0 {
		ramPct := float64(m.device.Sys.RAMSize-m.device.Sys.RAMFree) / float64(m.device.Sys.RAMSize) * 100
		parts = append(parts, m.kv("RAM", fmt.Sprintf("%.0f%%", ramPct)))
	}

	// Flash usage
	if m.device.Sys != nil && m.device.Sys.FSSize > 0 {
		fsPct := float64(m.device.Sys.FSSize-m.device.Sys.FSFree) / float64(m.device.Sys.FSSize) * 100
		parts = append(parts, m.kv("FS", fmt.Sprintf("%.0f%%", fsPct)))
	}

	// Firmware version
	if m.device.Info != nil && m.device.Info.Firmware != "" {
		parts = append(parts, m.kv("FW", m.device.Info.Firmware))
	}

	// Update available
	if m.device.Sys != nil && m.device.Sys.UpdateAvailable != "" {
		parts = append(parts, m.styles.Good.Render("▲ "+m.device.Sys.UpdateAvailable))
	}

	// Restart required
	if m.device.Sys != nil && m.device.Sys.RestartRequired {
		parts = append(parts, m.styles.Warning.Render("⟳ restart"))
	}

	// Last seen timestamp.
	if !m.device.UpdatedAt.IsZero() {
		parts = append(parts, m.kv("Seen", formatRelativeTime(m.device.UpdatedAt)))
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " │ ")
}

// buildComponentsHorizontal builds a compact horizontal component summary.
func (m Model) buildComponentsHorizontal() string {
	components := m.getComponents()
	if len(components) == 0 {
		return ""
	}

	// Show single component detail view if one is selected.
	if m.componentCursor >= 0 && m.componentCursor < len(components) {
		return m.renderSingleComponent(components[m.componentCursor])
	}

	// Build compact horizontal component list.
	parts := make([]string, 0, len(components))
	for i, comp := range components {
		stateStyle := m.styles.OffState
		stateChar := "○"
		if comp.State == "on" || comp.State == "On" || comp.State == "active" {
			stateStyle = m.styles.OnState
			stateChar = "●"
		}

		// Highlight selected component.
		name := comp.Name
		if m.componentCursor == i {
			name = m.styles.Selected.Render(name)
		}

		part := name + stateStyle.Render(stateChar)
		if comp.Power != nil && *comp.Power != 0 {
			part += m.styles.Power.Render(formatPowerShort(*comp.Power))
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, " ")
}

// renderSingleComponent renders detailed view of a single component.
func (m Model) renderSingleComponent(comp ComponentInfo) string {
	header := m.styles.Selected.Render(" " + comp.Name + " ")

	stateStyle := m.styles.OffState
	if comp.State == "on" || comp.State == "On" || comp.State == "active" {
		stateStyle = m.styles.OnState
	}

	lines := []string{
		header,
		"",
		m.kv("Type", comp.Type),
		m.kv("State", stateStyle.Render(comp.State)),
	}

	if comp.Power != nil {
		lines = append(lines, m.kv("Power", m.styles.Power.Render(formatPower(*comp.Power))))
	}

	if comp.Endpoint != "" {
		lines = append(lines, "", m.styles.Muted.Render("Endpoint: "+comp.Endpoint))
	}

	return strings.Join(lines, "\n")
}

// buildPowerLine builds a compact power summary line.
func (m Model) buildPowerLine() string {
	if m.device.Power == 0 && m.device.Voltage == 0 && m.device.Current == 0 {
		return ""
	}

	var parts []string

	if m.device.Power != 0 {
		parts = append(parts, m.styles.Power.Render(formatPower(m.device.Power)))
	}
	if m.device.Voltage != 0 {
		parts = append(parts, fmt.Sprintf("%.0fV", m.device.Voltage))
	}
	if m.device.Current != 0 {
		parts = append(parts, fmt.Sprintf("%.2fA", m.device.Current))
	}
	if m.device.TotalEnergy != 0 {
		parts = append(parts, formatEnergyShort(m.device.TotalEnergy))
	}

	if len(parts) == 0 {
		return ""
	}
	return m.styles.Section.Render("Power: ") + strings.Join(parts, " │ ")
}

// kv formats a key-value pair compactly with space after colon.
func (m Model) kv(label, value string) string {
	return m.styles.Label.Render(label+": ") + m.styles.Value.Render(value)
}

// inlineSection creates a compact inline section header.
func (m Model) inlineSection(name, content string) string {
	return m.styles.Section.Render(name+": ") + content
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

func (m Model) getComponents() []ComponentInfo {
	if m.device == nil || !m.device.Online {
		return nil
	}

	// Estimate capacity based on known components
	capacity := len(m.device.Switches) + len(m.device.Lights) + len(m.device.Covers) + len(m.device.Inputs)
	if m.device.Snapshot != nil {
		capacity += len(m.device.Snapshot.PM) + len(m.device.Snapshot.EM) + len(m.device.Snapshot.EM1)
	}
	components := make([]ComponentInfo, 0, capacity)

	components = m.appendSwitchComponents(components)
	components = m.appendLightComponents(components)
	components = m.appendCoverComponents(components)
	components = m.appendInputComponents(components)
	components = m.appendPowerMonitoringComponents(components)

	return components
}

func (m Model) appendSwitchComponents(components []ComponentInfo) []ComponentInfo {
	for _, sw := range m.device.Switches {
		state := "off"
		if sw.On {
			state = "on"
		}
		name := sw.Name
		if name == "" {
			name = fmt.Sprintf("Sw%d", sw.ID)
		}
		components = append(components, ComponentInfo{
			ID:       sw.ID,
			Name:     name,
			Type:     "Switch",
			State:    state,
			Endpoint: fmt.Sprintf("Switch.GetStatus?id=%d", sw.ID),
		})
	}
	return components
}

func (m Model) appendLightComponents(components []ComponentInfo) []ComponentInfo {
	for _, lt := range m.device.Lights {
		state := "off"
		if lt.On {
			state = "on"
		}
		name := lt.Name
		if name == "" {
			name = fmt.Sprintf("Lt%d", lt.ID)
		}
		components = append(components, ComponentInfo{
			ID:       lt.ID,
			Name:     name,
			Type:     "Light",
			State:    state,
			Endpoint: fmt.Sprintf("Light.GetStatus?id=%d", lt.ID),
		})
	}
	return components
}

func (m Model) appendCoverComponents(components []ComponentInfo) []ComponentInfo {
	for _, cv := range m.device.Covers {
		state := cv.State
		if state == "" {
			state = "stop"
		}
		if len(state) > 4 {
			state = state[:4]
		}
		name := cv.Name
		if name == "" {
			name = fmt.Sprintf("Cv%d", cv.ID)
		}
		components = append(components, ComponentInfo{
			ID:       cv.ID,
			Name:     name,
			Type:     "Cover",
			State:    state,
			Endpoint: fmt.Sprintf("Cover.GetStatus?id=%d", cv.ID),
		})
	}
	return components
}

func (m Model) appendInputComponents(components []ComponentInfo) []ComponentInfo {
	for _, inp := range m.device.Inputs {
		state := "low"
		if inp.State {
			state = "high"
		}
		inputType := inp.Type
		if inputType == "" {
			inputType = "Input"
		}
		name := inp.Name
		if name == "" {
			name = fmt.Sprintf("In%d", inp.ID)
		}
		components = append(components, ComponentInfo{
			ID:       inp.ID,
			Name:     name,
			Type:     inputType,
			State:    state,
			Endpoint: fmt.Sprintf("Input.GetStatus?id=%d", inp.ID),
		})
	}
	return components
}

func (m Model) appendPowerMonitoringComponents(components []ComponentInfo) []ComponentInfo {
	if m.device.Snapshot == nil {
		return components
	}

	for i, pm := range m.device.Snapshot.PM {
		power := pm.APower
		components = append(components, ComponentInfo{
			Name:     fmt.Sprintf("PM%d", i),
			Type:     "Power Meter",
			State:    "on",
			Power:    &power,
			Endpoint: fmt.Sprintf("PM.GetStatus?id=%d", i),
		})
	}

	for i, em := range m.device.Snapshot.EM {
		power := em.TotalActivePower
		components = append(components, ComponentInfo{
			Name:     fmt.Sprintf("EM%d", i),
			Type:     "Energy Meter",
			State:    "on",
			Power:    &power,
			Endpoint: fmt.Sprintf("EM.GetStatus?id=%d", i),
		})
	}

	for i, em1 := range m.device.Snapshot.EM1 {
		power := em1.ActPower
		components = append(components, ComponentInfo{
			Name:     fmt.Sprintf("EM1:%d", i),
			Type:     "Energy Meter 1",
			State:    "on",
			Power:    &power,
			Endpoint: fmt.Sprintf("EM1.GetStatus?id=%d", i),
		})
	}

	return components
}

// getMAC returns the MAC address from various sources.
func (m Model) getMAC() string {
	// Try device info first (populated from DeviceInfo)
	if m.device.Info != nil && m.device.Info.MAC != "" {
		return m.device.Info.MAC
	}
	// Fall back to device config
	if m.device.Device.MAC != "" {
		return m.device.Device.MAC
	}
	// Try Sys status
	if m.device.Sys != nil && m.device.Sys.MAC != "" {
		return m.device.Sys.MAC
	}
	return ""
}

// signalQualityShort returns a short signal quality indicator.
func (m Model) signalQualityShort(rssi int) string {
	switch {
	case rssi > -50:
		return m.styles.Good.Render("●●●●")
	case rssi > -60:
		return m.styles.Good.Render("●●●○")
	case rssi > -70:
		return m.styles.Fair.Render("●●○○")
	default:
		return m.styles.Weak.Render("●○○○")
	}
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
		return theme.StyledKeybindings(keys.FormatHints([]keys.Hint{{Key: "space", Desc: "toggle"}, {Key: "enter", Desc: "json"}}, keys.FooterHintWidth(m.width)))
	}
	return theme.StyledKeybindings(keys.FormatHints([]keys.Hint{{Key: "h/l", Desc: "select"}, {Key: "a", Desc: "all"}, {Key: "space", Desc: "toggle"}, {Key: "enter", Desc: "json"}}, keys.FooterHintWidth(m.width)))
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

// formatPowerShort formats power value compactly.
func formatPowerShort(value float64) string {
	absVal := value
	if absVal < 0 {
		absVal = -absVal
	}

	if absVal >= 1000 {
		return fmt.Sprintf("%.1fkW", value/1000)
	}
	return fmt.Sprintf("%.0fW", value)
}

// formatEnergyShort formats an energy value compactly.
func formatEnergyShort(wh float64) string {
	if wh >= 1000 {
		return fmt.Sprintf("%.1fkWh", wh/1000)
	}
	return fmt.Sprintf("%.0fWh", wh)
}

// formatUptimeShort formats seconds into compact duration.
func formatUptimeShort(seconds int) string {
	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60

	if d > 0 {
		return fmt.Sprintf("%dd%dh", d, h)
	}
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// formatRelativeTime formats a time as relative to now.
func formatRelativeTime(t time.Time) string {
	d := time.Since(t)

	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

// chipTypeForGeneration returns the chip type for a device generation.
func chipTypeForGeneration(gen int) string {
	switch gen {
	case 1:
		return "ESP8266"
	case 2, 3:
		return "ESP32"
	case 4:
		return "ESP32-C3"
	default:
		return "?"
	}
}

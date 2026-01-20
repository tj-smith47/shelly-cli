// Package energybars provides a visual energy/power bar component for the TUI.
package energybars

import (
	"cmp"
	"fmt"
	"image/color"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
)

// Bar represents a single energy bar (per-switch, not per-device).
type Bar struct {
	DeviceName string      // Device name for grouping
	SwitchName string      // Switch/input name (may be empty)
	Label      string      // Combined display label: "[Device] [Switch]"
	Value      float64     // Power value
	MaxVal     float64     // Max value for scaling (0 = auto)
	Unit       string      // Unit (W)
	Color      color.Color // Bar color
}

// Model represents the energy bars state.
type Model struct {
	helpers.Sizable
	bars       []Bar
	cache      *cache.Cache
	scroller   *panel.Scroller
	barHeight  int
	styles     Styles
	showTotal  bool
	focused    bool
	panelIndex int // For Shift+N hint
	loading    bool
}

// Styles for the energy bars.
type Styles struct {
	Container lipgloss.Style
	Header    lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	BarFill   lipgloss.Style
	BarEmpty  lipgloss.Style
	Total     lipgloss.Style
}

// DefaultStyles returns default styles for the energy bars.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: styles.PanelBorder().Padding(1, 2),
		Header: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text).
			Width(16),
		Value: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true).
			Width(10).
			Align(lipgloss.Right),
		BarFill: lipgloss.NewStyle().
			Foreground(colors.Warning),
		BarEmpty: lipgloss.NewStyle().
			Foreground(theme.Orange()), // Orange for empty bars, not black
		Total: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// New creates a new energy bars component.
func New(c *cache.Cache) Model {
	m := Model{
		Sizable:   helpers.NewSizableLoaderOnly(),
		cache:     c,
		scroller:  panel.NewScroller(0, 10), // Will be updated with actual counts
		barHeight: 1,
		styles:    DefaultStyles(),
		showTotal: true,
		loading:   true, // Start in loading state until cache is populated
	}
	m.Loader = m.Loader.SetMessage("Loading power data...")
	return m
}

// Init initializes the energy bars.
func (m Model) Init() tea.Cmd {
	if m.loading {
		return m.Loader.Tick()
	}
	return nil
}

// Update handles messages for the energy bars.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if navMsg, ok := msg.(messages.NavigationMsg); ok {
		return m.handleNavigation(navMsg), nil
	}

	if !m.loading {
		return m, nil
	}

	// Forward tick messages to loader when loading
	var cmd tea.Cmd
	m.Loader, cmd = m.Loader.Update(msg)

	// Auto-detect when cache has PM devices (data is ready)
	if m.hasPMDevicesInCache() {
		m.loading = false
	}

	return m, cmd
}

// handleNavigation handles NavigationMsg for scrolling.
func (m Model) handleNavigation(msg messages.NavigationMsg) Model {
	switch msg.Direction {
	case messages.NavDown:
		m.scroller.CursorDown()
	case messages.NavUp:
		m.scroller.CursorUp()
	case messages.NavPageDown:
		m.scroller.PageDown()
	case messages.NavPageUp:
		m.scroller.PageUp()
	case messages.NavHome:
		m.scroller.CursorToStart()
	case messages.NavEnd:
		m.scroller.CursorToEnd()
	case messages.NavLeft, messages.NavRight:
		// Horizontal navigation not applicable for energy bars
	}
	return m
}

// hasPMDevicesInCache checks if the cache has any PM-capable devices.
func (m Model) hasPMDevicesInCache() bool {
	if m.cache == nil {
		return false
	}
	devices := m.cache.GetOnlineDevices()
	for _, d := range devices {
		if hasPMCapability(d) {
			return true
		}
	}
	return false
}

// View renders the energy bars.
func (m Model) View() string {
	// Handle special states first
	if result, handled := m.checkSpecialStates(); handled {
		return result
	}

	// Collect bars from devices
	devices := m.cache.GetOnlineDevices()
	m.bars = m.collectBars(devices)
	if len(m.bars) == 0 {
		return m.renderNoData()
	}

	// Setup scroller with visible rows
	visibleRows := max(1, m.Height-4) // borders (2) + title/footer (2)
	m.scroller.SetItemCount(len(m.bars))
	m.scroller.SetVisibleRows(visibleRows)

	// Calculate layout dimensions
	labelWidth, barWidth := m.calculateLayout()

	// Render visible bars and calculate total
	content, totalPower := m.renderVisibleBars(labelWidth, barWidth)

	// Build panel with badge and content
	return m.buildPanel(content, totalPower)
}

// checkSpecialStates checks for loading or empty states (no data modification).
func (m Model) checkSpecialStates() (string, bool) {
	if m.loading {
		return m.renderLoading(), true
	}
	if m.cache == nil {
		return m.renderEmpty(), true
	}
	if len(m.cache.GetOnlineDevices()) == 0 {
		return m.renderEmpty(), true
	}
	return "", false
}

// calculateLayout computes label and bar widths based on content.
func (m Model) calculateLayout() (labelWidth, barWidth int) {
	maxLabelLen := 0
	for _, bar := range m.bars {
		maxLabelLen = max(maxLabelLen, len(bar.Label))
	}
	labelWidth = max(10, maxLabelLen+1)
	barWidth = max(5, m.Width-16-labelWidth)
	return labelWidth, barWidth
}

// renderVisibleBars renders only the visible portion and returns content and total.
func (m Model) renderVisibleBars(labelWidth, barWidth int) (renderedContent string, totalPower float64) {
	maxVal := max(100.0, m.findMaxValue())

	var sb strings.Builder
	start, end := m.scroller.VisibleRange()
	for i := start; i < end && i < len(m.bars); i++ {
		sb.WriteString(m.renderBar(m.bars[i], maxVal, barWidth, labelWidth) + "\n")
	}

	for _, bar := range m.bars {
		totalPower += bar.Value
	}
	return sb.String(), totalPower
}

// buildPanel creates the final rendered panel with badge and footer.
func (m Model) buildPanel(content string, totalPower float64) string {
	borderStyle := lipgloss.NewStyle().Foreground(theme.Yellow())
	barFillStyle := lipgloss.NewStyle().Foreground(theme.Orange())
	countInfo := fmt.Sprintf("%d switches", len(m.bars))
	if m.scroller.HasMore() || m.scroller.HasPrevious() {
		countInfo = m.scroller.ScrollInfoRange()
	}
	badge := borderStyle.Render(countInfo+" │ ") +
		barFillStyle.Render("██") + borderStyle.Render(" high ") +
		barFillStyle.Render("░░") + borderStyle.Render(" low")

	r := rendering.New(m.Width, m.Height).
		SetTitle("Power Consumption").
		SetBadge(badge).
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	m.applyBorderColor(r)

	if m.showTotal && len(m.bars) > 1 {
		r.SetFooter(fmt.Sprintf("Total: %s", formatValue(totalPower, "W")))
	}
	return r.SetContent(content).Render()
}

// applyBorderColor sets the appropriate border color for the panel.
// Uses the default blue focus color (from renderer) and yellow blur color.
func (m Model) applyBorderColor(r *rendering.Renderer) {
	r.SetBlurColor(theme.Yellow())
}

func (m Model) collectBars(devices []*cache.DeviceData) []Bar {
	var bars []Bar
	for _, d := range devices {
		if !hasPMCapability(d) {
			continue
		}
		bars = append(bars, collectDeviceBars(d)...)
	}

	// Sort by power value descending (highest power first)
	// Use Label as secondary sort key for stable ordering when power values are equal
	slices.SortFunc(bars, func(a, b Bar) int {
		if c := cmp.Compare(b.Value, a.Value); c != 0 {
			return c // Reversed for descending
		}
		return cmp.Compare(a.Label, b.Label) // Alphabetical for stability
	})

	return bars
}

// collectDeviceBars extracts bars for a single device based on its power source type.
func collectDeviceBars(d *cache.DeviceData) []Bar {
	deviceName := d.Device.Name

	switch {
	case len(d.SwitchPowers) > 0:
		return collectSwitchBars(deviceName, d)
	case len(d.PMPowers) > 0:
		return collectPMBars(deviceName, d.PMPowers)
	case len(d.EMPowers) > 0 || len(d.EM1Powers) > 0:
		return collectEMBars(deviceName, d.EMPowers, d.EM1Powers)
	default:
		return []Bar{{
			DeviceName: deviceName,
			SwitchName: "",
			Label:      deviceName,
			Value:      d.Power,
			MaxVal:     0,
			Unit:       "W",
			Color:      theme.Orange(),
		}}
	}
}

// collectSwitchBars creates bars for switch-integrated power monitoring (e.g., Plus 2PM).
func collectSwitchBars(deviceName string, d *cache.DeviceData) []Bar {
	// Build name lookup from Switches slice
	switchNames := make(map[int]string)
	for _, sw := range d.Switches {
		switchNames[sw.ID] = sw.Name
	}

	bars := make([]Bar, 0, len(d.SwitchPowers))
	for switchID, power := range d.SwitchPowers {
		switchName := switchNames[switchID]
		bars = append(bars, Bar{
			DeviceName: deviceName,
			SwitchName: switchName,
			Label:      formatSwitchLabel(deviceName, switchName, switchID, len(d.SwitchPowers)),
			Value:      power,
			MaxVal:     0,
			Unit:       "W",
			Color:      theme.Orange(),
		})
	}
	return bars
}

// collectPMBars creates bars for dedicated PM components.
func collectPMBars(deviceName string, pmPowers map[int]float64) []Bar {
	bars := make([]Bar, 0, len(pmPowers))
	for pmID, power := range pmPowers {
		label := deviceName
		if len(pmPowers) > 1 {
			label = fmt.Sprintf("%s PM%d", deviceName, pmID)
		}
		bars = append(bars, Bar{
			DeviceName: deviceName,
			SwitchName: fmt.Sprintf("PM%d", pmID),
			Label:      label,
			Value:      power,
			MaxVal:     0,
			Unit:       "W",
			Color:      theme.Orange(),
		})
	}
	return bars
}

// collectEMBars creates bars for energy meter components (EM and EM1).
func collectEMBars(deviceName string, emPowers, em1Powers map[int]float64) []Bar {
	bars := make([]Bar, 0, len(emPowers)+len(em1Powers))
	for emID, power := range emPowers {
		label := deviceName
		if len(emPowers) > 1 {
			label = fmt.Sprintf("%s EM%d", deviceName, emID)
		}
		bars = append(bars, Bar{
			DeviceName: deviceName,
			SwitchName: fmt.Sprintf("EM%d", emID),
			Label:      label,
			Value:      power,
			MaxVal:     0,
			Unit:       "W",
			Color:      theme.Orange(),
		})
	}
	for em1ID, power := range em1Powers {
		label := deviceName
		if len(em1Powers) > 1 {
			label = fmt.Sprintf("%s EM1:%d", deviceName, em1ID)
		}
		bars = append(bars, Bar{
			DeviceName: deviceName,
			SwitchName: fmt.Sprintf("EM1:%d", em1ID),
			Label:      label,
			Value:      power,
			MaxVal:     0,
			Unit:       "W",
			Color:      theme.Orange(),
		})
	}
	return bars
}

// formatSwitchLabel creates a display label for a switch.
// Format: "[Device Name] [Switch Name]" when there are multiple switches.
// If switch name is empty, shows "(Sw0)" format to identify the switch.
// If only one switch, just show device name.
// If switch name already starts with device name, uses switch name only to avoid duplication.
func formatSwitchLabel(deviceName, switchName string, switchID, switchCount int) string {
	if switchCount <= 1 {
		return deviceName
	}
	if switchName == "" {
		return fmt.Sprintf("%s (Sw%d)", deviceName, switchID)
	}
	// Avoid duplicating device name if switch name already includes it
	if strings.HasPrefix(strings.ToLower(switchName), strings.ToLower(deviceName)) {
		return switchName
	}
	return deviceName + " " + switchName
}

// hasPMCapability checks if a device has power monitoring capability.
// This checks both the snapshot (if PM/EM/EM1 components exist) and the model code.
func hasPMCapability(d *cache.DeviceData) bool {
	// Check snapshot for dedicated PM/EM components
	if d.Snapshot != nil && (len(d.Snapshot.PM) > 0 || len(d.Snapshot.EM) > 0 || len(d.Snapshot.EM1) > 0) {
		return true
	}

	// Check for switch-integrated power monitoring (e.g., Plus 2PM)
	// Devices like Plus 2PM have PM in the switch component (apower field),
	// not dedicated PM components. SwitchPowers is populated when switches report power.
	if len(d.SwitchPowers) > 0 {
		return true
	}

	// Also check cover-integrated power monitoring
	if len(d.CoverPowers) > 0 {
		return true
	}

	// Fallback: check device model for PM capability
	model := d.Device.Type // Use Type (raw model code) not Model (display name)
	if model == "" {
		model = d.Device.Model
	}

	return modelHasPM(model)
}

// modelHasPM checks if a device model code indicates power monitoring capability.
func modelHasPM(model string) bool {
	// Gen1 PM devices: SHSW-PM
	if strings.Contains(model, "-PM") {
		return true
	}

	// Gen2/Gen3 PM devices: SNSW-xxxPxxxx (P at position 8)
	// Examples: SNSW-001P16EU (Plus 1PM), SNSW-102P16EU (Plus 2PM)
	if strings.HasPrefix(model, "SNSW-") && len(model) >= 9 && model[8] == 'P' {
		return true
	}

	// Gen2/Gen3 Pro PM devices: SPSW-xxxPxxxxx (P at position 8)
	if strings.HasPrefix(model, "SPSW-") && len(model) >= 9 && model[8] == 'P' {
		return true
	}

	// Dimmers have power monitoring: SNDM-xxxx
	if strings.HasPrefix(model, "SNDM-") {
		return true
	}

	// Energy meters: SPEM (Pro EM), SNEM (Plus EM)
	if strings.HasPrefix(model, "SPEM-") || strings.HasPrefix(model, "SNEM-") {
		return true
	}

	return false
}

func (m Model) findMaxValue() float64 {
	var maxValue float64
	for _, bar := range m.bars {
		if bar.Value > maxValue {
			maxValue = bar.Value
		}
	}
	return maxValue
}

func (m Model) renderBar(bar Bar, maxVal float64, barWidth, labelWidth int) string {
	// Calculate fill percentage
	fillPct := bar.Value / maxVal
	if fillPct > 1 {
		fillPct = 1
	}
	if fillPct < 0 {
		fillPct = 0
	}

	fillCount := int(float64(barWidth) * fillPct)
	emptyCount := barWidth - fillCount

	// Build bar
	fillChar := "█"
	emptyChar := "░"

	fill := m.styles.BarFill.Foreground(bar.Color).Render(strings.Repeat(fillChar, fillCount))
	empty := m.styles.BarEmpty.Render(strings.Repeat(emptyChar, emptyCount))

	// Label - use full name, left-aligned with dynamic width
	labelStr := m.styles.Label.Width(labelWidth).Render(bar.Label)

	valueStr := formatValue(bar.Value, bar.Unit)
	valueRender := m.styles.Value.Width(10).Align(lipgloss.Right).Render(valueStr)

	return labelStr + " " + fill + empty + " " + valueRender
}

func (m Model) renderEmpty() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Power Consumption").
		SetFocused(false)
	centered := lipgloss.NewStyle().
		Width(m.Width-4).
		Height(m.Height-2).
		Align(lipgloss.Center, lipgloss.Center).
		Render("No devices online")
	return r.SetContent(centered).Render()
}

func (m Model) renderNoData() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Power Consumption").
		SetFocused(false)
	centered := lipgloss.NewStyle().
		Width(m.Width-4).
		Height(m.Height-2).
		Align(lipgloss.Center, lipgloss.Center).
		Render("No power data available")
	return r.SetContent(centered).Render()
}

func (m Model) renderLoading() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Power Consumption").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Use yellow borders for energy panels
	if m.focused {
		r.SetFocusColor(theme.Yellow())
	} else {
		r.SetBlurColor(theme.Yellow())
	}

	return r.SetContent(m.Loader.View()).Render()
}

// formatValue formats a power/energy value with appropriate units.
func formatValue(value float64, unit string) string {
	absVal := value
	if absVal < 0 {
		absVal = -absVal
	}

	if absVal >= 1000000 {
		return fmt.Sprintf("%.2f M%s", value/1000000, unit)
	}
	if absVal >= 1000 {
		return fmt.Sprintf("%.2f k%s", value/1000, unit)
	}
	return fmt.Sprintf("%.1f %s", value, unit)
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.ApplySize(width, height)
	return m
}

// ShowTotal enables or disables the total row.
func (m Model) ShowTotal(show bool) Model {
	m.showTotal = show
	return m
}

// SetBars manually sets the bars (for testing or custom display).
func (m Model) SetBars(bars []Bar) Model {
	m.bars = bars
	return m
}

// BarCount returns the number of bars.
func (m Model) BarCount() int {
	return len(m.bars)
}

// SetFocused sets whether this panel has focus.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return m
}

// SetLoading sets the loading state.
func (m Model) SetLoading(isLoading bool) Model {
	m.loading = isLoading
	return m
}

// StartLoading sets loading to true and returns a tick command.
func (m Model) StartLoading() (Model, tea.Cmd) {
	m.loading = true
	return m, m.Loader.Tick()
}

// IsLoading returns whether the component is in loading state.
func (m Model) IsLoading() bool {
	return m.loading
}

// ScrollUp scrolls the list up by one item.
func (m Model) ScrollUp() Model {
	m.scroller.CursorUp()
	return m
}

// ScrollDown scrolls the list down by one item.
func (m Model) ScrollDown() Model {
	m.scroller.CursorDown()
	return m
}

// PageUp scrolls up by one page.
func (m Model) PageUp() Model {
	m.scroller.PageUp()
	return m
}

// PageDown scrolls down by one page.
func (m Model) PageDown() Model {
	m.scroller.PageDown()
	return m
}

// CanScroll returns true if there are more items than visible rows.
func (m Model) CanScroll() bool {
	return m.scroller.HasMore() || m.scroller.HasPrevious()
}

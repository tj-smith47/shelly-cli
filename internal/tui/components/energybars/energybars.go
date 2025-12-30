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
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Bar represents a single energy bar.
type Bar struct {
	Label  string
	Value  float64
	MaxVal float64
	Unit   string
	Color  color.Color
}

// Model represents the energy bars state.
type Model struct {
	bars       []Bar
	cache      *cache.Cache
	width      int
	height     int
	barHeight  int
	styles     Styles
	showTotal  bool
	focused    bool
	panelIndex int // For Shift+N hint
	loading    bool
	loader     loading.Model
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
		Container: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Padding(1, 2),
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
	return Model{
		cache:     c,
		barHeight: 1,
		styles:    DefaultStyles(),
		showTotal: true,
		loading:   true, // Start in loading state until cache is populated
		loader: loading.New(
			loading.WithMessage("Loading power data..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
	}
}

// Init initializes the energy bars.
func (m Model) Init() tea.Cmd {
	if m.loading {
		return m.loader.Tick()
	}
	return nil
}

// Update handles messages for the energy bars.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.loading {
		return m, nil
	}

	// Forward tick messages to loader when loading
	var cmd tea.Cmd
	m.loader, cmd = m.loader.Update(msg)

	// Auto-detect when cache has PM devices (data is ready)
	if m.hasPMDevicesInCache() {
		m.loading = false
	}

	return m, cmd
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
	// Show loading indicator during initial load
	if m.loading {
		return m.renderLoading()
	}

	if m.cache == nil {
		return m.renderEmpty()
	}

	devices := m.cache.GetOnlineDevices()
	if len(devices) == 0 {
		return m.renderEmpty()
	}

	// Collect power data from devices
	m.bars = m.collectBars(devices)

	if len(m.bars) == 0 {
		return m.renderNoData()
	}

	var content strings.Builder

	// Find max value for scaling
	maxVal := m.findMaxValue()
	if maxVal < 100 {
		maxVal = 100 // Minimum scale for visibility
	}

	// Calculate max label width - never truncate names, shrink bars instead
	maxLabelLen := 0
	for _, bar := range m.bars {
		if len(bar.Label) > maxLabelLen {
			maxLabelLen = len(bar.Label)
		}
	}
	labelWidth := max(10, maxLabelLen+1) // +1 for spacing

	// Bar width calculation (shrinks to accommodate full names):
	// - Borders: 2 (left + right)
	// - Horizontal padding: 2 (1 each side inside border)
	// - Label: labelWidth (dynamic)
	// - Spaces: 2 (after label, after bar)
	// - Value: 10
	// Total overhead = 2 + 2 + labelWidth + 2 + 10 = 16 + labelWidth
	barWidth := m.width - 16 - labelWidth
	if barWidth < 5 {
		barWidth = 5 // Absolute minimum bar width
	}

	var totalPower float64
	for _, bar := range m.bars {
		content.WriteString(m.renderBar(bar, maxVal, barWidth, labelWidth) + "\n")
		totalPower += bar.Value
	}

	// Use rendering package for consistent embedded title styling
	// Show PM device count and legend in badge
	// Text in border color (yellow), Unicode chars in actual bar colors
	borderStyle := lipgloss.NewStyle().Foreground(theme.Yellow())
	barFillStyle := lipgloss.NewStyle().Foreground(theme.Orange())
	badge := borderStyle.Render(fmt.Sprintf("%d devices │ Legend: ", len(m.bars))) +
		barFillStyle.Render("██") + borderStyle.Render(" high ") +
		barFillStyle.Render("░░") + borderStyle.Render(" low")
	r := rendering.New(m.width, m.height).
		SetTitle("Power Consumption").
		SetBadge(badge).
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Use yellow borders for energy panels
	if m.focused {
		r.SetFocusColor(theme.Yellow())
	} else {
		r.SetBlurColor(theme.Yellow())
	}

	// Show total in footer when enabled and multiple devices
	if m.showTotal && len(m.bars) > 1 {
		r.SetFooter(fmt.Sprintf("Total: %s", formatValue(totalPower, "W")))
	}

	return r.SetContent(content.String()).Render()
}

func (m Model) collectBars(devices []*cache.DeviceData) []Bar {
	var bars []Bar
	for _, d := range devices {
		// Show all devices with power monitoring capability, even if power is 0
		// Check snapshot for PM/EM/EM1 components, or check model for PM capability
		hasPM := hasPMCapability(d)
		if hasPM {
			bars = append(bars, Bar{
				Label:  d.Device.Name,
				Value:  d.Power,
				MaxVal: 0, // Will be auto-calculated
				Unit:   "W",
				Color:  theme.Orange(),
			})
		}
	}

	// Sort by power value descending (highest power first)
	slices.SortFunc(bars, func(a, b Bar) int {
		return cmp.Compare(b.Value, a.Value) // Reversed for descending
	})

	return bars
}

// hasPMCapability checks if a device has power monitoring capability.
// This checks both the snapshot (if PM/EM/EM1 components exist) and the model code.
func hasPMCapability(d *cache.DeviceData) bool {
	// Check snapshot for actual PM components
	if d.Snapshot != nil && (len(d.Snapshot.PM) > 0 || len(d.Snapshot.EM) > 0 || len(d.Snapshot.EM1) > 0) {
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
	valueRender := m.styles.Value.Render(valueStr)

	return labelStr + " " + fill + empty + " " + valueRender
}

func (m Model) renderEmpty() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Power Consumption").
		SetFocused(false)
	centered := lipgloss.NewStyle().
		Width(m.width-4).
		Height(m.height-2).
		Align(lipgloss.Center, lipgloss.Center).
		Render("No devices online")
	return r.SetContent(centered).Render()
}

func (m Model) renderNoData() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Power Consumption").
		SetFocused(false)
	centered := lipgloss.NewStyle().
		Width(m.width-4).
		Height(m.height-2).
		Align(lipgloss.Center, lipgloss.Center).
		Render("No power data available")
	return r.SetContent(centered).Render()
}

func (m Model) renderLoading() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Power Consumption").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Use yellow borders for energy panels
	if m.focused {
		r.SetFocusColor(theme.Yellow())
	} else {
		r.SetBlurColor(theme.Yellow())
	}

	return r.SetContent(m.loader.View()).Render()
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
	m.width = width
	m.height = height
	m.loader = m.loader.SetSize(width-4, height-4)
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
	return m, m.loader.Tick()
}

// IsLoading returns whether the component is in loading state.
func (m Model) IsLoading() bool {
	return m.loading
}

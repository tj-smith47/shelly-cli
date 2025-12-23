// Package energybars provides a visual energy/power bar component for the TUI.
package energybars

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
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
	bars      []Bar
	cache     *cache.Cache
	width     int
	height    int
	barHeight int
	styles    Styles
	showTotal bool
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
			Foreground(colors.Muted),
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
	}
}

// Init initializes the energy bars.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the energy bars.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Energy bars are read-only, no key handling needed
	return m, nil
}

// View renders the energy bars.
func (m Model) View() string {
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

	// Header
	content.WriteString(m.styles.Header.Render("Power Consumption") + "\n\n")

	// Find max value for scaling
	maxVal := m.findMaxValue()
	if maxVal < 100 {
		maxVal = 100 // Minimum scale for visibility
	}

	// Render each bar
	barWidth := m.width - 34 // Label (16) + Value (10) + padding
	if barWidth < 10 {
		barWidth = 10
	}

	var totalPower float64
	for _, bar := range m.bars {
		content.WriteString(m.renderBar(bar, maxVal, barWidth) + "\n")
		totalPower += bar.Value
	}

	// Render total if enabled
	if m.showTotal && len(m.bars) > 1 {
		content.WriteString("\n" + m.renderTotal(totalPower))
	}

	return m.styles.Container.
		Width(m.width - 4).
		Height(m.height - 2).
		Render(content.String())
}

func (m Model) collectBars(devices []*cache.DeviceData) []Bar {
	var bars []Bar
	for _, d := range devices {
		if d.Power != 0 {
			bars = append(bars, Bar{
				Label:  d.Device.Name,
				Value:  d.Power,
				MaxVal: 0, // Will be auto-calculated
				Unit:   "W",
				Color:  theme.Orange(),
			})
		}
	}
	return bars
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

func (m Model) renderBar(bar Bar, maxVal float64, barWidth int) string {
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

	// Label and value
	label := bar.Label
	if len(label) > 14 {
		label = label[:11] + "..."
	}
	labelStr := m.styles.Label.Render(label)

	valueStr := formatValue(bar.Value, bar.Unit)
	valueRender := m.styles.Value.Render(valueStr)

	return labelStr + " " + fill + empty + " " + valueRender
}

func (m Model) renderTotal(total float64) string {
	label := m.styles.Label.Render("Total")
	valueStr := formatValue(total, "W")
	value := m.styles.Total.Render(valueStr)
	return label + " " + value
}

func (m Model) renderEmpty() string {
	return m.styles.Container.
		Width(m.width-4).
		Height(m.height-2).
		Align(lipgloss.Center, lipgloss.Center).
		Render("No devices online")
}

func (m Model) renderNoData() string {
	return m.styles.Container.
		Width(m.width-4).
		Height(m.height-2).
		Align(lipgloss.Center, lipgloss.Center).
		Render("No power data available")
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

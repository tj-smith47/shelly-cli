// Package energyhistory provides a sparkline-style energy history component.
package energyhistory

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/cache"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Sparkline characters for different heights (0-7).
var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// DataPoint represents a single historical data point.
type DataPoint struct {
	Value     float64
	Timestamp time.Time
}

// Model represents the energy history state.
type Model struct {
	cache    *cache.Cache
	mu       *sync.RWMutex
	history  map[string][]DataPoint // Device name -> history
	maxItems int                    // Max data points per device
	width    int
	height   int
	styles   Styles
}

// Styles for the energy history component.
type Styles struct {
	Container lipgloss.Style
	Header    lipgloss.Style
	Label     lipgloss.Style
	Sparkline lipgloss.Style
	Value     lipgloss.Style
	Time      lipgloss.Style
}

// DefaultStyles returns default styles for energy history.
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
		Sparkline: lipgloss.NewStyle().
			Foreground(colors.Secondary),
		Value: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),
		Time: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// New creates a new energy history component.
func New(c *cache.Cache) Model {
	return Model{
		cache:    c,
		mu:       &sync.RWMutex{},
		history:  make(map[string][]DataPoint),
		maxItems: 60, // 5 minutes at 5-second intervals
		styles:   DefaultStyles(),
	}
}

// Init initializes the energy history.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the energy history.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if updateMsg, ok := msg.(cache.DeviceUpdateMsg); ok {
		// Record power data when device is updated (for PM-capable devices)
		if updateMsg.Data != nil && updateMsg.Data.Online && hasPMCapability(updateMsg.Data) {
			m.addDataPoint(updateMsg.Name, updateMsg.Data.Power)
		}
	}
	return m, nil
}

// hasPMCapability checks if a device has power monitoring capability.
func hasPMCapability(d *cache.DeviceData) bool {
	// Check snapshot for actual PM components
	if d.Snapshot != nil && (len(d.Snapshot.PM) > 0 || len(d.Snapshot.EM) > 0 || len(d.Snapshot.EM1) > 0) {
		return true
	}

	// Fallback: check device model for PM capability
	model := d.Device.Type
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

func (m Model) addDataPoint(deviceName string, power float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	point := DataPoint{
		Value:     power,
		Timestamp: time.Now(),
	}

	history := m.history[deviceName]
	history = append(history, point)

	// Trim to max items
	if len(history) > m.maxItems {
		history = history[len(history)-m.maxItems:]
	}

	m.history[deviceName] = history
}

// View renders the energy history.
func (m Model) View() string {
	if m.cache == nil {
		return m.renderEmpty()
	}

	devices := m.cache.GetOnlineDevices()
	if len(devices) == 0 {
		return m.renderEmpty()
	}

	var content strings.Builder

	// Sparkline width - account for border (2) and padding
	sparkWidth := m.width - 42 // Label (16) + Value (12) + border/padding (14)
	if sparkWidth < 10 {
		sparkWidth = 10
	}
	if sparkWidth > m.maxItems {
		sparkWidth = m.maxItems
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	hasData := false
	for _, d := range devices {
		history := m.history[d.Device.Name]
		if len(history) > 0 {
			hasData = true
			content.WriteString(m.renderDeviceSparkline(d.Device.Name, history, sparkWidth) + "\n")
		}
	}

	if !hasData {
		return m.renderNoData()
	}

	// Use rendering package for consistent embedded title styling
	r := rendering.New(m.width, m.height).
		SetTitle("Energy History").
		SetBadge("5 min").
		SetFocused(false)

	return r.SetContent(content.String()).Render()
}

func (m Model) renderDeviceSparkline(name string, history []DataPoint, width int) string {
	// Truncate/pad label
	label := name
	if len(label) > 14 {
		label = label[:11] + "..."
	}
	labelStr := m.styles.Label.Render(label)

	// Generate sparkline
	sparkline := m.generateSparkline(history, width)
	sparkStr := m.styles.Sparkline.Render(sparkline)

	// Current value
	current := history[len(history)-1].Value
	valueStr := m.styles.Value.Render(formatValue(current, "W"))

	return labelStr + " " + sparkStr + " " + valueStr
}

func (m Model) generateSparkline(history []DataPoint, width int) string {
	if len(history) == 0 {
		return strings.Repeat(" ", width)
	}

	// Get the last 'width' points
	data := history
	if len(data) > width {
		data = data[len(data)-width:]
	}

	// Find min/max for scaling
	minVal, maxVal := data[0].Value, data[0].Value
	for _, d := range data {
		if d.Value < minVal {
			minVal = d.Value
		}
		if d.Value > maxVal {
			maxVal = d.Value
		}
	}

	// Avoid division by zero
	valRange := maxVal - minVal
	if valRange < 0.001 {
		valRange = 1
	}

	// Generate sparkline string
	var spark strings.Builder
	for _, d := range data {
		// Normalize to 0-7 range
		normalized := (d.Value - minVal) / valRange * 7
		idx := int(normalized)
		if idx > 7 {
			idx = 7
		}
		if idx < 0 {
			idx = 0
		}
		spark.WriteRune(sparkChars[idx])
	}

	// Pad if needed
	result := spark.String()
	if len(data) < width {
		result = strings.Repeat(" ", width-len(data)) + result
	}

	return result
}

func (m Model) renderEmpty() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Energy History").
		SetBadge("5 min").
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
		SetTitle("Energy History").
		SetBadge("5 min").
		SetFocused(false)
	centered := lipgloss.NewStyle().
		Width(m.width-4).
		Height(m.height-2).
		Align(lipgloss.Center, lipgloss.Center).
		Render("Collecting energy data...\nHistory will appear after a few updates.")
	return r.SetContent(centered).Render()
}

// formatValue formats a power value with appropriate units.
func formatValue(value float64, unit string) string {
	absVal := value
	if absVal < 0 {
		absVal = -absVal
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

// Clear clears all history data.
func (m Model) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.history)
}

// DeviceCount returns the number of devices with history.
func (m Model) DeviceCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.history)
}

// HistoryCount returns the total number of data points across all devices.
func (m Model) HistoryCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, h := range m.history {
		count += len(h)
	}
	return count
}

// Package energyhistory provides a sparkline-style energy history component.
package energyhistory

import (
	"cmp"
	"fmt"
	"slices"
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
	cache          *cache.Cache
	mu             *sync.RWMutex
	history        map[string][]DataPoint // Device name -> history
	maxItems       int                    // Max data points per device
	lastCollection time.Time              // Throttle data collection
	width          int
	height         int
	styles         Styles
	focused        bool
	panelIndex     int // For Shift+N hint
}

// Styles for the energy history component.
type Styles struct {
	Container     lipgloss.Style
	Header        lipgloss.Style
	Label         lipgloss.Style
	Sparkline     lipgloss.Style   // Fallback style (unused with gradient)
	SparkGradient [8]lipgloss.Style // Gradient colors for levels 0-7
	Value         lipgloss.Style
	Time          lipgloss.Style
}

// DefaultStyles returns default styles for energy history.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()

	// Gradient from cool (low) to warm (high): blue -> cyan -> green -> yellow -> orange -> red
	gradient := [8]lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#5c7cfa")), // 0: Blue (lowest)
		lipgloss.NewStyle().Foreground(lipgloss.Color("#22b8cf")), // 1: Cyan
		lipgloss.NewStyle().Foreground(lipgloss.Color("#20c997")), // 2: Teal
		lipgloss.NewStyle().Foreground(lipgloss.Color("#51cf66")), // 3: Green
		lipgloss.NewStyle().Foreground(lipgloss.Color("#94d82d")), // 4: Lime
		lipgloss.NewStyle().Foreground(lipgloss.Color("#fcc419")), // 5: Yellow
		lipgloss.NewStyle().Foreground(lipgloss.Color("#ff922b")), // 6: Orange
		lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b6b")), // 7: Red (highest)
	}

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
		SparkGradient: gradient,
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
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the energy history.
// Note: We only collect data via collectCurrentPower() at 5-second intervals
// to maintain the 5-minute window (60 points * 5 seconds = 5 minutes).
// We don't record on every DeviceUpdateMsg because cache updates can be more frequent.
func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// We intentionally don't record on DeviceUpdateMsg - see collectCurrentPower()
	_ = msg
	return *m, nil
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

func (m *Model) addDataPoint(deviceName string, power float64) {
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
func (m *Model) View() string {
	if m.cache == nil {
		return m.renderEmpty()
	}

	devices := m.cache.GetOnlineDevices()
	if len(devices) == 0 {
		return m.renderEmpty()
	}

	// Collect current power data from online devices with PM capability
	// This ensures we always record data on each render cycle
	m.collectCurrentPower(devices)

	var content strings.Builder

	// Calculate label width (same logic as renderDeviceSparkline)
	labelWidth := 16
	if m.width < 60 {
		labelWidth = 12
	}

	// Sparkline width calculation:
	// - Borders: 2 (left + right)
	// - Horizontal padding: 2 (1 each side inside border)
	// - Label: labelWidth (16 or 12)
	// - Spaces: 2 (after label, after sparkline)
	// - Value: 10
	// Total overhead = 2 + 2 + labelWidth + 2 + 10 = labelWidth + 16
	sparkWidth := m.width - labelWidth - 16
	if sparkWidth < 10 {
		sparkWidth = 10
	}
	// Don't cap at maxItems - generateSparkline will pad with low bars if we don't have enough data

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Sort devices by power (highest first) to match power consumption order
	slices.SortFunc(devices, func(a, b *cache.DeviceData) int {
		return cmp.Compare(b.Power, a.Power) // Descending
	})

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

	// Build gradient legend showing color scale
	legend := "Legend: " +
		m.styles.SparkGradient[0].Render("▁") +
		m.styles.SparkGradient[3].Render("▄") +
		m.styles.SparkGradient[7].Render("█")

	// Use rendering package for consistent embedded title styling
	r := rendering.New(m.width, m.height).
		SetTitle("Energy History").
		SetBadge(legend).
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex).
		SetFooter("5m·····now")

	// Use yellow borders for energy panels
	if m.focused {
		r.SetFocusColor(theme.Yellow())
	} else {
		r.SetBlurColor(theme.Yellow())
	}

	return r.SetContent(content.String()).Render()
}

// collectCurrentPower records power data from online devices with PM capability.
// Called during View() to ensure data is always collected on each render.
// Throttled to collect at most once every 5 seconds.
func (m *Model) collectCurrentPower(devices []*cache.DeviceData) {
	// Throttle: only collect every 5 seconds
	if time.Since(m.lastCollection) < 5*time.Second {
		return
	}
	m.lastCollection = time.Now()

	for _, d := range devices {
		if d.Online && hasPMCapability(d) {
			m.addDataPoint(d.Device.Name, d.Power)
		}
	}
}

func (m *Model) renderDeviceSparkline(name string, history []DataPoint, width int) string {
	// Label width based on available space (max 16, min truncated)
	labelWidth := 16
	if m.width < 60 {
		labelWidth = 12
	}

	// Truncate/pad label
	label := name
	maxLabel := labelWidth - 2 // Leave room for spacing
	if len(label) > maxLabel {
		label = label[:maxLabel-3] + "..."
	}
	labelStr := m.styles.Label.Width(labelWidth).Render(label)

	// Generate sparkline (already styled with gradient colors)
	sparkStr := m.generateSparkline(history, width)

	// Current value (right-aligned)
	current := history[len(history)-1].Value
	valueStr := m.styles.Value.Width(10).Align(lipgloss.Right).Render(formatValue(current, "W"))

	return labelStr + " " + sparkStr + " " + valueStr
}

func (m *Model) generateSparkline(history []DataPoint, width int) string {
	if len(history) == 0 {
		// Use lowest bar char for empty data (shows "no data" state)
		// Apply lowest gradient color for consistency
		lowestChar := m.styles.SparkGradient[0].Render(string(sparkChars[0]))
		var spark strings.Builder
		for range width {
			spark.WriteString(lowestChar)
		}
		return spark.String()
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

	// Scale calculation
	valRange := maxVal - minVal
	flatLine := false
	flatLevel := 0
	if valRange < 0.001 {
		// All values are the same - use appropriate fixed level
		flatLine = true
		if maxVal < 1.0 {
			flatLevel = 0 // Near zero: show blue (lowest)
		} else {
			flatLevel = 4 // Non-zero stable: show middle (lime)
		}
	}

	// Generate sparkline with gradient colors
	var spark strings.Builder

	// Pad at the start with lowest bar char if we don't have enough data
	if len(data) < width {
		padLevel := 0
		if flatLine {
			padLevel = flatLevel
		}
		padChar := m.styles.SparkGradient[padLevel].Render(string(sparkChars[padLevel]))
		for range width - len(data) {
			spark.WriteString(padChar)
		}
	}

	for _, d := range data {
		var idx int
		if flatLine {
			// All values same - use fixed level
			idx = flatLevel
		} else {
			// Normalize to 0-7 range
			normalized := (d.Value - minVal) / valRange * 7
			idx = int(normalized)
			if idx > 7 {
				idx = 7
			}
			if idx < 0 {
				idx = 0
			}
		}
		// Apply gradient color based on level
		spark.WriteString(m.styles.SparkGradient[idx].Render(string(sparkChars[idx])))
	}

	return spark.String()
}

func (m *Model) renderEmpty() string {
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

func (m *Model) renderNoData() string {
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
func (m *Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return *m
}

// Clear clears all history data.
func (m *Model) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.history)
}

// DeviceCount returns the number of devices with history.
func (m *Model) DeviceCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.history)
}

// HistoryCount returns the total number of data points across all devices.
func (m *Model) HistoryCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, h := range m.history {
		count += len(h)
	}
	return count
}

// SetFocused sets whether this panel has focus.
func (m *Model) SetFocused(focused bool) Model {
	m.focused = focused
	return *m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m *Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return *m
}

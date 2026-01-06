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
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
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
	loading        bool
	loader         loading.Model
}

// Styles for the energy history component.
type Styles struct {
	Container     lipgloss.Style
	Header        lipgloss.Style
	Label         lipgloss.Style
	Sparkline     lipgloss.Style    // Fallback style (unused with gradient)
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
		Container: styles.PanelBorder().Padding(1, 2),
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
		loading:  true, // Start in loading state until first data point
		loader: loading.New(
			loading.WithMessage("Collecting energy history..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
	}
}

// Init initializes the energy history.
func (m *Model) Init() tea.Cmd {
	if m.loading {
		return m.loader.Tick()
	}
	return nil
}

// Update handles messages for the energy history.
// Collects data from cache on DeviceUpdateMsg to ensure history is populated.
func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Forward tick messages to loader when loading
	if m.loading {
		result := generics.UpdateLoader(m.loader, msg, func(msg tea.Msg) bool {
			switch msg.(type) {
			case cache.DeviceUpdateMsg, cache.AllDevicesLoadedMsg:
				return true
			}
			return false
		})
		m.loader = result.Loader
		if result.Consumed {
			return *m, result.Cmd
		}
	}

	// Collect data on device updates - this is how history gets populated
	switch msg.(type) {
	case cache.DeviceUpdateMsg, cache.AllDevicesLoadedMsg:
		m.handleDeviceUpdate()
	}

	return *m, nil
}

// handleDeviceUpdate collects power data when device updates arrive.
func (m *Model) handleDeviceUpdate() {
	if m.cache == nil {
		return
	}

	devices := m.cache.GetOnlineDevices()
	if len(devices) == 0 {
		return
	}

	m.collectCurrentPower(devices)

	// Exit loading once we have data
	if !m.loading {
		return
	}

	m.mu.RLock()
	hasData := len(m.history) > 0
	m.mu.RUnlock()

	if hasData {
		m.loading = false
	}
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
	debug.TraceLock("energyhistory", "Lock", "addDataPoint:"+deviceName)
	m.mu.Lock()
	defer func() {
		m.mu.Unlock()
		debug.TraceUnlock("energyhistory", "Lock", "addDataPoint:"+deviceName)
	}()

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
	// Show loading indicator during initial data collection
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

	var content strings.Builder

	// Calculate max label width from actual device names - never truncate
	maxLabelLen := 0
	for _, d := range devices {
		if len(d.Device.Name) > maxLabelLen {
			maxLabelLen = len(d.Device.Name)
		}
	}
	labelWidth := max(10, maxLabelLen+1) // +1 for spacing

	// Sparkline width calculation (shrinks to accommodate full names):
	// - Borders: 2 (left + right)
	// - Horizontal padding: 2 (1 each side inside border)
	// - Label: labelWidth (dynamic)
	// - Spaces: 2 (after label, after sparkline)
	// - Value: 10
	// Total overhead = 2 + 2 + labelWidth + 2 + 10 = 16 + labelWidth
	sparkWidth := m.width - 16 - labelWidth
	if sparkWidth < 10 {
		sparkWidth = 10 // Absolute minimum sparkline width
	}

	debug.TraceLock("energyhistory", "RLock", "View")
	m.mu.RLock()
	defer func() {
		m.mu.RUnlock()
		debug.TraceUnlock("energyhistory", "RLock", "View")
	}()

	// Sort devices by power (highest first) to match power consumption order
	slices.SortFunc(devices, func(a, b *cache.DeviceData) int {
		return cmp.Compare(b.Power, a.Power) // Descending
	})

	hasData := false
	for _, d := range devices {
		history := m.history[d.Device.Name]
		if len(history) > 0 {
			hasData = true
			content.WriteString(m.renderDeviceSparkline(d.Device.Name, history, sparkWidth, labelWidth) + "\n")
		}
	}

	if !hasData {
		return m.renderNoData()
	}

	// Build legend: text in border color (yellow), Unicode chars in gradient colors
	borderStyle := lipgloss.NewStyle().Foreground(theme.Yellow())
	legend := borderStyle.Render("Legend:") + " " +
		m.styles.SparkGradient[0].Render("▁") + borderStyle.Render(" low ") +
		m.styles.SparkGradient[3].Render("▄") + borderStyle.Render(" mid ") +
		m.styles.SparkGradient[7].Render("█") + borderStyle.Render(" high")

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

func (m *Model) renderDeviceSparkline(name string, history []DataPoint, sparkWidth, labelWidth int) string {
	// Use full name with dynamic label width - no truncation
	labelStr := m.styles.Label.Width(labelWidth).Render(name)

	// Generate sparkline (already styled with gradient colors)
	sparkStr := m.generateSparkline(history, sparkWidth)

	// Current value (right-aligned)
	current := history[len(history)-1].Value
	valueStr := m.styles.Value.Width(10).Align(lipgloss.Right).Render(formatValue(current, "W"))

	return labelStr + " " + sparkStr + " " + valueStr
}

// sparklineParams holds sparkline generation parameters.
type sparklineParams struct {
	minVal, maxVal float64
	flatLine       bool
	flatLevel      int
}

// computeSparklineParams calculates min/max and flat line detection.
func computeSparklineParams(data []DataPoint) sparklineParams {
	minVal, maxVal := data[0].Value, data[0].Value
	for _, d := range data {
		minVal = min(minVal, d.Value)
		maxVal = max(maxVal, d.Value)
	}

	p := sparklineParams{minVal: minVal, maxVal: maxVal}
	valRange := maxVal - minVal
	if valRange < 0.001 {
		p.flatLine = true
		if maxVal < 1.0 {
			p.flatLevel = 0 // Near zero: show blue (lowest)
		} else {
			p.flatLevel = 4 // Non-zero stable: show middle (lime)
		}
	}
	return p
}

// normalizeToLevel converts a value to a sparkline level (0-7).
func normalizeToLevel(value float64, p sparklineParams) int {
	if p.flatLine {
		return p.flatLevel
	}
	valRange := p.maxVal - p.minVal
	normalized := (value - p.minVal) / valRange * 7
	return max(0, min(7, int(normalized)))
}

func (m *Model) generateSparkline(history []DataPoint, width int) string {
	if len(history) == 0 {
		lowestChar := m.styles.SparkGradient[0].Render(string(sparkChars[0]))
		return strings.Repeat(lowestChar, width)
	}

	// Scale data to fit width exactly - interpolates/compresses to fill full sparkline
	data := scaleDataToWidth(history, width)

	p := computeSparklineParams(data)

	var spark strings.Builder
	for _, d := range data {
		idx := normalizeToLevel(d.Value, p)
		spark.WriteString(m.styles.SparkGradient[idx].Render(string(sparkChars[idx])))
	}

	return spark.String()
}

// scaleDataToWidth scales the data points to fit the target width exactly.
// If we have more points than width, we compress by averaging groups.
// If we have fewer points than width, we stretch by interpolating between points.
// This ensures the sparkline is always fully filled with no dead padding.
func scaleDataToWidth(history []DataPoint, width int) []DataPoint {
	histLen := len(history)
	if histLen == 0 {
		return history
	}
	if histLen == width {
		return history
	}
	if histLen > width {
		return scaleDown(history, width)
	}
	return scaleUp(history, width)
}

// scaleDown compresses more data points into fewer by averaging groups.
func scaleDown(history []DataPoint, width int) []DataPoint {
	histLen := len(history)
	result := make([]DataPoint, width)
	ratio := float64(histLen) / float64(width)

	for i := range width {
		startIdx := int(float64(i) * ratio)
		endIdx := int(float64(i+1) * ratio)
		if endIdx > histLen {
			endIdx = histLen
		}
		if startIdx >= endIdx {
			startIdx = endIdx - 1
		}

		// Average the values in this bucket
		var sum float64
		count := 0
		for j := startIdx; j < endIdx; j++ {
			sum += history[j].Value
			count++
		}

		result[i] = DataPoint{
			Value:     sum / float64(count),
			Timestamp: history[endIdx-1].Timestamp,
		}
	}
	return result
}

// scaleUp stretches fewer data points to fill more width by interpolating.
func scaleUp(history []DataPoint, width int) []DataPoint {
	histLen := len(history)
	result := make([]DataPoint, width)

	for i := range width {
		// Map output position to source position
		srcPos := float64(i) * float64(histLen-1) / float64(width-1)
		lowIdx := int(srcPos)
		highIdx := lowIdx + 1

		if highIdx >= histLen {
			// At or beyond the last point
			result[i] = history[histLen-1]
			continue
		}

		// Linear interpolation between adjacent points
		frac := srcPos - float64(lowIdx)
		value := history[lowIdx].Value*(1-frac) + history[highIdx].Value*frac

		result[i] = DataPoint{
			Value:     value,
			Timestamp: history[highIdx].Timestamp,
		}
	}
	return result
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

func (m *Model) renderLoading() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Energy History").
		SetBadge("5 min").
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
	m.loader = helpers.SetLoaderSize(m.loader, width, height)
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

// SetLoading sets the loading state.
func (m *Model) SetLoading(isLoading bool) Model {
	m.loading = isLoading
	return *m
}

// StartLoading sets loading to true and returns a tick command.
func (m *Model) StartLoading() (Model, tea.Cmd) {
	m.loading = true
	return *m, m.loader.Tick()
}

// IsLoading returns whether the component is in loading state.
func (m *Model) IsLoading() bool {
	return m.loading
}

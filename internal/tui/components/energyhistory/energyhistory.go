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
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
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

// switchID uniquely identifies a switch for map indexing.
// This is the actual map key - only includes fields needed for uniqueness.
type switchID struct {
	DeviceName string
	SwitchIdx  int
}

// SwitchKey holds display information for a switch.
// This is NOT used as a map key - only for display purposes.
type SwitchKey struct {
	DeviceName  string
	SwitchID    int
	SwitchName  string // For display label (user-configured name)
	SwitchCount int    // Total number of switches on device (for label formatting)
}

// String returns the display label for this switch.
// Shows "(Sw0)" format when switch name is empty but device has multiple switches.
// If switch name already starts with device name, uses switch name only to avoid duplication.
func (k SwitchKey) String() string {
	if k.SwitchCount <= 1 {
		return k.DeviceName
	}
	if k.SwitchName == "" {
		return fmt.Sprintf("%s (Sw%d)", k.DeviceName, k.SwitchID)
	}
	// Avoid duplicating device name if switch name already includes it
	if strings.HasPrefix(strings.ToLower(k.SwitchName), strings.ToLower(k.DeviceName)) {
		return k.SwitchName
	}
	return k.DeviceName + " " + k.SwitchName
}

// toID returns the map key for this switch.
func (k SwitchKey) toID() switchID {
	return switchID{DeviceName: k.DeviceName, SwitchIdx: k.SwitchID}
}

// Model represents the energy history state.
type Model struct {
	helpers.Sizable
	cache          *cache.Cache
	mu             *sync.RWMutex
	history        map[switchID][]DataPoint // Per-switch history keyed by stable ID
	displayInfo    map[switchID]SwitchKey   // Display info for each switch (updated on each collection)
	hasInitialData map[switchID]bool        // Tracks switches that have initial data point
	maxItems       int                      // Max data points per switch
	lastCollection time.Time                // Throttle data collection
	scroller       *panel.Scroller          // For scrolling when many switches
	styles         Styles
	focused        bool
	panelIndex     int // For Shift+N hint
	loading        bool
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
	m := Model{
		Sizable:        helpers.NewSizableLoaderOnly(),
		cache:          c,
		mu:             &sync.RWMutex{},
		history:        make(map[switchID][]DataPoint),
		displayInfo:    make(map[switchID]SwitchKey),
		hasInitialData: make(map[switchID]bool),
		maxItems:       60,                       // 5 minutes at 5-second intervals
		scroller:       panel.NewScroller(0, 10), // Will be updated with actual counts
		styles:         DefaultStyles(),
		loading:        true, // Start in loading state until first data point
	}
	m.Loader = m.Loader.SetMessage("Collecting energy history...")
	return m
}

// Init initializes the energy history.
func (m *Model) Init() tea.Cmd {
	if m.loading {
		return m.Loader.Tick()
	}
	return nil
}

// Update handles messages for the energy history.
// Collects data from cache on DeviceUpdateMsg to ensure history is populated.
func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Handle navigation messages
	if navMsg, ok := msg.(messages.NavigationMsg); ok {
		m.handleNavigation(navMsg)
		return *m, nil
	}

	// Forward tick messages to loader when loading
	if m.loading {
		result := generics.UpdateLoader(m.Loader, msg, func(msg tea.Msg) bool {
			switch msg.(type) {
			case cache.DeviceUpdateMsg, cache.AllDevicesLoadedMsg:
				return true
			}
			return false
		})
		m.Loader = result.Loader
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

// handleNavigation handles NavigationMsg for scrolling.
func (m *Model) handleNavigation(msg messages.NavigationMsg) {
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
		// Not applicable for single-column scrolling
	}
}

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
	name := d.Device.Name

	// Check snapshot for dedicated PM/EM components
	if d.Snapshot != nil && (len(d.Snapshot.PM) > 0 || len(d.Snapshot.EM) > 0 || len(d.Snapshot.EM1) > 0) {
		debug.TraceEvent("energy: %s has PM (snapshot PM/EM components)", name)
		return true
	}

	// Check for switch-integrated power monitoring (e.g., Plus 2PM)
	// Devices like Plus 2PM have PM in the switch component (apower field),
	// not dedicated PM components. SwitchPowers is populated when switches report power.
	if len(d.SwitchPowers) > 0 {
		debug.TraceEvent("energy: %s has PM (switch-integrated, %d switches with power)", name, len(d.SwitchPowers))
		return true
	}

	// Also check cover-integrated power monitoring
	if len(d.CoverPowers) > 0 {
		debug.TraceEvent("energy: %s has PM (cover-integrated, %d covers with power)", name, len(d.CoverPowers))
		return true
	}

	// Fallback: check device model for PM capability
	model := d.Device.Type
	if model == "" {
		model = d.Device.Model
	}

	hasPM := modelHasPM(model)
	if hasPM {
		debug.TraceEvent("energy: %s has PM (model pattern: %s)", name, model)
	}
	return hasPM
}

// modelHasPM checks if a device model code or display name indicates power monitoring capability.
func modelHasPM(model string) bool {
	// Gen1 PM devices: SHSW-PM (type code format)
	if strings.Contains(model, "-PM") {
		return true
	}

	// Display names like "Shelly Plus 1PM", "Shelly Plus 2PM", "Shelly 1PM Gen3"
	// These have "PM" suffix or " PM" pattern (space before PM)
	if strings.HasSuffix(model, "PM") || strings.Contains(model, " PM") {
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

func (m *Model) addDataPoint(key SwitchKey, power float64) {
	id := key.toID()
	debug.TraceLock("energyhistory", "Lock", "addDataPoint:"+key.String())
	m.mu.Lock()
	defer func() {
		m.mu.Unlock()
		debug.TraceUnlock("energyhistory", "Lock", "addDataPoint:"+key.String())
	}()

	point := DataPoint{
		Value:     power,
		Timestamp: time.Now(),
	}

	history := m.history[id]
	history = append(history, point)

	// Trim to max items
	if len(history) > m.maxItems {
		history = history[len(history)-m.maxItems:]
	}

	m.history[id] = history
	// Update display info (may change as switch names/counts are discovered)
	m.displayInfo[id] = key
}

// historyEntry holds a switch key and its history for sorting/rendering.
type historyEntry struct {
	key     SwitchKey
	history []DataPoint
	power   float64 // Current power (latest value) for sorting
}

// View renders the energy history.
func (m *Model) View() string {
	// Handle special states first
	if result, handled := m.handleSpecialStates(); handled {
		return result
	}

	// Get sorted entries
	entries := m.getSortedEntries()
	if len(entries) == 0 {
		return m.renderNoData()
	}

	// Setup scroller
	visibleRows := max(1, m.Height-4)
	m.scroller.SetItemCount(len(entries))
	m.scroller.SetVisibleRows(visibleRows)

	// Calculate layout and render
	labelWidth, sparkWidth := m.calculateLayout(entries)
	content := m.renderVisibleEntries(entries, labelWidth, sparkWidth)

	return m.buildPanel(content, len(entries))
}

// handleSpecialStates checks for loading or empty states.
func (m *Model) handleSpecialStates() (string, bool) {
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

// getSortedEntries returns history entries sorted by power descending.
func (m *Model) getSortedEntries() []historyEntry {
	debug.TraceLock("energyhistory", "RLock", "View")
	m.mu.RLock()
	var entries []historyEntry
	for id, history := range m.history {
		if len(history) > 0 {
			// Get display info for this switch ID
			displayKey := m.displayInfo[id]
			entries = append(entries, historyEntry{
				key:     displayKey,
				history: history,
				power:   history[len(history)-1].Value,
			})
		}
	}
	m.mu.RUnlock()
	debug.TraceUnlock("energyhistory", "RLock", "View")

	// Sort by power descending, then by key label for stable ordering when power values are equal
	slices.SortFunc(entries, func(a, b historyEntry) int {
		if c := cmp.Compare(b.power, a.power); c != 0 {
			return c
		}
		return cmp.Compare(a.key.String(), b.key.String()) // Alphabetical for stability
	})
	return entries
}

// calculateLayout computes label and sparkline widths.
func (m *Model) calculateLayout(entries []historyEntry) (labelWidth, sparkWidth int) {
	maxLabelLen := 0
	for _, e := range entries {
		maxLabelLen = max(maxLabelLen, len(e.key.String()))
	}
	labelWidth = max(10, maxLabelLen+1)
	sparkWidth = max(10, m.Width-16-labelWidth)
	return labelWidth, sparkWidth
}

// renderVisibleEntries renders only the visible portion of entries.
func (m *Model) renderVisibleEntries(entries []historyEntry, labelWidth, sparkWidth int) string {
	var content strings.Builder
	start, end := m.scroller.VisibleRange()
	for i := start; i < end && i < len(entries); i++ {
		e := entries[i]
		content.WriteString(m.renderDeviceSparkline(e.key.String(), e.history, sparkWidth, labelWidth) + "\n")
	}
	return content.String()
}

// buildPanel creates the final rendered panel with badge.
func (m *Model) buildPanel(content string, entryCount int) string {
	borderStyle := lipgloss.NewStyle().Foreground(theme.Yellow())
	countInfo := fmt.Sprintf("%d switches", entryCount)
	if m.scroller.HasMore() || m.scroller.HasPrevious() {
		countInfo = m.scroller.ScrollInfoRange()
	}
	legend := borderStyle.Render(countInfo+" │ ") +
		m.styles.SparkGradient[0].Render("▁") + borderStyle.Render(" low ") +
		m.styles.SparkGradient[3].Render("▄") + borderStyle.Render(" mid ") +
		m.styles.SparkGradient[7].Render("█") + borderStyle.Render(" high")

	r := rendering.New(m.Width, m.Height).
		SetTitle("Energy History").
		SetBadge(legend).
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex).
		SetFooter("5m·····now")

	m.applyBorderColor(r)
	return r.SetContent(content).Render()
}

// applyBorderColor sets the appropriate border color for the panel.
// Uses the default blue focus color (from renderer) and yellow blur color.
func (m *Model) applyBorderColor(r *rendering.Renderer) {
	r.SetBlurColor(theme.Yellow())
}

// collectCurrentPower records power data per-switch from online devices with PM capability.
// Called during View() to ensure data is always collected on each render.
// Throttled to collect at most once every 5 seconds, except for initial data capture
// which happens immediately to prevent false ramp-up from 0.
func (m *Model) collectCurrentPower(devices []*cache.DeviceData) {
	throttleActive := time.Since(m.lastCollection) < 5*time.Second

	for _, d := range devices {
		if !d.Online || !hasPMCapability(d) {
			continue
		}

		deviceName := d.Device.Name

		// Build switch name lookup
		switchNames := make(map[int]string)
		for _, sw := range d.Switches {
			switchNames[sw.ID] = sw.Name
		}

		// Collect per-switch power data using switch for clarity
		switch {
		case len(d.SwitchPowers) > 0:
			switchCount := len(d.SwitchPowers)
			for switchID, power := range d.SwitchPowers {
				key := SwitchKey{
					DeviceName:  deviceName,
					SwitchID:    switchID,
					SwitchName:  switchNames[switchID],
					SwitchCount: switchCount,
				}
				m.collectDataForKey(key, power, throttleActive)
			}
		case len(d.PMPowers) > 0:
			pmCount := len(d.PMPowers)
			for pmID, power := range d.PMPowers {
				key := SwitchKey{
					DeviceName:  deviceName,
					SwitchID:    pmID,
					SwitchName:  fmt.Sprintf("PM%d", pmID),
					SwitchCount: pmCount,
				}
				m.collectDataForKey(key, power, throttleActive)
			}
		case len(d.EMPowers) > 0 || len(d.EM1Powers) > 0:
			emCount := len(d.EMPowers) + len(d.EM1Powers)
			for emID, power := range d.EMPowers {
				key := SwitchKey{
					DeviceName:  deviceName,
					SwitchID:    emID,
					SwitchName:  fmt.Sprintf("EM%d", emID),
					SwitchCount: emCount,
				}
				m.collectDataForKey(key, power, throttleActive)
			}
			for em1ID, power := range d.EM1Powers {
				key := SwitchKey{
					DeviceName:  deviceName,
					SwitchID:    em1ID,
					SwitchName:  fmt.Sprintf("EM1:%d", em1ID),
					SwitchCount: emCount,
				}
				m.collectDataForKey(key, power, throttleActive)
			}
		default:
			// Fallback: single aggregated power value
			key := SwitchKey{
				DeviceName:  deviceName,
				SwitchID:    0,
				SwitchName:  "",
				SwitchCount: 1,
			}
			m.collectDataForKey(key, d.Power, throttleActive)
		}
	}

	// Update throttle timestamp when not throttled
	if !throttleActive {
		m.lastCollection = time.Now()
	}
}

// collectDataForKey adds a data point for a switch key, respecting throttle and initial capture.
func (m *Model) collectDataForKey(key SwitchKey, power float64, throttleActive bool) {
	id := key.toID()
	m.mu.RLock()
	hasInitial := m.hasInitialData[id]
	m.mu.RUnlock()

	if !hasInitial {
		// First data point for this switch - capture immediately to prevent
		// false ramp-up from 0 when device was already ON before TUI started
		debug.TraceEvent("energy: %s initial capture: %.1fW (bypassing throttle)", key.String(), power)
		m.addDataPoint(key, power)
		m.mu.Lock()
		m.hasInitialData[id] = true
		m.mu.Unlock()
	} else if !throttleActive {
		// Subsequent data points respect the throttle
		m.addDataPoint(key, power)
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
	r := rendering.New(m.Width, m.Height).
		SetTitle("Energy History").
		SetBadge("5 min").
		SetFocused(false)
	centered := lipgloss.NewStyle().
		Width(m.Width-4).
		Height(m.Height-2).
		Align(lipgloss.Center, lipgloss.Center).
		Render("No devices online")
	return r.SetContent(centered).Render()
}

func (m *Model) renderNoData() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Energy History").
		SetBadge("5 min").
		SetFocused(false)
	centered := lipgloss.NewStyle().
		Width(m.Width-4).
		Height(m.Height-2).
		Align(lipgloss.Center, lipgloss.Center).
		Render("Collecting energy data...\nHistory will appear after a few updates.")
	return r.SetContent(centered).Render()
}

func (m *Model) renderLoading() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Energy History").
		SetBadge("5 min").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Use yellow blur border for energy panels (focus uses default blue)
	r.SetBlurColor(theme.Yellow())

	return r.SetContent(m.Loader.View()).Render()
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
	m.ApplySize(width, height)
	return *m
}

// Clear clears all history data.
func (m *Model) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.history)
	clear(m.displayInfo)
	clear(m.hasInitialData)
}

// DeviceCount returns the number of switches with history (renamed from devices).
func (m *Model) DeviceCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.history)
}

// SwitchCount returns the number of switches with history.
func (m *Model) SwitchCount() int {
	return m.DeviceCount()
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
	return *m, m.Loader.Tick()
}

// IsLoading returns whether the component is in loading state.
func (m *Model) IsLoading() bool {
	return m.loading
}

// ScrollUp scrolls the list up by one item.
func (m *Model) ScrollUp() Model {
	m.scroller.CursorUp()
	return *m
}

// ScrollDown scrolls the list down by one item.
func (m *Model) ScrollDown() Model {
	m.scroller.CursorDown()
	return *m
}

// PageUp scrolls up by one page.
func (m *Model) PageUp() Model {
	m.scroller.PageUp()
	return *m
}

// PageDown scrolls down by one page.
func (m *Model) PageDown() Model {
	m.scroller.PageDown()
	return *m
}

// CanScroll returns true if there are more items than visible rows.
func (m *Model) CanScroll() bool {
	return m.scroller.HasMore() || m.scroller.HasPrevious()
}

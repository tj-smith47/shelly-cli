package monitor

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// DeviceSensorData holds all sensor readings for a single device.
type DeviceSensorData struct {
	DeviceName  string
	SensorData  *model.SensorData
	HasExtPower bool // Whether external power is present (from DevicePower)
}

// EnvironmentModel displays environment sensor readings and safety alerts.
type EnvironmentModel struct {
	panel.Sizable
	devices  []DeviceSensorData
	focused  bool
	panelIdx int
	styles   EnvironmentStyles
}

// EnvironmentStyles holds styles for the environment panel.
type EnvironmentStyles struct {
	SectionHeader lipgloss.Style
	DeviceName    lipgloss.Style
	Value         lipgloss.Style
	Muted         lipgloss.Style
	Selected      lipgloss.Style

	// Temperature color ranges
	TempCold lipgloss.Style // <15°C (blue)
	TempOK   lipgloss.Style // 15-25°C (green)
	TempWarm lipgloss.Style // 25-35°C (orange)
	TempHot  lipgloss.Style // >35°C (red)

	// Humidity color ranges
	HumidOK      lipgloss.Style // 30-60% (green)
	HumidCaution lipgloss.Style // 20-30% or 60-80% (yellow)
	HumidBad     lipgloss.Style // <20% or >80% (red)

	// Battery color ranges
	BatteryGood lipgloss.Style // >40% (green)
	BatteryLow  lipgloss.Style // 20-40% (yellow)
	BatteryCrit lipgloss.Style // <20% (red)
	ExtPower    lipgloss.Style // External power indicator

	// Safety states
	AlarmOK    lipgloss.Style
	AlarmFire  lipgloss.Style // Red background for alarms
	AlarmMuted lipgloss.Style
}

// Temperature thresholds for color-coding.
const (
	tempColdThreshold = 15.0
	tempWarmThreshold = 25.0
	tempHotThreshold  = 35.0
)

// Humidity thresholds for color-coding.
const (
	humidLowBad     = 20.0
	humidLowCaution = 30.0
	humidHighOK     = 60.0
	humidHighBad    = 80.0
)

// Battery thresholds for color-coding.
const (
	batteryCritThreshold = 20
	batteryLowThreshold  = 40
)

// defaultEnvironmentStyles returns default styles for the environment panel.
func defaultEnvironmentStyles() EnvironmentStyles {
	colors := theme.GetSemanticColors()
	return EnvironmentStyles{
		SectionHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight),
		DeviceName: lipgloss.NewStyle().
			Foreground(colors.Text),
		Value: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Text),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Bold(true),

		// Temperature
		TempCold: lipgloss.NewStyle().Bold(true).Foreground(colors.Highlight), // blue
		TempOK:   lipgloss.NewStyle().Bold(true).Foreground(colors.Success),   // green
		TempWarm: lipgloss.NewStyle().Bold(true).Foreground(colors.Warning),   // orange
		TempHot:  lipgloss.NewStyle().Bold(true).Foreground(colors.Error),     // red

		// Humidity
		HumidOK:      lipgloss.NewStyle().Bold(true).Foreground(colors.Success), // green
		HumidCaution: lipgloss.NewStyle().Bold(true).Foreground(colors.Warning), // yellow
		HumidBad:     lipgloss.NewStyle().Bold(true).Foreground(colors.Error),   // red

		// Battery
		BatteryGood: lipgloss.NewStyle().Bold(true).Foreground(colors.Success), // green
		BatteryLow:  lipgloss.NewStyle().Bold(true).Foreground(colors.Warning), // yellow
		BatteryCrit: lipgloss.NewStyle().Bold(true).Foreground(colors.Error),   // red
		ExtPower:    lipgloss.NewStyle().Foreground(colors.Online),

		// Safety
		AlarmOK: lipgloss.NewStyle().Foreground(colors.Success),
		AlarmFire: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(colors.Error),
		AlarmMuted: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Warning),
	}
}

// NewEnvironment creates a new environment panel model.
func NewEnvironment() EnvironmentModel {
	return EnvironmentModel{
		Sizable: panel.NewSizable(5, panel.NewScroller(0, 10)),
		styles:  defaultEnvironmentStyles(),
	}
}

// Init returns the initial command.
func (m EnvironmentModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the environment panel.
func (m EnvironmentModel) Update(msg tea.Msg) (EnvironmentModel, tea.Cmd) {
	if navMsg, ok := msg.(messages.NavigationMsg); ok {
		return m.handleNavigation(navMsg), nil
	}
	return m, nil
}

// handleNavigation handles scrolling.
func (m EnvironmentModel) handleNavigation(msg messages.NavigationMsg) EnvironmentModel {
	m.Scroller.HandleNavigation(msg)
	return m
}

// SetDevices updates the environment panel with device sensor data.
func (m EnvironmentModel) SetDevices(statuses []DeviceStatus) EnvironmentModel {
	devices := make([]DeviceSensorData, 0, len(statuses))
	for _, s := range statuses {
		if !s.Online || s.Sensors == nil {
			continue
		}
		hasExtPower := false
		for _, dp := range s.Sensors.DevicePower {
			if dp.External.Present {
				hasExtPower = true
				break
			}
		}
		devices = append(devices, DeviceSensorData{
			DeviceName:  s.Name,
			SensorData:  s.Sensors,
			HasExtPower: hasExtPower,
		})
	}
	m.devices = devices

	// Count total display lines for scroller
	lines := m.countDisplayLines()
	m.Scroller.SetItemCount(lines)
	return m
}

// SetFocused sets whether this panel is focused.
func (m EnvironmentModel) SetFocused(focused bool) EnvironmentModel {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hints.
func (m EnvironmentModel) SetPanelIndex(idx int) EnvironmentModel {
	m.panelIdx = idx
	return m
}

// SetSize updates the component dimensions.
func (m EnvironmentModel) SetSize(width, height int) EnvironmentModel {
	m.ApplySize(width, height)
	return m
}

// IsFocused returns whether this panel is focused.
func (m EnvironmentModel) IsFocused() bool {
	return m.focused
}

// View renders the environment panel.
func (m EnvironmentModel) View() string {
	if m.Width < 10 || m.Height < 3 {
		return ""
	}

	content := m.renderContent()

	footer := keys.FormatHints([]keys.Hint{
		{Key: "j/k", Desc: "scroll"},
	}, keys.FooterHintWidth(m.Width))

	scrollInfo := ""
	if info := m.Scroller.ScrollInfo(); info != "" {
		scrollInfo = info
	}

	r := rendering.New(m.Width, m.Height).
		SetTitle("Environment").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIdx).
		SetFooter(footer).
		SetFooterBadge(scrollInfo).
		SetContent(content)

	return r.Render()
}

// renderContent builds the full environment panel content.
func (m EnvironmentModel) renderContent() string {
	var allLines []string

	// Gather sensor data grouped by type across all devices
	temps := m.collectTemperatures()
	humids := m.collectHumidities()
	illums := m.collectIlluminances()
	batts := m.collectBatteries()
	volts := m.collectVoltmeters()
	bthome := m.collectBTHome()
	floods := m.collectFloodSensors()
	smokes := m.collectSmokeSensors()

	contentWidth := m.Width - 6 // borders (2) + padding (2) + margin (2)
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Environment sections
	if len(temps) > 0 {
		allLines = append(allLines, m.renderSectionHeader("Temperature"))
		allLines = append(allLines, m.renderTemperatures(temps, contentWidth)...)
	}
	if len(humids) > 0 {
		allLines = append(allLines, m.renderSectionHeader("Humidity"))
		allLines = append(allLines, m.renderHumidities(humids, contentWidth)...)
	}
	if len(illums) > 0 {
		allLines = append(allLines, m.renderSectionHeader("Illuminance"))
		allLines = append(allLines, m.renderIlluminances(illums, contentWidth)...)
	}
	if len(batts) > 0 {
		allLines = append(allLines, m.renderSectionHeader("Battery"))
		allLines = append(allLines, m.renderBatteries(batts, contentWidth)...)
	}
	if len(volts) > 0 {
		allLines = append(allLines, m.renderSectionHeader("Voltmeter"))
		allLines = append(allLines, m.renderVoltmeters(volts, contentWidth)...)
	}
	if len(bthome) > 0 {
		allLines = append(allLines, m.renderSectionHeader("BTHome"))
		allLines = append(allLines, m.renderBTHome(bthome, contentWidth)...)
	}

	// Safety section — always visible
	allLines = append(allLines, m.renderSectionHeader("Safety"))
	if len(floods) == 0 && len(smokes) == 0 {
		allLines = append(allLines, m.styles.Muted.Render("  No safety sensors configured"))
	} else {
		if len(floods) > 0 {
			allLines = append(allLines, m.renderAlarmSensors("Flood", floods, contentWidth)...)
		}
		if len(smokes) > 0 {
			allLines = append(allLines, m.renderAlarmSensors("Smoke", smokes, contentWidth)...)
		}
	}

	if len(allLines) == 0 {
		return m.styles.Muted.Render("No sensor data available")
	}

	// Update scroller to match actual line count (handles initial render before SetDevices)
	m.Scroller.SetItemCount(len(allLines))

	// Apply scrolling
	startIdx, endIdx := m.Scroller.VisibleRange()
	if endIdx > len(allLines) {
		endIdx = len(allLines)
	}
	if startIdx > len(allLines) {
		startIdx = len(allLines)
	}

	return strings.Join(allLines[startIdx:endIdx], "\n")
}

// sensorEntry is a sortable sensor reading with device name.
type sensorEntry[T any] struct {
	DeviceName string
	Reading    T
}

// collectSensors gathers sensor readings from all devices, sorted by device name.
// getSensors extracts the sensor slice; include filters readings (nil means include all).
func collectSensors[T any](devices []DeviceSensorData, getSensors func(*model.SensorData) []T, include func(T) bool) []sensorEntry[T] {
	var entries []sensorEntry[T]
	for _, d := range devices {
		for _, s := range getSensors(d.SensorData) {
			if include != nil && !include(s) {
				continue
			}
			entries = append(entries, sensorEntry[T]{
				DeviceName: d.DeviceName,
				Reading:    s,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].DeviceName < entries[j].DeviceName
	})
	return entries
}

// collectTemperatures gathers all temperature readings sorted by device name.
func (m EnvironmentModel) collectTemperatures() []sensorEntry[model.TemperatureReading] {
	return collectSensors(m.devices,
		func(s *model.SensorData) []model.TemperatureReading { return s.Temperature },
		func(t model.TemperatureReading) bool { return t.TC != nil })
}

// collectHumidities gathers all humidity readings sorted by device name.
func (m EnvironmentModel) collectHumidities() []sensorEntry[model.HumidityReading] {
	return collectSensors(m.devices,
		func(s *model.SensorData) []model.HumidityReading { return s.Humidity },
		func(h model.HumidityReading) bool { return h.RH != nil })
}

// collectIlluminances gathers all illuminance readings sorted by device name.
func (m EnvironmentModel) collectIlluminances() []sensorEntry[model.IlluminanceReading] {
	return collectSensors(m.devices,
		func(s *model.SensorData) []model.IlluminanceReading { return s.Illuminance },
		func(il model.IlluminanceReading) bool { return il.Lux != nil })
}

// batteryEntry holds a battery reading with device context.
type batteryEntry struct {
	DeviceName  string
	Reading     model.DevicePowerReading
	HasExtPower bool
}

// collectBatteries gathers all battery readings sorted lowest-first.
func (m EnvironmentModel) collectBatteries() []batteryEntry {
	entries := make([]batteryEntry, 0, len(m.devices))
	for _, d := range m.devices {
		for _, dp := range d.SensorData.DevicePower {
			entries = append(entries, batteryEntry{
				DeviceName:  d.DeviceName,
				Reading:     dp,
				HasExtPower: d.HasExtPower,
			})
		}
	}
	// Sort lowest battery first
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Reading.Battery.Percent < entries[j].Reading.Battery.Percent
	})
	return entries
}

// collectVoltmeters gathers all voltmeter readings sorted by device name.
func (m EnvironmentModel) collectVoltmeters() []sensorEntry[model.VoltmeterReading] {
	return collectSensors(m.devices,
		func(s *model.SensorData) []model.VoltmeterReading { return s.Voltmeter },
		func(v model.VoltmeterReading) bool { return v.Voltage != nil })
}

// collectBTHome gathers all BTHome sensor readings sorted by device name.
func (m EnvironmentModel) collectBTHome() []sensorEntry[model.BTHomeSensorReading] {
	return collectSensors[model.BTHomeSensorReading](m.devices,
		func(s *model.SensorData) []model.BTHomeSensorReading { return s.BTHome },
		nil)
}

// alarmEntry holds an alarm sensor reading with device and type context.
type alarmEntry struct {
	DeviceName string
	Reading    model.AlarmSensorReading
}

// collectAlarmEntries gathers alarm sensor readings, sorted alarms-first then by device name.
func collectAlarmEntries(devices []DeviceSensorData, getSensors func(*model.SensorData) []model.AlarmSensorReading) []alarmEntry {
	entries := make([]alarmEntry, 0, len(devices))
	for _, d := range devices {
		for _, s := range getSensors(d.SensorData) {
			entries = append(entries, alarmEntry{
				DeviceName: d.DeviceName,
				Reading:    s,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Reading.Alarm != entries[j].Reading.Alarm {
			return entries[i].Reading.Alarm
		}
		return entries[i].DeviceName < entries[j].DeviceName
	})
	return entries
}

// collectFloodSensors gathers all flood sensor readings.
func (m EnvironmentModel) collectFloodSensors() []alarmEntry {
	return collectAlarmEntries(m.devices, func(s *model.SensorData) []model.AlarmSensorReading { return s.Flood })
}

// collectSmokeSensors gathers all smoke sensor readings.
func (m EnvironmentModel) collectSmokeSensors() []alarmEntry {
	return collectAlarmEntries(m.devices, func(s *model.SensorData) []model.AlarmSensorReading { return s.Smoke })
}

// renderSectionHeader renders a section divider.
func (m EnvironmentModel) renderSectionHeader(name string) string {
	return m.styles.SectionHeader.Render("  " + name)
}

// renderSensorLines renders sensor entries with a consistent format: "    name padding value".
// formatFn returns the fully-styled value string for each reading.
func renderSensorLines[T any](entries []sensorEntry[T], maxWidth, padding int,
	formatFn func(T) string, nameStyle lipgloss.Style) []string {
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		valueStr := formatFn(e.Reading)
		name := output.Truncate(e.DeviceName, maxWidth-padding)
		namePad := strings.Repeat(" ", max(0, maxWidth-padding-len(name)))
		lines = append(lines, fmt.Sprintf("    %s%s %s",
			nameStyle.Render(name), namePad, valueStr))
	}
	return lines
}

// renderTemperatures renders all temperature readings.
func (m EnvironmentModel) renderTemperatures(entries []sensorEntry[model.TemperatureReading], maxWidth int) []string {
	return renderSensorLines(entries, maxWidth, 16, func(t model.TemperatureReading) string {
		tc := *t.TC
		return m.tempStyle(tc).Render(fmt.Sprintf("%.1f°C", tc))
	}, m.styles.DeviceName)
}

// tempStyle returns the appropriate style for a temperature value.
func (m EnvironmentModel) tempStyle(tc float64) lipgloss.Style {
	switch {
	case tc < tempColdThreshold:
		return m.styles.TempCold
	case tc < tempWarmThreshold:
		return m.styles.TempOK
	case tc < tempHotThreshold:
		return m.styles.TempWarm
	default:
		return m.styles.TempHot
	}
}

// renderHumidities renders all humidity readings.
func (m EnvironmentModel) renderHumidities(entries []sensorEntry[model.HumidityReading], maxWidth int) []string {
	return renderSensorLines(entries, maxWidth, 16, func(h model.HumidityReading) string {
		rh := *h.RH
		return m.humidStyle(rh).Render(fmt.Sprintf("%.0f%%", rh))
	}, m.styles.DeviceName)
}

// humidStyle returns the appropriate style for a humidity value.
func (m EnvironmentModel) humidStyle(rh float64) lipgloss.Style {
	switch {
	case rh < humidLowBad || rh > humidHighBad:
		return m.styles.HumidBad
	case rh < humidLowCaution || rh > humidHighOK:
		return m.styles.HumidCaution
	default:
		return m.styles.HumidOK
	}
}

// renderIlluminances renders all illuminance readings.
func (m EnvironmentModel) renderIlluminances(entries []sensorEntry[model.IlluminanceReading], maxWidth int) []string {
	return renderSensorLines(entries, maxWidth, 20, func(il model.IlluminanceReading) string {
		lux := *il.Lux
		var valueStr string
		if il.Illumination != nil {
			valueStr = fmt.Sprintf("%.0f lux (%s)", lux, *il.Illumination)
		} else {
			valueStr = fmt.Sprintf("%.0f lux", lux)
		}
		return m.styles.Value.Render(valueStr)
	}, m.styles.DeviceName)
}

// renderBatteries renders all battery readings.
func (m EnvironmentModel) renderBatteries(entries []batteryEntry, maxWidth int) []string {
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		pct := e.Reading.Battery.Percent
		style := m.batteryStyle(pct)
		valueStr := fmt.Sprintf("%d%%", pct)

		extStr := ""
		if e.HasExtPower {
			extStr = " " + m.styles.ExtPower.Render("[ext]")
		}

		name := output.Truncate(e.DeviceName, maxWidth-20)
		namePad := strings.Repeat(" ", max(0, maxWidth-20-len(name)))
		lines = append(lines, fmt.Sprintf("    %s%s %s%s",
			m.styles.DeviceName.Render(name), namePad, style.Render(valueStr), extStr))
	}
	return lines
}

// batteryStyle returns the appropriate style for a battery percentage.
func (m EnvironmentModel) batteryStyle(pct int) lipgloss.Style {
	switch {
	case pct < batteryCritThreshold:
		return m.styles.BatteryCrit
	case pct < batteryLowThreshold:
		return m.styles.BatteryLow
	default:
		return m.styles.BatteryGood
	}
}

// renderVoltmeters renders all voltmeter readings.
func (m EnvironmentModel) renderVoltmeters(entries []sensorEntry[model.VoltmeterReading], maxWidth int) []string {
	return renderSensorLines(entries, maxWidth, 16, func(v model.VoltmeterReading) string {
		return m.styles.Value.Render(fmt.Sprintf("%.2fV", *v.Voltage))
	}, m.styles.DeviceName)
}

// renderBTHome renders all BTHome sensor readings.
func (m EnvironmentModel) renderBTHome(entries []sensorEntry[model.BTHomeSensorReading], maxWidth int) []string {
	return renderSensorLines(entries, maxWidth, 20, func(b model.BTHomeSensorReading) string {
		return m.styles.Value.Render(formatBTHomeValue(b.Value))
	}, m.styles.DeviceName)
}

// formatBTHomeValue formats a BTHome sensor value for display.
func formatBTHomeValue(v any) string {
	switch val := v.(type) {
	case float64:
		return fmt.Sprintf("%.1f", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case string:
		return val
	default:
		return fmt.Sprintf("%v", v)
	}
}

// renderAlarmSensors renders flood or smoke alarm sensor entries.
func (m EnvironmentModel) renderAlarmSensors(sensorType string, entries []alarmEntry, maxWidth int) []string {
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		var statusStr string
		switch {
		case e.Reading.Alarm:
			statusStr = m.styles.AlarmFire.Render(" ALARM ")
		case e.Reading.Mute:
			statusStr = m.styles.AlarmMuted.Render("MUTED")
		default:
			statusStr = m.styles.AlarmOK.Render("OK")
		}

		label := sensorType
		name := output.Truncate(e.DeviceName, maxWidth-24)
		namePad := strings.Repeat(" ", max(0, maxWidth-24-len(name)))

		lines = append(lines, fmt.Sprintf("    %s %s%s %s",
			m.styles.Muted.Render(label),
			m.styles.DeviceName.Render(name), namePad,
			statusStr))
	}
	return lines
}

// countSensorEntries counts matching sensor entries without allocating/sorting.
func countSensorEntries[T any](devices []DeviceSensorData, getSensors func(*model.SensorData) []T, include func(T) bool) int {
	count := 0
	for _, d := range devices {
		for _, s := range getSensors(d.SensorData) {
			if include == nil || include(s) {
				count++
			}
		}
	}
	return count
}

// countAlarmEntries counts alarm sensor entries without allocating.
func countAlarmEntries(devices []DeviceSensorData, getSensors func(*model.SensorData) []model.AlarmSensorReading) int {
	count := 0
	for _, d := range devices {
		count += len(getSensors(d.SensorData))
	}
	return count
}

// countDisplayLines counts the total number of display lines for scrolling.
// Uses counting helpers to avoid allocating/sorting slices just to count them.
func (m EnvironmentModel) countDisplayLines() int {
	count := 0

	addSection := func(n int) {
		if n > 0 {
			count += 1 + n // header + entries
		}
	}

	addSection(countSensorEntries(m.devices,
		func(s *model.SensorData) []model.TemperatureReading { return s.Temperature },
		func(t model.TemperatureReading) bool { return t.TC != nil }))
	addSection(countSensorEntries(m.devices,
		func(s *model.SensorData) []model.HumidityReading { return s.Humidity },
		func(h model.HumidityReading) bool { return h.RH != nil }))
	addSection(countSensorEntries(m.devices,
		func(s *model.SensorData) []model.IlluminanceReading { return s.Illuminance },
		func(il model.IlluminanceReading) bool { return il.Lux != nil }))

	// Battery: count DevicePower entries
	battCount := 0
	for _, d := range m.devices {
		battCount += len(d.SensorData.DevicePower)
	}
	addSection(battCount)

	addSection(countSensorEntries(m.devices,
		func(s *model.SensorData) []model.VoltmeterReading { return s.Voltmeter },
		func(v model.VoltmeterReading) bool { return v.Voltage != nil }))
	addSection(countSensorEntries[model.BTHomeSensorReading](m.devices,
		func(s *model.SensorData) []model.BTHomeSensorReading { return s.BTHome },
		nil))

	// Safety section is always present
	floodCount := countAlarmEntries(m.devices, func(s *model.SensorData) []model.AlarmSensorReading { return s.Flood })
	smokeCount := countAlarmEntries(m.devices, func(s *model.SensorData) []model.AlarmSensorReading { return s.Smoke })
	count++ // "Safety" header
	if floodCount == 0 && smokeCount == 0 {
		count++ // "No safety sensors configured"
	} else {
		count += floodCount + smokeCount
	}

	return count
}

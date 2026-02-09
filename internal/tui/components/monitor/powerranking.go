package monitor

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// TrendDirection represents the power consumption trend.
type TrendDirection int

const (
	// TrendStable means power has not changed significantly.
	TrendStable TrendDirection = iota
	// TrendRising means power has increased.
	TrendRising
	// TrendFalling means power has decreased.
	TrendFalling
)

// RankedDevice holds a device's power ranking data.
type RankedDevice struct {
	Name           string
	Address        string
	Type           string
	Online         bool
	Power          float64
	Trend          TrendDirection
	Error          error
	ConnectionType string

	// Health badges
	ChipTemp  *float64 // Component temperature (°C)
	WiFiRSSI  *float64 // WiFi signal strength (dBm)
	FSFree    int      // Filesystem free bytes
	FSSize    int      // Filesystem total bytes
	HasUpdate bool     // Firmware update available
}

// PowerRankingModel displays devices sorted by power consumption.
type PowerRankingModel struct {
	panel.Sizable
	devices    []RankedDevice
	prevPowers map[string]float64 // Previous power readings for trend detection
	focused    bool
	panelIdx   int
	styles     PowerRankingStyles
}

// PowerRankingStyles holds styles for the power ranking panel.
type PowerRankingStyles struct {
	Rank       lipgloss.Style
	DeviceName lipgloss.Style
	Power      lipgloss.Style
	PowerHigh  lipgloss.Style
	PowerMed   lipgloss.Style
	PowerLow   lipgloss.Style
	PowerZero  lipgloss.Style
	TrendUp    lipgloss.Style
	TrendDown  lipgloss.Style
	TrendFlat  lipgloss.Style
	Offline    lipgloss.Style
	ErrorText  lipgloss.Style
	Muted      lipgloss.Style
	Selected   lipgloss.Style
	Normal     lipgloss.Style
	Connection lipgloss.Style
}

// powerThresholdHigh is the threshold for "high" power coloring.
const powerThresholdHigh = 500.0

// powerThresholdMed is the threshold for "medium" power coloring.
const powerThresholdMed = 100.0

// Health badge thresholds.
const (
	chipTempWarn = 80.0  // °C - warn above this
	rssiWeak     = -75.0 // dBm - weak signal
	fsUsageWarn  = 90    // percent - warn above this
)

// defaultPowerRankingStyles returns default styles.
func defaultPowerRankingStyles() PowerRankingStyles {
	colors := theme.GetSemanticColors()
	return PowerRankingStyles{
		Rank: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Width(4),
		DeviceName: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Text),
		Power: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),
		PowerHigh: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
		PowerMed: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),
		PowerLow: lipgloss.NewStyle().
			Foreground(colors.Success).
			Bold(true),
		PowerZero: lipgloss.NewStyle().
			Foreground(colors.Muted),
		TrendUp: lipgloss.NewStyle().
			Foreground(colors.Error),
		TrendDown: lipgloss.NewStyle().
			Foreground(colors.Success),
		TrendFlat: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Offline: lipgloss.NewStyle().
			Foreground(colors.Offline),
		ErrorText: lipgloss.NewStyle().
			Foreground(colors.Error).
			Italic(true),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Bold(true),
		Normal: lipgloss.NewStyle(),
		Connection: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// NewPowerRanking creates a new power ranking model.
// Scroll offset accounts for borders, padding, and footer lines.
func NewPowerRanking() PowerRankingModel {
	return PowerRankingModel{
		Sizable:    panel.NewSizable(5, panel.NewScroller(0, 10)),
		devices:    []RankedDevice{},
		prevPowers: make(map[string]float64),
		styles:     defaultPowerRankingStyles(),
	}
}

// Init returns the initial command.
func (m PowerRankingModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the power ranking panel.
func (m PowerRankingModel) Update(msg tea.Msg) (PowerRankingModel, tea.Cmd) {
	if navMsg, ok := msg.(messages.NavigationMsg); ok {
		return m.handleNavigation(navMsg), nil
	}
	return m, nil
}

// handleNavigation handles scrolling/navigation.
func (m PowerRankingModel) handleNavigation(msg messages.NavigationMsg) PowerRankingModel {
	m.Scroller.HandleNavigation(msg)
	return m
}

// SetDevices updates the ranked device list from monitor statuses.
// Calculates trend from previous readings and sorts by power (highest first).
func (m PowerRankingModel) SetDevices(statuses []DeviceStatus) PowerRankingModel {
	ranked := make([]RankedDevice, 0, len(statuses))
	newPrevPowers := make(map[string]float64, len(statuses))

	for _, s := range statuses {
		trend := TrendStable
		if prev, ok := m.prevPowers[s.Name]; ok {
			diff := s.Power - prev
			switch {
			case diff > 1.0:
				trend = TrendRising
			case diff < -1.0:
				trend = TrendFalling
			}
		}

		ranked = append(ranked, RankedDevice{
			Name:           s.Name,
			Address:        s.Address,
			Type:           s.Type,
			Online:         s.Online,
			Power:          s.Power,
			Trend:          trend,
			Error:          s.Error,
			ConnectionType: s.ConnectionType,
			ChipTemp:       s.ChipTemp,
			WiFiRSSI:       s.WiFiRSSI,
			FSFree:         s.FSFree,
			FSSize:         s.FSSize,
			HasUpdate:      s.HasUpdate,
		})

		if s.Online {
			newPrevPowers[s.Name] = s.Power
		}
	}

	// Sort: online devices by power (desc), zero-power at bottom, offline at very bottom
	sort.Slice(ranked, func(i, j int) bool {
		di, dj := ranked[i], ranked[j]

		// Offline devices always last
		if di.Online != dj.Online {
			return di.Online
		}

		// Among online: sort by power descending
		if di.Online && dj.Online {
			// Zero-power after non-zero
			if (di.Power > 0) != (dj.Power > 0) {
				return di.Power > 0
			}
			return di.Power > dj.Power
		}

		// Among offline: sort by name
		return di.Name < dj.Name
	})

	m.devices = ranked
	m.prevPowers = newPrevPowers
	m.Scroller.SetItemCount(len(ranked))
	return m
}

// SetFocused sets whether this panel is focused.
func (m PowerRankingModel) SetFocused(focused bool) PowerRankingModel {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hints.
func (m PowerRankingModel) SetPanelIndex(idx int) PowerRankingModel {
	m.panelIdx = idx
	return m
}

// SetSize updates the component dimensions.
func (m PowerRankingModel) SetSize(width, height int) PowerRankingModel {
	m.ApplySize(width, height)
	return m
}

// SelectedDevice returns the currently selected device, if any.
func (m PowerRankingModel) SelectedDevice() *RankedDevice {
	cursor := m.Scroller.Cursor()
	if len(m.devices) == 0 || cursor < 0 || cursor >= len(m.devices) {
		return nil
	}
	return &m.devices[cursor]
}

// Devices returns the current ranked device list.
func (m PowerRankingModel) Devices() []RankedDevice {
	return m.devices
}

// IsFocused returns whether this panel is focused.
func (m PowerRankingModel) IsFocused() bool {
	return m.focused
}

// View renders the power ranking panel.
func (m PowerRankingModel) View() string {
	if m.Width < 10 || m.Height < 3 {
		return ""
	}

	content := m.renderContent()

	footer := keys.FormatHints([]keys.Hint{
		{Key: "j/k", Desc: "scroll"},
		{Key: "t", Desc: "toggle"},
		{Key: "c", Desc: "control"},
		{Key: "d", Desc: "detail"},
	}, keys.FooterHintWidth(m.Width))

	scrollInfo := ""
	if info := m.Scroller.ScrollInfo(); info != "" {
		scrollInfo = info
	}

	r := rendering.New(m.Width, m.Height).
		SetTitle("Power Ranking").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIdx).
		SetFooter(footer).
		SetFooterBadge(scrollInfo).
		SetContent(content)

	return r.Render()
}

// renderContent builds the power ranking list content.
func (m PowerRankingModel) renderContent() string {
	if len(m.devices) == 0 {
		return m.styles.Muted.Render("No devices to rank")
	}

	contentWidth := m.Width - 6 // borders (2) + padding (2) + margin (2)
	if contentWidth < 10 {
		contentWidth = 10
	}

	startIdx, endIdx := m.Scroller.VisibleRange()
	if endIdx > len(m.devices) {
		endIdx = len(m.devices)
	}

	var lines []string
	rank := 1
	for i := startIdx; i < endIdx; i++ {
		d := m.devices[i]
		isSelected := m.Scroller.IsCursorAt(i)
		line := m.renderDeviceRow(d, rank, isSelected, contentWidth)
		lines = append(lines, line)
		if d.Online && d.Power > 0 {
			rank++
		}
	}

	return strings.Join(lines, "\n")
}

// renderDeviceRow renders a single device row in the ranking.
func (m PowerRankingModel) renderDeviceRow(d RankedDevice, rank int, isSelected bool, maxWidth int) string {
	sel := "  "
	if isSelected {
		sel = "▶ "
	}

	if !d.Online {
		return m.renderOfflineRow(d, sel, maxWidth)
	}

	if d.Power == 0 {
		return m.renderZeroPowerRow(d, sel, maxWidth)
	}

	return m.renderPoweredRow(d, rank, sel, maxWidth)
}

// renderPoweredRow renders a device with active power consumption.
func (m PowerRankingModel) renderPoweredRow(d RankedDevice, rank int, sel string, maxWidth int) string {
	rankStr := m.styles.Rank.Render(fmt.Sprintf("%d.", rank))
	trendStr := m.trendIndicator(d.Trend)
	powerStr := m.powerStyled(d.Power)

	// Connection type indicator
	connStr := ""
	if d.ConnectionType == "ws" {
		connStr = m.styles.Connection.Render(" [ws]")
	}

	// Truncate name to fit
	nameWidth := maxWidth - 20 // rank(4) + power(10) + trend(2) + sel(2) + spacing(2)
	if nameWidth < 8 {
		nameWidth = 8
	}
	name := output.Truncate(d.Name, nameWidth)
	nameStr := m.styles.DeviceName.Render(name)

	// Pad name to alignment
	namePad := strings.Repeat(" ", max(0, nameWidth-len(name)))

	badges := m.healthBadges(d)
	line := sel + rankStr + " " + nameStr + namePad + " " + powerStr + " " + trendStr + connStr + badges
	return line
}

// renderZeroPowerRow renders a device with zero power.
func (m PowerRankingModel) renderZeroPowerRow(d RankedDevice, sel string, maxWidth int) string {
	nameWidth := maxWidth - 16
	if nameWidth < 8 {
		nameWidth = 8
	}
	name := output.Truncate(d.Name, nameWidth)
	namePad := strings.Repeat(" ", max(0, nameWidth-len(name)))

	badges := m.healthBadges(d)
	line := sel + m.styles.Muted.Render("─") + "  " +
		m.styles.Muted.Render(name) + namePad + " " +
		m.styles.PowerZero.Render("0W") + badges
	return line
}

// renderOfflineRow renders an offline device.
func (m PowerRankingModel) renderOfflineRow(d RankedDevice, sel string, maxWidth int) string {
	nameWidth := maxWidth - 20
	if nameWidth < 8 {
		nameWidth = 8
	}
	name := output.Truncate(d.Name, nameWidth)
	namePad := strings.Repeat(" ", max(0, nameWidth-len(name)))

	errMsg := "offline"
	if d.Error != nil {
		errMsg = output.Truncate(d.Error.Error(), 20)
	}

	line := sel + m.styles.Offline.Render("✗") + "  " +
		m.styles.Offline.Render(name) + namePad + " " +
		m.styles.ErrorText.Render(errMsg)
	return line
}

// trendIndicator returns the styled trend arrow.
func (m PowerRankingModel) trendIndicator(trend TrendDirection) string {
	switch trend {
	case TrendRising:
		return m.styles.TrendUp.Render("▲")
	case TrendFalling:
		return m.styles.TrendDown.Render("▼")
	default:
		return m.styles.TrendFlat.Render("─")
	}
}

// powerStyled returns styled power text based on magnitude.
func (m PowerRankingModel) powerStyled(watts float64) string {
	text := formatPower(watts)
	switch {
	case watts >= powerThresholdHigh:
		return m.styles.PowerHigh.Render(text)
	case watts >= powerThresholdMed:
		return m.styles.PowerMed.Render(text)
	default:
		return m.styles.PowerLow.Render(text)
	}
}

// healthBadges returns compact badge icons for device health warnings.
// Returns empty string if all health metrics are normal.
func (m PowerRankingModel) healthBadges(d RankedDevice) string {
	var badges []string

	// Chip temperature warning: >80°C
	if d.ChipTemp != nil && *d.ChipTemp >= chipTempWarn {
		badges = append(badges, m.styles.PowerHigh.Render("🌡"))
	}

	// WiFi RSSI warning: <-75dBm
	if d.WiFiRSSI != nil && *d.WiFiRSSI <= rssiWeak {
		badges = append(badges, m.styles.TrendUp.Render("📶"))
	}

	// Flash usage warning: >90%
	if d.FSSize > 0 {
		usedPct := 100 - (d.FSFree * 100 / d.FSSize)
		if usedPct >= fsUsageWarn {
			badges = append(badges, m.styles.TrendUp.Render("💾"))
		}
	}

	// Firmware update available
	if d.HasUpdate {
		badges = append(badges, m.styles.Muted.Render("⬆"))
	}

	// Solar return (negative power)
	if d.Power < 0 {
		badges = append(badges, m.styles.TrendDown.Render("☀"))
	}

	if len(badges) == 0 {
		return ""
	}
	return " " + strings.Join(badges, "")
}

// FooterText returns keybinding hints for the footer.
func (m PowerRankingModel) FooterText() string {
	return keys.FormatHints([]keys.Hint{
		{Key: "j/k", Desc: "scroll"},
		{Key: "t", Desc: "toggle"},
		{Key: "c", Desc: "control"},
		{Key: "d", Desc: "detail"},
		{Key: "x", Desc: "csv"},
		{Key: "X", Desc: "json"},
	}, keys.FooterHintWidth(m.Width))
}

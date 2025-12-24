// Package energy provides the energy dashboard view for the TUI.
package energy

import (
	"context"
	"fmt"
	"sort"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Deps holds the dependencies for the energy component.
type Deps struct {
	Ctx             context.Context
	Svc             *shelly.Service
	IOS             *iostreams.IOStreams
	RefreshInterval time.Duration
}

// validate ensures all required dependencies are set.
func (d Deps) validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	if d.IOS == nil {
		return fmt.Errorf("iostreams is required")
	}
	return nil
}

// DeviceEnergy represents energy data for a device.
type DeviceEnergy struct {
	Name        string
	Address     string
	Type        string
	Online      bool
	Power       float64 // Current power in watts
	TotalEnergy float64 // Total energy in Wh
	Voltage     float64
	Current     float64
	Frequency   float64
	Error       error
}

// UpdateMsg is sent when energy data is updated.
type UpdateMsg struct {
	Devices []DeviceEnergy
	Err     error
}

// RefreshTickMsg triggers periodic refresh.
type RefreshTickMsg struct{}

// Model holds the energy dashboard state.
type Model struct {
	ctx             context.Context
	svc             *shelly.Service
	ios             *iostreams.IOStreams
	devices         []DeviceEnergy
	loading         bool
	err             error
	width           int
	height          int
	styles          Styles
	refreshInterval time.Duration
	scrollOffset    int
	cursor          int // Currently selected row
}

// Styles for the energy component.
type Styles struct {
	Container     lipgloss.Style
	Header        lipgloss.Style
	Card          lipgloss.Style
	CardTitle     lipgloss.Style
	Value         lipgloss.Style
	Unit          lipgloss.Style
	Label         lipgloss.Style
	Positive      lipgloss.Style
	Negative      lipgloss.Style
	TotalHeader   lipgloss.Style
	TotalValue    lipgloss.Style
	DeviceName    lipgloss.Style
	DeviceAddress lipgloss.Style
	OnlineIcon    lipgloss.Style
	OfflineIcon   lipgloss.Style
	UpdatingIcon  lipgloss.Style
	Separator     lipgloss.Style
	Footer        lipgloss.Style
	SelectedRow   lipgloss.Style
}

// DefaultStyles returns default styles for energy.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Padding(1, 2),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight).
			MarginBottom(1),
		Card: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Padding(0, 1).
			MarginRight(1).
			MarginBottom(1),
		CardTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Text),
		Value: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Warning),
		Unit: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Positive: lipgloss.NewStyle().
			Foreground(colors.Success),
		Negative: lipgloss.NewStyle().
			Foreground(colors.Error),
		TotalHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Primary),
		TotalValue: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight),
		DeviceName: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Text).
			Width(18),
		DeviceAddress: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		OnlineIcon: lipgloss.NewStyle().
			Foreground(colors.Online),
		OfflineIcon: lipgloss.NewStyle().
			Foreground(colors.Offline),
		UpdatingIcon: lipgloss.NewStyle().
			Foreground(colors.Updating),
		Separator: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Footer: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		SelectedRow: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// New creates a new energy model.
func New(deps Deps) Model {
	if err := deps.validate(); err != nil {
		panic(fmt.Sprintf("energy: invalid deps: %v", err))
	}

	refreshInterval := deps.RefreshInterval
	if refreshInterval == 0 {
		refreshInterval = 5 * time.Second
	}

	return Model{
		ctx:             deps.Ctx,
		svc:             deps.Svc,
		ios:             deps.IOS,
		devices:         []DeviceEnergy{},
		loading:         true,
		styles:          DefaultStyles(),
		refreshInterval: refreshInterval,
	}
}

// Init returns the initial command for energy.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchEnergy(),
		m.scheduleRefresh(),
	)
}

// scheduleRefresh schedules the next refresh tick.
func (m Model) scheduleRefresh() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}

// fetchEnergy returns a command that fetches energy data.
func (m Model) fetchEnergy() tea.Cmd {
	return func() tea.Msg {
		deviceMap := config.ListDevices()
		if len(deviceMap) == 0 {
			return UpdateMsg{Devices: nil}
		}

		// Add timeout to prevent slow/offline devices from blocking UI
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		// Use errgroup for concurrent fetching
		// Limit to 3 concurrent requests to avoid overloading Gen1 (ESP8266) devices
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(3)

		results := make(chan DeviceEnergy, len(deviceMap))

		for _, d := range deviceMap {
			device := d
			g.Go(func() error {
				energy := m.fetchDeviceEnergy(gctx, device)
				results <- energy
				return nil
			})
		}

		go func() {
			if err := g.Wait(); err != nil {
				m.ios.DebugErr("energy errgroup wait", err)
			}
			close(results)
		}()

		devices := make([]DeviceEnergy, 0, len(deviceMap))
		for energy := range results {
			devices = append(devices, energy)
		}

		// Sort by name for consistent display
		sort.Slice(devices, func(i, j int) bool {
			return devices[i].Name < devices[j].Name
		})

		return UpdateMsg{Devices: devices}
	}
}

// fetchDeviceEnergy fetches energy data for a single device.
func (m Model) fetchDeviceEnergy(ctx context.Context, device model.Device) DeviceEnergy {
	energy := DeviceEnergy{
		Name:    device.Name,
		Address: device.Address,
		Type:    device.Type,
		Online:  false,
	}

	// Per-device timeout to prevent single slow device from blocking others
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Use GetMonitoringSnapshotAuto to handle both Gen1 and Gen2 devices
	snapshot, err := m.svc.GetMonitoringSnapshotAuto(ctx, device.Address)
	if err != nil {
		energy.Error = err
		return energy
	}

	energy.Online = true
	aggregateEnergyFromPM(&energy, snapshot.PM)
	aggregateEnergyFromEM(&energy, snapshot.EM)
	aggregateEnergyFromEM1(&energy, snapshot.EM1)

	return energy
}

// aggregateEnergyFromPM aggregates energy data from PM components.
func aggregateEnergyFromPM(energy *DeviceEnergy, pms []shelly.PMStatus) {
	for _, pm := range pms {
		energy.Power += pm.APower
		if energy.Voltage == 0 && pm.Voltage > 0 {
			energy.Voltage = pm.Voltage
		}
		if energy.Current == 0 && pm.Current > 0 {
			energy.Current = pm.Current
		}
		if pm.Freq != nil && energy.Frequency == 0 {
			energy.Frequency = *pm.Freq
		}
		if pm.AEnergy != nil {
			energy.TotalEnergy += pm.AEnergy.Total
		}
	}
}

// aggregateEnergyFromEM aggregates energy data from EM components (3-phase meters).
func aggregateEnergyFromEM(energy *DeviceEnergy, ems []shelly.EMStatus) {
	for _, em := range ems {
		energy.Power += em.TotalActivePower
		energy.Current += em.TotalCurrent
		if energy.Voltage == 0 && em.AVoltage > 0 {
			energy.Voltage = em.AVoltage
		}
		if em.AFreq != nil && energy.Frequency == 0 {
			energy.Frequency = *em.AFreq
		}
	}
}

// aggregateEnergyFromEM1 aggregates energy data from EM1 components (single-phase meters).
func aggregateEnergyFromEM1(energy *DeviceEnergy, em1s []shelly.EM1Status) {
	for _, em1 := range em1s {
		energy.Power += em1.ActPower
		if energy.Voltage == 0 && em1.Voltage > 0 {
			energy.Voltage = em1.Voltage
		}
		if energy.Current == 0 && em1.Current > 0 {
			energy.Current = em1.Current
		}
		if em1.Freq != nil && energy.Frequency == 0 {
			energy.Frequency = *em1.Freq
		}
	}
}

// Refresh returns a command to refresh energy data.
func (m Model) Refresh() tea.Cmd {
	return m.fetchEnergy()
}

// Update handles messages for energy.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UpdateMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.devices = msg.Devices
		return m, nil

	case RefreshTickMsg:
		return m, tea.Batch(
			m.fetchEnergy(),
			m.scheduleRefresh(),
		)

	case tea.KeyPressMsg:
		m = m.handleKeyPress(msg)
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) Model {
	switch msg.String() {
	case "j", "down":
		m = m.cursorDown()
	case "k", "up":
		m = m.cursorUp()
	case "g":
		m.cursor = 0
		m.scrollOffset = 0
	case "G":
		m = m.cursorToEnd()
	case "pgdown", "ctrl+d":
		m = m.pageDown()
	case "pgup", "ctrl+u":
		m = m.pageUp()
	}
	return m
}

func (m Model) cursorDown() Model {
	if m.cursor < len(m.devices)-1 {
		m.cursor++
		visibleDevices := m.visibleDevices()
		if m.cursor >= m.scrollOffset+visibleDevices {
			m.scrollOffset = m.cursor - visibleDevices + 1
		}
	}
	return m
}

func (m Model) cursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.scrollOffset {
			m.scrollOffset = m.cursor
		}
	}
	return m
}

func (m Model) cursorToEnd() Model {
	if len(m.devices) > 0 {
		m.cursor = len(m.devices) - 1
		maxOffset := len(m.devices) - m.visibleDevices()
		if maxOffset < 0 {
			maxOffset = 0
		}
		m.scrollOffset = maxOffset
	}
	return m
}

func (m Model) pageDown() Model {
	if len(m.devices) == 0 {
		return m
	}
	visibleDevices := m.visibleDevices()
	m.cursor += visibleDevices
	if m.cursor >= len(m.devices) {
		m.cursor = len(m.devices) - 1
	}
	if m.cursor >= m.scrollOffset+visibleDevices {
		m.scrollOffset = m.cursor - visibleDevices + 1
	}
	maxOffset := len(m.devices) - visibleDevices
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	return m
}

func (m Model) pageUp() Model {
	if len(m.devices) == 0 {
		return m
	}
	visibleDevices := m.visibleDevices()
	m.cursor -= visibleDevices
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	return m
}

// visibleDevices calculates how many device rows can be displayed.
func (m Model) visibleDevices() int {
	availableHeight := m.height - 10 // Account for header, summary, footer
	if availableHeight < 1 {
		return 3
	}
	return availableHeight / 2 // Each device takes ~2 lines
}

// SetSize sets the component size.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// View renders the energy dashboard.
func (m Model) View() string {
	if m.loading {
		loadingText := m.styles.UpdatingIcon.Render("◐ ") + "Fetching energy data..."
		return m.styles.Container.
			Width(m.width-4).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(loadingText)
	}

	if m.err != nil {
		return m.styles.Container.
			Width(m.width - 4).
			Render(theme.StatusError().Render("Error: " + m.err.Error()))
	}

	if len(m.devices) == 0 {
		return m.styles.Container.
			Width(m.width-4).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No devices with energy monitoring.\nAdd devices with power metering to see energy data.")
	}

	// Build the energy dashboard
	header := m.styles.Header.Render("Energy Dashboard")

	// Calculate totals
	var totalPower, totalEnergy float64
	var onlineCount, offlineCount int
	for _, d := range m.devices {
		if d.Online {
			onlineCount++
			totalPower += d.Power
			totalEnergy += d.TotalEnergy
		} else {
			offlineCount++
		}
	}

	// Summary cards row
	summary := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderSummaryCard("Current Power", m.formatPower(totalPower)),
		m.renderSummaryCard("Total Energy", m.formatEnergy(totalEnergy)),
		m.renderSummaryCard("Devices", fmt.Sprintf("%d online / %d total", onlineCount, len(m.devices))),
	)

	// Device rows with scrolling
	visibleDevices := m.visibleDevices()
	startIdx := m.scrollOffset
	endIdx := startIdx + visibleDevices
	if endIdx > len(m.devices) {
		endIdx = len(m.devices)
	}

	var deviceRows string
	for i := startIdx; i < endIdx; i++ {
		d := m.devices[i]
		isSelected := i == m.cursor
		deviceRows += m.renderDeviceRow(d, isSelected)
		if i < endIdx-1 {
			deviceRows += m.styles.Separator.Render("─────────────────────────────────────────────────") + "\n"
		}
	}

	if deviceRows == "" {
		deviceRows = m.styles.Label.Render("No active power consumption detected.")
	}

	// Footer with scroll info
	scrollInfo := ""
	if len(m.devices) > visibleDevices {
		scrollInfo = fmt.Sprintf(" [%d-%d of %d]", startIdx+1, endIdx, len(m.devices))
	}
	footer := m.styles.Footer.Render(
		fmt.Sprintf("Last updated: %s  j/k: scroll%s", time.Now().Format("15:04:05"), scrollInfo),
	)

	content := lipgloss.JoinVertical(lipgloss.Left, header, summary, "", deviceRows, footer)
	return m.styles.Container.Width(m.width - 4).Render(content)
}

func (m Model) renderSummaryCard(label, value string) string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		m.styles.Label.Render(label),
		m.styles.TotalValue.Render(value),
	)
	return m.styles.Card.Width(22).Render(content)
}

func (m Model) renderDeviceRow(d DeviceEnergy, isSelected bool) string {
	line1 := m.renderDeviceHeader(d, isSelected)
	line2 := m.renderDeviceMetrics(d)
	return m.applyRowStyle(line1, line2, isSelected)
}

// renderDeviceHeader renders the first line with status icon and device info.
func (m Model) renderDeviceHeader(d DeviceEnergy, isSelected bool) string {
	selIndicator := "  "
	if isSelected {
		selIndicator = "▶ "
	}

	statusIcon := m.styles.OfflineIcon.Render("○")
	if d.Online {
		statusIcon = m.styles.OnlineIcon.Render("●")
	}

	return fmt.Sprintf("%s%s %s  %s",
		selIndicator,
		statusIcon,
		m.styles.DeviceName.Render(d.Name),
		m.styles.DeviceAddress.Render(d.Address),
	)
}

// renderDeviceMetrics renders the second line with power metrics or error.
func (m Model) renderDeviceMetrics(d DeviceEnergy) string {
	if !d.Online {
		return "    " + theme.StatusError().Render(m.formatError(d.Error))
	}

	metrics := m.buildMetricsList(d)
	if len(metrics) == 0 {
		return "    " + m.styles.Label.Render("no power data")
	}
	return "    " + m.joinMetrics(metrics)
}

// formatError formats an error message, truncating if necessary.
func (m Model) formatError(err error) string {
	if err == nil {
		return "unreachable"
	}
	errMsg := err.Error()
	if len(errMsg) > 40 {
		return errMsg[:40] + "..."
	}
	return errMsg
}

// buildMetricsList builds the list of metric strings for a device.
func (m Model) buildMetricsList(d DeviceEnergy) []string {
	var metrics []string

	if d.Power < 0 {
		metrics = append(metrics, m.styles.Positive.Render(m.formatPower(d.Power)+" (generating)"))
	} else if d.Power > 0 {
		metrics = append(metrics, m.styles.Value.Render(m.formatPower(d.Power)))
	}

	if d.Voltage > 0 {
		metrics = append(metrics, m.styles.Label.Render("V: ")+m.styles.Value.Render(fmt.Sprintf("%.1f", d.Voltage)))
	}
	if d.Current > 0 {
		metrics = append(metrics, m.styles.Label.Render("A: ")+m.styles.Value.Render(fmt.Sprintf("%.2f", d.Current)))
	}
	if d.TotalEnergy > 0 {
		metrics = append(metrics, m.styles.Label.Render("Total: ")+m.styles.Value.Render(m.formatEnergy(d.TotalEnergy)))
	}

	return metrics
}

// joinMetrics joins metrics with separators.
func (m Model) joinMetrics(metrics []string) string {
	result := ""
	for i, metric := range metrics {
		if i > 0 {
			result += m.styles.Separator.Render(" │ ")
		}
		result += metric
	}
	return result
}

// applyRowStyle applies the appropriate style to the row content.
func (m Model) applyRowStyle(line1, line2 string, isSelected bool) string {
	content := line1 + "\n" + line2
	if isSelected {
		return m.styles.SelectedRow.Render(content) + "\n"
	}
	return content + "\n"
}

func (m Model) formatPower(watts float64) string {
	if watts >= 1000 || watts <= -1000 {
		return fmt.Sprintf("%.2f kW", watts/1000)
	}
	return fmt.Sprintf("%.1f W", watts)
}

func (m Model) formatEnergy(wh float64) string {
	if wh >= 1000000 {
		return fmt.Sprintf("%.2f MWh", wh/1000000)
	}
	if wh >= 1000 {
		return fmt.Sprintf("%.2f kWh", wh/1000)
	}
	return fmt.Sprintf("%.1f Wh", wh)
}

// TotalPower returns the total power consumption across all devices.
func (m Model) TotalPower() float64 {
	var total float64
	for _, d := range m.devices {
		if d.Online {
			total += d.Power
		}
	}
	return total
}

// DeviceCount returns the number of devices.
func (m Model) DeviceCount() int {
	return len(m.devices)
}

// SelectedDevice returns the currently selected device, if any.
func (m Model) SelectedDevice() *DeviceEnergy {
	if len(m.devices) == 0 {
		return nil
	}
	if m.cursor < 0 || m.cursor >= len(m.devices) {
		return nil
	}
	return &m.devices[m.cursor]
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.cursor
}

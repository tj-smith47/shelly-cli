// Package monitor provides real-time device monitoring for the TUI.
package monitor

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

// Deps holds the dependencies for the monitor component.
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

// DeviceStatus represents the real-time status of a device.
type DeviceStatus struct {
	Name      string
	Address   string
	Type      string
	Online    bool
	Power     float64
	Voltage   float64
	Current   float64
	Frequency float64
	UpdatedAt time.Time
	Error     error
}

// StatusUpdateMsg is sent when device status is updated.
type StatusUpdateMsg struct {
	Statuses []DeviceStatus
	Err      error
}

// RefreshTickMsg triggers periodic refresh.
type RefreshTickMsg struct{}

// Model holds the monitor state.
type Model struct {
	ctx             context.Context
	svc             *shelly.Service
	ios             *iostreams.IOStreams
	statuses        []DeviceStatus
	loading         bool
	err             error
	width           int
	height          int
	styles          Styles
	refreshInterval time.Duration
	scrollOffset    int
	cursor          int // Currently selected row
}

// Styles for the monitor component.
type Styles struct {
	Container    lipgloss.Style
	Header       lipgloss.Style
	Row          lipgloss.Style
	SelectedRow  lipgloss.Style
	OnlineIcon   lipgloss.Style
	OfflineIcon  lipgloss.Style
	UpdatingIcon lipgloss.Style
	DeviceName   lipgloss.Style
	Address      lipgloss.Style
	Power        lipgloss.Style
	Voltage      lipgloss.Style
	Current      lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	LastUpdated  lipgloss.Style
	Separator    lipgloss.Style
}

// DefaultStyles returns default styles for the monitor.
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
		Row: lipgloss.NewStyle().
			MarginBottom(1),
		SelectedRow: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true).
			MarginBottom(1),
		OnlineIcon: lipgloss.NewStyle().
			Foreground(colors.Online),
		OfflineIcon: lipgloss.NewStyle().
			Foreground(colors.Offline),
		UpdatingIcon: lipgloss.NewStyle().
			Foreground(colors.Updating),
		DeviceName: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Text),
		Address: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Power: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true),
		Voltage: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Current: lipgloss.NewStyle().
			Foreground(colors.Primary),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Width(8),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		LastUpdated: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Separator: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// New creates a new monitor model.
func New(deps Deps) Model {
	if err := deps.validate(); err != nil {
		panic(fmt.Sprintf("monitor: invalid deps: %v", err))
	}

	refreshInterval := deps.RefreshInterval
	if refreshInterval == 0 {
		refreshInterval = 5 * time.Second
	}

	return Model{
		ctx:             deps.Ctx,
		svc:             deps.Svc,
		ios:             deps.IOS,
		statuses:        []DeviceStatus{},
		loading:         true,
		styles:          DefaultStyles(),
		refreshInterval: refreshInterval,
	}
}

// Init returns the initial command for the monitor.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchStatuses(),
		m.scheduleRefresh(),
	)
}

// scheduleRefresh schedules the next refresh tick.
func (m Model) scheduleRefresh() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}

// fetchStatuses returns a command that fetches device statuses.
func (m Model) fetchStatuses() tea.Cmd {
	return func() tea.Msg {
		deviceMap := config.ListDevices()
		if len(deviceMap) == 0 {
			return StatusUpdateMsg{Statuses: nil}
		}

		// Add timeout to prevent slow/offline devices from blocking UI
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		// Use errgroup for concurrent status fetching
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(10)

		results := make(chan DeviceStatus, len(deviceMap))

		for _, d := range deviceMap {
			device := d
			g.Go(func() error {
				status := m.checkDeviceStatus(gctx, device)
				results <- status
				return nil
			})
		}

		go func() {
			if err := g.Wait(); err != nil {
				m.ios.DebugErr("monitor errgroup wait", err)
			}
			close(results)
		}()

		statuses := make([]DeviceStatus, 0, len(deviceMap))
		for status := range results {
			statuses = append(statuses, status)
		}

		// Sort by name for consistent display
		sort.Slice(statuses, func(i, j int) bool {
			return statuses[i].Name < statuses[j].Name
		})

		return StatusUpdateMsg{Statuses: statuses}
	}
}

// checkDeviceStatus fetches the real-time status of a single device.
func (m Model) checkDeviceStatus(ctx context.Context, device model.Device) DeviceStatus {
	status := DeviceStatus{
		Name:      device.Name,
		Address:   device.Address,
		Type:      device.Type,
		Online:    false,
		UpdatedAt: time.Now(),
	}

	// Per-device timeout to prevent single slow device from blocking others
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	snapshot, err := m.svc.GetMonitoringSnapshot(ctx, device.Address)
	if err != nil {
		status.Error = err
		return status
	}

	status.Online = true
	aggregateStatusFromPM(&status, snapshot.PM)
	aggregateStatusFromEM(&status, snapshot.EM)
	aggregateStatusFromEM1(&status, snapshot.EM1)

	return status
}

// aggregateStatusFromPM aggregates status data from PM components.
func aggregateStatusFromPM(status *DeviceStatus, pms []shelly.PMStatus) {
	for _, pm := range pms {
		status.Power += pm.APower
		if status.Voltage == 0 && pm.Voltage > 0 {
			status.Voltage = pm.Voltage
		}
		if status.Current == 0 && pm.Current > 0 {
			status.Current = pm.Current
		}
		if pm.Freq != nil && status.Frequency == 0 {
			status.Frequency = *pm.Freq
		}
	}
}

// aggregateStatusFromEM aggregates status data from EM components (3-phase meters).
func aggregateStatusFromEM(status *DeviceStatus, ems []shelly.EMStatus) {
	for _, em := range ems {
		status.Power += em.TotalActivePower
		status.Current += em.TotalCurrent
		if status.Voltage == 0 && em.AVoltage > 0 {
			status.Voltage = em.AVoltage
		}
		if em.AFreq != nil && status.Frequency == 0 {
			status.Frequency = *em.AFreq
		}
	}
}

// aggregateStatusFromEM1 aggregates status data from EM1 components (single-phase meters).
func aggregateStatusFromEM1(status *DeviceStatus, em1s []shelly.EM1Status) {
	for _, em1 := range em1s {
		status.Power += em1.ActPower
		if status.Voltage == 0 && em1.Voltage > 0 {
			status.Voltage = em1.Voltage
		}
		if status.Current == 0 && em1.Current > 0 {
			status.Current = em1.Current
		}
		if em1.Freq != nil && status.Frequency == 0 {
			status.Frequency = *em1.Freq
		}
	}
}

// Refresh returns a command to refresh the monitor.
func (m Model) Refresh() tea.Cmd {
	return m.fetchStatuses()
}

// Update handles messages for the monitor.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusUpdateMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.statuses = msg.Statuses
		return m, nil

	case RefreshTickMsg:
		return m, tea.Batch(
			m.fetchStatuses(),
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
	if m.cursor < len(m.statuses)-1 {
		m.cursor++
		visibleRows := m.visibleRows()
		if m.cursor >= m.scrollOffset+visibleRows {
			m.scrollOffset = m.cursor - visibleRows + 1
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
	if len(m.statuses) > 0 {
		m.cursor = len(m.statuses) - 1
		maxOffset := len(m.statuses) - m.visibleRows()
		if maxOffset < 0 {
			maxOffset = 0
		}
		m.scrollOffset = maxOffset
	}
	return m
}

func (m Model) pageDown() Model {
	if len(m.statuses) == 0 {
		return m
	}
	visibleRows := m.visibleRows()
	m.cursor += visibleRows
	if m.cursor >= len(m.statuses) {
		m.cursor = len(m.statuses) - 1
	}
	if m.cursor >= m.scrollOffset+visibleRows {
		m.scrollOffset = m.cursor - visibleRows + 1
	}
	maxOffset := len(m.statuses) - visibleRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	return m
}

func (m Model) pageUp() Model {
	if len(m.statuses) == 0 {
		return m
	}
	visibleRows := m.visibleRows()
	m.cursor -= visibleRows
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	return m
}

// SetSize sets the component size.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// visibleRows calculates how many device rows can be displayed.
// Each device card takes 3 lines (2 content + margin) + 1 separator = 4 lines total.
func (m Model) visibleRows() int {
	// Account for: header (2), footer (2), container padding (2) = 6 lines overhead
	availableHeight := m.height - 6
	// Each card: 2 lines content + 1 margin + 1 separator = 4 lines
	// Last card has no separator, so for N cards: 4N - 1 lines
	// Solving for N: N = (availableHeight + 1) / 4
	rowHeight := 4
	visibleRows := (availableHeight + 1) / rowHeight
	if visibleRows < 1 {
		return 1
	}
	return visibleRows
}

// View renders the monitor.
func (m Model) View() string {
	if m.loading {
		loadingText := m.styles.UpdatingIcon.Render("◐ ") + "Fetching device statuses..."
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

	if len(m.statuses) == 0 {
		return m.styles.Container.
			Width(m.width-4).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No devices to monitor.\nUse 'shelly device add' to add devices.")
	}

	// Build the monitor view
	header := m.styles.Header.Render("Real-Time Device Monitor")

	// Calculate visible rows
	visibleRows := m.visibleRows()

	// Apply scroll offset
	startIdx := m.scrollOffset
	endIdx := startIdx + visibleRows
	if endIdx > len(m.statuses) {
		endIdx = len(m.statuses)
	}

	var rows string
	for i := startIdx; i < endIdx; i++ {
		s := m.statuses[i]
		isSelected := i == m.cursor
		rows += m.renderDeviceCard(s, isSelected)
		if i < endIdx-1 {
			rows += m.styles.Separator.Render("─────────────────────────────────────────") + "\n"
		}
	}

	// Scroll indicator
	scrollInfo := ""
	if len(m.statuses) > visibleRows {
		scrollInfo = m.styles.LastUpdated.Render(
			fmt.Sprintf(" [%d-%d of %d] ", startIdx+1, endIdx, len(m.statuses)),
		)
	}

	footer := m.styles.LastUpdated.Render(
		fmt.Sprintf("Last updated: %s", time.Now().Format("15:04:05")),
	) + scrollInfo

	content := lipgloss.JoinVertical(lipgloss.Left, header, rows, footer)
	return m.styles.Container.Width(m.width - 4).Render(content)
}

// renderDeviceCard renders a single device status card.
func (m Model) renderDeviceCard(s DeviceStatus, isSelected bool) string {
	// Choose row style based on selection
	rowStyle := m.styles.Row
	if isSelected {
		rowStyle = m.styles.SelectedRow
	}

	// Selection indicator
	selIndicator := "  "
	if isSelected {
		selIndicator = "▶ "
	}

	statusIcon := m.styles.OfflineIcon.Render("○")
	if s.Online {
		statusIcon = m.styles.OnlineIcon.Render("●")
	}

	// First line: selection indicator, icon, name, address, type
	line1 := fmt.Sprintf("%s%s %s  %s  %s",
		selIndicator,
		statusIcon,
		m.styles.DeviceName.Render(s.Name),
		m.styles.Address.Render(s.Address),
		m.styles.Address.Render(s.Type),
	)

	if !s.Online {
		errMsg := "unreachable"
		if s.Error != nil {
			errMsg = s.Error.Error()
			if len(errMsg) > 40 {
				errMsg = errMsg[:40] + "..."
			}
		}
		line2 := "    " + theme.StatusError().Render(errMsg)
		return rowStyle.Render(line1+"\n"+line2) + "\n"
	}

	// Second line: metrics
	metrics := []string{}

	if s.Power > 0 || s.Power < 0 {
		metrics = append(metrics, m.styles.Power.Render(fmt.Sprintf("%.1fW", s.Power)))
	}
	if s.Voltage > 0 {
		metrics = append(metrics, m.styles.Voltage.Render(fmt.Sprintf("%.1fV", s.Voltage)))
	}
	if s.Current > 0 {
		metrics = append(metrics, m.styles.Current.Render(fmt.Sprintf("%.2fA", s.Current)))
	}
	if s.Frequency > 0 {
		metrics = append(metrics, m.styles.Value.Render(fmt.Sprintf("%.1fHz", s.Frequency)))
	}

	line2 := "    "
	if len(metrics) > 0 {
		for i, metric := range metrics {
			if i > 0 {
				line2 += m.styles.Separator.Render(" │ ")
			}
			line2 += metric
		}
	} else {
		line2 += m.styles.LastUpdated.Render("no power data")
	}

	return rowStyle.Render(line1+"\n"+line2) + "\n"
}

// StatusCount returns the count of online/offline devices.
func (m Model) StatusCount() (online, offline int) {
	for _, s := range m.statuses {
		if s.Online {
			online++
		} else {
			offline++
		}
	}
	return online, offline
}

// SelectedDevice returns the currently selected device, if any.
func (m Model) SelectedDevice() *DeviceStatus {
	if len(m.statuses) == 0 {
		return nil
	}
	if m.cursor < 0 || m.cursor >= len(m.statuses) {
		return nil
	}
	return &m.statuses[m.cursor]
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.cursor
}

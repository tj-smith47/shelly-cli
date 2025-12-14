// Package devicelist provides the device list component for the TUI.
package devicelist

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Deps holds the dependencies for the device list component.
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

// DeviceStatus represents a device with its live status.
type DeviceStatus struct {
	Device  model.Device
	Online  bool
	Power   float64
	Voltage float64
	Error   error
}

// DevicesLoadedMsg signals that devices were loaded.
type DevicesLoadedMsg struct {
	Devices []DeviceStatus
	Err     error
}

// RefreshTickMsg triggers periodic refresh.
type RefreshTickMsg struct{}

// Model holds the device list state.
type Model struct {
	ctx             context.Context
	svc             *shelly.Service
	ios             *iostreams.IOStreams
	table           table.Model
	devices         []DeviceStatus // All devices
	filtered        []DeviceStatus // Filtered devices for display
	filter          string         // Current filter string
	loading         bool
	err             error
	width           int
	height          int
	styles          Styles
	refreshInterval time.Duration
}

// Styles for the device list component.
type Styles struct {
	Table       lipgloss.Style
	Online      lipgloss.Style
	Offline     lipgloss.Style
	StatusOK    lipgloss.Style
	StatusError lipgloss.Style
}

// DefaultStyles returns default styles for the device list.
func DefaultStyles() Styles {
	return Styles{
		Table: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.BrightBlack()),
		Online: lipgloss.NewStyle().
			Foreground(theme.Green()),
		Offline: lipgloss.NewStyle().
			Foreground(theme.Red()),
		StatusOK: lipgloss.NewStyle().
			Foreground(theme.Green()),
		StatusError: lipgloss.NewStyle().
			Foreground(theme.Red()),
	}
}

// New creates a new device list model.
func New(deps Deps) Model {
	if err := deps.validate(); err != nil {
		panic(fmt.Sprintf("devicelist: invalid deps: %v", err))
	}

	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Address", Width: 16},
		{Title: "Gen", Width: 5},
		{Title: "Type", Width: 15},
		{Title: "Status", Width: 10},
		{Title: "Power", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.BrightBlack()).
		BorderBottom(true).
		Bold(true).
		Foreground(theme.Cyan())
	s.Selected = s.Selected.
		Foreground(theme.Fg()).
		Background(theme.BrightBlack()).
		Bold(true)
	t.SetStyles(s)

	refreshInterval := deps.RefreshInterval
	if refreshInterval == 0 {
		refreshInterval = 5 * time.Second
	}

	return Model{
		ctx:             deps.Ctx,
		svc:             deps.Svc,
		ios:             deps.IOS,
		table:           t,
		loading:         true,
		styles:          DefaultStyles(),
		refreshInterval: refreshInterval,
	}
}

// Init returns the initial command to fetch devices.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchDevices(),
		m.scheduleRefresh(),
	)
}

// scheduleRefresh schedules the next refresh tick.
func (m Model) scheduleRefresh() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(time.Time) tea.Msg {
		return RefreshTickMsg{}
	})
}

// fetchDevices returns a command that loads devices and their status.
func (m Model) fetchDevices() tea.Cmd {
	return func() tea.Msg {
		deviceMap := config.ListDevices()
		if len(deviceMap) == 0 {
			return DevicesLoadedMsg{Devices: nil}
		}

		// Use errgroup to fetch status concurrently
		g, ctx := errgroup.WithContext(m.ctx)
		g.SetLimit(10)

		results := make(chan DeviceStatus, len(deviceMap))

		for _, d := range deviceMap {
			device := d
			g.Go(func() error {
				status := m.checkDeviceStatus(ctx, device)
				results <- status
				return nil
			})
		}

		// Wait for all goroutines and close channel
		go func() {
			if err := g.Wait(); err != nil {
				m.ios.DebugErr("devicelist errgroup wait", err)
			}
			close(results)
		}()

		// Collect results
		devices := make([]DeviceStatus, 0, len(deviceMap))
		for status := range results {
			devices = append(devices, status)
		}

		// Sort by name for consistent display
		sort.Slice(devices, func(i, j int) bool {
			return devices[i].Device.Name < devices[j].Device.Name
		})

		return DevicesLoadedMsg{Devices: devices}
	}
}

// checkDeviceStatus checks the status of a single device.
func (m Model) checkDeviceStatus(ctx context.Context, device model.Device) DeviceStatus {
	status := DeviceStatus{
		Device: device,
		Online: false,
	}

	snapshot, err := m.svc.GetMonitoringSnapshot(ctx, device.Address)
	if err != nil {
		status.Error = err
		return status
	}

	status.Online = true

	// Sum up power from all PM components
	for _, pm := range snapshot.PM {
		status.Power += pm.APower
	}

	// Also check EM components (3-phase meters)
	for _, em := range snapshot.EM {
		status.Power += em.TotalActivePower
	}

	// Also check EM1 components (single-phase meters)
	for _, em1 := range snapshot.EM1 {
		status.Power += em1.ActPower
	}

	return status
}

// Refresh returns a command to refresh the device list.
func (m Model) Refresh() tea.Cmd {
	return m.fetchDevices()
}

// Update handles messages for the device list.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DevicesLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.devices = msg.Devices
		m = m.applyFilter()
		return m, nil

	case RefreshTickMsg:
		return m, tea.Batch(
			m.fetchDevices(),
			m.scheduleRefresh(),
		)

	case tea.KeyPressMsg:
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// SetFilter sets the filter string and re-applies it to the device list.
func (m Model) SetFilter(filter string) Model {
	m.filter = filter
	return m.applyFilter()
}

// applyFilter filters the device list based on the current filter string.
func (m Model) applyFilter() Model {
	if m.filter == "" {
		m.filtered = m.devices
	} else {
		filterLower := strings.ToLower(m.filter)
		m.filtered = make([]DeviceStatus, 0, len(m.devices))
		for _, d := range m.devices {
			// Match against name, address, type, or model
			if strings.Contains(strings.ToLower(d.Device.Name), filterLower) ||
				strings.Contains(strings.ToLower(d.Device.Address), filterLower) ||
				strings.Contains(strings.ToLower(d.Device.Type), filterLower) ||
				strings.Contains(strings.ToLower(d.Device.Model), filterLower) {
				m.filtered = append(m.filtered, d)
			}
		}
	}
	m.table.SetRows(m.devicesToRows())
	return m
}

// devicesToRows converts filtered devices to table rows.
func (m Model) devicesToRows() []table.Row {
	rows := make([]table.Row, 0, len(m.filtered))
	for _, d := range m.filtered {
		gen := "?"
		if d.Device.Generation > 0 {
			gen = fmt.Sprintf("%d", d.Device.Generation)
		}

		var statusStr string
		switch {
		case d.Error != nil:
			statusStr = m.styles.Offline.Render("offline")
		case d.Online:
			statusStr = m.styles.Online.Render("online")
		default:
			statusStr = m.styles.Offline.Render("offline")
		}

		powerStr := "-"
		if d.Online && d.Power > 0 {
			powerStr = fmt.Sprintf("%.1fW", d.Power)
		}

		rows = append(rows, table.Row{
			d.Device.Name,
			d.Device.Address,
			gen,
			d.Device.Type,
			statusStr,
			powerStr,
		})
	}
	return rows
}

// SetSize sets the component size.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	m.table.SetWidth(width - 2)
	m.table.SetHeight(height - 2)
	return m
}

// View renders the device list.
func (m Model) View() string {
	if m.loading {
		return m.styles.Table.
			Width(m.width-2).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Fetching device status...")
	}

	if m.err != nil {
		return m.styles.StatusError.Render("Error: " + m.err.Error())
	}

	if len(m.devices) == 0 {
		return m.styles.Table.
			Width(m.width-2).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No devices registered.\nUse 'shelly device add' to add devices.")
	}

	if len(m.filtered) == 0 && m.filter != "" {
		return m.styles.Table.
			Width(m.width-2).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("No devices match filter %q.\nPress / to clear or modify filter.", m.filter))
	}

	return m.styles.Table.Render(m.table.View())
}

// SelectedDevice returns the currently selected device, if any.
func (m Model) SelectedDevice() *DeviceStatus {
	if len(m.filtered) == 0 {
		return nil
	}

	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.filtered) {
		return nil
	}

	return &m.filtered[idx]
}

// DeviceCount returns the number of filtered devices.
func (m Model) DeviceCount() int {
	return len(m.filtered)
}

// TotalDeviceCount returns the total number of devices (before filtering).
func (m Model) TotalDeviceCount() int {
	return len(m.devices)
}

// Filter returns the current filter string.
func (m Model) Filter() string {
	return m.filter
}

// Package firmware provides TUI components for firmware management.
package firmware

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// Deps holds the dependencies for the Firmware component.
type Deps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d Deps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	return nil
}

// DeviceFirmware represents firmware status for a device.
type DeviceFirmware struct {
	Name      string
	Address   string
	Current   string
	Available string
	HasUpdate bool
	Updating  bool
	Selected  bool
	Checked   bool
	Err       error
}

// CheckCompleteMsg signals that firmware check completed.
type CheckCompleteMsg struct {
	Results []DeviceFirmware
}

// UpdateCompleteMsg signals that firmware update completed.
type UpdateCompleteMsg struct {
	Name    string
	Success bool
	Err     error
}

// Model displays firmware management.
type Model struct {
	ctx      context.Context
	svc      *shelly.Service
	devices  []DeviceFirmware
	cursor   int
	scroll   int
	checking bool
	updating bool
	err      error
	width    int
	height   int
	focused  bool
	styles   Styles
}

// Styles holds styles for the Firmware component.
type Styles struct {
	HasUpdate lipgloss.Style
	UpToDate  lipgloss.Style
	Unknown   lipgloss.Style
	Updating  lipgloss.Style
	Selected  lipgloss.Style
	Cursor    lipgloss.Style
	Label     lipgloss.Style
	Error     lipgloss.Style
	Muted     lipgloss.Style
	Button    lipgloss.Style
	Version   lipgloss.Style
}

// DefaultStyles returns the default styles for the Firmware component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		HasUpdate: lipgloss.NewStyle().
			Foreground(colors.Warning),
		UpToDate: lipgloss.NewStyle().
			Foreground(colors.Online),
		Unknown: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Updating: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Selected: lipgloss.NewStyle().
			Foreground(colors.Online),
		Cursor: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Button: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Version: lipgloss.NewStyle().
			Foreground(colors.Text),
	}
}

// New creates a new Firmware model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("firmware: invalid deps: %v", err))
	}

	return Model{
		ctx:    deps.Ctx,
		svc:    deps.Svc,
		styles: DefaultStyles(),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// LoadDevices loads registered devices.
func (m Model) LoadDevices() Model {
	cfg := config.Get()
	if cfg == nil {
		return m
	}

	m.devices = make([]DeviceFirmware, 0, len(cfg.Devices))
	for name, dev := range cfg.Devices {
		m.devices = append(m.devices, DeviceFirmware{
			Name:    name,
			Address: dev.Address,
		})
	}

	return m
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focus state.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// CheckAll starts a firmware check on all devices.
func (m Model) CheckAll() (Model, tea.Cmd) {
	if m.checking || len(m.devices) == 0 {
		return m, nil
	}

	m.checking = true
	m.err = nil
	return m, m.checkAllDevices()
}

func (m Model) checkAllDevices() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 120*time.Second)
		defer cancel()

		var (
			results []DeviceFirmware
			mu      sync.Mutex
		)

		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(5)

		for _, dev := range m.devices {
			device := dev
			g.Go(func() error {
				info, checkErr := m.svc.CheckFirmware(gctx, device.Name)
				mu.Lock()
				result := DeviceFirmware{
					Name:    device.Name,
					Address: device.Address,
					Checked: true,
					Err:     checkErr,
				}
				if info != nil {
					result.Current = info.Current
					result.Available = info.Available
					result.HasUpdate = info.HasUpdate
				}
				results = append(results, result)
				mu.Unlock()
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			_ = err // Individual errors captured per device
		}

		return CheckCompleteMsg{Results: results}
	}
}

// UpdateSelected starts firmware update on selected devices.
func (m Model) UpdateSelected() (Model, tea.Cmd) {
	if m.updating || m.checking {
		return m, nil
	}

	selected := m.selectedDevices()
	if len(selected) == 0 {
		m.err = fmt.Errorf("no devices selected")
		return m, nil
	}

	m.updating = true
	m.err = nil

	// Mark selected devices as updating
	for i := range m.devices {
		if m.devices[i].Selected && m.devices[i].HasUpdate {
			m.devices[i].Updating = true
		}
	}

	return m, m.updateDevices(selected)
}

func (m Model) updateDevices(devices []DeviceFirmware) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 300*time.Second)
		defer cancel()

		// Update devices sequentially to avoid overwhelming the network
		for _, dev := range devices {
			if !dev.HasUpdate {
				continue
			}

			err := m.svc.UpdateFirmwareStable(ctx, dev.Name)
			// We send individual updates - though for TUI we'll just send final result
			_ = err
		}

		// Return a message indicating updates are done
		return updateBatchComplete{}
	}
}

type updateBatchComplete struct{}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CheckCompleteMsg:
		m.checking = false
		// Merge results with existing devices
		for _, result := range msg.Results {
			for i := range m.devices {
				if m.devices[i].Name == result.Name {
					m.devices[i] = result
					break
				}
			}
		}
		return m, nil

	case UpdateCompleteMsg:
		for i := range m.devices {
			if m.devices[i].Name == msg.Name {
				m.devices[i].Updating = false
				if msg.Success {
					m.devices[i].HasUpdate = false
				} else {
					m.devices[i].Err = msg.Err
				}
				break
			}
		}
		return m, nil

	case updateBatchComplete:
		m.updating = false
		// Clear updating status
		for i := range m.devices {
			m.devices[i].Updating = false
			m.devices[i].Selected = false
		}
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m = m.cursorDown()
	case "k", "up":
		m = m.cursorUp()
	case "space":
		m = m.toggleSelection()
	case "a":
		m = m.selectAllWithUpdates()
	case "n":
		m = m.selectNone()
	case "c", "r":
		return m.CheckAll()
	case "u", "enter":
		return m.UpdateSelected()
	}

	return m, nil
}

func (m Model) cursorDown() Model {
	if m.cursor < len(m.devices)-1 {
		m.cursor++
		m = m.ensureVisible()
	}
	return m
}

func (m Model) cursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
		m = m.ensureVisible()
	}
	return m
}

func (m Model) ensureVisible() Model {
	visible := m.visibleRows()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	} else if m.cursor >= m.scroll+visible {
		m.scroll = m.cursor - visible + 1
	}
	return m
}

func (m Model) visibleRows() int {
	rows := m.height - 10
	if rows < 1 {
		return 1
	}
	return rows
}

func (m Model) toggleSelection() Model {
	if len(m.devices) > 0 && m.cursor < len(m.devices) {
		m.devices[m.cursor].Selected = !m.devices[m.cursor].Selected
	}
	return m
}

func (m Model) selectAllWithUpdates() Model {
	for i := range m.devices {
		if m.devices[i].HasUpdate {
			m.devices[i].Selected = true
		}
	}
	return m
}

func (m Model) selectNone() Model {
	for i := range m.devices {
		m.devices[i].Selected = false
	}
	return m
}

func (m Model) selectedDevices() []DeviceFirmware {
	selected := make([]DeviceFirmware, 0)
	for _, d := range m.devices {
		if d.Selected {
			selected = append(selected, d)
		}
	}
	return selected
}

// View renders the Firmware component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Firmware").
		SetFocused(m.focused)

	var content strings.Builder

	// Action buttons
	content.WriteString(m.renderActions())
	content.WriteString("\n\n")

	// Device list
	if len(m.devices) == 0 {
		content.WriteString(m.styles.Muted.Render("No devices registered"))
	} else {
		content.WriteString(m.renderDeviceList())
	}

	// Error display with categorized messaging and retry hint
	if m.err != nil {
		msg, hint := tuierrors.FormatError(m.err)
		content.WriteString("\n")
		content.WriteString(m.styles.Error.Render(msg))
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render("  " + hint))
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render("  Press 'r' to retry"))
	}

	// Status indicator
	if m.checking {
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render("Checking firmware..."))
	} else if m.updating {
		content.WriteString("\n")
		content.WriteString(m.styles.Updating.Render("Updating devices..."))
	}

	// Help text
	content.WriteString("\n\n")
	content.WriteString(m.styles.Muted.Render("space: select | a: all updates | c: check | u: update"))

	r.SetContent(content.String())
	return r.Render()
}

func (m Model) renderActions() string {
	parts := make([]string, 0, 2)

	// Check button
	checkStyle := m.styles.Button
	if m.checking {
		checkStyle = m.styles.Muted
	}
	parts = append(parts, checkStyle.Render("[c] Check All"))

	// Update button
	updateStyle := m.styles.Button
	selectedCount := len(m.selectedDevices())
	if m.updating || selectedCount == 0 {
		updateStyle = m.styles.Muted
	}
	parts = append(parts, updateStyle.Render(fmt.Sprintf("[u] Update (%d)", selectedCount)))

	return strings.Join(parts, "  ")
}

func (m Model) renderDeviceList() string {
	var content strings.Builder

	// Summary
	updateCount := 0
	for _, d := range m.devices {
		if d.HasUpdate {
			updateCount++
		}
	}
	content.WriteString(m.styles.Label.Render(
		fmt.Sprintf("Devices (%d with updates):", updateCount),
	))
	content.WriteString("\n\n")

	visible := m.visibleRows()
	endIdx := m.scroll + visible
	if endIdx > len(m.devices) {
		endIdx = len(m.devices)
	}

	for i := m.scroll; i < endIdx; i++ {
		device := m.devices[i]
		isCursor := i == m.cursor
		content.WriteString(m.renderDeviceLine(device, isCursor))
		if i < endIdx-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(m.devices) > visible {
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render(
			fmt.Sprintf("[%d/%d]", m.cursor+1, len(m.devices)),
		))
	}

	return content.String()
}

func (m Model) renderDeviceLine(device DeviceFirmware, isCursor bool) string {
	cursor := "  "
	if isCursor {
		cursor = "▶ "
	}

	// Checkbox
	var checkbox string
	if device.Selected {
		checkbox = m.styles.Selected.Render("[✓]")
	} else {
		checkbox = m.styles.Muted.Render("[ ]")
	}

	// Status indicator
	var status string
	switch {
	case device.Updating:
		status = m.styles.Updating.Render("↻")
	case device.Err != nil:
		status = m.styles.Error.Render("✗")
	case !device.Checked:
		status = m.styles.Unknown.Render("?")
	case device.HasUpdate:
		status = m.styles.HasUpdate.Render("↑")
	default:
		status = m.styles.UpToDate.Render("✓")
	}

	// Name
	nameStr := device.Name

	// Version info
	var versionStr string
	if device.Checked && device.Err == nil {
		if device.HasUpdate {
			versionStr = m.styles.Version.Render(device.Current) +
				m.styles.Muted.Render(" → ") +
				m.styles.HasUpdate.Render(device.Available)
		} else if device.Current != "" {
			versionStr = m.styles.UpToDate.Render(device.Current)
		}
	} else if device.Err != nil {
		versionStr = m.styles.Error.Render("error")
	}

	line := fmt.Sprintf("%s%s %s %s", cursor, checkbox, status, nameStr)
	if versionStr != "" {
		line += " " + versionStr
	}

	if isCursor {
		return m.styles.Cursor.Render(line)
	}
	return line
}

// Devices returns the device list.
func (m Model) Devices() []DeviceFirmware {
	return m.devices
}

// Checking returns whether a check is in progress.
func (m Model) Checking() bool {
	return m.checking
}

// Updating returns whether updates are in progress.
func (m Model) Updating() bool {
	return m.updating
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.cursor
}

// SelectedCount returns the number of selected devices.
func (m Model) SelectedCount() int {
	count := 0
	for _, d := range m.devices {
		if d.Selected {
			count++
		}
	}
	return count
}

// UpdateCount returns the number of devices with updates.
func (m Model) UpdateCount() int {
	count := 0
	for _, d := range m.devices {
		if d.HasUpdate {
			count++
		}
	}
	return count
}

// Refresh reloads devices and clears state.
func (m Model) Refresh() (Model, tea.Cmd) {
	m.err = nil
	m = m.LoadDevices()
	return m.CheckAll()
}

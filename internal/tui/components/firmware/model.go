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
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
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
	ctx          context.Context
	svc          *shelly.Service
	devices      []DeviceFirmware
	scroller     *panel.Scroller
	checking     bool
	updating     bool
	err          error
	width        int
	height       int
	focused      bool
	panelIndex   int
	styles       Styles
	checkLoader  loading.Model
	updateLoader loading.Model
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
		ctx:      deps.Ctx,
		svc:      deps.Svc,
		scroller: panel.NewScroller(0, 10),
		styles:   DefaultStyles(),
		checkLoader: loading.New(
			loading.WithMessage("Checking firmware..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(false, false),
		),
		updateLoader: loading.New(
			loading.WithMessage("Updating devices..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(false, false),
		),
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

	m.scroller.SetItemCount(len(m.devices))
	return m
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	visibleRows := height - 10
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
	// Update loader sizes
	m.checkLoader = m.checkLoader.SetSize(width-4, height-4)
	m.updateLoader = m.updateLoader.SetSize(width-4, height-4)
	return m
}

// SetFocused sets the focus state.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return m
}

// CheckAll starts a firmware check on all devices.
func (m Model) CheckAll() (Model, tea.Cmd) {
	if m.checking || len(m.devices) == 0 {
		return m, nil
	}

	m.checking = true
	m.err = nil
	return m, tea.Batch(m.checkLoader.Tick(), m.checkAllDevices())
}

func (m Model) checkAllDevices() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 120*time.Second)
		defer cancel()

		var (
			results []DeviceFirmware
			mu      sync.Mutex
		)

		// Rate limiting is handled at the service layer
		g, gctx := errgroup.WithContext(ctx)

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
			// Individual errors are captured per device in results
			iostreams.DebugErr("firmware check batch", err)
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

	return m, tea.Batch(m.updateLoader.Tick(), m.updateDevices(selected))
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
			if err != nil {
				// Log but continue with remaining devices
				iostreams.DebugErr("firmware update "+dev.Name, err)
			}
		}

		// Return a message indicating updates are done
		return updateBatchComplete{}
	}
}

type updateBatchComplete struct{}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Forward tick messages to the appropriate loader
	var cmd tea.Cmd
	m, cmd = m.updateLoaders(msg)
	if cmd != nil {
		return m, cmd
	}

	switch msg := msg.(type) {
	case CheckCompleteMsg:
		return m.handleCheckComplete(msg), nil
	case UpdateCompleteMsg:
		return m.handleUpdateComplete(msg), nil
	case updateBatchComplete:
		return m.handleBatchComplete(), nil
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

// updateLoaders forwards tick messages to the appropriate loader.
// Returns updated model and command if the message was consumed.
func (m Model) updateLoaders(msg tea.Msg) (Model, tea.Cmd) {
	if m.checking {
		var cmd tea.Cmd
		m.checkLoader, cmd = m.checkLoader.Update(msg)
		if _, ok := msg.(CheckCompleteMsg); !ok && cmd != nil {
			return m, cmd
		}
	}
	if m.updating {
		var cmd tea.Cmd
		m.updateLoader, cmd = m.updateLoader.Update(msg)
		switch msg.(type) {
		case updateBatchComplete, UpdateCompleteMsg:
			// Pass through
		default:
			if cmd != nil {
				return m, cmd
			}
		}
	}
	return m, nil
}

func (m Model) handleCheckComplete(msg CheckCompleteMsg) Model {
	m.checking = false
	for _, result := range msg.Results {
		for i := range m.devices {
			if m.devices[i].Name == result.Name {
				m.devices[i] = result
				break
			}
		}
	}
	return m
}

func (m Model) handleUpdateComplete(msg UpdateCompleteMsg) Model {
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
	return m
}

func (m Model) handleBatchComplete() Model {
	m.updating = false
	for i := range m.devices {
		m.devices[i].Updating = false
		m.devices[i].Selected = false
	}
	return m
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.scroller.CursorDown()
	case "k", "up":
		m.scroller.CursorUp()
	case "g":
		m.scroller.CursorToStart()
	case "G":
		m.scroller.CursorToEnd()
	case "ctrl+d", "pgdown":
		m.scroller.PageDown()
	case "ctrl+u", "pgup":
		m.scroller.PageUp()
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

func (m Model) toggleSelection() Model {
	cursor := m.scroller.Cursor()
	if len(m.devices) > 0 && cursor < len(m.devices) {
		m.devices[cursor].Selected = !m.devices[cursor].Selected
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
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused {
		r.SetFooter("c:check u:update spc:sel a:all")
	}

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
		content.WriteString(m.checkLoader.View())
	} else if m.updating {
		content.WriteString("\n")
		content.WriteString(m.updateLoader.View())
	}

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

	start, end := m.scroller.VisibleRange()
	for i := start; i < end; i++ {
		device := m.devices[i]
		isCursor := m.scroller.IsCursorAt(i)
		content.WriteString(m.renderDeviceLine(device, isCursor))
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))

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
	return m.scroller.Cursor()
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

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll g/G:top/bottom enter:update u:update-all"
}

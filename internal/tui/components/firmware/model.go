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

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cachestatus"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// Deps holds the dependencies for the Firmware component.
type Deps struct {
	Ctx       context.Context
	Svc       *shelly.Service
	FileCache *cache.FileCache
}

// Validate ensures all required dependencies are set.
func (d Deps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	// FileCache is optional - caching is disabled if nil
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

// IsSelected implements generics.Selectable.
func (d *DeviceFirmware) IsSelected() bool { return d.Selected }

// SetSelected implements generics.Selectable.
func (d *DeviceFirmware) SetSelected(v bool) { d.Selected = v }

// Selection helpers for value slices.
func deviceFirmwareGet(d *DeviceFirmware) bool       { return d.Selected }
func deviceFirmwareSet(d *DeviceFirmware, v bool)    { d.Selected = v }
func deviceFirmwareHasUpdate(d *DeviceFirmware) bool { return d.HasUpdate }

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
	fileCache    *cache.FileCache
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
	cacheStatus  cachestatus.Model
	lastChecked  time.Time
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
			Foreground(colors.Text),
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
		iostreams.DebugErr("firmware component init", err)
		panic(fmt.Sprintf("firmware: invalid deps: %v", err))
	}

	return Model{
		ctx:       deps.Ctx,
		svc:       deps.Svc,
		fileCache: deps.FileCache,
		scroller:  panel.NewScroller(0, 10),
		styles:    DefaultStyles(),
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
		cacheStatus: cachestatus.New(),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// LoadDevices loads registered devices and their cached firmware status.
func (m Model) LoadDevices() Model {
	cfg := config.Get()
	if cfg == nil {
		return m
	}

	m.devices = make([]DeviceFirmware, 0, len(cfg.Devices))
	var oldestCache time.Time

	for name, dev := range cfg.Devices {
		df := DeviceFirmware{
			Name:    name,
			Address: dev.Address,
		}

		// Try to load cached firmware status
		if cached := m.loadCachedFirmware(name); cached != nil {
			df.Current = cached.Current
			df.Available = cached.Available
			df.HasUpdate = cached.HasUpdate
			df.Checked = true

			if oldestCache.IsZero() || cached.CachedAt.Before(oldestCache) {
				oldestCache = cached.CachedAt
			}
		}

		m.devices = append(m.devices, df)
	}

	// Set cache status if we loaded any cached data
	if !oldestCache.IsZero() {
		m.lastChecked = oldestCache
		m.cacheStatus = m.cacheStatus.SetUpdatedAt(oldestCache)
	}

	m.scroller.SetItemCount(len(m.devices))
	return m
}

// cachedFirmwareInfo holds cached firmware info with metadata.
type cachedFirmwareInfo struct {
	Current   string
	Available string
	HasUpdate bool
	CachedAt  time.Time
}

// loadCachedFirmware loads firmware info from file cache for a device.
func (m Model) loadCachedFirmware(name string) *cachedFirmwareInfo {
	if m.fileCache == nil {
		return nil
	}
	entry, err := m.fileCache.Get(name, cache.TypeFirmware)
	if err != nil || entry == nil {
		return nil
	}
	var info shelly.FirmwareInfo
	if err := entry.Unmarshal(&info); err != nil {
		return nil
	}
	return &cachedFirmwareInfo{
		Current:   info.Current,
		Available: info.Available,
		HasUpdate: info.HasUpdate,
		CachedAt:  entry.CachedAt,
	}
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
		// Don't use errgroup.WithContext - it would cancel all on first error
		var g errgroup.Group

		for _, dev := range m.devices {
			device := dev
			g.Go(func() error {
				// Per-device timeout to avoid blocking on unreachable devices
				deviceCtx, deviceCancel := context.WithTimeout(ctx, 15*time.Second)
				defer deviceCancel()

				info, checkErr := m.svc.CheckFirmware(deviceCtx, device.Name)
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

			// Per-device timeout to avoid blocking on unreachable devices
			// Firmware updates can take longer, so use 60 seconds
			deviceCtx, deviceCancel := context.WithTimeout(ctx, 60*time.Second)
			err := m.svc.UpdateFirmwareStable(deviceCtx, dev.Name)
			deviceCancel()
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
	// Update cache status spinner
	if m.cacheStatus.IsRefreshing() {
		var cmd tea.Cmd
		m.cacheStatus, cmd = m.cacheStatus.Update(msg)
		if cmd != nil {
			return m, cmd
		}
	}
	return m, nil
}

func (m Model) handleCheckComplete(msg CheckCompleteMsg) Model {
	m.checking = false
	m.lastChecked = time.Now()
	m.cacheStatus = m.cacheStatus.SetUpdatedAt(m.lastChecked)

	for _, result := range msg.Results {
		for i := range m.devices {
			if m.devices[i].Name == result.Name {
				m.devices[i] = result

				// Cache successful results to file cache
				if m.fileCache != nil && result.Err == nil && result.Checked {
					info := shelly.FirmwareInfo{
						Current:   result.Current,
						Available: result.Available,
						HasUpdate: result.HasUpdate,
					}
					if err := m.fileCache.Set(result.Name, cache.TypeFirmware, info, cache.TTLFirmware); err != nil {
						iostreams.DebugErr("cache firmware "+result.Name, err)
					}
				}
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
		if m.devices[i].Updating {
			// Invalidate cache for updated devices
			if m.fileCache != nil {
				iostreams.DebugErr("invalidate firmware cache "+m.devices[i].Name,
					m.fileCache.Invalidate(m.devices[i].Name, cache.TypeFirmware))
			}
		}
		m.devices[i].Updating = false
		m.devices[i].Selected = false
	}
	return m
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Handle navigation keys first
	if keys.HandleScrollNavigation(msg.String(), m.scroller) {
		return m, nil
	}

	// Handle action keys
	switch msg.String() {
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
	generics.ToggleAtFunc(m.devices, m.scroller.Cursor(), deviceFirmwareGet, deviceFirmwareSet)
	return m
}

func (m Model) selectAllWithUpdates() Model {
	generics.SelectWhereFunc(m.devices, deviceFirmwareHasUpdate, deviceFirmwareSet)
	return m
}

func (m Model) selectNone() Model {
	generics.SelectNoneFunc(m.devices, deviceFirmwareSet)
	return m
}

func (m Model) selectedDevices() []DeviceFirmware {
	return generics.Filter(m.devices, func(d DeviceFirmware) bool { return d.Selected })
}

// View renders the Firmware component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Firmware").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings and cache status when focused
	if m.focused {
		footer := "c:check u:update spc:sel a:all r:refresh"
		if cs := m.cacheStatus.View(); cs != "" {
			footer += " | " + cs
		}
		r.SetFooter(footer)
	}

	var content strings.Builder

	// Action buttons
	content.WriteString(m.renderActions())
	content.WriteString("\n")

	// Status indicator (under actions, above device list)
	if m.checking {
		content.WriteString(m.checkLoader.View())
		content.WriteString("\n")
	} else if m.updating {
		content.WriteString(m.updateLoader.View())
		content.WriteString("\n")
	}

	content.WriteString("\n")

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

	// Summary - use CountWhereFunc for consistency
	updateCount := generics.CountWhereFunc(m.devices, deviceFirmwareHasUpdate)
	content.WriteString(m.styles.Label.Render(
		fmt.Sprintf("Devices (%d with updates):", updateCount),
	))
	content.WriteString("\n\n")

	content.WriteString(generics.RenderScrollableList(generics.ListRenderConfig[DeviceFirmware]{
		Items:    m.devices,
		Scroller: m.scroller,
		RenderItem: func(device DeviceFirmware, _ int, isCursor bool) string {
			return m.renderDeviceLine(device, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoAlways,
	}))

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
	return generics.CountSelectedFunc(m.devices, deviceFirmwareGet)
}

// UpdateCount returns the number of devices with updates.
func (m Model) UpdateCount() int {
	return generics.CountWhereFunc(m.devices, deviceFirmwareHasUpdate)
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

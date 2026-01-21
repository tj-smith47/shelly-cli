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
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
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
	Name        string
	Address     string
	Current     string
	Available   string
	HasUpdate   bool
	CanRollback bool
	Updating    bool
	RollingBack bool
	Selected    bool
	Checked     bool
	Err         error
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

// RollbackCompleteMsg signals that firmware rollback completed.
type RollbackCompleteMsg struct {
	Name    string
	Success bool
	Err     error
}

// Model displays firmware management.
type Model struct {
	helpers.Sizable
	ctx                context.Context
	svc                *shelly.Service
	fileCache          *cache.FileCache
	devices            []DeviceFirmware
	checking           bool
	updating           bool
	err                error
	focused            bool
	panelIndex         int
	confirmingRollback bool   // True when showing rollback confirmation
	rollbackDevice     string // Device name pending rollback
	styles             Styles
	updateLoader       loading.Model
	cacheStatus        cachestatus.Model
	lastChecked        time.Time
	lastResults        []UpdateResult
	showSummary        bool
	stagedPercent      int
	currentStage       int
	totalStages        int
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
	Warning   lipgloss.Style
	Confirm   lipgloss.Style
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
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Confirm: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
	}
}

// New creates a new Firmware model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("firmware component init", err)
		panic(fmt.Sprintf("firmware: invalid deps: %v", err))
	}

	m := Model{
		Sizable:   helpers.NewSizable(10, panel.NewScroller(0, 10)),
		ctx:       deps.Ctx,
		svc:       deps.Svc,
		fileCache: deps.FileCache,
		styles:    DefaultStyles(),
		updateLoader: loading.New(
			loading.WithMessage("Updating devices..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(false, false),
		),
		cacheStatus: cachestatus.New(),
	}
	m.Loader = m.Loader.SetMessage("Checking firmware...")
	return m
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

	m.Scroller.SetItemCount(len(m.devices))
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
	resized := m.ApplySizeWithExtraLoaders(width, height, m.updateLoader)
	m.updateLoader = resized[0]
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
	return m, tea.Batch(m.Loader.Tick(), m.checkAllDevices())
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

		results := make([]UpdateResult, 0, len(devices))

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

			result := UpdateResult{Name: dev.Name, Success: err == nil, Err: err}
			results = append(results, result)

			if err != nil {
				// Log but continue with remaining devices
				iostreams.DebugErr("firmware update "+dev.Name, err)
			}
		}

		// Return a message with update results for summary
		return updateBatchComplete{Results: results}
	}
}

// updateBatchComplete contains the batch update results for summary display.
type updateBatchComplete struct {
	Results []UpdateResult
}

// UpdateResult contains the result of a single device update.
type UpdateResult struct {
	Name    string
	Success bool
	Err     error
}

// StagedUpdateCompleteMsg signals staged update completion.
type StagedUpdateCompleteMsg struct {
	Stage       int
	TotalStages int
	Results     []UpdateResult
}

// RollbackCurrent starts the rollback confirmation for the device at cursor.
func (m Model) RollbackCurrent() (Model, tea.Cmd) {
	if m.updating || m.checking || len(m.devices) == 0 {
		return m, nil
	}

	cursor := m.Scroller.Cursor()
	if cursor >= len(m.devices) {
		return m, nil
	}

	// Start rollback confirmation
	m.confirmingRollback = true
	m.rollbackDevice = m.devices[cursor].Name
	return m, nil
}

// executeRollback performs the actual rollback after confirmation.
func (m Model) executeRollback(deviceName string) (Model, tea.Cmd) {
	// Find and update device state
	for i := range m.devices {
		if m.devices[i].Name == deviceName {
			m.devices[i].RollingBack = true
			break
		}
	}
	m.updating = true
	m.err = nil

	return m, tea.Batch(m.updateLoader.Tick(), m.rollbackDeviceCmd(deviceName))
}

func (m Model) rollbackDeviceCmd(deviceName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 60*time.Second)
		defer cancel()

		// First check if rollback is available
		status, err := m.svc.GetFirmwareStatus(ctx, deviceName)
		if err != nil {
			return RollbackCompleteMsg{Name: deviceName, Success: false, Err: fmt.Errorf("failed to check status: %w", err)}
		}
		if !status.CanRollback {
			return RollbackCompleteMsg{Name: deviceName, Success: false, Err: fmt.Errorf("rollback not available")}
		}

		// Perform rollback
		err = m.svc.RollbackFirmware(ctx, deviceName)
		if err != nil {
			return RollbackCompleteMsg{Name: deviceName, Success: false, Err: err}
		}

		return RollbackCompleteMsg{Name: deviceName, Success: true}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Handle rollback confirmation
	if m.confirmingRollback {
		return m.handleRollbackConfirmation(msg)
	}

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
	case RollbackCompleteMsg:
		return m.handleRollbackComplete(msg), nil
	case updateBatchComplete:
		return m.handleBatchComplete(msg), nil
	case StagedUpdateCompleteMsg:
		return m.handleStagedUpdateComplete(msg)
	case messages.NavigationMsg:
		return m.handleNavigationMsg(msg)
	case messages.ToggleEnableRequestMsg, messages.ScanRequestMsg, messages.RefreshRequestMsg:
		return m.handleActionMsg(msg)
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleRollbackConfirmation(msg tea.Msg) (Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "y", "Y":
			m.confirmingRollback = false
			deviceName := m.rollbackDevice
			m.rollbackDevice = ""
			return m.executeRollback(deviceName)
		case "n", "N", "esc":
			m.confirmingRollback = false
			m.rollbackDevice = ""
			return m, nil
		}
	}
	return m, nil
}

func (m Model) handleNavigationMsg(msg messages.NavigationMsg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	return m.handleNavigation(msg)
}

func (m Model) handleActionMsg(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	switch msg.(type) {
	case messages.ToggleEnableRequestMsg:
		m = m.toggleSelection()
		return m, nil
	case messages.ScanRequestMsg, messages.RefreshRequestMsg:
		return m.CheckAll()
	}
	return m, nil
}

// updateLoaders forwards tick messages to the appropriate loader.
// Returns updated model and command if the message was consumed.
func (m Model) updateLoaders(msg tea.Msg) (Model, tea.Cmd) {
	if m.checking {
		result := generics.UpdateLoader(m.Loader, msg, func(msg tea.Msg) bool {
			_, ok := msg.(CheckCompleteMsg)
			return ok
		})
		m.Loader = result.Loader
		if result.Consumed {
			return m, result.Cmd
		}
	}
	if m.updating {
		result := generics.UpdateLoader(m.updateLoader, msg, func(msg tea.Msg) bool {
			switch msg.(type) {
			case updateBatchComplete, UpdateCompleteMsg, RollbackCompleteMsg, StagedUpdateCompleteMsg:
				return true
			}
			return false
		})
		m.updateLoader = result.Loader
		if result.Consumed {
			return m, result.Cmd
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

// updateDeviceFromResult updates a device's status based on update result.
func (m Model) updateDeviceFromResult(device *DeviceFirmware, resultMap map[string]UpdateResult) {
	// Invalidate cache for updated devices
	if m.fileCache != nil {
		iostreams.DebugErr("invalidate firmware cache "+device.Name,
			m.fileCache.Invalidate(device.Name, cache.TypeFirmware))
	}
	// Update device status based on result
	if result, ok := resultMap[device.Name]; ok {
		if result.Success {
			device.HasUpdate = false
		} else {
			device.Err = result.Err
		}
	}
}

func (m Model) handleBatchComplete(msg updateBatchComplete) Model {
	m.updating = false
	m.lastResults = msg.Results
	m.showSummary = len(msg.Results) > 0

	// Build a map for quick lookup of results
	resultMap := make(map[string]UpdateResult)
	for _, r := range msg.Results {
		resultMap[r.Name] = r
	}

	for i := range m.devices {
		if m.devices[i].Updating {
			m.updateDeviceFromResult(&m.devices[i], resultMap)
		}
		m.devices[i].Updating = false
		m.devices[i].Selected = false
	}
	return m
}

func (m Model) handleStagedUpdateComplete(msg StagedUpdateCompleteMsg) (Model, tea.Cmd) {
	// Accumulate results
	m.lastResults = append(m.lastResults, msg.Results...)
	m.currentStage = msg.Stage + 1

	// Update device status for completed batch
	resultMap := make(map[string]UpdateResult)
	for _, r := range msg.Results {
		resultMap[r.Name] = r
	}
	for i := range m.devices {
		if result, ok := resultMap[m.devices[i].Name]; ok {
			m.devices[i].Updating = false
			if result.Success {
				m.devices[i].HasUpdate = false
			} else {
				m.devices[i].Err = result.Err
			}
			// Invalidate cache
			if m.fileCache != nil && result.Success {
				iostreams.DebugErr("invalidate firmware cache "+m.devices[i].Name,
					m.fileCache.Invalidate(m.devices[i].Name, cache.TypeFirmware))
			}
		}
	}

	// Continue with next batch
	selected := m.selectedDevices()
	devicesPerStage := max(1, len(selected)*m.stagedPercent/100)
	startIdx := msg.Stage * devicesPerStage

	// Mark next batch as updating
	endIdx := min(startIdx+devicesPerStage, len(selected))
	for i := startIdx; i < endIdx; i++ {
		idx := m.findDeviceIndex(selected[i].Name)
		if idx >= 0 && m.devices[idx].HasUpdate {
			m.devices[idx].Updating = true
		}
	}

	// Update loader message
	m.updateLoader = m.updateLoader.SetMessage(
		fmt.Sprintf("Updating stage %d/%d...", m.currentStage, m.totalStages),
	)

	return m, tea.Batch(m.updateLoader.Tick(), m.updateDevicesStaged(selected, startIdx, devicesPerStage))
}

func (m Model) handleRollbackComplete(msg RollbackCompleteMsg) Model {
	m.updating = false
	idx := m.findDeviceIndex(msg.Name)
	if idx < 0 {
		return m
	}

	m.devices[idx].RollingBack = false
	if !msg.Success {
		m.devices[idx].Err = msg.Err
		return m
	}

	// Invalidate cache so next check gets fresh data
	if m.fileCache != nil {
		iostreams.DebugErr("invalidate firmware cache "+msg.Name,
			m.fileCache.Invalidate(msg.Name, cache.TypeFirmware))
	}
	// Clear version info - will be refreshed on next check
	m.devices[idx].Current = ""
	m.devices[idx].Available = ""
	m.devices[idx].HasUpdate = false
	m.devices[idx].Checked = false
	return m
}

// findDeviceIndex returns the index of a device by name, or -1 if not found.
func (m Model) findDeviceIndex(name string) int {
	for i := range m.devices {
		if m.devices[i].Name == name {
			return i
		}
	}
	return -1
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Component-specific keys not in context system
	switch msg.String() {
	case "a":
		m = m.selectAllWithUpdates()
	case "n":
		m = m.selectNone()
	case "R":
		return m.RollbackCurrent()
	case "u", "enter":
		return m.UpdateSelected()
	case "U":
		// Select all with updates and update
		m = m.selectAllWithUpdates()
		return m.UpdateSelected()
	case "s":
		// Dismiss update summary
		m.showSummary = false
		m.lastResults = nil
	case "S":
		// Staged rollout (update 25% at a time)
		return m.UpdateSelectedStaged(25)
	}

	return m, nil
}

func (m Model) handleNavigation(msg messages.NavigationMsg) (Model, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		m.Scroller.CursorUp()
	case messages.NavDown:
		m.Scroller.CursorDown()
	case messages.NavPageUp:
		m.Scroller.PageUp()
	case messages.NavPageDown:
		m.Scroller.PageDown()
	case messages.NavHome:
		m.Scroller.CursorToStart()
	case messages.NavEnd:
		m.Scroller.CursorToEnd()
	case messages.NavLeft, messages.NavRight:
		// Not applicable for this component
	}
	return m, nil
}

// UpdateSelectedStaged starts a staged firmware update.
// percentPerStage determines what percentage of devices to update per stage.
func (m Model) UpdateSelectedStaged(percentPerStage int) (Model, tea.Cmd) {
	if m.updating || m.checking {
		return m, nil
	}

	selected := m.selectedDevices()
	if len(selected) == 0 {
		m.err = fmt.Errorf("no devices selected")
		return m, nil
	}

	// Calculate stages
	devicesPerStage := max(1, len(selected)*percentPerStage/100)
	m.totalStages = (len(selected) + devicesPerStage - 1) / devicesPerStage
	m.currentStage = 1
	m.stagedPercent = percentPerStage

	// Get first batch
	endIdx := min(devicesPerStage, len(selected))
	firstBatch := selected[:endIdx]

	m.updating = true
	m.err = nil

	// Mark first batch as updating
	batchNames := make(map[string]bool)
	for _, d := range firstBatch {
		batchNames[d.Name] = true
	}
	for i := range m.devices {
		if batchNames[m.devices[i].Name] && m.devices[i].HasUpdate {
			m.devices[i].Updating = true
		}
	}

	return m, tea.Batch(m.updateLoader.Tick(), m.updateDevicesStaged(selected, 0, devicesPerStage))
}

func (m Model) updateDevicesStaged(allDevices []DeviceFirmware, startIdx, batchSize int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 300*time.Second)
		defer cancel()

		endIdx := min(startIdx+batchSize, len(allDevices))
		batch := allDevices[startIdx:endIdx]

		results := make([]UpdateResult, 0, len(batch))

		for _, dev := range batch {
			if !dev.HasUpdate {
				continue
			}

			deviceCtx, deviceCancel := context.WithTimeout(ctx, 60*time.Second)
			err := m.svc.UpdateFirmwareStable(deviceCtx, dev.Name)
			deviceCancel()

			result := UpdateResult{Name: dev.Name, Success: err == nil, Err: err}
			results = append(results, result)

			if err != nil {
				iostreams.DebugErr("firmware update "+dev.Name, err)
			}
		}

		// Check if there are more batches
		if endIdx >= len(allDevices) {
			// All done
			return updateBatchComplete{Results: results}
		}

		// More batches to go - return staged update message
		return StagedUpdateCompleteMsg{
			Stage:       (startIdx / batchSize) + 1,
			TotalStages: (len(allDevices) + batchSize - 1) / batchSize,
			Results:     results,
		}
	}
}

func (m Model) toggleSelection() Model {
	generics.ToggleAtFunc(m.devices, m.Scroller.Cursor(), deviceFirmwareGet, deviceFirmwareSet)
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
	r := rendering.New(m.Width, m.Height).
		SetTitle("Firmware").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings and cache status when focused
	if m.focused {
		footer := "c:check u:update U:all S:staged R:rollback spc:sel a:sel-all"
		if m.showSummary {
			footer = "s:dismiss " + footer
		}
		if cs := m.cacheStatus.View(); cs != "" {
			footer += " | " + cs
		}
		r.SetFooter(footer)
	}

	// Show rollback confirmation
	if m.confirmingRollback {
		var content strings.Builder
		content.WriteString(m.styles.Confirm.Render("Rollback firmware on: " + m.rollbackDevice + "?"))
		content.WriteString("\n\n")
		content.WriteString(m.styles.Warning.Render("This will revert to the previous firmware version."))
		content.WriteString("\n\n")
		content.WriteString(m.styles.Muted.Render("Press Y to confirm, N or Esc to cancel"))
		r.SetContent(content.String())
		return r.Render()
	}

	var content strings.Builder

	// Action buttons
	content.WriteString(m.renderActions())
	content.WriteString("\n")

	// Status indicator (under actions, above device list)
	if m.checking {
		content.WriteString(m.Loader.View())
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

	// Update summary display
	if m.showSummary && len(m.lastResults) > 0 {
		content.WriteString(m.renderUpdateSummary())
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

func (m Model) renderUpdateSummary() string {
	var content strings.Builder

	successCount := 0
	failCount := 0
	for _, r := range m.lastResults {
		if r.Success {
			successCount++
		} else {
			failCount++
		}
	}

	content.WriteString("\n")
	content.WriteString(m.styles.Label.Render("Update Summary:"))
	content.WriteString("\n")

	if successCount > 0 {
		content.WriteString(m.styles.UpToDate.Render(fmt.Sprintf("  ✓ %d succeeded", successCount)))
		content.WriteString("\n")
	}
	if failCount > 0 {
		content.WriteString(m.styles.Error.Render(fmt.Sprintf("  ✗ %d failed", failCount)))
		content.WriteString("\n")
		// Show failed devices
		for _, r := range m.lastResults {
			if !r.Success {
				errMsg := "unknown error"
				if r.Err != nil {
					errMsg = r.Err.Error()
				}
				content.WriteString(m.styles.Muted.Render(fmt.Sprintf("    - %s: %s", r.Name, errMsg)))
				content.WriteString("\n")
			}
		}
	}

	content.WriteString(m.styles.Muted.Render("  Press 's' to dismiss summary"))
	return content.String()
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
		Scroller: m.Scroller,
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
	case device.RollingBack:
		status = m.styles.Updating.Render("↺")
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
	return m.Scroller.Cursor()
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

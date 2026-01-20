// Package backup provides TUI components for backup and restore operations.
package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	shellybackup "github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// Deps holds the dependencies for the Backup component.
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

// Mode represents the backup view mode.
type Mode int

// Mode constants.
const (
	ModeExport Mode = iota
	ModeImport
)

// String returns the mode name.
func (m Mode) String() string {
	switch m {
	case ModeExport:
		return "Export"
	case ModeImport:
		return "Import"
	default:
		return "Unknown"
	}
}

// DeviceBackup represents a device available for backup operations.
type DeviceBackup struct {
	Name      string
	Address   string
	Selected  bool
	Exporting bool
	Exported  bool
	FilePath  string
	Err       error
}

// IsSelected implements generics.Selectable.
func (d *DeviceBackup) IsSelected() bool { return d.Selected }

// SetSelected implements generics.Selectable.
func (d *DeviceBackup) SetSelected(v bool) { d.Selected = v }

// Selection helpers for value slices.
func deviceBackupGet(d *DeviceBackup) bool    { return d.Selected }
func deviceBackupSet(d *DeviceBackup, v bool) { d.Selected = v }

// File represents a backup file available for import.
type File struct {
	Name       string
	Path       string
	DeviceName string
	DeviceID   string
	Timestamp  time.Time
	Selected   bool
}

// ExportCompleteMsg signals that export operation completed.
type ExportCompleteMsg struct {
	Results []ExportResult
}

// ExportResult holds the result of a single device export.
type ExportResult struct {
	Name     string
	FilePath string
	Success  bool
	Err      error
}

// ImportCompleteMsg signals that import operation completed.
type ImportCompleteMsg struct {
	Name    string
	Success bool
	Err     error
}

// Model displays backup and restore operations.
type Model struct {
	helpers.Sizable
	ctx          context.Context
	svc          *shelly.Service
	mode         Mode
	devices      []DeviceBackup
	backupFiles  []File
	exporting    bool
	importing    bool
	backupDir    string
	err          error
	focused      bool
	panelIndex   int
	styles       Styles
	importLoader loading.Model // Extra loader for import step
}

// Styles holds styles for the Backup component.
type Styles struct {
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Cursor     lipgloss.Style
	Success    lipgloss.Style
	Failure    lipgloss.Style
	InProgress lipgloss.Style
	Label      lipgloss.Style
	Error      lipgloss.Style
	Muted      lipgloss.Style
	Button     lipgloss.Style
	ModeActive lipgloss.Style
}

// DefaultStyles returns the default styles for the Backup component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Selected: lipgloss.NewStyle().
			Foreground(colors.Online),
		Unselected: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Cursor: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(colors.Online),
		Failure: lipgloss.NewStyle().
			Foreground(colors.Error),
		InProgress: lipgloss.NewStyle().
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
		ModeActive: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// New creates a new Backup model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("backup component init", err)
		panic(fmt.Sprintf("backup: invalid deps: %v", err))
	}

	// Default backup directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	backupDir := filepath.Join(homeDir, ".shelly", "backups")

	m := Model{
		Sizable:   helpers.NewSizable(10, panel.NewScroller(0, 10)),
		ctx:       deps.Ctx,
		svc:       deps.Svc,
		mode:      ModeExport,
		backupDir: backupDir,
		styles:    DefaultStyles(),
		importLoader: loading.New(
			loading.WithMessage("Importing backup..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(false, false),
		),
	}
	m.Loader = m.Loader.SetMessage("Exporting backups...")
	return m
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

	m.devices = make([]DeviceBackup, 0, len(cfg.Devices))
	for name, dev := range cfg.Devices {
		m.devices = append(m.devices, DeviceBackup{
			Name:    name,
			Address: dev.Address,
		})
	}

	// Update scroller if in export mode
	if m.mode == ModeExport {
		m.Scroller.SetItemCount(len(m.devices))
	}

	return m
}

// LoadBackupFiles loads available backup files.
func (m Model) LoadBackupFiles() Model {
	m.backupFiles = nil

	fs := config.Fs()

	// List backup files in backup directory
	entries, err := afero.ReadDir(fs, m.backupDir)
	if err != nil {
		// Directory may not exist yet
		return m
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		path := filepath.Join(m.backupDir, name)

		// Parse backup to get device info
		data, err := afero.ReadFile(fs, path)
		if err != nil {
			continue
		}

		bkp, err := shellybackup.Validate(data)
		if err != nil {
			continue
		}

		deviceInfo := bkp.Device()
		m.backupFiles = append(m.backupFiles, File{
			Name:       name,
			Path:       path,
			DeviceName: deviceInfo.Name,
			DeviceID:   deviceInfo.ID,
			Timestamp:  entry.ModTime(),
		})
	}

	// Update scroller if in import mode
	if m.mode == ModeImport {
		m.Scroller.SetItemCount(len(m.backupFiles))
	}

	return m
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	resized := m.ApplySizeWithExtraLoaders(width, height, m.importLoader)
	m.importLoader = resized[0]
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

// SetBackupDir sets the backup directory.
func (m Model) SetBackupDir(dir string) Model {
	m.backupDir = dir
	return m
}

// ExportSelected exports backups for selected devices.
func (m Model) ExportSelected() (Model, tea.Cmd) {
	if m.exporting || m.mode != ModeExport {
		return m, nil
	}

	selected := m.selectedDevices()
	if len(selected) == 0 {
		m.err = fmt.Errorf("no devices selected")
		return m, nil
	}

	m.exporting = true
	m.err = nil

	// Mark selected devices as exporting
	for i := range m.devices {
		if m.devices[i].Selected {
			m.devices[i].Exporting = true
		}
	}

	return m, tea.Batch(m.Loader.Tick(), m.exportDevices(selected))
}

func (m Model) exportDevices(devices []DeviceBackup) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 180*time.Second)
		defer cancel()

		// Ensure backup directory exists
		if err := config.Fs().MkdirAll(m.backupDir, 0o750); err != nil {
			return ExportCompleteMsg{Results: []ExportResult{{
				Name:    "all",
				Success: false,
				Err:     fmt.Errorf("failed to create backup directory: %w", err),
			}}}
		}

		var (
			results []ExportResult
			mu      sync.Mutex
		)

		// Rate limiting is handled at the service layer
		g, gctx := errgroup.WithContext(ctx)

		for _, dev := range devices {
			device := dev
			g.Go(func() error {
				// Capture results per device rather than failing the group
				result := ExportResult{Name: device.Name}

				bkp, err := m.svc.CreateBackup(gctx, device.Name, shellybackup.Options{})
				if err != nil {
					result.Err = err
					mu.Lock()
					results = append(results, result)
					mu.Unlock()
					return nil //nolint:nilerr // Errors captured per device, not propagated to group
				}

				// Generate filename
				deviceInfo := bkp.Device()
				filename := shellybackup.GenerateFilename(deviceInfo.Name, deviceInfo.ID, false)
				filePath := filepath.Join(m.backupDir, filename)

				// Save backup
				data, err := json.MarshalIndent(bkp.Backup, "", "  ")
				if err != nil {
					result.Err = fmt.Errorf("failed to serialize backup: %w", err)
					mu.Lock()
					results = append(results, result)
					mu.Unlock()
					return nil
				}

				if err := shellybackup.SaveToFile(data, filePath); err != nil {
					result.Err = err
					mu.Lock()
					results = append(results, result)
					mu.Unlock()
					return nil //nolint:nilerr // Errors captured per device, not propagated to group
				}

				result.FilePath = filePath
				result.Success = true
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			// Individual errors are captured per device in results
			iostreams.DebugErr("backup export batch", err)
		}

		return ExportCompleteMsg{Results: results}
	}
}

// ImportSelected imports the selected backup file.
func (m Model) ImportSelected() (Model, tea.Cmd) {
	if m.importing || m.mode != ModeImport {
		return m, nil
	}

	cursor := m.Scroller.Cursor()
	if cursor >= len(m.backupFiles) {
		m.err = fmt.Errorf("no backup file selected")
		return m, nil
	}

	backupFile := m.backupFiles[cursor]
	m.importing = true
	m.err = nil

	return m, tea.Batch(m.importLoader.Tick(), m.importBackup(backupFile))
}

func (m Model) importBackup(backupFile File) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 120*time.Second)
		defer cancel()

		// Load and validate backup file
		bkp, err := shellybackup.LoadAndValidate(backupFile.Path)
		if err != nil {
			return ImportCompleteMsg{
				Name:    backupFile.Name,
				Success: false,
				Err:     err,
			}
		}

		// Find matching device by ID
		deviceInfo := bkp.Device()
		targetDevice := ""

		cfg := config.Get()
		if cfg != nil {
			for name, dev := range cfg.Devices {
				// Try to match by device info (we could improve this matching)
				_ = dev
				if name == deviceInfo.Name || name == deviceInfo.ID {
					targetDevice = name
					break
				}
			}
		}

		if targetDevice == "" {
			return ImportCompleteMsg{
				Name:    backupFile.Name,
				Success: false,
				Err:     fmt.Errorf("no matching device found for %s", deviceInfo.Name),
			}
		}

		// Restore backup
		_, err = m.svc.RestoreBackup(ctx, targetDevice, bkp, shellybackup.RestoreOptions{})
		if err != nil {
			return ImportCompleteMsg{
				Name:    backupFile.Name,
				Success: false,
				Err:     err,
			}
		}

		return ImportCompleteMsg{
			Name:    backupFile.Name,
			Success: true,
		}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Forward tick messages to loaders when active
	if m.exporting {
		var cmd tea.Cmd
		m.Loader, cmd = m.Loader.Update(msg)
		switch msg.(type) {
		case ExportCompleteMsg:
			// Pass through to main switch below
		default:
			if cmd != nil {
				return m, cmd
			}
		}
	}
	if m.importing {
		var cmd tea.Cmd
		m.importLoader, cmd = m.importLoader.Update(msg)
		switch msg.(type) {
		case ImportCompleteMsg:
			// Pass through to main switch below
		default:
			if cmd != nil {
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case ExportCompleteMsg:
		return m.handleExportComplete(msg)
	case ImportCompleteMsg:
		return m.handleImportComplete(msg)
	case messages.NavigationMsg:
		return m.handleNavigationMsg(msg)
	case messages.ToggleEnableRequestMsg, messages.RunRequestMsg, messages.RefreshRequestMsg:
		return m.handleActionMsg(msg)
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Component-specific keys not in context system
	switch msg.String() {
	case "a":
		m = m.selectAll()
	case "n":
		m = m.selectNone()
	case "1":
		m.mode = ModeExport
		m.Scroller.SetItemCount(len(m.devices))
		m.Scroller.CursorToStart()
	case "2":
		m.mode = ModeImport
		m = m.LoadBackupFiles()
		m.Scroller.CursorToStart()
	}

	return m, nil
}

func (m Model) handleExportComplete(msg ExportCompleteMsg) (Model, tea.Cmd) {
	m.exporting = false
	for _, result := range msg.Results {
		for i := range m.devices {
			if m.devices[i].Name != result.Name {
				continue
			}
			m.devices[i].Exporting = false
			m.devices[i].Exported = result.Success
			m.devices[i].FilePath = result.FilePath
			m.devices[i].Err = result.Err
			break
		}
	}
	// Reload backup files after export
	m = m.LoadBackupFiles()
	return m, nil
}

func (m Model) handleImportComplete(msg ImportCompleteMsg) (Model, tea.Cmd) {
	m.importing = false
	if !msg.Success {
		m.err = msg.Err
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
	case messages.RunRequestMsg:
		if m.mode == ModeExport {
			return m.ExportSelected()
		}
		return m.ImportSelected()
	case messages.RefreshRequestMsg:
		if m.mode == ModeImport {
			m = m.LoadBackupFiles()
		}
		return m, nil
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

func (m Model) toggleSelection() Model {
	if m.mode == ModeExport {
		generics.ToggleAtFunc(m.devices, m.Scroller.Cursor(), deviceBackupGet, deviceBackupSet)
	}
	return m
}

func (m Model) selectAll() Model {
	if m.mode == ModeExport {
		generics.SelectAllFunc(m.devices, deviceBackupSet)
	}
	return m
}

func (m Model) selectNone() Model {
	if m.mode == ModeExport {
		generics.SelectNoneFunc(m.devices, deviceBackupSet)
	}
	return m
}

func (m Model) selectedDevices() []DeviceBackup {
	return generics.Filter(m.devices, func(d DeviceBackup) bool { return d.Selected })
}

// View renders the Backup component.
func (m Model) View() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Backup & Restore").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused {
		if m.mode == ModeExport {
			r.SetFooter(theme.StyledKeybindings("spc:sel a:all x:export 2:import"))
		} else {
			r.SetFooter(theme.StyledKeybindings("enter:import r:refresh 1:export"))
		}
	}

	var content strings.Builder

	// Mode selector
	content.WriteString(m.renderModeSelector())
	content.WriteString("\n\n")

	// Content based on mode
	if m.mode == ModeExport {
		content.WriteString(m.renderExportView())
	} else {
		content.WriteString(m.renderImportView())
	}

	// Error display with categorized messaging and retry hint
	if m.err != nil {
		msg, hint := tuierrors.FormatError(m.err)
		content.WriteString("\n")
		content.WriteString(m.styles.Error.Render(msg))
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render("  " + hint))
		content.WriteString("\n")
		if m.mode == ModeImport {
			content.WriteString(m.styles.Muted.Render("  Press 'r' to refresh"))
		} else {
			content.WriteString(m.styles.Muted.Render("  Press 'x' to retry export"))
		}
	}

	// Status indicator with animated loader
	if m.exporting {
		content.WriteString("\n")
		content.WriteString(m.Loader.View())
	} else if m.importing {
		content.WriteString("\n")
		content.WriteString(m.importLoader.View())
	}

	r.SetContent(content.String())
	return r.Render()
}

func (m Model) renderModeSelector() string {
	modes := []struct {
		mode Mode
		key  string
		name string
	}{
		{ModeExport, "1", "Export"},
		{ModeImport, "2", "Import"},
	}

	parts := make([]string, 0, len(modes))
	for _, mode := range modes {
		style := m.styles.Muted
		if mode.mode == m.mode {
			style = m.styles.ModeActive
		}
		parts = append(parts, style.Render(fmt.Sprintf("[%s] %s", mode.key, mode.name)))
	}

	return m.styles.Label.Render("Mode: ") + strings.Join(parts, " ")
}

func (m Model) renderExportView() string {
	if len(m.devices) == 0 {
		return m.styles.Muted.Render("No devices registered")
	}

	var content strings.Builder

	selectedCount := len(m.selectedDevices())
	content.WriteString(m.styles.Label.Render(
		fmt.Sprintf("Devices (%d selected):", selectedCount),
	))
	content.WriteString("\n\n")

	content.WriteString(generics.RenderScrollableList(generics.ListRenderConfig[DeviceBackup]{
		Items:    m.devices,
		Scroller: m.Scroller,
		RenderItem: func(device DeviceBackup, _ int, isCursor bool) string {
			return m.renderExportDeviceLine(device, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoWhenNeeded,
	}))

	return content.String()
}

func (m Model) renderExportDeviceLine(device DeviceBackup, isCursor bool) string {
	cursor := "  "
	if isCursor {
		cursor = "▶ "
	}

	// Checkbox
	var checkbox string
	if device.Selected {
		checkbox = m.styles.Selected.Render("[✓]")
	} else {
		checkbox = m.styles.Unselected.Render("[ ]")
	}

	// Status indicator
	var status string
	switch {
	case device.Exporting:
		status = m.styles.InProgress.Render("↻")
	case device.Err != nil:
		status = m.styles.Failure.Render("✗")
	case device.Exported:
		status = m.styles.Success.Render("✓")
	default:
		status = m.styles.Muted.Render("○")
	}

	line := fmt.Sprintf("%s%s %s %s", cursor, checkbox, status, device.Name)

	if device.Exported && device.FilePath != "" {
		filename := filepath.Base(device.FilePath)
		line += m.styles.Muted.Render(fmt.Sprintf(" → %s", filename))
	}

	if isCursor {
		return m.styles.Cursor.Render(line)
	}
	return line
}

func (m Model) renderImportView() string {
	if len(m.backupFiles) == 0 {
		return m.styles.Muted.Render("No backup files found in " + m.backupDir)
	}

	var content strings.Builder

	content.WriteString(m.styles.Label.Render(
		fmt.Sprintf("Backup files (%d):", len(m.backupFiles)),
	))
	content.WriteString("\n\n")

	content.WriteString(generics.RenderScrollableList(generics.ListRenderConfig[File]{
		Items:    m.backupFiles,
		Scroller: m.Scroller,
		RenderItem: func(backupFile File, _ int, isCursor bool) string {
			return m.renderBackupFileLine(backupFile, isCursor)
		},
		ScrollStyle:    m.styles.Muted,
		ScrollInfoMode: generics.ScrollInfoWhenNeeded,
	}))

	return content.String()
}

func (m Model) renderBackupFileLine(backupFile File, isCursor bool) string {
	cursor := "  "
	if isCursor {
		cursor = "▶ "
	}

	// Device name and timestamp
	deviceStr := backupFile.DeviceName
	if deviceStr == "" {
		deviceStr = backupFile.DeviceID
	}
	if deviceStr == "" {
		deviceStr = "Unknown"
	}

	timeStr := backupFile.Timestamp.Format("2006-01-02 15:04")

	line := fmt.Sprintf("%s%s %s %s",
		cursor,
		m.styles.Label.Render(deviceStr),
		m.styles.Muted.Render(timeStr),
		m.styles.Muted.Render(backupFile.Name),
	)

	if isCursor {
		return m.styles.Cursor.Render(line)
	}
	return line
}

// Devices returns the device list.
func (m Model) Devices() []DeviceBackup {
	return m.devices
}

// BackupFiles returns the backup file list.
func (m Model) BackupFiles() []File {
	return m.backupFiles
}

// Mode returns the current mode.
func (m Model) Mode() Mode {
	return m.mode
}

// Exporting returns whether export is in progress.
func (m Model) Exporting() bool {
	return m.exporting
}

// Importing returns whether import is in progress.
func (m Model) Importing() bool {
	return m.importing
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
	return generics.CountSelectedFunc(m.devices, deviceBackupGet)
}

// BackupDir returns the backup directory path.
func (m Model) BackupDir() string {
	return m.backupDir
}

// Refresh reloads devices and backup files.
func (m Model) Refresh() (Model, tea.Cmd) {
	m.err = nil
	m = m.LoadDevices()
	m = m.LoadBackupFiles()
	return m, nil
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	if m.mode == ModeExport {
		return "j/k:scroll g/G:top/bottom enter:backup"
	}
	return "j/k:scroll g/G:top/bottom enter:restore"
}

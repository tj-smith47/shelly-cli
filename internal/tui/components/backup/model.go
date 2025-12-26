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
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
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
	ctx         context.Context
	svc         *shelly.Service
	mode        Mode
	devices     []DeviceBackup
	backupFiles []File
	scroller    *panel.Scroller
	exporting   bool
	importing   bool
	backupDir   string
	err         error
	width       int
	height      int
	focused     bool
	panelIndex  int
	styles      Styles
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
			Foreground(colors.Muted),
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
		panic(fmt.Sprintf("backup: invalid deps: %v", err))
	}

	// Default backup directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	backupDir := filepath.Join(homeDir, ".shelly", "backups")

	return Model{
		ctx:       deps.Ctx,
		svc:       deps.Svc,
		mode:      ModeExport,
		scroller:  panel.NewScroller(0, 10),
		backupDir: backupDir,
		styles:    DefaultStyles(),
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

	m.devices = make([]DeviceBackup, 0, len(cfg.Devices))
	for name, dev := range cfg.Devices {
		m.devices = append(m.devices, DeviceBackup{
			Name:    name,
			Address: dev.Address,
		})
	}

	// Update scroller if in export mode
	if m.mode == ModeExport {
		m.scroller.SetItemCount(len(m.devices))
	}

	return m
}

// LoadBackupFiles loads available backup files.
func (m Model) LoadBackupFiles() Model {
	m.backupFiles = nil

	// List backup files in backup directory
	entries, err := os.ReadDir(m.backupDir)
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
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Parse backup to get device info
		data, err := os.ReadFile(path) //nolint:gosec // G304: path is constructed from known backup directory
		if err != nil {
			continue
		}

		backup, err := shelly.ValidateBackup(data)
		if err != nil {
			continue
		}

		deviceInfo := backup.Device()
		m.backupFiles = append(m.backupFiles, File{
			Name:       name,
			Path:       path,
			DeviceName: deviceInfo.Name,
			DeviceID:   deviceInfo.ID,
			Timestamp:  info.ModTime(),
		})
	}

	// Update scroller if in import mode
	if m.mode == ModeImport {
		m.scroller.SetItemCount(len(m.backupFiles))
	}

	return m
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	// Calculate visible rows: height - mode selector - header - borders
	visibleRows := height - 10
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
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

	return m, m.exportDevices(selected)
}

func (m Model) exportDevices(devices []DeviceBackup) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 180*time.Second)
		defer cancel()

		// Ensure backup directory exists
		if err := os.MkdirAll(m.backupDir, 0o750); err != nil {
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

				backup, err := m.svc.CreateBackup(gctx, device.Name, shelly.BackupOptions{})
				if err != nil {
					result.Err = err
					mu.Lock()
					results = append(results, result)
					mu.Unlock()
					return nil //nolint:nilerr // Errors captured per device, not propagated to group
				}

				// Generate filename
				deviceInfo := backup.Device()
				filename := m.svc.GenerateBackupFilename(deviceInfo.Name, deviceInfo.ID, false)
				filePath := filepath.Join(m.backupDir, filename)

				// Save backup
				data, err := json.MarshalIndent(backup.Backup, "", "  ")
				if err != nil {
					result.Err = fmt.Errorf("failed to serialize backup: %w", err)
					mu.Lock()
					results = append(results, result)
					mu.Unlock()
					return nil
				}

				if err := m.svc.SaveBackupToFile(data, filePath); err != nil {
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

	cursor := m.scroller.Cursor()
	if cursor >= len(m.backupFiles) {
		m.err = fmt.Errorf("no backup file selected")
		return m, nil
	}

	backupFile := m.backupFiles[cursor]
	m.importing = true
	m.err = nil

	return m, m.importBackup(backupFile)
}

func (m Model) importBackup(backupFile File) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 120*time.Second)
		defer cancel()

		// Load backup file
		data, err := m.svc.LoadBackupFromFile(backupFile.Path)
		if err != nil {
			return ImportCompleteMsg{
				Name:    backupFile.Name,
				Success: false,
				Err:     err,
			}
		}

		// Validate backup
		backup, err := shelly.ValidateBackup(data)
		if err != nil {
			return ImportCompleteMsg{
				Name:    backupFile.Name,
				Success: false,
				Err:     err,
			}
		}

		// Find matching device by ID
		deviceInfo := backup.Device()
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
		_, err = m.svc.RestoreBackup(ctx, targetDevice, backup, shelly.RestoreOptions{})
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
	switch msg := msg.(type) {
	case ExportCompleteMsg:
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

	case ImportCompleteMsg:
		m.importing = false
		if !msg.Success {
			m.err = msg.Err
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
		m = m.selectAll()
	case "n":
		m = m.selectNone()
	case "1":
		m.mode = ModeExport
		m.scroller.SetItemCount(len(m.devices))
		m.scroller.CursorToStart()
	case "2":
		m.mode = ModeImport
		m = m.LoadBackupFiles()
		m.scroller.CursorToStart()
	case "enter", "x":
		if m.mode == ModeExport {
			return m.ExportSelected()
		}
		return m.ImportSelected()
	case "r":
		if m.mode == ModeImport {
			m = m.LoadBackupFiles()
		}
	}

	return m, nil
}

func (m Model) toggleSelection() Model {
	cursor := m.scroller.Cursor()
	if m.mode == ModeExport && len(m.devices) > 0 && cursor < len(m.devices) {
		m.devices[cursor].Selected = !m.devices[cursor].Selected
	}
	return m
}

func (m Model) selectAll() Model {
	if m.mode == ModeExport {
		for i := range m.devices {
			m.devices[i].Selected = true
		}
	}
	return m
}

func (m Model) selectNone() Model {
	if m.mode == ModeExport {
		for i := range m.devices {
			m.devices[i].Selected = false
		}
	}
	return m
}

func (m Model) selectedDevices() []DeviceBackup {
	selected := make([]DeviceBackup, 0)
	for _, d := range m.devices {
		if d.Selected {
			selected = append(selected, d)
		}
	}
	return selected
}

// View renders the Backup component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Backup & Restore").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused {
		if m.mode == ModeExport {
			r.SetFooter("spc:sel a:all x:export 2:import")
		} else {
			r.SetFooter("enter:import r:refresh 1:export")
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

	// Status indicator
	if m.exporting {
		content.WriteString("\n")
		content.WriteString(m.styles.InProgress.Render("Exporting backups..."))
	} else if m.importing {
		content.WriteString("\n")
		content.WriteString(m.styles.InProgress.Render("Importing backup..."))
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

	startIdx, endIdx := m.scroller.VisibleRange()
	if endIdx > len(m.devices) {
		endIdx = len(m.devices)
	}

	for i := startIdx; i < endIdx; i++ {
		device := m.devices[i]
		isCursor := m.scroller.IsCursorAt(i)
		content.WriteString(m.renderExportDeviceLine(device, isCursor))
		if i < endIdx-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	if m.scroller.HasMore() || m.scroller.HasPrevious() {
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))
	}

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

	startIdx, endIdx := m.scroller.VisibleRange()
	if endIdx > len(m.backupFiles) {
		endIdx = len(m.backupFiles)
	}

	for i := startIdx; i < endIdx; i++ {
		backupFile := m.backupFiles[i]
		isCursor := m.scroller.IsCursorAt(i)
		content.WriteString(m.renderBackupFileLine(backupFile, isCursor))
		if i < endIdx-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	if m.scroller.HasMore() || m.scroller.HasPrevious() {
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))
	}

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

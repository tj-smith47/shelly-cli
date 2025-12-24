package views

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/backup"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/batch"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/discovery"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/firmware"
	"github.com/tj-smith47/shelly-cli/internal/tui/layout"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// ManagePanel identifies which panel is focused in the Manage view.
type ManagePanel int

// Manage panel constants.
const (
	ManagePanelDiscovery ManagePanel = iota
	ManagePanelBatch
	ManagePanelFirmware
	ManagePanelBackup
)

// ManageDeps holds dependencies for the manage view.
type ManageDeps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d ManageDeps) Validate() error {
	if d.Ctx == nil {
		return errNilContext
	}
	if d.Svc == nil {
		return errNilService
	}
	return nil
}

// Manage is the manage view that composes Discovery, Batch, Firmware, and Backup components.
// This provides local device administration (not Shelly Cloud Fleet).
type Manage struct {
	ctx context.Context
	svc *shelly.Service
	id  ViewID

	// Component models
	discovery discovery.Model
	batch     batch.Model
	firmware  firmware.Model
	backup    backup.Model

	// State
	focusedPanel ManagePanel
	width        int
	height       int
	styles       ManageStyles

	// Layout calculator for flexible panel sizing
	layoutCalc *layout.TwoColumnLayout
}

// ManageStyles holds styles for the manage view.
type ManageStyles struct {
	Panel       lipgloss.Style
	PanelActive lipgloss.Style
	Title       lipgloss.Style
	Muted       lipgloss.Style
}

// DefaultManageStyles returns default styles for the manage view.
func DefaultManageStyles() ManageStyles {
	colors := theme.GetSemanticColors()
	return ManageStyles{
		Panel: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder),
		PanelActive: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// NewManage creates a new manage view.
func NewManage(deps ManageDeps) *Manage {
	if err := deps.Validate(); err != nil {
		panic("manage: " + err.Error())
	}

	discoveryDeps := discovery.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	batchDeps := batch.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	firmwareDeps := firmware.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	backupDeps := backup.Deps{Ctx: deps.Ctx, Svc: deps.Svc}

	// Create flexible layout with 40/60 column split (left/right)
	layoutCalc := layout.NewTwoColumnLayout(0.4, 1)

	// Configure left column panels (Discovery, Firmware, Backup) with expansion on focus
	layoutCalc.LeftColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(ManagePanelDiscovery), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(ManagePanelFirmware), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(ManagePanelBackup), MinHeight: 5, ExpandOnFocus: true},
	}

	// Configure right column (Batch takes full height)
	layoutCalc.RightColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(ManagePanelBatch), MinHeight: 10, ExpandOnFocus: true},
	}

	m := &Manage{
		ctx:          deps.Ctx,
		svc:          deps.Svc,
		id:           tabs.TabManage,
		discovery:    discovery.New(discoveryDeps),
		batch:        batch.New(batchDeps),
		firmware:     firmware.New(firmwareDeps),
		backup:       backup.New(backupDeps),
		focusedPanel: ManagePanelDiscovery,
		styles:       DefaultManageStyles(),
		layoutCalc:   layoutCalc,
	}

	// Initialize focus states so the default focused panel (Discovery) receives key events
	m.updateFocusStates()

	// Load device lists for batch, firmware, and backup
	m.batch = m.batch.LoadDevices()
	m.firmware = m.firmware.LoadDevices()
	m.backup = m.backup.LoadDevices()

	return m
}

// Init returns the initial command.
func (m *Manage) Init() tea.Cmd {
	return tea.Batch(
		m.discovery.Init(),
		m.batch.Init(),
		m.firmware.Init(),
		m.backup.Init(),
	)
}

// ID returns the view ID.
func (m *Manage) ID() ViewID {
	return m.id
}

// SetSize sets the view dimensions.
func (m *Manage) SetSize(width, height int) View {
	m.width = width
	m.height = height

	if m.isNarrow() {
		// Narrow mode: all components get full width, full height
		contentWidth := width - 4
		contentHeight := height - 6

		m.discovery = m.discovery.SetSize(contentWidth, contentHeight)
		m.batch = m.batch.SetSize(contentWidth, contentHeight)
		m.firmware = m.firmware.SetSize(contentWidth, contentHeight)
		m.backup = m.backup.SetSize(contentWidth, contentHeight)
		return m
	}

	// Update layout with new dimensions and focus
	m.layoutCalc.SetSize(width, height)
	m.layoutCalc.SetFocus(layout.PanelID(m.focusedPanel))

	// Calculate panel dimensions using flexible layout
	dims := m.layoutCalc.Calculate()

	// Apply sizes to left column components (with border adjustment)
	if d, ok := dims[layout.PanelID(ManagePanelDiscovery)]; ok {
		cw, ch := d.ContentDimensions(2)
		m.discovery = m.discovery.SetSize(cw, ch)
	}
	if d, ok := dims[layout.PanelID(ManagePanelFirmware)]; ok {
		cw, ch := d.ContentDimensions(2)
		m.firmware = m.firmware.SetSize(cw, ch)
	}
	if d, ok := dims[layout.PanelID(ManagePanelBackup)]; ok {
		cw, ch := d.ContentDimensions(2)
		m.backup = m.backup.SetSize(cw, ch)
	}

	// Apply size to right column (Batch)
	if d, ok := dims[layout.PanelID(ManagePanelBatch)]; ok {
		cw, ch := d.ContentDimensions(2)
		m.batch = m.batch.SetSize(cw, ch)
	}

	return m
}

// Update handles messages.
func (m *Manage) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle keyboard input
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		m.handleKeyPress(keyMsg)
	}

	// Update all components (they handle their own messages)
	cmd := m.updateComponents(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Manage) handleKeyPress(msg tea.KeyPressMsg) {
	switch msg.String() {
	case keyTab:
		m.focusNext()
	case keyShiftTab:
		m.focusPrev()
	}
}

func (m *Manage) focusNext() {
	panels := []ManagePanel{ManagePanelDiscovery, ManagePanelBatch, ManagePanelFirmware, ManagePanelBackup}
	for i, p := range panels {
		if p == m.focusedPanel {
			m.focusedPanel = panels[(i+1)%len(panels)]
			break
		}
	}
	m.updateFocusStates()
}

func (m *Manage) focusPrev() {
	panels := []ManagePanel{ManagePanelDiscovery, ManagePanelBatch, ManagePanelFirmware, ManagePanelBackup}
	for i, p := range panels {
		if p == m.focusedPanel {
			prevIdx := (i - 1 + len(panels)) % len(panels)
			m.focusedPanel = panels[prevIdx]
			break
		}
	}
	m.updateFocusStates()
}

func (m *Manage) updateFocusStates() {
	m.discovery = m.discovery.SetFocused(m.focusedPanel == ManagePanelDiscovery)
	m.batch = m.batch.SetFocused(m.focusedPanel == ManagePanelBatch)
	m.firmware = m.firmware.SetFocused(m.focusedPanel == ManagePanelFirmware)
	m.backup = m.backup.SetFocused(m.focusedPanel == ManagePanelBackup)

	// Recalculate layout with new focus (panels resize on focus change)
	if m.layoutCalc != nil && m.width > 0 && m.height > 0 {
		m.SetSize(m.width, m.height)
	}
}

func (m *Manage) updateComponents(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Only update the focused component for key messages
	if _, ok := msg.(tea.KeyPressMsg); ok {
		switch m.focusedPanel {
		case ManagePanelDiscovery:
			m.discovery, cmd = m.discovery.Update(msg)
		case ManagePanelBatch:
			m.batch, cmd = m.batch.Update(msg)
		case ManagePanelFirmware:
			m.firmware, cmd = m.firmware.Update(msg)
		case ManagePanelBackup:
			m.backup, cmd = m.backup.Update(msg)
		}
		cmds = append(cmds, cmd)
	} else {
		// For non-key messages (async results), update all components
		m.discovery, cmd = m.discovery.Update(msg)
		cmds = append(cmds, cmd)
		m.batch, cmd = m.batch.Update(msg)
		cmds = append(cmds, cmd)
		m.firmware, cmd = m.firmware.Update(msg)
		cmds = append(cmds, cmd)
		m.backup, cmd = m.backup.Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// isNarrow returns true if the view should use narrow/vertical layout.
func (m *Manage) isNarrow() bool {
	return m.width < 80
}

// View renders the manage view.
func (m *Manage) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	if m.isNarrow() {
		return m.renderNarrowLayout()
	}

	return m.renderStandardLayout()
}

func (m *Manage) renderNarrowLayout() string {
	// In narrow mode, show only the focused panel at full width
	// Components already have embedded titles from rendering.New()
	switch m.focusedPanel {
	case ManagePanelDiscovery:
		return m.discovery.View()
	case ManagePanelBatch:
		return m.batch.View()
	case ManagePanelFirmware:
		return m.firmware.View()
	case ManagePanelBackup:
		return m.backup.View()
	default:
		return m.discovery.View()
	}
}

func (m *Manage) renderStandardLayout() string {
	// Render panels (components already have embedded titles)
	leftColumn := lipgloss.JoinVertical(lipgloss.Left,
		m.discovery.View(),
		m.firmware.View(),
		m.backup.View(),
	)

	// Join left column with right column (batch)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, " ", m.batch.View())
}

// Refresh reloads all components.
func (m *Manage) Refresh() tea.Cmd {
	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.discovery, cmd = m.discovery.Refresh()
	cmds = append(cmds, cmd)

	m.batch, cmd = m.batch.Refresh()
	cmds = append(cmds, cmd)

	m.firmware, cmd = m.firmware.Refresh()
	cmds = append(cmds, cmd)

	m.backup, cmd = m.backup.Refresh()
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// FocusedPanel returns the currently focused panel.
func (m *Manage) FocusedPanel() ManagePanel {
	return m.focusedPanel
}

// Discovery returns the discovery component.
func (m *Manage) Discovery() discovery.Model {
	return m.discovery
}

// Batch returns the batch component.
func (m *Manage) Batch() batch.Model {
	return m.batch
}

// Firmware returns the firmware component.
func (m *Manage) Firmware() firmware.Model {
	return m.firmware
}

// Backup returns the backup component.
func (m *Manage) Backup() backup.Model {
	return m.backup
}

// StatusSummary returns a status summary string.
func (m *Manage) StatusSummary() string {
	var parts []string

	// Discovery status
	if m.discovery.Scanning() {
		parts = append(parts, "Scanning...")
	} else {
		devices := m.discovery.Devices()
		if len(devices) > 0 {
			parts = append(parts, m.styles.Muted.Render(
				strings.ReplaceAll("Discovered: %d", "%d", string(rune('0'+len(devices)))),
			))
		}
	}

	// Firmware status
	if m.firmware.Checking() {
		parts = append(parts, "Checking firmware...")
	} else if m.firmware.Updating() {
		parts = append(parts, "Updating firmware...")
	} else if count := m.firmware.UpdateCount(); count > 0 {
		parts = append(parts, m.styles.Title.Render(
			strings.ReplaceAll("Updates: %d", "%d", string(rune('0'+count))),
		))
	}

	// Batch status
	if m.batch.Executing() {
		parts = append(parts, "Executing batch operation...")
	}

	// Backup status
	if m.backup.Exporting() {
		parts = append(parts, "Exporting backups...")
	} else if m.backup.Importing() {
		parts = append(parts, "Importing backup...")
	}

	if len(parts) == 0 {
		return m.styles.Muted.Render("Device management ready")
	}

	return strings.Join(parts, " | ")
}

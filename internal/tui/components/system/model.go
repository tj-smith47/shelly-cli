// Package system provides TUI components for managing device system settings.
package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Deps holds the dependencies for the System component.
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

// StatusLoadedMsg signals that system status was loaded.
type StatusLoadedMsg struct {
	Status *shelly.SysStatus
	Config *shelly.SysConfig
	Err    error
}

// SettingField represents a configurable field.
type SettingField int

// Setting field constants.
const (
	FieldName SettingField = iota
	FieldTimezone
	FieldEcoMode
	FieldDiscoverable
	FieldCount
)

// Model displays system settings for a device.
type Model struct {
	ctx        context.Context
	svc        *shelly.Service
	device     string
	status     *shelly.SysStatus
	config     *shelly.SysConfig
	cursor     SettingField
	loading    bool
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int // 1-based panel index for Shift+N hotkey hint
	styles     Styles
}

// Styles holds styles for the System component.
type Styles struct {
	Label        lipgloss.Style
	Value        lipgloss.Style
	ValueEnabled lipgloss.Style
	ValueMuted   lipgloss.Style
	Selected     lipgloss.Style
	Error        lipgloss.Style
	Muted        lipgloss.Style
	Title        lipgloss.Style
}

// DefaultStyles returns the default styles for the System component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		ValueEnabled: lipgloss.NewStyle().
			Foreground(colors.Online),
		ValueMuted: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// New creates a new System model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("system: invalid deps: %v", err))
	}

	return Model{
		ctx:     deps.Ctx,
		svc:     deps.Svc,
		loading: false,
		styles:  DefaultStyles(),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display system settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.status = nil
	m.config = nil
	m.cursor = FieldName
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, m.fetchStatus()
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetSysStatus(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}

		config, configErr := m.svc.GetSysConfig(ctx, m.device)
		if configErr != nil {
			return StatusLoadedMsg{Status: status, Err: configErr}
		}

		return StatusLoadedMsg{Status: status, Config: config}
	}
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

// SetPanelIndex sets the 1-based panel index for Shift+N hotkey hint.
func (m Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.status = msg.Status
		m.config = msg.Config
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
	case "r":
		if !m.loading && m.device != "" {
			m.loading = true
			return m, m.fetchStatus()
		}
	case "t":
		return m.toggleCurrentField()
	}

	return m, nil
}

func (m Model) cursorDown() Model {
	if m.cursor < FieldCount-1 {
		m.cursor++
	}
	return m
}

func (m Model) cursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
	}
	return m
}

func (m Model) toggleCurrentField() (Model, tea.Cmd) {
	if m.config == nil || m.device == "" {
		return m, nil
	}

	switch m.cursor {
	case FieldEcoMode:
		newVal := !m.config.EcoMode
		return m, m.setEcoMode(newVal)
	case FieldDiscoverable:
		newVal := !m.config.Discoverable
		return m, m.setDiscoverable(newVal)
	case FieldName, FieldTimezone, FieldCount:
		// These fields are not toggleable
	}

	return m, nil
}

func (m Model) setEcoMode(enable bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SetSysEcoMode(ctx, m.device, enable)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}

		// Refresh status after change
		status, err := m.svc.GetSysStatus(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}
		config, err := m.svc.GetSysConfig(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Status: status, Err: err}
		}
		return StatusLoadedMsg{Status: status, Config: config}
	}
}

func (m Model) setDiscoverable(discoverable bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SetSysDiscoverable(ctx, m.device, discoverable)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}

		// Refresh status after change
		status, err := m.svc.GetSysStatus(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}
		config, err := m.svc.GetSysConfig(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Status: status, Err: err}
		}
		return StatusLoadedMsg{Status: status, Config: config}
	}
}

// View renders the System component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("System").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	if m.device == "" {
		r.SetContent(m.styles.Muted.Render("No device selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.styles.Muted.Render("Loading system settings..."))
		return r.Render()
	}

	if m.err != nil {
		r.SetContent(m.styles.Error.Render("Error: " + m.err.Error()))
		return r.Render()
	}

	var content strings.Builder

	// System status section
	content.WriteString(m.renderStatus())
	content.WriteString("\n\n")

	// Configuration section
	content.WriteString(m.styles.Title.Render("Settings"))
	content.WriteString("\n")
	content.WriteString(m.renderSettings())

	// Help text
	content.WriteString("\n\n")
	content.WriteString(m.styles.Muted.Render("t: toggle | r: refresh"))

	r.SetContent(content.String())
	return r.Render()
}

func (m Model) renderStatus() string {
	if m.status == nil {
		return m.styles.Muted.Render("No status available")
	}

	var content strings.Builder

	// Uptime
	content.WriteString(m.styles.Label.Render("Uptime:    "))
	content.WriteString(m.styles.Value.Render(formatUptime(m.status.Uptime)))
	content.WriteString("\n")

	// Memory
	content.WriteString(m.styles.Label.Render("RAM:       "))
	ramUsed := m.status.RAMSize - m.status.RAMFree
	ramPct := float64(ramUsed) / float64(m.status.RAMSize) * 100
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d/%d KB (%.0f%%)",
		ramUsed/1024, m.status.RAMSize/1024, ramPct)))
	content.WriteString("\n")

	// Filesystem
	content.WriteString(m.styles.Label.Render("Storage:   "))
	fsUsed := m.status.FSSize - m.status.FSFree
	fsPct := float64(fsUsed) / float64(m.status.FSSize) * 100
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d/%d KB (%.0f%%)",
		fsUsed/1024, m.status.FSSize/1024, fsPct)))
	content.WriteString("\n")

	// Time
	if m.status.Time != "" {
		content.WriteString(m.styles.Label.Render("Time:      "))
		content.WriteString(m.styles.Value.Render(m.status.Time))
		content.WriteString("\n")
	}

	// Update available
	if m.status.UpdateAvailable != "" {
		content.WriteString(m.styles.Label.Render("Update:    "))
		content.WriteString(m.styles.ValueEnabled.Render(m.status.UpdateAvailable + " available"))
		content.WriteString("\n")
	}

	// Restart required
	if m.status.RestartRequired {
		content.WriteString(m.styles.Error.Render("⚠ Restart required"))
	}

	return content.String()
}

func (m Model) renderSettings() string {
	if m.config == nil {
		return m.styles.Muted.Render("No configuration available")
	}

	var content strings.Builder

	// Name field
	content.WriteString(m.renderSettingLine(FieldName, "Name", m.config.Name))
	content.WriteString("\n")

	// Timezone field
	tz := m.config.Timezone
	if tz == "" {
		tz = "(not set)"
	}
	content.WriteString(m.renderSettingLine(FieldTimezone, "Timezone", tz))
	content.WriteString("\n")

	// Eco Mode field
	ecoModeVal := "Disabled"
	if m.config.EcoMode {
		ecoModeVal = "Enabled"
	}
	content.WriteString(m.renderToggleLine(FieldEcoMode, "Eco Mode", m.config.EcoMode, ecoModeVal))
	content.WriteString("\n")

	// Discoverable field
	discVal := "Hidden"
	if m.config.Discoverable {
		discVal = "Visible"
	}
	content.WriteString(m.renderToggleLine(FieldDiscoverable, "Discoverable", m.config.Discoverable, discVal))

	return content.String()
}

func (m Model) renderSettingLine(field SettingField, label, value string) string {
	selector := "  "
	if m.cursor == field {
		selector = "▶ "
	}

	labelWidth := 14
	paddedLabel := fmt.Sprintf("%-*s", labelWidth, label+":")

	line := selector + m.styles.Label.Render(paddedLabel) + m.styles.Value.Render(value)

	if m.cursor == field {
		return m.styles.Selected.Render(line)
	}
	return line
}

func (m Model) renderToggleLine(field SettingField, label string, enabled bool, value string) string {
	selector := "  "
	if m.cursor == field {
		selector = "▶ "
	}

	labelWidth := 14
	paddedLabel := fmt.Sprintf("%-*s", labelWidth, label+":")

	var valueStyle lipgloss.Style
	if enabled {
		valueStyle = m.styles.ValueEnabled
	} else {
		valueStyle = m.styles.ValueMuted
	}

	line := selector + m.styles.Label.Render(paddedLabel) + valueStyle.Render(value)

	if m.cursor == field {
		return m.styles.Selected.Render(line)
	}
	return line
}

func formatUptime(seconds int) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	mins := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

// Status returns the current system status.
func (m Model) Status() *shelly.SysStatus {
	return m.status
}

// Config returns the current system config.
func (m Model) Config() *shelly.SysConfig {
	return m.config
}

// Device returns the current device address.
func (m Model) Device() string {
	return m.device
}

// Loading returns whether the component is loading.
func (m Model) Loading() bool {
	return m.loading
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Refresh triggers a refresh of the system status.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, m.fetchStatus()
}

// Package cloud provides TUI components for managing device cloud settings.
package cloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// EditSaveResultMsg signals a save operation completed.
type EditSaveResultMsg struct {
	Enabled bool
	Err     error
}

// EditOpenedMsg signals the edit modal was opened.
type EditOpenedMsg struct{}

// EditClosedMsg signals the edit modal was closed.
type EditClosedMsg struct {
	Saved bool
}

// EditModel represents the cloud configuration edit modal.
type EditModel struct {
	ctx     context.Context
	svc     *shelly.Service
	device  string
	visible bool
	saving  bool
	err     error
	width   int
	height  int
	styles  EditStyles

	// Cloud state
	connected bool   // Current connection status
	enabled   bool   // Current enabled state
	server    string // Cloud server URL

	// Pending change
	pendingEnabled bool
}

// EditStyles holds styles for the edit modal.
type EditStyles struct {
	Overlay        lipgloss.Style
	Modal          lipgloss.Style
	Title          lipgloss.Style
	Label          lipgloss.Style
	Value          lipgloss.Style
	Error          lipgloss.Style
	Help           lipgloss.Style
	StatusOn       lipgloss.Style
	StatusOff      lipgloss.Style
	ToggleEnabled  lipgloss.Style
	ToggleDisabled lipgloss.Style
	Warning        lipgloss.Style
}

// DefaultEditStyles returns the default edit modal styles.
func DefaultEditStyles() EditStyles {
	colors := theme.GetSemanticColors()
	return EditStyles{
		Overlay: lipgloss.NewStyle().
			Background(lipgloss.Color("#000000")),
		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder).
			Background(colors.Background).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			MarginBottom(1),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Width(12),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted),
		StatusOn: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		StatusOff: lipgloss.NewStyle().
			Foreground(colors.Offline).
			Bold(true),
		ToggleEnabled: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		ToggleDisabled: lipgloss.NewStyle().
			Foreground(colors.Offline).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
	}
}

// NewEditModel creates a new cloud configuration edit modal.
func NewEditModel(ctx context.Context, svc *shelly.Service) EditModel {
	return EditModel{
		ctx:    ctx,
		svc:    svc,
		styles: DefaultEditStyles(),
	}
}

// Show displays the edit modal with the given device and cloud state.
func (m EditModel) Show(device string, connected, enabled bool, server string) EditModel {
	m.device = device
	m.visible = true
	m.saving = false
	m.err = nil
	m.connected = connected
	m.enabled = enabled
	m.server = server
	m.pendingEnabled = enabled

	return m
}

// Hide hides the edit modal.
func (m EditModel) Hide() EditModel {
	m.visible = false
	return m
}

// Visible returns whether the modal is visible.
func (m EditModel) Visible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m EditModel) SetSize(width, height int) EditModel {
	m.width = width
	m.height = height
	return m
}

// Init returns the initial command.
func (m EditModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case EditSaveResultMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - update state and close modal
		m.enabled = msg.Enabled
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case "enter":
		return m.save()

	case "t", " ":
		// Toggle cloud enabled state
		if !m.saving {
			m.pendingEnabled = !m.pendingEnabled
			return m, nil
		}
	}

	return m, nil
}

func (m EditModel) save() (EditModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	// Check if anything changed
	if m.pendingEnabled == m.enabled {
		// No changes, just close
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	}

	m.saving = true
	m.err = nil

	return m, m.createSaveCmd()
}

func (m EditModel) createSaveCmd() tea.Cmd {
	newEnabled := m.pendingEnabled
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SetCloudEnabled(ctx, m.device, newEnabled)
		return EditSaveResultMsg{Enabled: newEnabled, Err: err}
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(m.styles.Title.Render("Cloud Configuration"))
	content.WriteString("\n\n")

	// Connection status
	content.WriteString(m.renderConnectionStatus())
	content.WriteString("\n\n")

	// Server (read-only)
	if m.server != "" {
		content.WriteString(m.renderServer())
		content.WriteString("\n\n")
	}

	// Enable toggle
	content.WriteString(m.renderToggle())
	content.WriteString("\n")

	// Show change indicator if modified
	if m.pendingEnabled != m.enabled {
		content.WriteString(m.renderChangeIndicator())
		content.WriteString("\n")
	}

	// Error display
	if m.err != nil {
		content.WriteString("\n")
		content.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
		content.WriteString("\n")
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(m.renderHelpText())

	// Render modal box
	modalContent := content.String()
	modalWidth := min(50, m.width-4)
	modal := m.styles.Modal.Width(modalWidth).Render(modalContent)

	// Center the modal
	return m.centerModal(modal)
}

func (m EditModel) renderConnectionStatus() string {
	var content strings.Builder

	content.WriteString(m.styles.Label.Render("Status:"))
	content.WriteString(" ")

	if m.connected {
		content.WriteString(m.styles.StatusOn.Render("● Connected"))
	} else {
		content.WriteString(m.styles.StatusOff.Render("○ Disconnected"))
	}

	return content.String()
}

func (m EditModel) renderServer() string {
	var content strings.Builder

	content.WriteString(m.styles.Label.Render("Server:"))
	content.WriteString(" ")
	content.WriteString(m.styles.Value.Render(m.server))

	return content.String()
}

func (m EditModel) renderToggle() string {
	var content strings.Builder

	content.WriteString(m.styles.Label.Render("Enabled:"))
	content.WriteString(" ")

	if m.pendingEnabled {
		content.WriteString(m.styles.ToggleEnabled.Render("[●] ON "))
	} else {
		content.WriteString(m.styles.ToggleDisabled.Render("[ ] OFF"))
	}

	return content.String()
}

func (m EditModel) renderChangeIndicator() string {
	var msg string
	if m.pendingEnabled {
		msg = "Will enable cloud connection"
	} else {
		msg = "Will disable cloud connection"
	}
	return m.styles.Warning.Render(fmt.Sprintf("  ⚡ %s", msg))
}

func (m EditModel) renderHelpText() string {
	if m.saving {
		return m.styles.Help.Render("Saving...")
	}
	return m.styles.Help.Render("t/Space: Toggle | Enter: Save | Esc: Cancel")
}

func (m EditModel) centerModal(modal string) string {
	lines := strings.Split(modal, "\n")
	modalHeight := len(lines)
	modalWidth := 0
	for _, line := range lines {
		if lipgloss.Width(line) > modalWidth {
			modalWidth = lipgloss.Width(line)
		}
	}

	// Calculate centering
	topPad := (m.height - modalHeight) / 2
	leftPad := (m.width - modalWidth) / 2

	if topPad < 0 {
		topPad = 0
	}
	if leftPad < 0 {
		leftPad = 0
	}

	// Build centered output
	var result strings.Builder
	for range topPad {
		result.WriteString("\n")
	}

	padding := strings.Repeat(" ", leftPad)
	for _, line := range lines {
		result.WriteString(padding)
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

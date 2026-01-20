// Package cloud provides TUI components for managing device cloud settings.
package cloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// EditSaveResultMsg is an alias for the shared save result message.
type EditSaveResultMsg = messages.SaveResultMsg

// EditOpenedMsg is an alias for the shared edit opened message.
type EditOpenedMsg = messages.EditOpenedMsg

// EditClosedMsg is an alias for the shared edit closed message.
type EditClosedMsg = messages.EditClosedMsg

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
	styles  editmodal.Styles

	// Cloud state
	connected bool   // Current connection status
	enabled   bool   // Current enabled state
	server    string // Cloud server URL

	// Pending change
	pendingEnabled bool
}

// NewEditModel creates a new cloud configuration edit modal.
func NewEditModel(ctx context.Context, svc *shelly.Service) EditModel {
	return EditModel{
		ctx:    ctx,
		svc:    svc,
		styles: editmodal.DefaultStyles().WithLabelWidth(12),
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

	return m.handleMessage(msg)
}

func (m EditModel) handleMessage(msg tea.Msg) (EditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SaveResultMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - update state and close modal
		m.enabled = m.pendingEnabled
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	// Action messages from context system
	case messages.NavigationMsg:
		// Single toggle component - navigation not applicable
		return m, nil
	case messages.ToggleEnableRequestMsg:
		if !m.saving {
			m.pendingEnabled = !m.pendingEnabled
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	// Modal-specific keys not covered by action messages
	switch msg.String() {
	case keyconst.KeyEsc, "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case keyconst.KeyEnter:
		return m.save()

	case "t":
		// Toggle cloud enabled state (t key for "toggle")
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
		if err != nil {
			return messages.NewSaveError(nil, err)
		}
		return messages.NewSaveResult(nil)
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	// Build footer
	footer := "Space: Toggle | Ctrl+S: Save | Esc: Cancel"
	if m.saving {
		footer = "Saving..."
	}

	// Use common modal helper
	r := rendering.NewModal(m.width, m.height, "Cloud Configuration", footer)

	// Build content
	var content strings.Builder

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

	// Show change indicator if modified
	if m.pendingEnabled != m.enabled {
		content.WriteString("\n")
		content.WriteString(m.renderChangeIndicator())
	}

	// Error display
	if m.err != nil {
		content.WriteString("\n")
		msg, _ := tuierrors.FormatError(m.err)
		content.WriteString(m.styles.Error.Render(msg))
	}

	return r.SetContent(content.String()).Render()
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
		content.WriteString(m.styles.StatusOn.Render("[●] ON "))
	} else {
		content.WriteString(m.styles.StatusOff.Render("[ ] OFF"))
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

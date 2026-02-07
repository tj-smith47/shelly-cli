// Package cloud provides TUI components for managing device cloud settings.
package cloud

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

// EditSaveResultMsg is an alias for the shared save result message.
type EditSaveResultMsg = messages.SaveResultMsg

// EditOpenedMsg is an alias for the shared edit opened message.
type EditOpenedMsg = messages.EditOpenedMsg

// EditClosedMsg is an alias for the shared edit closed message.
type EditClosedMsg = messages.EditClosedMsg

// EditModel represents the cloud configuration edit modal.
type EditModel struct {
	editmodal.Base

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
		Base: editmodal.Base{
			Ctx:    ctx,
			Svc:    svc,
			Styles: editmodal.DefaultStyles().WithLabelWidth(12),
		},
	}
}

// Show displays the edit modal with the given device and cloud state.
func (m EditModel) Show(device string, connected, enabled bool, server string) EditModel {
	m.Base.Show(device, 0)
	m.connected = connected
	m.enabled = enabled
	m.server = server
	m.pendingEnabled = enabled

	return m
}

// Hide hides the edit modal.
func (m EditModel) Hide() EditModel {
	m.Base.Hide()
	return m
}

// Visible returns whether the modal is visible.
func (m EditModel) Visible() bool {
	return m.Base.Visible()
}

// SetSize sets the modal dimensions.
func (m EditModel) SetSize(width, height int) EditModel {
	m.Base.SetSize(width, height)
	return m
}

// Update handles messages.
func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	if !m.Visible() {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m EditModel) handleMessage(msg tea.Msg) (EditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SaveResultMsg:
		saved, cmd := m.HandleSaveResult(msg)
		if saved {
			m.enabled = m.pendingEnabled
		}
		return m, cmd

	// Action messages from context system
	case messages.NavigationMsg:
		// Single toggle component - navigation not applicable
		return m, nil
	case messages.ToggleEnableRequestMsg:
		if !m.Saving {
			m.pendingEnabled = !m.pendingEnabled
		}
		return m, nil
	case tea.KeyPressMsg:
		action := m.HandleKey(msg)
		if action != editmodal.ActionNone {
			return m.applyAction(action)
		}
		return m.handleCustomKey(msg)
	}

	return m, nil
}

func (m EditModel) applyAction(action editmodal.KeyAction) (EditModel, tea.Cmd) {
	switch action {
	case editmodal.ActionNone:
		return m, nil
	case editmodal.ActionClose:
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	case editmodal.ActionSave:
		return m.save()
	case editmodal.ActionNext, editmodal.ActionNavDown:
		// Single toggle, no navigation needed
		return m, nil
	case editmodal.ActionPrev, editmodal.ActionNavUp:
		return m, nil
	}
	return m, nil
}

func (m EditModel) handleCustomKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	if msg.String() == "t" && !m.Saving {
		m.pendingEnabled = !m.pendingEnabled
	}

	return m, nil
}

func (m EditModel) save() (EditModel, tea.Cmd) {
	if m.Saving {
		return m, nil
	}

	// Check if anything changed
	if m.pendingEnabled == m.enabled {
		// No changes, just close
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	}

	m.StartSave()
	newEnabled := m.pendingEnabled
	cmd := m.SaveCmd(func(ctx context.Context) error {
		return m.Svc.SetCloudEnabled(ctx, m.Device, newEnabled)
	})

	return m, cmd
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.Visible() {
		return ""
	}

	footer := m.RenderSavingFooter("Space: Toggle | Ctrl+S: Save | Esc: Cancel")

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
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n")
		content.WriteString(errStr)
	}

	return m.RenderModal("Cloud Configuration", content.String(), footer)
}

func (m EditModel) renderConnectionStatus() string {
	var content strings.Builder

	content.WriteString(m.Styles.Label.Render("Status:"))
	content.WriteString(" ")

	if m.connected {
		content.WriteString(m.Styles.StatusOn.Render("● Connected"))
	} else {
		content.WriteString(m.Styles.StatusOff.Render("○ Disconnected"))
	}

	return content.String()
}

func (m EditModel) renderServer() string {
	var content strings.Builder

	content.WriteString(m.Styles.Label.Render("Server:"))
	content.WriteString(" ")
	content.WriteString(m.Styles.Value.Render(m.server))

	return content.String()
}

func (m EditModel) renderToggle() string {
	var content strings.Builder

	content.WriteString(m.Styles.Label.Render("Enabled:"))
	content.WriteString(" ")

	if m.pendingEnabled {
		content.WriteString(m.Styles.StatusOn.Render("[●] ON "))
	} else {
		content.WriteString(m.Styles.StatusOff.Render("[ ] OFF"))
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
	return m.Styles.Warning.Render(fmt.Sprintf("  ⚡ %s", msg))
}

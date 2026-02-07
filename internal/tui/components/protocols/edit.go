// Package protocols provides TUI components for managing device protocol settings.
package protocols

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

// MQTTEditField represents a field in the MQTT edit form.
type MQTTEditField int

// MQTT edit field constants.
const (
	MQTTFieldEnable MQTTEditField = iota
	MQTTFieldServer
	MQTTFieldUser
	MQTTFieldPassword
	MQTTFieldClientID
	MQTTFieldTopicPrefix
	MQTTFieldTLS
	MQTTFieldCount
)

// TLSOption represents TLS configuration options.
type TLSOption int

// TLS option constants.
const (
	TLSNone TLSOption = iota
	TLSNoVerify
	TLSDefaultCA
	TLSUserCA
)

// MQTTEditSaveResultMsg is an alias for the shared save result message.
type MQTTEditSaveResultMsg = messages.SaveResultMsg

// MQTTEditOpenedMsg is an alias for the shared edit opened message.
type MQTTEditOpenedMsg = messages.EditOpenedMsg

// MQTTEditClosedMsg is an alias for the shared edit closed message.
type MQTTEditClosedMsg = messages.EditClosedMsg

// MQTTEditModel represents the MQTT configuration edit modal.
type MQTTEditModel struct {
	editmodal.Base

	// Original config for comparison
	originalData *MQTTData

	// Form inputs
	enableToggle     form.Toggle
	serverInput      form.TextInput
	userInput        form.TextInput
	passwordInput    form.Password
	clientIDInput    form.TextInput
	topicPrefixInput form.TextInput
	tlsDropdown      form.Select
}

// NewMQTTEditModel creates a new MQTT configuration edit modal.
func NewMQTTEditModel(ctx context.Context, svc *shelly.Service) MQTTEditModel {
	enableToggle := form.NewToggle(
		form.WithToggleOnLabel("Enabled"),
		form.WithToggleOffLabel("Disabled"),
	)

	serverInput := form.NewTextInput(
		form.WithPlaceholder("mqtt.example.com:1883"),
		form.WithCharLimit(256),
		form.WithWidth(30),
		form.WithHelp("Hostname:port (default 1883 for TCP, 8883 for TLS)"),
	)

	userInput := form.NewTextInput(
		form.WithPlaceholder("username"),
		form.WithCharLimit(64),
		form.WithWidth(30),
	)

	passwordInput := form.NewPassword(
		form.WithPasswordPlaceholder("password"),
		form.WithPasswordCharLimit(64),
		form.WithPasswordWidth(30),
		form.WithPasswordHelp("Ctrl+T to show/hide"),
	)

	clientIDInput := form.NewTextInput(
		form.WithPlaceholder("(auto: device ID)"),
		form.WithCharLimit(128),
		form.WithWidth(30),
		form.WithHelp("Leave empty to use device ID"),
	)

	topicPrefixInput := form.NewTextInput(
		form.WithPlaceholder("(auto: device ID)"),
		form.WithCharLimit(300),
		form.WithWidth(30),
		form.WithHelp("Max 300 chars, no $, #, +, %, ?"),
	)

	tlsDropdown := form.NewSelect(
		form.WithSelectOptions([]string{"No TLS", "TLS (no verify)", "TLS (default CA)", "TLS (user CA)"}),
		form.WithSelectSelected(0),
		form.WithSelectMaxVisible(4),
	)

	return MQTTEditModel{
		Base: editmodal.Base{
			Ctx:    ctx,
			Svc:    svc,
			Styles: editmodal.DefaultStyles(),
		},
		enableToggle:     enableToggle,
		serverInput:      serverInput,
		userInput:        userInput,
		passwordInput:    passwordInput,
		clientIDInput:    clientIDInput,
		topicPrefixInput: topicPrefixInput,
		tlsDropdown:      tlsDropdown,
	}
}

// Show displays the edit modal with the given device and MQTT data.
func (m MQTTEditModel) Show(device string, data *MQTTData) MQTTEditModel {
	m.Base.Show(device, int(MQTTFieldCount))
	m.originalData = data

	// Populate form fields from current data
	if data != nil {
		m.enableToggle = m.enableToggle.SetValue(data.Enable)
		m.serverInput = m.serverInput.SetValue(data.Server)
		m.userInput = m.userInput.SetValue(data.User)
		m.clientIDInput = m.clientIDInput.SetValue(data.ClientID)
		m.topicPrefixInput = m.topicPrefixInput.SetValue(data.TopicPrefix)
		// Note: Password is write-only, we don't have access to it
		m.passwordInput = m.passwordInput.Reset()

		// Set TLS dropdown based on SSLCA value
		m.tlsDropdown = m.setTLSFromSSLCA(data.SSLCA)
	} else {
		m.enableToggle = m.enableToggle.SetValue(false)
		m.serverInput = m.serverInput.SetValue("")
		m.userInput = m.userInput.SetValue("")
		m.passwordInput = m.passwordInput.Reset()
		m.clientIDInput = m.clientIDInput.SetValue("")
		m.topicPrefixInput = m.topicPrefixInput.SetValue("")
		m.tlsDropdown = m.tlsDropdown.SetSelected(0)
	}

	// Focus first field
	m.enableToggle = m.enableToggle.Focus()

	return m
}

// setTLSFromSSLCA converts SSLCA config value to dropdown selection.
func (m MQTTEditModel) setTLSFromSSLCA(sslca string) form.Select {
	switch sslca {
	case "":
		return m.tlsDropdown.SetSelected(0) // No TLS
	case "*":
		return m.tlsDropdown.SetSelected(1) // TLS no verify
	case "ca.pem":
		return m.tlsDropdown.SetSelected(2) // TLS default CA
	case "user_ca.pem":
		return m.tlsDropdown.SetSelected(3) // TLS user CA
	default:
		return m.tlsDropdown.SetSelected(0)
	}
}

// getSSLCAFromDropdown converts dropdown selection to SSLCA config value.
func (m MQTTEditModel) getSSLCAFromDropdown() string {
	switch m.tlsDropdown.Selected() {
	case 1:
		return "*"
	case 2:
		return "ca.pem"
	case 3:
		return "user_ca.pem"
	default:
		return ""
	}
}

// Hide hides the edit modal.
func (m MQTTEditModel) Hide() MQTTEditModel {
	m.Base.Hide()
	m = m.blurAllFields()
	return m
}

// IsVisible returns whether the modal is visible.
func (m MQTTEditModel) IsVisible() bool {
	return m.Visible()
}

// SetSize sets the modal dimensions.
func (m MQTTEditModel) SetSize(width, height int) MQTTEditModel {
	m.Base.SetSize(width, height)
	return m
}

// Init returns the initial command.
func (m MQTTEditModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m MQTTEditModel) Update(msg tea.Msg) (MQTTEditModel, tea.Cmd) {
	if !m.IsVisible() {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m MQTTEditModel) handleMessage(msg tea.Msg) (MQTTEditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SaveResultMsg:
		_, cmd := m.HandleSaveResult(msg)
		return m, cmd

	// Action messages from context system
	case messages.NavigationMsg:
		if m.Saving {
			return m, nil
		}
		action := m.HandleNavigation(msg)
		return m.applyAction(action)
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m MQTTEditModel) applyAction(action editmodal.KeyAction) (MQTTEditModel, tea.Cmd) {
	switch action {
	case editmodal.ActionClose:
		m = m.Hide()
		return m, func() tea.Msg { return MQTTEditClosedMsg{Saved: false} }
	case editmodal.ActionSave:
		return m.save()
	case editmodal.ActionNext, editmodal.ActionNavDown:
		m = m.moveFocus(m.NextField())
		return m, nil
	case editmodal.ActionPrev, editmodal.ActionNavUp:
		m = m.moveFocus(m.PrevField())
		return m, nil
	case editmodal.ActionNone:
		// No action to take
	}
	return m, nil
}

func (m MQTTEditModel) handleKey(msg tea.KeyPressMsg) (MQTTEditModel, tea.Cmd) {
	// Handle dropdown expansion separately
	if MQTTEditField(m.Cursor) == MQTTFieldTLS && m.tlsDropdown.IsExpanded() {
		var cmd tea.Cmd
		m.tlsDropdown, cmd = m.tlsDropdown.Update(msg)
		return m, cmd
	}

	if m.Saving {
		return m, nil
	}

	// This modal has text inputs and password fields, so Enter should NOT trigger save.
	// Only Ctrl+S saves; Enter is forwarded to focused inputs.
	switch msg.String() {
	case keyconst.KeyEsc, "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return MQTTEditClosedMsg{Saved: false} }

	case keyconst.KeyCtrlS:
		return m.save()

	case keyconst.KeyEnter:
		// If on TLS dropdown and not expanded, expand it
		if MQTTEditField(m.Cursor) == MQTTFieldTLS && !m.tlsDropdown.IsExpanded() {
			m.tlsDropdown = m.tlsDropdown.Expand()
			return m, nil
		}
		// Forward Enter to focused input (not save)
		return m.updateFocusedInput(msg)

	case keyconst.KeyTab:
		m = m.moveFocus(m.NextField())
		return m, nil

	case keyconst.KeyShiftTab:
		m = m.moveFocus(m.PrevField())
		return m, nil
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m MQTTEditModel) updateFocusedInput(msg tea.Msg) (MQTTEditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch MQTTEditField(m.Cursor) {
	case MQTTFieldEnable:
		m.enableToggle, cmd = m.enableToggle.Update(msg)
	case MQTTFieldServer:
		m.serverInput, cmd = m.serverInput.Update(msg)
	case MQTTFieldUser:
		m.userInput, cmd = m.userInput.Update(msg)
	case MQTTFieldPassword:
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	case MQTTFieldClientID:
		m.clientIDInput, cmd = m.clientIDInput.Update(msg)
	case MQTTFieldTopicPrefix:
		m.topicPrefixInput, cmd = m.topicPrefixInput.Update(msg)
	case MQTTFieldTLS:
		m.tlsDropdown, cmd = m.tlsDropdown.Update(msg)
	case MQTTFieldCount:
		// No input to update
	}

	return m, cmd
}

// moveFocus blurs the old field and focuses the new one.
func (m MQTTEditModel) moveFocus(oldCursor, newCursor int) MQTTEditModel {
	m = m.blurField(MQTTEditField(oldCursor))
	m = m.focusField(MQTTEditField(newCursor))
	return m
}

func (m MQTTEditModel) blurField(field MQTTEditField) MQTTEditModel {
	switch field {
	case MQTTFieldEnable:
		m.enableToggle = m.enableToggle.Blur()
	case MQTTFieldServer:
		m.serverInput = m.serverInput.Blur()
	case MQTTFieldUser:
		m.userInput = m.userInput.Blur()
	case MQTTFieldPassword:
		m.passwordInput = m.passwordInput.Blur()
	case MQTTFieldClientID:
		m.clientIDInput = m.clientIDInput.Blur()
	case MQTTFieldTopicPrefix:
		m.topicPrefixInput = m.topicPrefixInput.Blur()
	case MQTTFieldTLS:
		m.tlsDropdown = m.tlsDropdown.Blur()
	case MQTTFieldCount:
		// Nothing
	}
	return m
}

func (m MQTTEditModel) focusField(field MQTTEditField) MQTTEditModel {
	switch field {
	case MQTTFieldEnable:
		m.enableToggle = m.enableToggle.Focus()
	case MQTTFieldServer:
		m.serverInput, _ = m.serverInput.Focus()
	case MQTTFieldUser:
		m.userInput, _ = m.userInput.Focus()
	case MQTTFieldPassword:
		m.passwordInput, _ = m.passwordInput.Focus()
	case MQTTFieldClientID:
		m.clientIDInput, _ = m.clientIDInput.Focus()
	case MQTTFieldTopicPrefix:
		m.topicPrefixInput, _ = m.topicPrefixInput.Focus()
	case MQTTFieldTLS:
		m.tlsDropdown = m.tlsDropdown.Focus()
	case MQTTFieldCount:
		// Nothing
	}
	return m
}

func (m MQTTEditModel) blurAllFields() MQTTEditModel {
	m.enableToggle = m.enableToggle.Blur()
	m.serverInput = m.serverInput.Blur()
	m.userInput = m.userInput.Blur()
	m.passwordInput = m.passwordInput.Blur()
	m.clientIDInput = m.clientIDInput.Blur()
	m.topicPrefixInput = m.topicPrefixInput.Blur()
	m.tlsDropdown = m.tlsDropdown.Blur()
	return m
}

func (m MQTTEditModel) save() (MQTTEditModel, tea.Cmd) {
	if m.Saving {
		return m, nil
	}

	// Validate server address when enabling
	if m.enableToggle.Value() && m.serverInput.Value() == "" {
		m.Err = fmt.Errorf("server address is required when enabling MQTT")
		return m, nil
	}

	// Validate topic prefix characters
	topicPrefix := m.topicPrefixInput.Value()
	if topicPrefix != "" {
		if strings.HasPrefix(topicPrefix, "$") {
			m.Err = fmt.Errorf("topic prefix cannot start with $")
			return m, nil
		}
		for _, forbidden := range []string{"#", "+", "%", "?"} {
			if strings.Contains(topicPrefix, forbidden) {
				m.Err = fmt.Errorf("topic prefix cannot contain %s", forbidden)
				return m, nil
			}
		}
	}

	m.StartSave()

	enabled := m.enableToggle.Value()
	params := shelly.MQTTSetConfigParams{
		Enable:      &enabled,
		Server:      m.serverInput.Value(),
		User:        m.userInput.Value(),
		Password:    m.passwordInput.Value(),
		ClientID:    m.clientIDInput.Value(),
		TopicPrefix: m.topicPrefixInput.Value(),
		SSLCA:       m.getSSLCAFromDropdown(),
	}

	device := m.Device
	cmd := m.SaveCmd(func(ctx context.Context) error {
		return m.Svc.SetMQTTConfigFull(ctx, device, params)
	})

	return m, cmd
}

// View renders the edit modal.
func (m MQTTEditModel) View() string {
	if !m.IsVisible() {
		return ""
	}

	// Build footer
	footer := m.RenderSavingFooter("Tab: Next | Ctrl+S: Save | Esc: Cancel")

	// Build content
	var content strings.Builder

	// Connection status (if we have original data)
	if m.originalData != nil {
		content.WriteString(m.renderConnectionStatus())
		content.WriteString("\n\n")
	}

	// Form fields
	content.WriteString(m.renderFormFields())

	return m.RenderModal("MQTT Configuration", content.String(), footer)
}

func (m MQTTEditModel) renderConnectionStatus() string {
	var content strings.Builder

	content.WriteString(m.Styles.Label.Render("Status:"))
	content.WriteString(" ")

	switch {
	case m.originalData.Connected:
		content.WriteString(m.Styles.StatusOn.Render("● Connected"))
	case m.originalData.Enable:
		content.WriteString(m.Styles.Warning.Render("◐ Enabled (disconnected)"))
	default:
		content.WriteString(m.Styles.StatusOff.Render("○ Disabled"))
	}

	return content.String()
}

func (m MQTTEditModel) renderFormFields() string {
	var content strings.Builder

	// Enable toggle
	content.WriteString(m.RenderField(int(MQTTFieldEnable), "Enable:", m.enableToggle.View()))
	content.WriteString("\n\n")

	// Server
	content.WriteString(m.RenderField(int(MQTTFieldServer), "Server:", m.serverInput.View()))
	content.WriteString("\n\n")

	// User
	content.WriteString(m.RenderField(int(MQTTFieldUser), "Username:", m.userInput.View()))
	content.WriteString("\n\n")

	// Password
	content.WriteString(m.RenderField(int(MQTTFieldPassword), "Password:", m.passwordInput.View()))
	content.WriteString("\n\n")

	// Client ID
	content.WriteString(m.RenderField(int(MQTTFieldClientID), "Client ID:", m.clientIDInput.View()))
	content.WriteString("\n\n")

	// Topic prefix
	content.WriteString(m.RenderField(int(MQTTFieldTopicPrefix), "Topic Prefix:", m.topicPrefixInput.View()))
	content.WriteString("\n\n")

	// TLS settings
	content.WriteString(m.RenderField(int(MQTTFieldTLS), "TLS:", m.tlsDropdown.View()))

	// Error display
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n\n")
		content.WriteString(errStr)
	}

	return content.String()
}

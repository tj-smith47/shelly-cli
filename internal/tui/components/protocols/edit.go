// Package protocols provides TUI components for managing device protocol settings.
package protocols

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
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
	ctx     context.Context
	svc     *shelly.Service
	device  string
	visible bool
	cursor  MQTTEditField
	saving  bool
	err     error
	width   int
	height  int
	styles  editmodal.Styles

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
		ctx:              ctx,
		svc:              svc,
		styles:           editmodal.DefaultStyles(),
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
	m.device = device
	m.visible = true
	m.cursor = MQTTFieldEnable
	m.saving = false
	m.err = nil
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
	m.visible = false
	m = m.blurAllFields()
	return m
}

// IsVisible returns whether the modal is visible.
func (m MQTTEditModel) IsVisible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m MQTTEditModel) SetSize(width, height int) MQTTEditModel {
	m.width = width
	m.height = height
	return m
}

// Init returns the initial command.
func (m MQTTEditModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m MQTTEditModel) Update(msg tea.Msg) (MQTTEditModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case messages.SaveResultMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - close modal
		m = m.Hide()
		return m, func() tea.Msg { return MQTTEditClosedMsg{Saved: true} }

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m MQTTEditModel) handleKey(msg tea.KeyPressMsg) (MQTTEditModel, tea.Cmd) {
	key := msg.String()

	// Handle dropdown expansion separately
	if m.cursor == MQTTFieldTLS && m.tlsDropdown.IsExpanded() {
		var cmd tea.Cmd
		m.tlsDropdown, cmd = m.tlsDropdown.Update(msg)
		return m, cmd
	}

	switch key {
	case "esc", "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return MQTTEditClosedMsg{Saved: false} }

	case "enter":
		// If on TLS dropdown and not expanded, expand it
		if m.cursor == MQTTFieldTLS && !m.tlsDropdown.IsExpanded() {
			m.tlsDropdown = m.tlsDropdown.Expand()
			return m, nil
		}
		return m.save()

	case "tab":
		return m.nextField(), nil

	case "shift+tab":
		return m.prevField(), nil
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m MQTTEditModel) updateFocusedInput(msg tea.Msg) (MQTTEditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.cursor {
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

func (m MQTTEditModel) nextField() MQTTEditModel {
	m = m.blurCurrentField()
	if m.cursor < MQTTFieldCount-1 {
		m.cursor++
	}
	m = m.focusCurrentField()
	return m
}

func (m MQTTEditModel) prevField() MQTTEditModel {
	m = m.blurCurrentField()
	if m.cursor > 0 {
		m.cursor--
	}
	m = m.focusCurrentField()
	return m
}

func (m MQTTEditModel) blurCurrentField() MQTTEditModel {
	switch m.cursor {
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

func (m MQTTEditModel) focusCurrentField() MQTTEditModel {
	switch m.cursor {
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
	if m.saving {
		return m, nil
	}

	// Validate server address when enabling
	if m.enableToggle.Value() && m.serverInput.Value() == "" {
		m.err = fmt.Errorf("server address is required when enabling MQTT")
		return m, nil
	}

	// Validate topic prefix characters
	topicPrefix := m.topicPrefixInput.Value()
	if topicPrefix != "" {
		if strings.HasPrefix(topicPrefix, "$") {
			m.err = fmt.Errorf("topic prefix cannot start with $")
			return m, nil
		}
		for _, forbidden := range []string{"#", "+", "%", "?"} {
			if strings.Contains(topicPrefix, forbidden) {
				m.err = fmt.Errorf("topic prefix cannot contain %s", forbidden)
				return m, nil
			}
		}
	}

	m.saving = true
	m.err = nil

	return m, m.createSaveCmd()
}

func (m MQTTEditModel) createSaveCmd() tea.Cmd {
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

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SetMQTTConfigFull(ctx, m.device, params)
		if err != nil {
			return messages.NewSaveError(nil, err)
		}
		return messages.NewSaveResult(nil)
	}
}

// View renders the edit modal.
func (m MQTTEditModel) View() string {
	if !m.visible {
		return ""
	}

	// Build footer
	footer := "Tab: Next | Enter: Save | Esc: Cancel"
	if m.saving {
		footer = "Saving..."
	}

	// Use common modal helper
	r := rendering.NewModal(m.width, m.height, "MQTT Configuration", footer)

	// Build content
	var content strings.Builder

	// Connection status (if we have original data)
	if m.originalData != nil {
		content.WriteString(m.renderConnectionStatus())
		content.WriteString("\n\n")
	}

	// Form fields
	content.WriteString(m.renderFormFields())

	return r.SetContent(content.String()).Render()
}

func (m MQTTEditModel) renderConnectionStatus() string {
	var content strings.Builder

	content.WriteString(m.styles.Label.Render("Status:"))
	content.WriteString(" ")

	switch {
	case m.originalData.Connected:
		content.WriteString(m.styles.StatusOn.Render("● Connected"))
	case m.originalData.Enable:
		content.WriteString(m.styles.Warning.Render("◐ Enabled (disconnected)"))
	default:
		content.WriteString(m.styles.StatusOff.Render("○ Disabled"))
	}

	return content.String()
}

func (m MQTTEditModel) renderFormFields() string {
	var content strings.Builder

	// Enable toggle
	content.WriteString(m.renderField(MQTTFieldEnable, "Enable:", m.enableToggle.View()))
	content.WriteString("\n\n")

	// Server
	content.WriteString(m.renderField(MQTTFieldServer, "Server:", m.serverInput.View()))
	content.WriteString("\n\n")

	// User
	content.WriteString(m.renderField(MQTTFieldUser, "Username:", m.userInput.View()))
	content.WriteString("\n\n")

	// Password
	content.WriteString(m.renderField(MQTTFieldPassword, "Password:", m.passwordInput.View()))
	content.WriteString("\n\n")

	// Client ID
	content.WriteString(m.renderField(MQTTFieldClientID, "Client ID:", m.clientIDInput.View()))
	content.WriteString("\n\n")

	// Topic prefix
	content.WriteString(m.renderField(MQTTFieldTopicPrefix, "Topic Prefix:", m.topicPrefixInput.View()))
	content.WriteString("\n\n")

	// TLS settings
	content.WriteString(m.renderField(MQTTFieldTLS, "TLS:", m.tlsDropdown.View()))

	// Error display
	if m.err != nil {
		content.WriteString("\n\n")
		content.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
	}

	return content.String()
}

func (m MQTTEditModel) renderField(field MQTTEditField, label, input string) string {
	var selector, labelStr string

	if m.cursor == field {
		selector = m.styles.Selector.Render("▶ ")
		labelStr = m.styles.LabelFocus.Render(label)
	} else {
		selector = "  "
		labelStr = m.styles.Label.Render(label)
	}

	return selector + labelStr + " " + input
}

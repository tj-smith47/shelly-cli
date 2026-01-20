// Package webhooks provides TUI components for managing device webhooks.
package webhooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// EditField represents a field in the webhook edit form.
type EditField int

// Edit field constants.
const (
	EditFieldName EditField = iota
	EditFieldEvent
	EditFieldURLs
	EditFieldEnable
	EditFieldCount
)

// EditSaveResultMsg is an alias for the shared save result message.
type EditSaveResultMsg = messages.SaveResultMsg

// EditOpenedMsg is an alias for the shared edit opened message.
type EditOpenedMsg = messages.EditOpenedMsg

// EditClosedMsg is an alias for the shared edit closed message.
type EditClosedMsg = messages.EditClosedMsg

// EditModel represents the webhook edit modal.
type EditModel struct {
	ctx       context.Context
	svc       *shelly.Service
	device    string
	webhookID int
	visible   bool
	cursor    EditField
	saving    bool
	err       error
	width     int
	height    int
	styles    editmodal.Styles

	// Form inputs
	nameInput   form.TextInput
	eventInput  form.TextInput
	urlsInput   form.TextArea
	enableInput form.Toggle
}

// NewEditModel creates a new webhook edit modal.
func NewEditModel(ctx context.Context, svc *shelly.Service) EditModel {
	nameInput := form.NewTextInput(
		form.WithPlaceholder("Webhook Name"),
		form.WithCharLimit(64),
		form.WithWidth(40),
		form.WithHelp("Optional descriptive name"),
	)

	eventInput := form.NewTextInput(
		form.WithPlaceholder("switch.on"),
		form.WithCharLimit(64),
		form.WithWidth(40),
		form.WithHelp("Event type (e.g., switch.on, input.toggle_on)"),
	)

	urlsInput := form.NewTextArea(
		form.WithTextAreaPlaceholder("http://example.com/webhook"),
		form.WithTextAreaCharLimit(1024),
		form.WithTextAreaDimensions(40, 4),
		form.WithTextAreaHelp("One URL per line"),
	)

	enableInput := form.NewToggle(
		form.WithToggleLabel("Enable"),
		form.WithToggleValue(true),
		form.WithToggleOnLabel("Enabled"),
		form.WithToggleOffLabel("Disabled"),
	)

	return EditModel{
		ctx:         ctx,
		svc:         svc,
		styles:      editmodal.DefaultStyles().WithLabelWidth(8),
		nameInput:   nameInput,
		eventInput:  eventInput,
		urlsInput:   urlsInput,
		enableInput: enableInput,
	}
}

// Show displays the edit modal for an existing webhook.
func (m EditModel) Show(device string, webhook *Webhook) EditModel {
	m.device = device
	m.webhookID = webhook.ID
	m.visible = true
	m.cursor = EditFieldName
	m.saving = false
	m.err = nil

	// Set input values from webhook
	m.nameInput = m.nameInput.SetValue(webhook.Name)
	m.eventInput = m.eventInput.SetValue(webhook.Event)
	m.urlsInput = m.urlsInput.SetValue(strings.Join(webhook.URLs, "\n"))
	m.enableInput = m.enableInput.SetValue(webhook.Enable)

	// Focus first input
	m.nameInput, _ = m.nameInput.Focus()
	m.eventInput = m.eventInput.Blur()
	m.urlsInput = m.urlsInput.Blur()
	m.enableInput = m.enableInput.Blur()

	return m
}

// ShowCreate displays the edit modal for creating a new webhook.
func (m EditModel) ShowCreate(device string) EditModel {
	m.device = device
	m.webhookID = 0 // 0 indicates new webhook
	m.visible = true
	m.cursor = EditFieldName
	m.saving = false
	m.err = nil

	// Set default values for new webhook
	m.nameInput = m.nameInput.SetValue("")
	m.eventInput = m.eventInput.SetValue("")
	m.urlsInput = m.urlsInput.SetValue("")
	m.enableInput = m.enableInput.SetValue(true)

	// Focus first input
	m.nameInput, _ = m.nameInput.Focus()
	m.eventInput = m.eventInput.Blur()
	m.urlsInput = m.urlsInput.Blur()
	m.enableInput = m.enableInput.Blur()

	return m
}

// IsCreating returns true if creating a new webhook (vs editing).
func (m EditModel) IsCreating() bool {
	return m.webhookID == 0
}

// Hide hides the edit modal.
func (m EditModel) Hide() EditModel {
	m.visible = false
	m.nameInput = m.nameInput.Blur()
	m.eventInput = m.eventInput.Blur()
	m.urlsInput = m.urlsInput.Blur()
	m.enableInput = m.enableInput.Blur()
	return m
}

// IsVisible returns whether the modal is visible.
func (m EditModel) IsVisible() bool {
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
		// Success - close modal
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	// Action messages from context system
	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleNavigation(msg messages.NavigationMsg) (EditModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		return m.prevField(), nil
	case messages.NavDown:
		return m.nextField(), nil
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this form
	}
	return m, nil
}

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	// Modal-specific keys not covered by action messages
	switch msg.String() {
	case keyconst.KeyEsc, "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case keyconst.KeyCtrlS:
		return m.save()

	case keyconst.KeyTab:
		return m.nextField(), nil

	case keyconst.KeyShiftTab:
		return m.prevField(), nil
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) updateFocusedInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.cursor {
	case EditFieldName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case EditFieldEvent:
		m.eventInput, cmd = m.eventInput.Update(msg)
	case EditFieldURLs:
		m.urlsInput, cmd = m.urlsInput.Update(msg)
	case EditFieldEnable:
		m.enableInput, cmd = m.enableInput.Update(msg)
	case EditFieldCount:
		// No-op
	}

	return m, cmd
}

func (m EditModel) nextField() EditModel {
	m = m.blurCurrentField()
	if m.cursor < EditFieldCount-1 {
		m.cursor++
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) prevField() EditModel {
	m = m.blurCurrentField()
	if m.cursor > 0 {
		m.cursor--
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) blurCurrentField() EditModel {
	switch m.cursor {
	case EditFieldName:
		m.nameInput = m.nameInput.Blur()
	case EditFieldEvent:
		m.eventInput = m.eventInput.Blur()
	case EditFieldURLs:
		m.urlsInput = m.urlsInput.Blur()
	case EditFieldEnable:
		m.enableInput = m.enableInput.Blur()
	case EditFieldCount:
		// No-op
	}
	return m
}

func (m EditModel) focusCurrentField() EditModel {
	switch m.cursor {
	case EditFieldName:
		m.nameInput, _ = m.nameInput.Focus()
	case EditFieldEvent:
		m.eventInput, _ = m.eventInput.Focus()
	case EditFieldURLs:
		m.urlsInput, _ = m.urlsInput.Focus()
	case EditFieldEnable:
		m.enableInput = m.enableInput.Focus()
	case EditFieldCount:
		// No-op
	}
	return m
}

func (m EditModel) save() (EditModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	// Validate event
	event := strings.TrimSpace(m.eventInput.Value())
	if event == "" {
		m.err = fmt.Errorf("event is required")
		return m, nil
	}

	// Parse URLs
	urlsStr := strings.TrimSpace(m.urlsInput.Value())
	if urlsStr == "" {
		m.err = fmt.Errorf("at least one URL is required")
		return m, nil
	}

	var urls []string
	for _, line := range strings.Split(urlsStr, "\n") {
		url := strings.TrimSpace(line)
		if url != "" {
			urls = append(urls, url)
		}
	}

	if len(urls) == 0 {
		m.err = fmt.Errorf("at least one URL is required")
		return m, nil
	}

	m.saving = true
	m.err = nil

	return m, m.createSaveCmd(event, urls)
}

func (m EditModel) createSaveCmd(event string, urls []string) tea.Cmd {
	name := strings.TrimSpace(m.nameInput.Value())
	enable := m.enableInput.Value()

	// Check if we're creating a new webhook or updating existing
	if m.webhookID == 0 {
		return m.createNewWebhookCmd(event, urls, name, enable)
	}
	return m.updateWebhookCmd(event, urls, name, enable)
}

func (m EditModel) createNewWebhookCmd(event string, urls []string, name string, enable bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		webhookID, err := m.svc.CreateWebhook(ctx, m.device, shelly.CreateWebhookParams{
			Event:  event,
			URLs:   urls,
			Name:   name,
			Enable: enable,
		})
		if err != nil {
			return messages.NewSaveError(0, err)
		}
		return messages.NewSaveResult(webhookID)
	}
}

func (m EditModel) updateWebhookCmd(event string, urls []string, name string, enable bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.UpdateWebhook(ctx, m.device, m.webhookID, shelly.UpdateWebhookParams{
			Event:  event,
			URLs:   urls,
			Name:   name,
			Enable: &enable,
		})
		if err != nil {
			return messages.NewSaveError(m.webhookID, err)
		}
		return messages.NewSaveResult(m.webhookID)
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	// Build footer
	footer := "Tab: Next | Ctrl+S: Save | Esc: Cancel"
	if m.saving {
		footer = "Saving..."
	}

	// Use appropriate title based on create/edit mode
	title := "Edit Webhook"
	if m.IsCreating() {
		title = "New Webhook"
	}

	// Use common modal helper
	r := rendering.NewModal(m.width, m.height, title, footer)

	// Build content
	return r.SetContent(m.renderFormFields()).Render()
}

func (m EditModel) renderFormFields() string {
	var content strings.Builder

	// Name field
	content.WriteString(m.renderField(EditFieldName, "Name:", m.nameInput.View()))
	content.WriteString("\n\n")

	// Event field
	content.WriteString(m.renderField(EditFieldEvent, "Event:", m.eventInput.View()))
	content.WriteString("\n\n")

	// URLs field
	content.WriteString(m.renderField(EditFieldURLs, "URLs:", m.urlsInput.View()))
	content.WriteString("\n\n")

	// Enable field
	content.WriteString(m.renderField(EditFieldEnable, "Status:", m.enableInput.View()))
	content.WriteString("\n")

	// Error display
	if m.err != nil {
		content.WriteString("\n\n")
		msg, _ := tuierrors.FormatError(m.err)
		content.WriteString(m.styles.Error.Render(msg))
	}

	return content.String()
}

func (m EditModel) renderField(field EditField, label, input string) string {
	var selector, labelStr string

	if m.cursor == field {
		selector = m.styles.Selector.Render("> ")
		labelStr = m.styles.LabelFocus.Render(label)
	} else {
		selector = "  "
		labelStr = m.styles.Label.Render(label)
	}

	return selector + labelStr + " " + input
}

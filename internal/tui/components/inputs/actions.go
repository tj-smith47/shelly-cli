// Package inputs provides TUI components for managing device input settings.
package inputs

import (
	"context"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Key constant for reuse.
const keyEnter = "enter"

// Input event types available for actions.
var inputEventTypes = []string{
	"single_push",
	"double_push",
	"triple_push",
	"long_push",
	"btn_down",
	"btn_up",
}

// ActionSavedMsg signals that an action was saved.
type ActionSavedMsg struct {
	InputID   int
	EventType string
	Err       error
}

// ActionDeletedMsg signals that an action was deleted.
type ActionDeletedMsg struct {
	InputID   int
	EventType string
	Err       error
}

// ActionModalOpenedMsg signals that the action modal was opened.
type ActionModalOpenedMsg struct{}

// ActionModalClosedMsg signals that the action modal was closed.
type ActionModalClosedMsg struct {
	Changed bool
}

// ActionField represents a field in the action modal.
type ActionField int

// Action field constants.
const (
	ActionFieldEvent ActionField = iota
	ActionFieldURL
	ActionFieldCount
)

// ActionModalStyles holds styles for the action modal.
type ActionModalStyles struct {
	Title       lipgloss.Style
	Label       lipgloss.Style
	LabelFocus  lipgloss.Style
	Selector    lipgloss.Style
	Info        lipgloss.Style
	Error       lipgloss.Style
	Success     lipgloss.Style
	Muted       lipgloss.Style
	ActionItem  lipgloss.Style
	ActionEvent lipgloss.Style
}

// DefaultActionModalStyles returns the default styles for the action modal.
func DefaultActionModalStyles() ActionModalStyles {
	colors := theme.GetSemanticColors()
	return ActionModalStyles{
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text),
		LabelFocus: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Selector: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Info: lipgloss.NewStyle().
			Foreground(colors.Info),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Success: lipgloss.NewStyle().
			Foreground(colors.Online),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		ActionItem: lipgloss.NewStyle().
			Foreground(colors.Text),
		ActionEvent: lipgloss.NewStyle().
			Foreground(colors.Warning),
	}
}

// ActionModal represents the modal for configuring input actions.
type ActionModal struct {
	ctx     context.Context
	svc     *shelly.Service
	device  string
	inputID int
	visible bool
	cursor  ActionField
	saving  bool
	err     error
	width   int
	height  int
	styles  ActionModalStyles
	changed bool // Whether any changes were made

	// Existing actions for this input
	actions []InputAction

	// Form inputs
	eventDropdown form.Select
	urlInput      form.TextInput
}

// NewActionModal creates a new action modal.
func NewActionModal(ctx context.Context, svc *shelly.Service) ActionModal {
	eventDropdown := form.NewSelect(
		form.WithSelectOptions(inputEventTypes),
		form.WithSelectHelp("Select event type to configure"),
	)

	urlInput := form.NewTextInput(
		form.WithPlaceholder("http://example.com/webhook"),
		form.WithCharLimit(256),
		form.WithWidth(40),
		form.WithHelp("URL to call when this event occurs (empty to clear)"),
	)

	return ActionModal{
		ctx:           ctx,
		svc:           svc,
		styles:        DefaultActionModalStyles(),
		eventDropdown: eventDropdown,
		urlInput:      urlInput,
	}
}

// Show displays the action modal for an input.
func (m ActionModal) Show(device string, inputID int, actions []InputAction) ActionModal {
	m.device = device
	m.inputID = inputID
	m.visible = true
	m.cursor = ActionFieldEvent
	m.saving = false
	m.err = nil
	m.changed = false
	m.actions = actions

	// Reset form
	m.eventDropdown = m.eventDropdown.SetSelected(0)
	m.urlInput = m.urlInput.SetValue("")

	// Focus event dropdown
	m = m.blurAllInputs()
	m.eventDropdown = m.eventDropdown.Focus()

	// If there's an existing action for the first event, populate URL
	m = m.populateURLForSelectedEvent()

	return m
}

// Hide hides the modal.
func (m ActionModal) Hide() ActionModal {
	m.visible = false
	return m
}

// Visible returns whether the modal is visible.
func (m ActionModal) Visible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m ActionModal) SetSize(width, height int) ActionModal {
	m.width = width
	m.height = height
	return m
}

// Init returns the initial command.
func (m ActionModal) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m ActionModal) Update(msg tea.Msg) (ActionModal, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case ActionSavedMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.changed = true
		// Update local actions list
		m = m.updateLocalActions(msg.EventType)
		return m, nil

	case ActionDeletedMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.changed = true
		// Remove from local actions list
		m = m.removeLocalAction(msg.EventType)
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m ActionModal) populateURLForSelectedEvent() ActionModal {
	selectedEvent := m.eventDropdown.SelectedValue()
	fullEvent := "input." + selectedEvent

	for _, action := range m.actions {
		if action.Event == fullEvent && len(action.URLs) > 0 {
			m.urlInput = m.urlInput.SetValue(action.URLs[0])
			return m
		}
	}
	m.urlInput = m.urlInput.SetValue("")
	return m
}

func (m ActionModal) updateLocalActions(eventType string) ActionModal {
	fullEvent := "input." + eventType
	url := strings.TrimSpace(m.urlInput.Value())

	// Find and update or add the action
	found := false
	for i := range m.actions {
		if m.actions[i].Event == fullEvent {
			m.actions[i].URLs = []string{url}
			m.actions[i].Enable = true
			found = true
			break
		}
	}
	if !found && url != "" {
		m.actions = append(m.actions, InputAction{
			Event:  fullEvent,
			URLs:   []string{url},
			Enable: true,
		})
	}
	return m
}

func (m ActionModal) removeLocalAction(eventType string) ActionModal {
	fullEvent := "input." + eventType
	newActions := make([]InputAction, 0, len(m.actions))
	for _, action := range m.actions {
		if action.Event != fullEvent {
			newActions = append(newActions, action)
		}
	}
	m.actions = newActions
	m.urlInput = m.urlInput.SetValue("")
	return m
}

func (m ActionModal) handleKey(msg tea.KeyPressMsg) (ActionModal, tea.Cmd) {
	key := msg.String()

	// Block all input while saving
	if m.saving {
		return m, nil
	}

	switch key {
	case "esc", "ctrl+[":
		m.visible = false
		return m, func() tea.Msg { return ActionModalClosedMsg{Changed: m.changed} }
	case "tab", "down":
		return m.nextField(), nil
	case "shift+tab", "up":
		return m.prevField(), nil
	case keyEnter:
		return m.handleEnter()
	case "ctrl+s":
		return m.save()
	case "d", "ctrl+d":
		return m.deleteAction()
	case " ":
		return m.handleSpace()
	}

	// Forward to focused input (Select handles its own navigation)
	return m.updateFocusedInput(msg)
}

func (m ActionModal) handleEnter() (ActionModal, tea.Cmd) {
	// Handle dropdown collapse or save
	if m.cursor == ActionFieldEvent && m.eventDropdown.IsExpanded() {
		m.eventDropdown = m.eventDropdown.Collapse()
		m = m.populateURLForSelectedEvent()
		return m, nil
	}
	return m.save()
}

func (m ActionModal) handleSpace() (ActionModal, tea.Cmd) {
	// Space expands/collapses dropdown when on event field
	if m.cursor == ActionFieldEvent {
		if m.eventDropdown.IsExpanded() {
			m.eventDropdown = m.eventDropdown.Collapse()
			m = m.populateURLForSelectedEvent()
		} else {
			m.eventDropdown = m.eventDropdown.Expand()
		}
	}
	return m, nil
}

func (m ActionModal) updateFocusedInput(msg tea.Msg) (ActionModal, tea.Cmd) {
	var cmd tea.Cmd

	switch m.cursor {
	case ActionFieldEvent:
		oldSelected := m.eventDropdown.SelectedValue()
		m.eventDropdown, cmd = m.eventDropdown.Update(msg)
		// Refresh URL when selection changes
		if m.eventDropdown.SelectedValue() != oldSelected {
			m = m.populateURLForSelectedEvent()
		}
	case ActionFieldURL:
		m.urlInput, cmd = m.urlInput.Update(msg)
	case ActionFieldCount:
		// Sentinel, no input
	}

	return m, cmd
}

func (m ActionModal) nextField() ActionModal {
	m = m.blurCurrentField()
	m.cursor++
	if m.cursor >= ActionFieldCount {
		m.cursor = 0
	}
	m = m.focusCurrentField()
	return m
}

func (m ActionModal) prevField() ActionModal {
	m = m.blurCurrentField()
	m.cursor--
	if m.cursor < 0 {
		m.cursor = ActionFieldCount - 1
	}
	m = m.focusCurrentField()
	return m
}

func (m ActionModal) blurCurrentField() ActionModal {
	switch m.cursor {
	case ActionFieldEvent:
		m.eventDropdown = m.eventDropdown.Blur()
	case ActionFieldURL:
		m.urlInput = m.urlInput.Blur()
	case ActionFieldCount:
		// Sentinel
	}
	return m
}

func (m ActionModal) focusCurrentField() ActionModal {
	switch m.cursor {
	case ActionFieldEvent:
		m.eventDropdown = m.eventDropdown.Focus()
	case ActionFieldURL:
		m.urlInput, _ = m.urlInput.Focus()
	case ActionFieldCount:
		// Sentinel
	}
	return m
}

func (m ActionModal) blurAllInputs() ActionModal {
	m.eventDropdown = m.eventDropdown.Blur()
	m.urlInput = m.urlInput.Blur()
	return m
}

func (m ActionModal) save() (ActionModal, tea.Cmd) {
	m.err = nil
	url := strings.TrimSpace(m.urlInput.Value())
	eventType := m.eventDropdown.SelectedValue()

	// If URL is empty, treat as delete
	if url == "" {
		return m.deleteAction()
	}

	m.saving = true
	return m, m.createSaveCmd(eventType, url)
}

func (m ActionModal) createSaveCmd(eventType, url string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		fullEvent := "input." + eventType

		// Check if webhook already exists for this event/input combination
		hooks, err := m.svc.ListWebhooks(ctx, m.device)
		if err != nil {
			return ActionSavedMsg{InputID: m.inputID, EventType: eventType, Err: err}
		}

		var existingID *int
		for _, h := range hooks {
			if h.Event == fullEvent && h.Cid == m.inputID {
				id := h.ID
				existingID = &id
				break
			}
		}

		if existingID != nil {
			// Update existing webhook
			enable := true
			err = m.svc.UpdateWebhook(ctx, m.device, *existingID, shelly.UpdateWebhookParams{
				Event:  fullEvent,
				URLs:   []string{url},
				Enable: &enable,
			})
		} else {
			// Create new webhook
			_, err = m.svc.CreateWebhook(ctx, m.device, shelly.CreateWebhookParams{
				Event:  fullEvent,
				URLs:   []string{url},
				Enable: true,
				Cid:    m.inputID,
			})
		}

		return ActionSavedMsg{InputID: m.inputID, EventType: eventType, Err: err}
	}
}

func (m ActionModal) deleteAction() (ActionModal, tea.Cmd) {
	m.err = nil
	eventType := m.eventDropdown.SelectedValue()
	fullEvent := "input." + eventType

	// Find webhook ID for this event
	var webhookID *int
	for _, action := range m.actions {
		if action.Event == fullEvent {
			id := action.WebhookID
			webhookID = &id
			break
		}
	}

	if webhookID == nil {
		// No action to delete
		return m, nil
	}

	m.saving = true
	return m, m.createDeleteCmd(eventType, *webhookID)
}

func (m ActionModal) createDeleteCmd(eventType string, webhookID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.DeleteWebhook(ctx, m.device, webhookID)
		return ActionDeletedMsg{InputID: m.inputID, EventType: eventType, Err: err}
	}
}

// View renders the action modal.
func (m ActionModal) View() string {
	if !m.visible {
		return ""
	}

	// Build footer
	footer := "Tab: Next | Enter: Save | d: Delete | Esc: Close"
	if m.saving {
		footer = "Saving..."
	}

	// Use common modal helper
	r := rendering.NewModal(m.width, m.height, "Configure Input Actions", footer)

	var content strings.Builder

	// Input info
	content.WriteString("  ")
	content.WriteString(m.styles.Info.Render("Input: "))
	content.WriteString(m.styles.Selector.Render(strconv.Itoa(m.inputID)))
	content.WriteString("\n\n")

	// Existing actions summary
	if len(m.actions) > 0 {
		content.WriteString("  ")
		content.WriteString(m.styles.Muted.Render("Configured actions:"))
		content.WriteString("\n")
		for _, action := range m.actions {
			eventName := strings.TrimPrefix(action.Event, "input.")
			content.WriteString("    ")
			content.WriteString(m.styles.ActionEvent.Render(eventName))
			if len(action.URLs) > 0 {
				content.WriteString(" → ")
				content.WriteString(m.styles.Muted.Render(truncateURL(action.URLs[0], 30)))
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Form fields
	content.WriteString(m.renderFormFields())

	// Error message
	if m.err != nil {
		content.WriteString("\n")
		content.WriteString(m.styles.Error.Render(m.err.Error()))
	}

	return r.SetContent(content.String()).Render()
}

func (m ActionModal) renderFormFields() string {
	var content strings.Builder

	// Event type
	content.WriteString(m.renderField(ActionFieldEvent, "Event Type:", m.eventDropdown.View()))
	content.WriteString("\n\n")

	// URL
	content.WriteString(m.renderField(ActionFieldURL, "Webhook URL:", m.urlInput.View()))
	content.WriteString("\n\n")

	// Help text
	content.WriteString("  ")
	content.WriteString(m.styles.Muted.Render("(Leave URL empty to delete action)"))

	return content.String()
}

func (m ActionModal) renderField(field ActionField, label, value string) string {
	var selector, labelStr string
	if m.cursor == field {
		selector = m.styles.Selector.Render("▶ ")
		labelStr = m.styles.LabelFocus.Render(label)
	} else {
		selector = "  "
		labelStr = m.styles.Label.Render(label)
	}
	return selector + labelStr + " " + value
}

func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}

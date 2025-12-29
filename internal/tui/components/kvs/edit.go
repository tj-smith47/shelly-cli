// Package kvs provides TUI components for browsing device key-value store.
package kvs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	shellykvs "github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
)

// EditField represents a field in the KVS edit form.
type EditField int

// Edit field constants.
const (
	EditFieldKey EditField = iota
	EditFieldValue
	EditFieldCount
)

// EditSaveResultMsg signals a save operation completed.
type EditSaveResultMsg struct {
	Key string
	Err error
}

// EditOpenedMsg signals the edit modal was opened.
type EditOpenedMsg struct{}

// EditClosedMsg signals the edit modal was closed.
type EditClosedMsg struct {
	Saved bool
}

// EditModel represents the KVS edit modal.
type EditModel struct {
	ctx     context.Context
	svc     *shellykvs.Service
	device  string
	visible bool
	isNew   bool // true if creating new entry, false if editing existing
	cursor  EditField
	saving  bool
	err     error
	width   int
	height  int
	styles  EditStyles

	// Original item for comparison (nil for new entries)
	original *Item

	// Form inputs
	keyInput   form.TextInput
	valueInput form.TextArea
}

// EditStyles holds styles for the KVS edit modal.
type EditStyles struct {
	Modal      lipgloss.Style
	Title      lipgloss.Style
	Label      lipgloss.Style
	LabelFocus lipgloss.Style
	Error      lipgloss.Style
	Help       lipgloss.Style
	Selector   lipgloss.Style
	Warning    lipgloss.Style
}

// DefaultEditStyles returns the default edit modal styles.
func DefaultEditStyles() EditStyles {
	colors := theme.GetSemanticColors()
	return EditStyles{
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
			Width(8),
		LabelFocus: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Width(8),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Selector: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
	}
}

// NewEditModel creates a new KVS edit modal.
func NewEditModel(ctx context.Context, svc *shellykvs.Service) EditModel {
	keyInput := form.NewTextInput(
		form.WithPlaceholder("my_key"),
		form.WithCharLimit(256),
		form.WithWidth(40),
		form.WithHelp("Key must be alphanumeric with underscores"),
	)

	valueInput := form.NewTextArea(
		form.WithTextAreaPlaceholder("Enter value (JSON or plain text)"),
		form.WithTextAreaCharLimit(4096),
		form.WithTextAreaDimensions(40, 6),
		form.WithTextAreaHelp("JSON will be parsed, otherwise stored as string"),
	)

	return EditModel{
		ctx:        ctx,
		svc:        svc,
		styles:     DefaultEditStyles(),
		keyInput:   keyInput,
		valueInput: valueInput,
	}
}

// ShowNew displays the edit modal for creating a new entry.
func (m EditModel) ShowNew(device string) EditModel {
	m.device = device
	m.visible = true
	m.isNew = true
	m.cursor = EditFieldKey
	m.saving = false
	m.err = nil
	m.original = nil

	// Clear inputs
	m.keyInput = m.keyInput.SetValue("")
	m.valueInput = m.valueInput.Reset()

	// Focus key input
	m.keyInput, _ = m.keyInput.Focus()
	m.valueInput = m.valueInput.Blur()

	return m
}

// ShowEdit displays the edit modal for editing an existing entry.
func (m EditModel) ShowEdit(device string, item *Item) EditModel {
	m.device = device
	m.visible = true
	m.isNew = false
	m.cursor = EditFieldValue // Skip to value since key is read-only when editing
	m.saving = false
	m.err = nil
	m.original = item

	// Set key (will be read-only since isNew is false)
	m.keyInput = m.keyInput.SetValue(item.Key)

	// Format value as JSON if it's a complex type
	valueStr := formatValueForEdit(item.Value)
	m.valueInput = m.valueInput.SetValue(valueStr)

	// Focus value input
	m.keyInput = m.keyInput.Blur()
	m.valueInput, _ = m.valueInput.Focus()

	return m
}

// formatValueForEdit formats a value for editing in the textarea.
func formatValueForEdit(value any) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%v", v)
	case map[string]any, []any:
		jsonBytes, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Hide hides the edit modal.
func (m EditModel) Hide() EditModel {
	m.visible = false
	m.keyInput = m.keyInput.Blur()
	m.valueInput = m.valueInput.Blur()
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

	switch msg := msg.(type) {
	case EditSaveResultMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - close modal
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case "ctrl+s":
		return m.save()

	case "tab":
		return m.nextField(), nil

	case "shift+tab":
		return m.prevField(), nil
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) updateFocusedInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.cursor {
	case EditFieldKey:
		// Key is only editable when creating new entry
		if m.isNew {
			m.keyInput, cmd = m.keyInput.Update(msg)
		}
	case EditFieldValue:
		m.valueInput, cmd = m.valueInput.Update(msg)
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
	// Skip key field when editing (it's read-only)
	if !m.isNew && m.cursor == EditFieldKey {
		m.cursor = EditFieldValue
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) prevField() EditModel {
	m = m.blurCurrentField()
	if m.cursor > 0 {
		m.cursor--
	}
	// Skip key field when editing (it's read-only)
	if !m.isNew && m.cursor == EditFieldKey {
		m.cursor = EditFieldValue
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) blurCurrentField() EditModel {
	switch m.cursor {
	case EditFieldKey:
		m.keyInput = m.keyInput.Blur()
	case EditFieldValue:
		m.valueInput = m.valueInput.Blur()
	case EditFieldCount:
		// No-op
	}
	return m
}

func (m EditModel) focusCurrentField() EditModel {
	switch m.cursor {
	case EditFieldKey:
		m.keyInput, _ = m.keyInput.Focus()
	case EditFieldValue:
		m.valueInput, _ = m.valueInput.Focus()
	case EditFieldCount:
		// No-op
	}
	return m
}

func (m EditModel) save() (EditModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	// Validate key
	key := strings.TrimSpace(m.keyInput.Value())
	if key == "" {
		m.err = fmt.Errorf("key is required")
		return m, nil
	}

	// Validate key format (alphanumeric and underscores only)
	for _, c := range key {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_') {
			m.err = fmt.Errorf("key must contain only alphanumeric characters and underscores")
			return m, nil
		}
	}

	// Parse value
	valueStr := strings.TrimSpace(m.valueInput.Value())
	if valueStr == "" {
		m.err = fmt.Errorf("value is required")
		return m, nil
	}

	m.saving = true
	m.err = nil

	return m, m.createSaveCmd(key, valueStr)
}

func (m EditModel) createSaveCmd(key, valueStr string) tea.Cmd {
	// Parse value - try JSON first, fall back to string
	value := shellykvs.ParseValue(valueStr)

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.Set(ctx, m.device, key, value)
		return EditSaveResultMsg{Key: key, Err: err}
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title
	if m.isNew {
		content.WriteString(m.styles.Title.Render("New KVS Entry"))
	} else {
		content.WriteString(m.styles.Title.Render("Edit KVS Entry"))
	}
	content.WriteString("\n\n")

	// Form fields
	content.WriteString(m.renderFormFields())

	// Render modal box
	modalContent := content.String()
	modalWidth := min(60, m.width-4)
	modal := m.styles.Modal.Width(modalWidth).Render(modalContent)

	// Center the modal
	return m.centerModal(modal)
}

func (m EditModel) renderFormFields() string {
	var content strings.Builder

	// Key field
	keyLabel := "Key:"
	if !m.isNew {
		keyLabel = "Key: (read-only)"
	}
	content.WriteString(m.renderField(EditFieldKey, keyLabel, m.keyInput.View()))
	content.WriteString("\n\n")

	// Value field
	content.WriteString(m.renderField(EditFieldValue, "Value:", m.valueInput.View()))
	content.WriteString("\n")

	// Error display
	if m.err != nil {
		content.WriteString("\n")
		content.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
		content.WriteString("\n")
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(m.renderHelpText())

	return content.String()
}

func (m EditModel) renderField(field EditField, label, input string) string {
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

func (m EditModel) renderHelpText() string {
	if m.saving {
		return m.styles.Help.Render("Saving...")
	}
	return m.styles.Help.Render("Tab: Next | Ctrl+S: Save | Esc: Cancel")
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

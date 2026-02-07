// Package kvs provides TUI components for browsing device key-value store.
package kvs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/output/jsonfmt"
	shellykvs "github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

// EditField represents a field in the KVS edit form.
type EditField int

// Edit field constants.
const (
	EditFieldKey EditField = iota
	EditFieldValue
	EditFieldCount
)

// EditSaveResultMsg is an alias for the shared save result message.
type EditSaveResultMsg = messages.SaveResultMsg

// EditOpenedMsg is an alias for the shared edit opened message.
type EditOpenedMsg = messages.EditOpenedMsg

// EditClosedMsg is an alias for the shared edit closed message.
type EditClosedMsg = messages.EditClosedMsg

// EditModel represents the KVS edit modal.
type EditModel struct {
	editmodal.Base

	svc   *shellykvs.Service
	isNew bool // true if creating new entry, false if editing existing

	// Original item for comparison (nil for new entries)
	original *Item

	// Form inputs
	keyInput   form.TextInput
	valueInput form.TextArea
}

// NewEditModel creates a new KVS edit modal.
func NewEditModel(ctx context.Context, svc *shellykvs.Service) EditModel {
	keyInput := form.NewTextInput(
		form.WithPlaceholder("my_key"),
		form.WithCharLimit(42), // Shelly KVS max key length
		form.WithWidth(40),
		form.WithHelp("Max 42 characters"),
	)

	valueInput := form.NewTextArea(
		form.WithTextAreaPlaceholder("Enter value (JSON or plain text)"),
		form.WithTextAreaCharLimit(253), // Shelly KVS max value length
		form.WithTextAreaDimensions(40, 6),
		form.WithTextAreaHelp("Max 253 chars; JSON parsed, otherwise stored as string"),
	)

	return EditModel{
		Base: editmodal.Base{
			Ctx:    ctx,
			Styles: editmodal.DefaultStyles().WithLabelWidth(12),
		},
		svc:        svc,
		keyInput:   keyInput,
		valueInput: valueInput,
	}
}

// ShowNew displays the edit modal for creating a new entry.
func (m EditModel) ShowNew(device string) EditModel {
	m.Show(device, int(EditFieldCount))
	m.isNew = true
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
	m.Show(device, int(EditFieldCount))
	m.SetCursor(int(EditFieldValue)) // Skip to value since key is read-only when editing
	m.isNew = false
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
		// Try to parse as JSON and pretty-print
		return jsonfmt.PrettyString(v)
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
	m.Base.Hide()
	m.keyInput = m.keyInput.Blur()
	m.valueInput = m.valueInput.Blur()
	return m
}

// Visible returns whether the modal is visible.
func (m EditModel) Visible() bool {
	return m.Base.Visible()
}

// SetSize sets the modal dimensions.
func (m EditModel) SetSize(width, height int) EditModel {
	m.Base.SetSize(width, height)
	inputWidth := m.InputWidth()
	m.keyInput = m.keyInput.SetWidth(inputWidth)
	// Value textarea gets more height for JSON content
	valueHeight := 10
	if height > 40 {
		valueHeight = 15
	}
	m.valueInput = m.valueInput.SetDimensions(inputWidth, valueHeight)
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
		_, cmd := m.HandleSaveResult(msg)
		if cmd != nil {
			// Success - modal already hidden by HandleSaveResult
			return m, cmd
		}
		// Error - modal stays open, Err is set by HandleSaveResult
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	if m.Saving {
		return m, nil
	}

	switch msg.String() {
	case keyconst.KeyEsc, "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case keyconst.KeyCtrlS:
		return m.save()

	case keyconst.KeyCtrlF:
		// Format JSON in value field
		if EditField(m.Cursor) == EditFieldValue {
			return m.formatValueField(), nil
		}
		return m, nil

	case keyconst.KeyTab:
		return m.nextField(), nil

	case keyconst.KeyShiftTab:
		return m.prevField(), nil
	}

	// Forward to focused input (including Enter for text area)
	return m.updateFocusedInput(msg)
}

func (m EditModel) updateFocusedInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch EditField(m.Cursor) {
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
	oldCursor, newCursor := m.NextField()
	_ = oldCursor
	// Skip key field when editing (it's read-only)
	if !m.isNew && EditField(newCursor) == EditFieldKey {
		m.SetCursor(int(EditFieldValue))
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) prevField() EditModel {
	m = m.blurCurrentField()
	oldCursor, newCursor := m.PrevField()
	_ = oldCursor
	// Skip key field when editing (it's read-only)
	if !m.isNew && EditField(newCursor) == EditFieldKey {
		m.SetCursor(int(EditFieldValue))
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) formatValueField() EditModel {
	valueStr := m.valueInput.Value()
	if valueStr == "" {
		return m
	}

	// Try to parse and reformat as JSON
	formatted := jsonfmt.PrettyString(valueStr)
	if formatted != valueStr {
		m.valueInput = m.valueInput.SetValue(formatted)
	}
	return m
}

func (m EditModel) blurCurrentField() EditModel {
	switch EditField(m.Cursor) {
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
	switch EditField(m.Cursor) {
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
	if m.Saving {
		return m, nil
	}

	// Validate key
	key := strings.TrimSpace(m.keyInput.Value())
	if key == "" {
		m.Err = fmt.Errorf("key is required")
		return m, nil
	}

	// Validate key length (Shelly KVS max key length is 42 characters)
	if len(key) > 42 {
		m.Err = fmt.Errorf("key must be 42 characters or less")
		return m, nil
	}

	// Parse value
	valueStr := strings.TrimSpace(m.valueInput.Value())
	if valueStr == "" {
		m.Err = fmt.Errorf("value is required")
		return m, nil
	}

	m.StartSave()

	// Parse value - try JSON first, fall back to string
	value := shellykvs.ParseValue(valueStr)
	svc := m.svc
	device := m.Device

	cmd := m.SaveCmdWithID(key, func(ctx context.Context) error {
		return svc.Set(ctx, device, key, value)
	})

	return m, cmd
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.Visible() {
		return ""
	}

	// Build title
	title := "Edit KVS Entry"
	if m.isNew {
		title = "New KVS Entry"
	}

	// Build footer with keybindings
	footer := m.RenderSavingFooter("Tab: Next | Ctrl+F: Format | Ctrl+S: Save | Esc: Cancel")

	return m.RenderModal(title, m.renderFormFields(), footer)
}

func (m EditModel) renderFormFields() string {
	var content strings.Builder

	// Key field
	keyLabel := "Key:"
	if !m.isNew {
		keyLabel = "Key (RO):"
	}
	content.WriteString(m.renderField(EditFieldKey, keyLabel, m.keyInput.View()))
	content.WriteString("\n\n")

	// Value field
	content.WriteString(m.renderField(EditFieldValue, "Value:", m.valueInput.View()))
	content.WriteString("\n")

	// Error display
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n")
		content.WriteString(errStr)
	}

	return content.String()
}

func (m EditModel) renderField(field EditField, label, input string) string {
	var selector, labelStr string

	if EditField(m.Cursor) == field {
		selector = m.Styles.Selector.Render("▶ ")
		labelStr = m.Styles.LabelFocus.Render(label)
	} else {
		selector = "  "
		labelStr = m.Styles.Label.Render(label)
	}

	prefix := selector + labelStr + " "

	// Handle multi-line inputs by indenting subsequent lines
	lines := strings.Split(input, "\n")
	if len(lines) <= 1 {
		return prefix + input
	}

	// Calculate indent for subsequent lines (selector width + label width + space)
	indent := strings.Repeat(" ", 2+12+1) // 2 for selector, 12 for label width, 1 for space

	var result strings.Builder
	result.WriteString(prefix + lines[0])
	for i := 1; i < len(lines); i++ {
		result.WriteString("\n")
		result.WriteString(indent + lines[i])
	}
	return result.String()
}

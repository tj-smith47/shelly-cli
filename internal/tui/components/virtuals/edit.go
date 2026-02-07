// Package virtuals provides TUI components for managing virtual components.
package virtuals

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

// EditField represents a field in the virtual component edit form.
type EditField int

// Edit field constants.
const (
	EditFieldType EditField = iota
	EditFieldName
	EditFieldValue
	EditFieldCount
)

// EditSaveResultMsg is an alias for the shared save result message.
type EditSaveResultMsg = messages.SaveResultMsg

// EditOpenedMsg is an alias for the shared edit opened message.
type EditOpenedMsg = messages.EditOpenedMsg

// EditClosedMsg is an alias for the shared edit closed message.
type EditClosedMsg = messages.EditClosedMsg

// EditModel represents the virtual component edit modal.
type EditModel struct {
	editmodal.Base

	isNew bool

	// Current virtual being edited
	virtual *Virtual

	// Form inputs
	typeDropdown form.Select
	nameInput    form.TextInput
	boolToggle   form.Toggle
	numberInput  form.TextInput
	textInput    form.TextInput
	enumDropdown form.Select
}

// NewEditModel creates a new virtual component edit modal.
func NewEditModel(ctx context.Context, svc *shelly.Service) EditModel {
	typeOptions := []string{"Boolean", "Number", "Text", "Enum", "Button"}

	nameInput := form.NewTextInput(
		form.WithPlaceholder("Component Name"),
		form.WithCharLimit(64),
		form.WithWidth(30),
		form.WithHelp("Display name for the component"),
	)

	boolToggle := form.NewToggle(
		form.WithToggleLabel("Value"),
		form.WithToggleValue(false),
		form.WithToggleOnLabel("ON"),
		form.WithToggleOffLabel("OFF"),
	)

	numberInput := form.NewTextInput(
		form.WithPlaceholder("0"),
		form.WithCharLimit(20),
		form.WithWidth(20),
		form.WithHelp("Numeric value"),
	)

	textInput := form.NewTextInput(
		form.WithPlaceholder("Enter text"),
		form.WithCharLimit(256),
		form.WithWidth(30),
		form.WithHelp("Text value"),
	)

	return EditModel{
		Base: editmodal.Base{
			Ctx:    ctx,
			Svc:    svc,
			Styles: editmodal.DefaultStyles().WithLabelWidth(10),
		},
		typeDropdown: form.NewSelect(form.WithSelectOptions(typeOptions)),
		nameInput:    nameInput,
		boolToggle:   boolToggle,
		numberInput:  numberInput,
		textInput:    textInput,
		enumDropdown: form.NewSelect(),
	}
}

// ShowNew displays the edit modal for creating a new virtual component.
func (m EditModel) ShowNew(device string) EditModel {
	m.Show(device, int(EditFieldCount))
	m.isNew = true
	m.virtual = nil

	// Reset inputs
	m.typeDropdown = m.typeDropdown.SetSelected(0)
	m.nameInput = m.nameInput.SetValue("")
	m.boolToggle = m.boolToggle.SetValue(false)
	m.numberInput = m.numberInput.SetValue("0")
	m.textInput = m.textInput.SetValue("")

	// Focus type selector
	m.typeDropdown = m.typeDropdown.Focus()

	return m
}

// ShowEdit displays the edit modal for editing an existing virtual component.
func (m EditModel) ShowEdit(device string, v *Virtual) EditModel {
	m.Show(device, int(EditFieldCount))
	m.SetCursor(int(EditFieldValue)) // Skip to value since type is fixed
	m.isNew = false
	m.virtual = v

	// Set name
	m.nameInput = m.nameInput.SetValue(v.Name)

	// Set value based on type
	switch v.Type {
	case shelly.VirtualBoolean:
		if v.BoolValue != nil {
			m.boolToggle = m.boolToggle.SetValue(*v.BoolValue)
		}
		m.boolToggle = m.boolToggle.Focus()
	case shelly.VirtualNumber:
		if v.NumValue != nil {
			m.numberInput = m.numberInput.SetValue(fmt.Sprintf("%g", *v.NumValue))
		}
		m.numberInput, _ = m.numberInput.Focus()
	case shelly.VirtualText:
		if v.StrValue != nil {
			m.textInput = m.textInput.SetValue(*v.StrValue)
		}
		m.textInput, _ = m.textInput.Focus()
	case shelly.VirtualEnum:
		// Build options from the enum's options
		selectedIdx := 0
		for i, opt := range v.Options {
			if v.StrValue != nil && *v.StrValue == opt {
				selectedIdx = i
			}
		}
		m.enumDropdown = form.NewSelect(form.WithSelectOptions(v.Options))
		m.enumDropdown = m.enumDropdown.SetSelected(selectedIdx)
		m.enumDropdown = m.enumDropdown.Focus()
	case shelly.VirtualButton, shelly.VirtualGroup:
		// Buttons and groups have no editable value, focus toggle as default
		m.boolToggle = m.boolToggle.Focus()
	}

	return m
}

// Hide hides the edit modal.
func (m EditModel) Hide() EditModel {
	m.Base.Hide()
	m.typeDropdown = m.typeDropdown.Blur()
	m.nameInput = m.nameInput.Blur()
	m.boolToggle = m.boolToggle.Blur()
	m.numberInput = m.numberInput.Blur()
	m.textInput = m.textInput.Blur()
	m.enumDropdown = m.enumDropdown.Blur()
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

// Init returns the initial command.
func (m EditModel) Init() tea.Cmd {
	return nil
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
	if m.Saving {
		return m, nil
	}
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
	if m.Saving {
		return m, nil
	}

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

	if m.isNew && EditField(m.Cursor) == EditFieldType {
		m.typeDropdown, cmd = m.typeDropdown.Update(msg)
		return m, cmd
	}

	if EditField(m.Cursor) == EditFieldName {
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	}

	if EditField(m.Cursor) == EditFieldValue {
		return m.updateValueInput(msg)
	}

	return m, nil
}

func (m EditModel) updateValueInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd
	vType := m.getVirtualType()

	switch vType {
	case shelly.VirtualBoolean:
		m.boolToggle, cmd = m.boolToggle.Update(msg)
	case shelly.VirtualNumber:
		m.numberInput, cmd = m.numberInput.Update(msg)
	case shelly.VirtualText:
		m.textInput, cmd = m.textInput.Update(msg)
	case shelly.VirtualEnum:
		m.enumDropdown, cmd = m.enumDropdown.Update(msg)
	case shelly.VirtualButton, shelly.VirtualGroup:
		// No input to update for buttons/groups
	}

	return m, cmd
}

func (m EditModel) getVirtualType() shelly.VirtualComponentType {
	if !m.isNew && m.virtual != nil {
		return m.virtual.Type
	}

	// Get type from dropdown
	selected := m.typeDropdown.SelectedValue()

	switch selected {
	case "Boolean":
		return shelly.VirtualBoolean
	case "Number":
		return shelly.VirtualNumber
	case "Text":
		return shelly.VirtualText
	case "Enum":
		return shelly.VirtualEnum
	case "Button":
		return shelly.VirtualButton
	default:
		return shelly.VirtualBoolean
	}
}

func (m EditModel) nextField() EditModel {
	m = m.blurCurrentField()

	switch {
	case m.isNew && EditField(m.Cursor) < EditFieldCount-1:
		// New: Type -> Name -> Value
		m.Cursor++
	case !m.isNew && EditField(m.Cursor) != EditFieldValue:
		// Edit: Jump directly to Value (type is fixed)
		m.SetCursor(int(EditFieldValue))
	}

	m = m.focusCurrentField()
	return m
}

func (m EditModel) prevField() EditModel {
	m = m.blurCurrentField()

	if m.isNew {
		if m.Cursor > 0 {
			m.Cursor--
		}
	} else {
		// For edit, stay on value
		m.SetCursor(int(EditFieldValue))
	}

	m = m.focusCurrentField()
	return m
}

func (m EditModel) blurCurrentField() EditModel {
	switch EditField(m.Cursor) {
	case EditFieldType:
		m.typeDropdown = m.typeDropdown.Blur()
	case EditFieldName:
		m.nameInput = m.nameInput.Blur()
	case EditFieldValue:
		m.boolToggle = m.boolToggle.Blur()
		m.numberInput = m.numberInput.Blur()
		m.textInput = m.textInput.Blur()
		m.enumDropdown = m.enumDropdown.Blur()
	case EditFieldCount:
		// No-op
	}
	return m
}

func (m EditModel) focusCurrentField() EditModel {
	switch EditField(m.Cursor) {
	case EditFieldType:
		m.typeDropdown = m.typeDropdown.Focus()
	case EditFieldName:
		m.nameInput, _ = m.nameInput.Focus()
	case EditFieldValue:
		m = m.focusValueInput()
	case EditFieldCount:
		// No-op
	}
	return m
}

func (m EditModel) focusValueInput() EditModel {
	vType := m.getVirtualType()
	switch vType {
	case shelly.VirtualBoolean:
		m.boolToggle = m.boolToggle.Focus()
	case shelly.VirtualNumber:
		m.numberInput, _ = m.numberInput.Focus()
	case shelly.VirtualText:
		m.textInput, _ = m.textInput.Focus()
	case shelly.VirtualEnum:
		m.enumDropdown = m.enumDropdown.Focus()
	case shelly.VirtualButton, shelly.VirtualGroup:
		// No focusable input for buttons/groups
	}
	return m
}

func (m EditModel) save() (EditModel, tea.Cmd) {
	if m.Saving {
		return m, nil
	}

	vType := m.getVirtualType()

	// For buttons, no value to save when editing
	if !m.isNew && vType == shelly.VirtualButton {
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	}

	m.StartSave()

	if m.isNew {
		return m, m.createNewComponent(vType)
	}
	return m, m.updateExistingComponent(vType)
}

func (m EditModel) createNewComponent(vType shelly.VirtualComponentType) tea.Cmd {
	name := strings.TrimSpace(m.nameInput.Value())
	device := m.Device

	return m.SaveCmdWithID(nil, func(ctx context.Context) error {
		params := shelly.AddVirtualComponentParams{
			Type: vType,
			Name: name,
		}

		id, err := m.Svc.AddVirtualComponent(ctx, device, params)
		if err != nil {
			return err
		}

		// Now set the initial value
		switch vType {
		case shelly.VirtualBoolean:
			val := m.boolToggle.Value()
			return m.Svc.SetVirtualBoolean(ctx, device, id, val)
		case shelly.VirtualNumber:
			val, parseErr := strconv.ParseFloat(strings.TrimSpace(m.numberInput.Value()), 64)
			if parseErr != nil {
				return fmt.Errorf("invalid number: %w", parseErr)
			}
			return m.Svc.SetVirtualNumber(ctx, device, id, val)
		case shelly.VirtualText:
			val := m.textInput.Value()
			return m.Svc.SetVirtualText(ctx, device, id, val)
		case shelly.VirtualEnum:
			selectedVal := m.enumDropdown.SelectedValue()
			if selectedVal != "" {
				return m.Svc.SetVirtualEnum(ctx, device, id, selectedVal)
			}
		case shelly.VirtualButton, shelly.VirtualGroup:
			// No initial value to set for buttons/groups
		}

		return nil
	})
}

func (m EditModel) updateExistingComponent(vType shelly.VirtualComponentType) tea.Cmd {
	if m.virtual == nil {
		return func() tea.Msg {
			return messages.NewSaveError(nil, fmt.Errorf("no component to update"))
		}
	}

	id := m.virtual.ID
	key := m.virtual.Key
	device := m.Device

	return m.SaveCmdWithID(key, func(ctx context.Context) error {
		switch vType {
		case shelly.VirtualBoolean:
			val := m.boolToggle.Value()
			return m.Svc.SetVirtualBoolean(ctx, device, id, val)
		case shelly.VirtualNumber:
			val, parseErr := strconv.ParseFloat(strings.TrimSpace(m.numberInput.Value()), 64)
			if parseErr != nil {
				return fmt.Errorf("invalid number: %w", parseErr)
			}
			return m.Svc.SetVirtualNumber(ctx, device, id, val)
		case shelly.VirtualText:
			val := m.textInput.Value()
			return m.Svc.SetVirtualText(ctx, device, id, val)
		case shelly.VirtualEnum:
			selectedVal := m.enumDropdown.SelectedValue()
			if selectedVal != "" {
				return m.Svc.SetVirtualEnum(ctx, device, id, selectedVal)
			}
		case shelly.VirtualButton, shelly.VirtualGroup:
			// Buttons are triggered, not edited; groups contain other components
		}

		return nil
	})
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.Visible() {
		return ""
	}

	// Build title and footer
	title := "Edit Virtual Component"
	if m.isNew {
		title = "New Virtual Component"
	}

	footer := m.RenderSavingFooter("Tab: Next | Ctrl+S: Save | Esc: Cancel")

	return m.RenderModal(title, m.renderFormFields(), footer)
}

func (m EditModel) renderFormFields() string {
	var content strings.Builder
	vType := m.getVirtualType()

	// Type field (only for new)
	if m.isNew {
		content.WriteString(m.renderField(EditFieldType, "Type:", m.typeDropdown.View()))
		content.WriteString("\n\n")

		// Name field
		content.WriteString(m.renderField(EditFieldName, "Name:", m.nameInput.View()))
		content.WriteString("\n\n")
	} else if m.virtual != nil {
		// Show type info for existing
		typeStr := virtualTypeLabel(m.virtual.Type)
		content.WriteString(m.Styles.Label.Render("Type:") + " " + m.Styles.Info.Render(typeStr))
		content.WriteString("\n\n")

		// Show name
		name := m.virtual.Name
		if name == "" {
			name = fmt.Sprintf("#%d", m.virtual.ID)
		}
		content.WriteString(m.Styles.Label.Render("Name:") + " " + m.Styles.Info.Render(name))
		content.WriteString("\n\n")
	}

	// Value field based on type
	content.WriteString(m.renderValueField(vType))
	content.WriteString("\n")

	// Range info for numbers
	if rangeInfo := m.buildRangeInfo(vType); rangeInfo != "" {
		content.WriteString("\n" + m.Styles.Info.Render(rangeInfo))
	}

	// Error display
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n\n")
		content.WriteString(errStr)
	}

	return content.String()
}

func (m EditModel) renderValueField(vType shelly.VirtualComponentType) string {
	var valueView string

	switch vType {
	case shelly.VirtualBoolean:
		valueView = m.boolToggle.View()
	case shelly.VirtualNumber:
		valueView = m.numberInput.View()
	case shelly.VirtualText:
		valueView = m.textInput.View()
	case shelly.VirtualEnum:
		valueView = m.enumDropdown.View()
	case shelly.VirtualButton:
		valueView = m.Styles.Muted.Render("[Button - press Enter to trigger]")
	default:
		valueView = m.Styles.Muted.Render("N/A")
	}

	return m.renderField(EditFieldValue, "Value:", valueView)
}

func (m EditModel) renderField(field EditField, label, input string) string {
	var selector, labelStr string

	if EditField(m.Cursor) == field {
		selector = m.Styles.Selector.Render("> ")
		labelStr = m.Styles.LabelFocus.Render(label)
	} else {
		selector = "  "
		labelStr = m.Styles.Label.Render(label)
	}

	return selector + labelStr + " " + input
}

func (m EditModel) buildRangeInfo(vType shelly.VirtualComponentType) string {
	if vType != shelly.VirtualNumber || m.isNew || m.virtual == nil {
		return ""
	}
	if m.virtual.Min == nil && m.virtual.Max == nil {
		return ""
	}

	rangeInfo := "Range: "
	if m.virtual.Min != nil {
		rangeInfo += fmt.Sprintf("%g", *m.virtual.Min)
	} else {
		rangeInfo += "-∞"
	}
	rangeInfo += " to "
	if m.virtual.Max != nil {
		rangeInfo += fmt.Sprintf("%g", *m.virtual.Max)
	} else {
		rangeInfo += "∞"
	}
	if m.virtual.Unit != nil {
		rangeInfo += " " + *m.virtual.Unit
	}
	return rangeInfo
}

func virtualTypeLabel(t shelly.VirtualComponentType) string {
	switch t {
	case shelly.VirtualBoolean:
		return "Boolean"
	case shelly.VirtualNumber:
		return "Number"
	case shelly.VirtualText:
		return "Text"
	case shelly.VirtualEnum:
		return "Enum"
	case shelly.VirtualButton:
		return "Button"
	case shelly.VirtualGroup:
		return "Group"
	default:
		return "Unknown"
	}
}

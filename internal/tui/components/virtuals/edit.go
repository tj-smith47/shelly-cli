// Package virtuals provides TUI components for managing virtual components.
package virtuals

import (
	"context"
	"fmt"
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

// EditField represents a field in the virtual component edit form.
type EditField int

// Edit field constants.
const (
	EditFieldType EditField = iota
	EditFieldName
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

// EditModel represents the virtual component edit modal.
type EditModel struct {
	ctx     context.Context
	svc     *shelly.Service
	device  string
	isNew   bool
	visible bool
	cursor  EditField
	saving  bool
	err     error
	width   int
	height  int
	styles  EditStyles

	// Current virtual being edited
	virtual *Virtual

	// Form inputs
	typeDropdown form.Dropdown
	nameInput    form.TextInput
	boolToggle   form.Toggle
	numberInput  form.TextInput
	textInput    form.TextInput
	enumDropdown form.Dropdown
}

// EditStyles holds styles for the virtual component edit modal.
type EditStyles struct {
	Modal      lipgloss.Style
	Title      lipgloss.Style
	Label      lipgloss.Style
	LabelFocus lipgloss.Style
	Error      lipgloss.Style
	Help       lipgloss.Style
	Selector   lipgloss.Style
	Info       lipgloss.Style
	Muted      lipgloss.Style
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
			Foreground(colors.Text).
			Width(10),
		LabelFocus: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Width(10),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Selector: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Info: lipgloss.NewStyle().
			Foreground(colors.Info),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
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
		ctx:          ctx,
		svc:          svc,
		styles:       DefaultEditStyles(),
		typeDropdown: form.NewDropdown(form.WithDropdownOptions(typeOptions)),
		nameInput:    nameInput,
		boolToggle:   boolToggle,
		numberInput:  numberInput,
		textInput:    textInput,
		enumDropdown: form.NewDropdown(),
	}
}

// ShowNew displays the edit modal for creating a new virtual component.
func (m EditModel) ShowNew(device string) EditModel {
	m.device = device
	m.visible = true
	m.isNew = true
	m.cursor = EditFieldType
	m.saving = false
	m.err = nil
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
	m.device = device
	m.visible = true
	m.isNew = false
	m.cursor = EditFieldValue // Skip to value since type is fixed
	m.saving = false
	m.err = nil
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
		m.enumDropdown = form.NewDropdown(form.WithDropdownOptions(v.Options))
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
	m.visible = false
	m.typeDropdown = m.typeDropdown.Blur()
	m.nameInput = m.nameInput.Blur()
	m.boolToggle = m.boolToggle.Blur()
	m.numberInput = m.numberInput.Blur()
	m.textInput = m.textInput.Blur()
	m.enumDropdown = m.enumDropdown.Blur()
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
	case "esc", "ctrl+[":
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

	if m.isNew && m.cursor == EditFieldType {
		m.typeDropdown, cmd = m.typeDropdown.Update(msg)
		return m, cmd
	}

	if m.cursor == EditFieldName {
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	}

	if m.cursor == EditFieldValue {
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
	case m.isNew && m.cursor < EditFieldCount-1:
		// New: Type -> Name -> Value
		m.cursor++
	case !m.isNew && m.cursor != EditFieldValue:
		// Edit: Jump directly to Value (type is fixed)
		m.cursor = EditFieldValue
	}

	m = m.focusCurrentField()
	return m
}

func (m EditModel) prevField() EditModel {
	m = m.blurCurrentField()

	if m.isNew {
		if m.cursor > 0 {
			m.cursor--
		}
	} else {
		// For edit, stay on value
		m.cursor = EditFieldValue
	}

	m = m.focusCurrentField()
	return m
}

func (m EditModel) blurCurrentField() EditModel {
	switch m.cursor {
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
	switch m.cursor {
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
	if m.saving {
		return m, nil
	}

	vType := m.getVirtualType()

	// For buttons, no value to save when editing
	if !m.isNew && vType == shelly.VirtualButton {
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	}

	m.saving = true
	m.err = nil

	if m.isNew {
		return m, m.createNewComponent(vType)
	}
	return m, m.updateExistingComponent(vType)
}

func (m EditModel) createNewComponent(vType shelly.VirtualComponentType) tea.Cmd {
	name := strings.TrimSpace(m.nameInput.Value())

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		params := shelly.AddVirtualComponentParams{
			Type: vType,
			Name: name,
		}

		id, err := m.svc.AddVirtualComponent(ctx, m.device, params)
		if err != nil {
			return EditSaveResultMsg{Err: err}
		}

		// Now set the initial value
		switch vType {
		case shelly.VirtualBoolean:
			val := m.boolToggle.Value()
			err = m.svc.SetVirtualBoolean(ctx, m.device, id, val)
		case shelly.VirtualNumber:
			val, parseErr := strconv.ParseFloat(strings.TrimSpace(m.numberInput.Value()), 64)
			if parseErr != nil {
				return EditSaveResultMsg{Err: fmt.Errorf("invalid number: %w", parseErr)}
			}
			err = m.svc.SetVirtualNumber(ctx, m.device, id, val)
		case shelly.VirtualText:
			val := m.textInput.Value()
			err = m.svc.SetVirtualText(ctx, m.device, id, val)
		case shelly.VirtualEnum:
			selectedVal := m.enumDropdown.SelectedValue()
			if selectedVal != "" {
				err = m.svc.SetVirtualEnum(ctx, m.device, id, selectedVal)
			}
		case shelly.VirtualButton, shelly.VirtualGroup:
			// No initial value to set for buttons/groups
		}

		return EditSaveResultMsg{Key: fmt.Sprintf("virtual:%d", id), Err: err}
	}
}

func (m EditModel) updateExistingComponent(vType shelly.VirtualComponentType) tea.Cmd {
	if m.virtual == nil {
		return func() tea.Msg {
			return EditSaveResultMsg{Err: fmt.Errorf("no component to update")}
		}
	}

	id := m.virtual.ID

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		var err error
		switch vType {
		case shelly.VirtualBoolean:
			val := m.boolToggle.Value()
			err = m.svc.SetVirtualBoolean(ctx, m.device, id, val)
		case shelly.VirtualNumber:
			val, parseErr := strconv.ParseFloat(strings.TrimSpace(m.numberInput.Value()), 64)
			if parseErr != nil {
				return EditSaveResultMsg{Err: fmt.Errorf("invalid number: %w", parseErr)}
			}
			err = m.svc.SetVirtualNumber(ctx, m.device, id, val)
		case shelly.VirtualText:
			val := m.textInput.Value()
			err = m.svc.SetVirtualText(ctx, m.device, id, val)
		case shelly.VirtualEnum:
			selectedVal := m.enumDropdown.SelectedValue()
			if selectedVal != "" {
				err = m.svc.SetVirtualEnum(ctx, m.device, id, selectedVal)
			}
		case shelly.VirtualButton, shelly.VirtualGroup:
			// Buttons are triggered, not edited; groups contain other components
		}

		return EditSaveResultMsg{Key: m.virtual.Key, Err: err}
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	// Build title and footer
	title := "Edit Virtual Component"
	if m.isNew {
		title = "New Virtual Component"
	}

	footer := "Tab: Next | Ctrl+S: Save | Esc: Cancel"
	if m.saving {
		footer = "Saving..."
	}

	// Use common modal helper
	r := rendering.NewModal(m.width, m.height, title, footer)

	// Build content
	return r.SetContent(m.renderFormFields()).Render()
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
		content.WriteString(m.styles.Label.Render("Type:") + " " + m.styles.Info.Render(typeStr))
		content.WriteString("\n\n")

		// Show name
		name := m.virtual.Name
		if name == "" {
			name = fmt.Sprintf("#%d", m.virtual.ID)
		}
		content.WriteString(m.styles.Label.Render("Name:") + " " + m.styles.Info.Render(name))
		content.WriteString("\n\n")
	}

	// Value field based on type
	content.WriteString(m.renderValueField(vType))
	content.WriteString("\n")

	// Range info for numbers
	if rangeInfo := m.buildRangeInfo(vType); rangeInfo != "" {
		content.WriteString("\n" + m.styles.Info.Render(rangeInfo))
	}

	// Error display
	if m.err != nil {
		content.WriteString("\n\n")
		content.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
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
		valueView = m.styles.Muted.Render("[Button - press Enter to trigger]")
	default:
		valueView = m.styles.Muted.Render("N/A")
	}

	return m.renderField(EditFieldValue, "Value:", valueView)
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

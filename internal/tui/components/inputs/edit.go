// Package inputs provides TUI components for managing device input settings.
package inputs

import (
	"context"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

// Input type constants.
const (
	inputTypeButton = "button"
	inputTypeSwitch = "switch"
)

// EditField represents a field in the input edit form.
type EditField int

// Edit field constants.
const (
	EditFieldName EditField = iota
	EditFieldType
	EditFieldEnable
	EditFieldInvert
	EditFieldFactoryReset
	EditFieldCount
)

// EditSaveResultMsg is an alias for the shared save result message.
type EditSaveResultMsg = messages.SaveResultMsg

// EditOpenedMsg is an alias for the shared edit opened message.
type EditOpenedMsg = messages.EditOpenedMsg

// EditClosedMsg is an alias for the shared edit closed message.
type EditClosedMsg = messages.EditClosedMsg

// EditModel represents the input edit modal.
type EditModel struct {
	editmodal.Base

	inputID int
	loading bool

	// Original config for comparison
	original *model.InputConfig

	// Form inputs
	nameInput          form.TextInput
	typeDropdown       form.Select
	enableToggle       form.Toggle
	invertToggle       form.Toggle
	factoryResetToggle form.Toggle
}

// configLoadedMsg signals that input config was fetched.
type configLoadedMsg struct {
	config *model.InputConfig
}

// NewEditModel creates a new input edit modal.
func NewEditModel(ctx context.Context, svc *shelly.Service) EditModel {
	nameInput := form.NewTextInput(
		form.WithPlaceholder("Input name"),
		form.WithCharLimit(64),
		form.WithWidth(30),
		form.WithHelp("Optional display name for this input"),
	)

	typeDropdown := form.NewSelect(
		form.WithSelectOptions([]string{inputTypeButton, inputTypeSwitch}),
		form.WithSelectHelp("Input type (button=momentary, switch=toggle)"),
	)

	enableToggle := form.NewToggle(
		form.WithToggleLabel("Enable"),
		form.WithToggleHelp("Enable or disable this input"),
	)

	invertToggle := form.NewToggle(
		form.WithToggleLabel("Invert"),
		form.WithToggleHelp("Invert the input logic"),
	)

	factoryResetToggle := form.NewToggle(
		form.WithToggleLabel("Factory Reset"),
		form.WithToggleHelp("Allow factory reset via 5 toggles in 60s"),
	)

	return EditModel{
		Base: editmodal.Base{
			Ctx:    ctx,
			Svc:    svc,
			Styles: editmodal.DefaultStyles().WithLabelWidth(12),
		},
		nameInput:          nameInput,
		typeDropdown:       typeDropdown,
		enableToggle:       enableToggle,
		invertToggle:       invertToggle,
		factoryResetToggle: factoryResetToggle,
	}
}

// Show displays the edit modal for an existing input.
func (m EditModel) Show(device string, inputID int) (EditModel, tea.Cmd) {
	m.Base.Show(device, int(EditFieldCount))
	m.inputID = inputID
	m.loading = true
	m.original = nil

	// Blur all inputs first
	m = m.blurAllInputs()

	// Focus name input
	m.nameInput, _ = m.nameInput.Focus()

	// Fetch the current config
	return m, m.fetchConfig()
}

func (m EditModel) fetchConfig() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.Ctx, 30*time.Second)
		defer cancel()

		config, err := m.Svc.InputGetConfig(ctx, m.Base.Device, m.inputID)
		if err != nil {
			return messages.NewSaveError(m.inputID, err)
		}

		return configLoadedMsg{config: config}
	}
}

// Hide hides the edit modal.
func (m EditModel) Hide() EditModel {
	m.Base.Hide()
	m.nameInput = m.nameInput.Blur()
	m.typeDropdown = m.typeDropdown.Blur()
	m.enableToggle = m.enableToggle.Blur()
	m.invertToggle = m.invertToggle.Blur()
	m.factoryResetToggle = m.factoryResetToggle.Blur()
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
	m.nameInput = m.nameInput.SetWidth(inputWidth)
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
	case configLoadedMsg:
		m.loading = false
		m.original = msg.config
		m = m.populateFromConfig(msg.config)
		return m, nil

	case messages.SaveResultMsg:
		_, cmd := m.HandleSaveResult(msg)
		return m, cmd

	case messages.NavigationMsg:
		if m.Saving {
			return m, nil
		}
		return m.applyAction(m.HandleNavigation(msg))

	case messages.ToggleEnableRequestMsg:
		return m.handleSpace()

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) applyAction(action editmodal.KeyAction) (EditModel, tea.Cmd) {
	switch action {
	case editmodal.ActionClose:
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
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

func (m EditModel) populateFromConfig(cfg *model.InputConfig) EditModel {
	if cfg.Name != nil {
		m.nameInput = m.nameInput.SetValue(*cfg.Name)
	} else {
		m.nameInput = m.nameInput.SetValue("")
	}

	// Set type dropdown
	typeIdx := 0 // Default to button
	switch cfg.Type {
	case inputTypeButton:
		typeIdx = 0
	case inputTypeSwitch:
		typeIdx = 1
	}
	m.typeDropdown = m.typeDropdown.SetSelected(typeIdx)

	// Set toggles
	m.enableToggle = m.enableToggle.SetValue(cfg.Enable)
	m.invertToggle = m.invertToggle.SetValue(cfg.Invert)
	m.factoryResetToggle = m.factoryResetToggle.SetValue(cfg.FactoryReset)

	return m
}

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	if m.Saving {
		return m, nil
	}

	switch msg.String() {
	case keyconst.KeyEsc, "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case keyconst.KeyTab:
		m = m.moveFocus(m.NextField())
		return m, nil

	case keyconst.KeyShiftTab:
		m = m.moveFocus(m.PrevField())
		return m, nil

	case keyconst.KeyEnter:
		// Enter goes to focused input (text input / select dropdown), NOT save
		if EditField(m.Cursor) == EditFieldType {
			if m.typeDropdown.IsExpanded() {
				m.typeDropdown = m.typeDropdown.Collapse()
				return m, nil
			}
		}
		// Forward to focused input
		return m.updateFocusedInput(msg)

	case keyconst.KeyCtrlS:
		return m.save()
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleSpace() (EditModel, tea.Cmd) {
	switch EditField(m.Cursor) {
	case EditFieldType:
		if m.typeDropdown.IsExpanded() {
			m.typeDropdown = m.typeDropdown.Collapse()
		} else {
			m.typeDropdown = m.typeDropdown.Expand()
		}
	case EditFieldEnable:
		m.enableToggle = m.enableToggle.Toggle()
	case EditFieldInvert:
		m.invertToggle = m.invertToggle.Toggle()
	case EditFieldFactoryReset:
		m.factoryResetToggle = m.factoryResetToggle.Toggle()
	case EditFieldName, EditFieldCount:
		// No-op for text input and count sentinel
	}
	return m, nil
}

func (m EditModel) updateFocusedInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch EditField(m.Cursor) {
	case EditFieldName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case EditFieldType:
		m.typeDropdown, cmd = m.typeDropdown.Update(msg)
	case EditFieldEnable:
		m.enableToggle, cmd = m.enableToggle.Update(msg)
	case EditFieldInvert:
		m.invertToggle, cmd = m.invertToggle.Update(msg)
	case EditFieldFactoryReset:
		m.factoryResetToggle, cmd = m.factoryResetToggle.Update(msg)
	case EditFieldCount:
		// Count is a sentinel, no input to update
	}

	return m, cmd
}

// moveFocus blurs the old field and focuses the new one.
func (m EditModel) moveFocus(oldCursor, newCursor int) EditModel {
	m = m.blurField(EditField(oldCursor))
	m = m.focusField(EditField(newCursor))
	return m
}

func (m EditModel) blurField(field EditField) EditModel {
	switch field {
	case EditFieldName:
		m.nameInput = m.nameInput.Blur()
	case EditFieldType:
		m.typeDropdown = m.typeDropdown.Blur()
	case EditFieldEnable:
		m.enableToggle = m.enableToggle.Blur()
	case EditFieldInvert:
		m.invertToggle = m.invertToggle.Blur()
	case EditFieldFactoryReset:
		m.factoryResetToggle = m.factoryResetToggle.Blur()
	case EditFieldCount:
		// Count is a sentinel, no field to blur
	}
	return m
}

func (m EditModel) focusField(field EditField) EditModel {
	switch field {
	case EditFieldName:
		m.nameInput, _ = m.nameInput.Focus()
	case EditFieldType:
		m.typeDropdown = m.typeDropdown.Focus()
	case EditFieldEnable:
		m.enableToggle = m.enableToggle.Focus()
	case EditFieldInvert:
		m.invertToggle = m.invertToggle.Focus()
	case EditFieldFactoryReset:
		m.factoryResetToggle = m.factoryResetToggle.Focus()
	case EditFieldCount:
		// Count is a sentinel, no field to focus
	}
	return m
}

func (m EditModel) blurAllInputs() EditModel {
	m.nameInput = m.nameInput.Blur()
	m.typeDropdown = m.typeDropdown.Blur()
	m.enableToggle = m.enableToggle.Blur()
	m.invertToggle = m.invertToggle.Blur()
	m.factoryResetToggle = m.factoryResetToggle.Blur()
	return m
}

func (m EditModel) save() (EditModel, tea.Cmd) {
	// Build config from form values
	name := strings.TrimSpace(m.nameInput.Value())
	var namePtr *string
	if name != "" {
		namePtr = &name
	}

	selectedType := m.typeDropdown.SelectedValue()
	if selectedType == "" {
		selectedType = inputTypeButton // Default
	}

	cfg := &model.InputConfig{
		ID:           m.inputID,
		Name:         namePtr,
		Type:         selectedType,
		Enable:       m.enableToggle.Value(),
		Invert:       m.invertToggle.Value(),
		FactoryReset: m.factoryResetToggle.Value(),
	}

	m.StartSave()

	inputID := m.inputID
	cmd := m.SaveCmdWithID(inputID, func(ctx context.Context) error {
		return m.Svc.InputSetConfig(ctx, m.Base.Device, inputID, cfg)
	})
	return m, cmd
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.Visible() {
		return ""
	}

	footer := m.RenderSavingFooter("Tab: Next | Ctrl+S: Save | Esc: Cancel")

	// Build content
	var content strings.Builder

	// Input ID info (indented to align with form fields)
	content.WriteString("  ")
	content.WriteString(m.Styles.Info.Render("Input ID: "))
	content.WriteString(m.Styles.Selector.Render(strconv.Itoa(m.inputID)))
	content.WriteString("\n\n")

	// Form fields
	content.WriteString(m.renderFormFields())

	// Error display
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n")
		content.WriteString(errStr)
	}

	return m.RenderModal("Edit Input", content.String(), footer)
}

func (m EditModel) renderFormFields() string {
	var content strings.Builder

	// Name
	content.WriteString(m.RenderField(int(EditFieldName), "Name:", m.nameInput.View()))
	content.WriteString("\n\n")

	// Type
	content.WriteString(m.RenderField(int(EditFieldType), "Type:", m.typeDropdown.View()))
	content.WriteString("\n\n")

	// Enable
	content.WriteString(m.RenderField(int(EditFieldEnable), "Enabled:", m.enableToggle.View()))
	content.WriteString("\n\n")

	// Invert
	content.WriteString(m.RenderField(int(EditFieldInvert), "Invert Logic:", m.invertToggle.View()))
	content.WriteString("\n\n")

	// Factory Reset
	content.WriteString(m.RenderField(int(EditFieldFactoryReset), "Factory Reset:", m.factoryResetToggle.View()))

	return content.String()
}

// InputID returns the current input ID being edited.
func (m EditModel) InputID() int {
	return m.inputID
}

// Device returns the current device.
func (m EditModel) Device() string {
	return m.Base.Device
}

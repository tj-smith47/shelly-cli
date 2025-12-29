// Package inputs provides TUI components for managing device input settings.
package inputs

import (
	"context"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
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

// EditSaveResultMsg signals a save operation completed.
type EditSaveResultMsg struct {
	InputID int
	Err     error
}

// EditOpenedMsg signals the edit modal was opened.
type EditOpenedMsg struct{}

// EditClosedMsg signals the edit modal was closed.
type EditClosedMsg struct {
	Saved bool
}

// EditModel represents the input edit modal.
type EditModel struct {
	ctx     context.Context
	svc     *shelly.Service
	device  string
	inputID int
	visible bool
	cursor  EditField
	saving  bool
	err     error
	width   int
	height  int
	styles  EditStyles

	// Original config for comparison
	original *model.InputConfig

	// Form inputs
	nameInput          form.TextInput
	typeDropdown       form.Dropdown
	enableToggle       form.Toggle
	invertToggle       form.Toggle
	factoryResetToggle form.Toggle
}

// EditStyles holds styles for the input edit modal.
type EditStyles struct {
	Modal      lipgloss.Style
	Title      lipgloss.Style
	Label      lipgloss.Style
	LabelFocus lipgloss.Style
	Error      lipgloss.Style
	Help       lipgloss.Style
	Selector   lipgloss.Style
	Warning    lipgloss.Style
	Info       lipgloss.Style
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
			Width(14),
		LabelFocus: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Width(14),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Selector: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Info: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// NewEditModel creates a new input edit modal.
func NewEditModel(ctx context.Context, svc *shelly.Service) EditModel {
	nameInput := form.NewTextInput(
		form.WithPlaceholder("Input name"),
		form.WithCharLimit(64),
		form.WithWidth(30),
		form.WithHelp("Optional display name for this input"),
	)

	typeDropdown := form.NewDropdown(
		form.WithDropdownOptions([]string{"button", "switch"}),
		form.WithDropdownHelp("Input type (button=momentary, switch=toggle)"),
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
		ctx:                ctx,
		svc:                svc,
		styles:             DefaultEditStyles(),
		nameInput:          nameInput,
		typeDropdown:       typeDropdown,
		enableToggle:       enableToggle,
		invertToggle:       invertToggle,
		factoryResetToggle: factoryResetToggle,
	}
}

// Show displays the edit modal for an existing input.
func (m EditModel) Show(device string, inputID int) (EditModel, tea.Cmd) {
	m.device = device
	m.inputID = inputID
	m.visible = true
	m.cursor = EditFieldName
	m.saving = false
	m.err = nil
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
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		config, err := m.svc.InputGetConfig(ctx, m.device, m.inputID)
		if err != nil {
			return EditSaveResultMsg{InputID: m.inputID, Err: err}
		}

		return configLoadedMsg{config: config}
	}
}

type configLoadedMsg struct {
	config *model.InputConfig
}

// Hide hides the edit modal.
func (m EditModel) Hide() EditModel {
	m.visible = false
	return m
}

// Visible returns whether the modal is visible.
func (m EditModel) Visible() bool {
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
	case configLoadedMsg:
		m.original = msg.config
		m = m.populateFromConfig(msg.config)
		return m, nil

	case EditSaveResultMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Save successful, close the modal
		m.visible = false
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
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
	case "button":
		typeIdx = 0
	case "switch":
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
	key := msg.String()

	switch key {
	case "esc":
		if m.saving {
			return m, nil
		}
		m.visible = false
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case "tab", "down":
		if m.saving {
			return m, nil
		}
		m = m.nextField()
		return m, nil

	case "shift+tab", "up":
		if m.saving {
			return m, nil
		}
		m = m.prevField()
		return m, nil

	case "enter":
		if m.saving {
			return m, nil
		}
		// Handle dropdown selection or save
		if m.cursor == EditFieldType {
			if m.typeDropdown.IsExpanded() {
				m.typeDropdown = m.typeDropdown.Collapse()
				return m, nil
			}
		}
		return m.save()

	case "ctrl+s":
		if m.saving {
			return m, nil
		}
		return m.save()

	case " ":
		// Space toggles for toggles and opens dropdown
		return m.handleSpace()
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleSpace() (EditModel, tea.Cmd) {
	switch m.cursor {
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

	switch m.cursor {
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

func (m EditModel) nextField() EditModel {
	m = m.blurCurrentField()
	m.cursor++
	if m.cursor >= EditFieldCount {
		m.cursor = 0
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) prevField() EditModel {
	m = m.blurCurrentField()
	m.cursor--
	if m.cursor < 0 {
		m.cursor = EditFieldCount - 1
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) blurCurrentField() EditModel {
	switch m.cursor {
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

func (m EditModel) focusCurrentField() EditModel {
	switch m.cursor {
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
	m.err = nil

	// Build config from form values
	name := strings.TrimSpace(m.nameInput.Value())
	var namePtr *string
	if name != "" {
		namePtr = &name
	}

	selectedType := m.typeDropdown.SelectedValue()
	if selectedType == "" {
		selectedType = "button" // Default
	}

	cfg := &model.InputConfig{
		ID:           m.inputID,
		Name:         namePtr,
		Type:         selectedType,
		Enable:       m.enableToggle.Value(),
		Invert:       m.invertToggle.Value(),
		FactoryReset: m.factoryResetToggle.Value(),
	}

	m.saving = true
	return m, m.createSaveCmd(cfg)
}

func (m EditModel) createSaveCmd(cfg *model.InputConfig) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.InputSetConfig(ctx, m.device, m.inputID, cfg)
		return EditSaveResultMsg{InputID: m.inputID, Err: err}
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(m.styles.Title.Render("Edit Input"))
	content.WriteString("\n\n")

	// Input ID info
	content.WriteString(m.styles.Info.Render("Input ID: "))
	content.WriteString(m.styles.Selector.Render(strconv.Itoa(m.inputID)))
	content.WriteString("\n\n")

	// Form fields
	content.WriteString(m.renderFormFields())
	content.WriteString("\n\n")

	// Error message
	if m.err != nil {
		content.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
		content.WriteString("\n\n")
	}

	// Saving indicator
	if m.saving {
		content.WriteString(m.styles.Info.Render("Saving..."))
		content.WriteString("\n\n")
	}

	// Help
	content.WriteString(m.styles.Help.Render("Tab: next field • Enter/Ctrl+S: save • Esc: cancel"))

	// Calculate modal dimensions
	modalWidth := 50
	if m.width > 0 && m.width < modalWidth+10 {
		modalWidth = m.width - 10
	}

	modal := m.styles.Modal.Width(modalWidth).Render(content.String())

	// Center the modal
	return m.centerModal(modal)
}

func (m EditModel) renderFormFields() string {
	var content strings.Builder

	// Name
	content.WriteString(m.renderField(EditFieldName, "Name:", m.nameInput.View()))
	content.WriteString("\n\n")

	// Type
	content.WriteString(m.renderField(EditFieldType, "Type:", m.typeDropdown.View()))
	content.WriteString("\n\n")

	// Enable
	content.WriteString(m.renderField(EditFieldEnable, "Enabled:", m.enableToggle.View()))
	content.WriteString("\n\n")

	// Invert
	content.WriteString(m.renderField(EditFieldInvert, "Invert Logic:", m.invertToggle.View()))
	content.WriteString("\n\n")

	// Factory Reset
	content.WriteString(m.renderField(EditFieldFactoryReset, "Factory Reset:", m.factoryResetToggle.View()))

	return content.String()
}

func (m EditModel) renderField(field EditField, label, value string) string {
	labelStyle := m.styles.Label
	if m.cursor == field {
		labelStyle = m.styles.LabelFocus
	}
	return labelStyle.Render(label) + " " + value
}

func (m EditModel) centerModal(modal string) string {
	if m.width == 0 || m.height == 0 {
		return modal
	}

	lines := strings.Split(modal, "\n")
	modalHeight := len(lines)
	modalWidth := 0
	for _, line := range lines {
		if lipgloss.Width(line) > modalWidth {
			modalWidth = lipgloss.Width(line)
		}
	}

	// Calculate padding
	topPadding := (m.height - modalHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}
	leftPadding := (m.width - modalWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	// Build centered output
	var result strings.Builder
	for range topPadding {
		result.WriteString("\n")
	}
	for _, line := range lines {
		result.WriteString(strings.Repeat(" ", leftPadding))
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}

// InputID returns the current input ID being edited.
func (m EditModel) InputID() int {
	return m.inputID
}

// Device returns the current device.
func (m EditModel) Device() string {
	return m.device
}

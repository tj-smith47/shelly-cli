// Package scripts provides TUI components for managing device scripts.
package scripts

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// CreateField represents a field in the script create form.
type CreateField int

// Create field constants.
const (
	CreateFieldName CreateField = iota
	CreateFieldTemplate
	CreateFieldCount
)

// CreatedMsg signals that a new script was created.
type CreatedMsg struct {
	Device   string
	ScriptID int
	Name     string
	Err      error
}

// CreateModel represents the script create modal.
type CreateModel struct {
	ctx      context.Context
	svc      *automation.Service
	device   string
	visible  bool
	cursor   CreateField
	saving   bool
	err      error
	width    int
	height   int
	styles   editmodal.Styles
	tplNames []string

	// Form inputs
	nameInput        form.TextInput
	templateDropdown form.Select
}

// NewCreateModel creates a new script create modal.
func NewCreateModel(ctx context.Context, svc *automation.Service) CreateModel {
	nameInput := form.NewTextInput(
		form.WithPlaceholder("my_script"),
		form.WithCharLimit(64),
		form.WithWidth(30),
		form.WithHelp("Name for the new script"),
	)

	// Build template options from built-in templates
	templates := automation.ListAllScriptTemplates()
	tplNames := make([]string, 0, len(templates)+1)
	tplNames = append(tplNames, "(empty script)")
	for name := range templates {
		tplNames = append(tplNames, name)
	}

	templateDropdown := form.NewSelect(
		form.WithSelectOptions(tplNames),
		form.WithSelectHelp("Start from a template"),
	)

	return CreateModel{
		ctx:              ctx,
		svc:              svc,
		styles:           editmodal.DefaultStyles().WithLabelWidth(10),
		nameInput:        nameInput,
		templateDropdown: templateDropdown,
		tplNames:         tplNames,
	}
}

// Show displays the create modal.
func (m CreateModel) Show(device string) CreateModel {
	m.device = device
	m.visible = true
	m.cursor = CreateFieldName
	m.saving = false
	m.err = nil

	// Reset inputs
	m.nameInput = m.nameInput.SetValue("")
	m.templateDropdown = m.templateDropdown.SetSelected(0)

	// Focus name input
	m.nameInput, _ = m.nameInput.Focus()
	m.templateDropdown = m.templateDropdown.Blur()

	return m
}

// Hide hides the create modal.
func (m CreateModel) Hide() CreateModel {
	m.visible = false
	m.nameInput = m.nameInput.Blur()
	m.templateDropdown = m.templateDropdown.Blur()
	return m
}

// IsVisible returns whether the modal is visible.
func (m CreateModel) IsVisible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m CreateModel) SetSize(width, height int) CreateModel {
	m.width = width
	m.height = height
	return m
}

// Init returns the initial command.
func (m CreateModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m CreateModel) Update(msg tea.Msg) (CreateModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case CreatedMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - close modal
		m = m.Hide()
		return m, func() tea.Msg { return messages.EditClosedMsg{Saved: true} }

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m CreateModel) handleKey(msg tea.KeyPressMsg) (CreateModel, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return messages.EditClosedMsg{Saved: false} }

	case "ctrl+s", keyconst.KeyEnter:
		if m.cursor == CreateFieldTemplate && m.templateDropdown.IsExpanded() {
			m.templateDropdown = m.templateDropdown.Collapse()
			return m, nil
		}
		return m.save()

	case "tab":
		return m.nextField(), nil

	case "shift+tab":
		return m.prevField(), nil

	case " ":
		if m.cursor == CreateFieldTemplate {
			if m.templateDropdown.IsExpanded() {
				m.templateDropdown = m.templateDropdown.Collapse()
			} else {
				m.templateDropdown = m.templateDropdown.Expand()
			}
			return m, nil
		}
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m CreateModel) updateFocusedInput(msg tea.Msg) (CreateModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.cursor {
	case CreateFieldName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case CreateFieldTemplate:
		m.templateDropdown, cmd = m.templateDropdown.Update(msg)
	case CreateFieldCount:
		// No-op
	}

	return m, cmd
}

func (m CreateModel) nextField() CreateModel {
	m = m.blurCurrentField()
	if m.cursor < CreateFieldCount-1 {
		m.cursor++
	}
	m = m.focusCurrentField()
	return m
}

func (m CreateModel) prevField() CreateModel {
	m = m.blurCurrentField()
	if m.cursor > 0 {
		m.cursor--
	}
	m = m.focusCurrentField()
	return m
}

func (m CreateModel) blurCurrentField() CreateModel {
	switch m.cursor {
	case CreateFieldName:
		m.nameInput = m.nameInput.Blur()
	case CreateFieldTemplate:
		m.templateDropdown = m.templateDropdown.Blur()
	case CreateFieldCount:
		// No-op
	}
	return m
}

func (m CreateModel) focusCurrentField() CreateModel {
	switch m.cursor {
	case CreateFieldName:
		m.nameInput, _ = m.nameInput.Focus()
	case CreateFieldTemplate:
		m.templateDropdown = m.templateDropdown.Focus()
	case CreateFieldCount:
		// No-op
	}
	return m
}

func (m CreateModel) save() (CreateModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	// Validate name
	name := strings.TrimSpace(m.nameInput.Value())
	if name == "" {
		m.err = fmt.Errorf("script name is required")
		return m, nil
	}

	m.saving = true
	m.err = nil

	return m, m.createSaveCmd(name)
}

func (m CreateModel) createSaveCmd(name string) tea.Cmd {
	// Get selected template
	selectedTpl := m.templateDropdown.SelectedValue()
	var code string
	if selectedTpl != "(empty script)" {
		tpl, ok := automation.GetScriptTemplate(selectedTpl)
		if ok {
			code = tpl.Code
		}
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		// Create the script
		scriptID, err := m.svc.CreateScript(ctx, m.device, name)
		if err != nil {
			return CreatedMsg{Device: m.device, Err: err}
		}

		// If we have template code, upload it
		if code != "" {
			if err := m.svc.UpdateScriptCode(ctx, m.device, scriptID, code, false); err != nil {
				return CreatedMsg{Device: m.device, ScriptID: scriptID, Err: err}
			}
		}

		return CreatedMsg{Device: m.device, ScriptID: scriptID, Name: name}
	}
}

// View renders the create modal.
func (m CreateModel) View() string {
	if !m.visible {
		return ""
	}

	// Build footer
	footer := "Tab: Next | Ctrl+S: Save | Esc: Cancel"
	if m.saving {
		footer = "Creating..."
	}

	// Use common modal helper
	r := rendering.NewModal(m.width, m.height, "New Script", footer)

	// Build content
	return r.SetContent(m.renderFormFields()).Render()
}

func (m CreateModel) renderFormFields() string {
	var content strings.Builder

	// Name field
	content.WriteString(m.renderField(CreateFieldName, "Name:", m.nameInput.View()))
	content.WriteString("\n\n")

	// Template field
	content.WriteString(m.renderField(CreateFieldTemplate, "Template:", m.templateDropdown.View()))
	content.WriteString("\n")

	// Error display
	if m.err != nil {
		content.WriteString("\n\n")
		msg, _ := tuierrors.FormatError(m.err)
		content.WriteString(m.styles.Error.Render(msg))
	}

	return content.String()
}

func (m CreateModel) renderField(field CreateField, label, input string) string {
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

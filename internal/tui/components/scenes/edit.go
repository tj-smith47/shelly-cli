package scenes

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

// EditMode indicates what mode the edit modal is in.
type EditMode int

const (
	// EditModeCreate is for creating a new scene.
	EditModeCreate EditMode = iota
	// EditModeEdit is for editing an existing scene.
	EditModeEdit
	// EditModeDelete is for confirming scene deletion.
	EditModeDelete
)

// EditSavedMsg signals that a scene was saved.
type EditSavedMsg struct {
	SceneName string
	Err       error
}

// EditClosedMsg signals that the edit modal was closed.
type EditClosedMsg struct {
	Saved bool
}

// EditField identifies which field is focused in the editor.
type EditField int

const (
	// EditFieldName is the name field.
	EditFieldName EditField = iota
	// EditFieldDescription is the description field.
	EditFieldDescription
	// EditFieldActions is the actions list.
	EditFieldActions
	// EditFieldSave is the save button.
	EditFieldSave
	// EditFieldCancel is the cancel button.
	EditFieldCancel
)

// EditModel is the scene edit modal.
type EditModel struct {
	mode         EditMode
	originalName string // Original name for editing (to detect renames)
	nameInput    form.TextInput
	descInput    form.TextInput
	actions      []config.SceneAction
	actionCursor int
	focusedField EditField
	visible      bool
	width        int
	height       int
	styles       EditStyles
	err          error
}

// EditStyles holds styles for the edit modal.
type EditStyles struct {
	Title       lipgloss.Style
	Label       lipgloss.Style
	Input       lipgloss.Style
	InputFocus  lipgloss.Style
	Button      lipgloss.Style
	ButtonFocus lipgloss.Style
	Error       lipgloss.Style
	Muted       lipgloss.Style
	Action      lipgloss.Style
	ActionFocus lipgloss.Style
	Box         lipgloss.Style
}

// DefaultEditStyles returns the default styles for the edit modal.
func DefaultEditStyles() EditStyles {
	colors := theme.GetSemanticColors()
	return EditStyles{
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text),
		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(colors.Muted).
			Padding(0, 1),
		InputFocus: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(colors.Highlight).
			Padding(0, 1),
		Button: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(colors.Muted).
			Padding(0, 2),
		ButtonFocus: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(colors.Highlight).
			Background(colors.Highlight).
			Foreground(colors.Primary).
			Padding(0, 2),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Action: lipgloss.NewStyle().
			Foreground(colors.Text),
		ActionFocus: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight),
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(colors.Muted).
			Padding(1, 2),
	}
}

// NewEditModel creates a new scene edit modal.
func NewEditModel() EditModel {
	return EditModel{
		nameInput: form.NewTextInput(
			form.WithLabel("Name"),
			form.WithPlaceholder("Scene name"),
			form.WithWidth(30),
		),
		descInput: form.NewTextInput(
			form.WithLabel("Description"),
			form.WithPlaceholder("Description (optional)"),
			form.WithWidth(30),
		),
		actions:      []config.SceneAction{},
		focusedField: EditFieldName,
		styles:       DefaultEditStyles(),
	}
}

// SetSize sets the modal dimensions.
func (m EditModel) SetSize(width, height int) EditModel {
	m.width = width
	m.height = height
	return m
}

// ShowCreate shows the modal in create mode.
func (m EditModel) ShowCreate() EditModel {
	m.mode = EditModeCreate
	m.originalName = ""
	m.nameInput = form.NewTextInput(
		form.WithLabel("Name"),
		form.WithPlaceholder("Scene name"),
		form.WithWidth(30),
	)
	m.nameInput, _ = m.nameInput.Focus()
	m.descInput = form.NewTextInput(
		form.WithLabel("Description"),
		form.WithPlaceholder("Description (optional)"),
		form.WithWidth(30),
	)
	m.actions = []config.SceneAction{}
	m.actionCursor = 0
	m.focusedField = EditFieldName
	m.visible = true
	m.err = nil
	return m
}

// ShowEdit shows the modal in edit mode with existing scene data.
func (m EditModel) ShowEdit(scene config.Scene) EditModel {
	m.mode = EditModeEdit
	m.originalName = scene.Name
	m.nameInput = form.NewTextInput(
		form.WithLabel("Name"),
		form.WithPlaceholder("Scene name"),
		form.WithWidth(30),
	)
	m.nameInput = m.nameInput.SetValue(scene.Name)
	m.nameInput, _ = m.nameInput.Focus()
	m.descInput = form.NewTextInput(
		form.WithLabel("Description"),
		form.WithPlaceholder("Description (optional)"),
		form.WithWidth(30),
	)
	m.descInput = m.descInput.SetValue(scene.Description)
	m.actions = make([]config.SceneAction, len(scene.Actions))
	copy(m.actions, scene.Actions)
	m.actionCursor = 0
	m.focusedField = EditFieldName
	m.visible = true
	m.err = nil
	return m
}

// ShowDelete shows the modal in delete confirmation mode.
func (m EditModel) ShowDelete(scene config.Scene) EditModel {
	m.mode = EditModeDelete
	m.originalName = scene.Name
	m.visible = true
	m.err = nil
	return m
}

// Hide hides the modal.
func (m EditModel) Hide() EditModel {
	m.visible = false
	return m
}

// Visible returns whether the modal is visible.
func (m EditModel) Visible() bool {
	return m.visible
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
	case EditSavedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.visible = false
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	// Action messages from context system
	case messages.NavigationMsg:
		// Only handle navigation for actions field
		if m.focusedField == EditFieldActions {
			return m.handleActionsNavigation(msg)
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Update text inputs
	if m.focusedField == EditFieldName {
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	}
	if m.focusedField == EditFieldDescription {
		var cmd tea.Cmd
		m.descInput, cmd = m.descInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m EditModel) handleActionsNavigation(msg messages.NavigationMsg) (EditModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		if m.actionCursor > 0 {
			m.actionCursor--
		}
	case messages.NavDown:
		if m.actionCursor < len(m.actions)-1 {
			m.actionCursor++
		}
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable
	}
	return m, nil
}

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	key := msg.String()

	// Handle delete mode
	if m.mode == EditModeDelete {
		return m.handleDeleteKey(key)
	}

	// Handle common keys
	switch key {
	case keyconst.KeyEsc:
		m.visible = false
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	case keyconst.KeyTab:
		return m.focusNext(), nil
	case keyconst.KeyShiftTab:
		return m.focusPrev(), nil
	case keyconst.KeyEnter:
		return m.handleEnter()
	}

	// Handle field-specific keys
	switch m.focusedField {
	case EditFieldActions:
		return m.handleActionsKey(key)
	case EditFieldName:
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	case EditFieldDescription:
		var cmd tea.Cmd
		m.descInput, cmd = m.descInput.Update(msg)
		return m, cmd
	case EditFieldSave, EditFieldCancel:
		// Buttons don't handle character input
		return m, nil
	}

	return m, nil
}

func (m EditModel) handleDeleteKey(key string) (EditModel, tea.Cmd) {
	switch key {
	case keyconst.KeyEsc, "n":
		m.visible = false
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	case keyconst.KeyEnter, "y":
		return m, m.performDelete()
	}
	return m, nil
}

func (m EditModel) handleActionsKey(key string) (EditModel, tea.Cmd) {
	// Modal-specific keys not covered by NavigationMsg
	if key == "d" {
		// Delete selected action
		if len(m.actions) > 0 && m.actionCursor < len(m.actions) {
			m.actions = append(m.actions[:m.actionCursor], m.actions[m.actionCursor+1:]...)
			if m.actionCursor >= len(m.actions) && m.actionCursor > 0 {
				m.actionCursor--
			}
		}
	}
	return m, nil
}

func (m EditModel) handleEnter() (EditModel, tea.Cmd) {
	switch m.focusedField {
	case EditFieldSave:
		return m, m.save()
	case EditFieldCancel:
		m.visible = false
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }
	default:
		return m.focusNext(), nil
	}
}

func (m EditModel) focusNext() EditModel {
	m.nameInput = m.nameInput.Blur()
	m.descInput = m.descInput.Blur()

	switch m.focusedField {
	case EditFieldName:
		m.focusedField = EditFieldDescription
		m.descInput, _ = m.descInput.Focus()
	case EditFieldDescription:
		m.focusedField = EditFieldActions
	case EditFieldActions:
		m.focusedField = EditFieldSave
	case EditFieldSave:
		m.focusedField = EditFieldCancel
	case EditFieldCancel:
		m.focusedField = EditFieldName
		m.nameInput, _ = m.nameInput.Focus()
	}
	return m
}

func (m EditModel) focusPrev() EditModel {
	m.nameInput = m.nameInput.Blur()
	m.descInput = m.descInput.Blur()

	switch m.focusedField {
	case EditFieldName:
		m.focusedField = EditFieldCancel
	case EditFieldDescription:
		m.focusedField = EditFieldName
		m.nameInput, _ = m.nameInput.Focus()
	case EditFieldActions:
		m.focusedField = EditFieldDescription
		m.descInput, _ = m.descInput.Focus()
	case EditFieldSave:
		m.focusedField = EditFieldActions
	case EditFieldCancel:
		m.focusedField = EditFieldSave
	}
	return m
}

func (m EditModel) save() tea.Cmd {
	name := strings.TrimSpace(m.nameInput.Value())
	if name == "" {
		return func() tea.Msg {
			return EditSavedMsg{SceneName: "", Err: fmt.Errorf("scene name is required")}
		}
	}

	if m.mode == EditModeCreate {
		return m.saveCreate(name)
	}
	return m.saveUpdate(name)
}

func (m EditModel) saveCreate(name string) tea.Cmd {
	actions := m.actions
	desc := strings.TrimSpace(m.descInput.Value())
	return func() tea.Msg {
		err := config.CreateScene(name, desc)
		if err == nil && len(actions) > 0 {
			err = config.SetSceneActions(name, actions)
		}
		return EditSavedMsg{SceneName: name, Err: err}
	}
}

func (m EditModel) saveUpdate(name string) tea.Cmd {
	originalName := m.originalName
	actions := m.actions
	desc := strings.TrimSpace(m.descInput.Value())
	return func() tea.Msg {
		var err error
		if originalName != name {
			err = config.UpdateScene(originalName, name, desc)
		} else if desc != "" {
			err = config.UpdateScene(originalName, "", desc)
		}
		if err == nil {
			err = config.SetSceneActions(name, actions)
		}
		return EditSavedMsg{SceneName: name, Err: err}
	}
}

func (m EditModel) performDelete() tea.Cmd {
	name := m.originalName
	return func() tea.Msg {
		err := config.DeleteScene(name)
		return EditSavedMsg{SceneName: name, Err: err}
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	if m.mode == EditModeDelete {
		return m.renderDeleteConfirmation()
	}

	return m.renderEditForm()
}

func (m EditModel) renderDeleteConfirmation() string {
	var content strings.Builder

	content.WriteString(m.styles.Title.Render("Delete Scene"))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("Are you sure you want to delete scene %q?\n\n", m.originalName))
	content.WriteString(m.styles.Muted.Render("Press y to confirm, n or Esc to cancel"))

	box := m.styles.Box.
		Width(min(60, m.width-4)).
		Render(content.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m EditModel) renderEditForm() string {
	var content strings.Builder

	// Title
	title := "Create Scene"
	if m.mode == EditModeEdit {
		title = "Edit Scene"
	}
	content.WriteString(m.styles.Title.Render(title))
	content.WriteString("\n\n")

	// Error message
	if m.err != nil {
		content.WriteString(m.styles.Error.Render(m.err.Error()))
		content.WriteString("\n\n")
	}

	// Name field
	nameStyle := m.styles.Input
	if m.focusedField == EditFieldName {
		nameStyle = m.styles.InputFocus
	}
	content.WriteString(nameStyle.Width(40).Render(m.nameInput.View()))
	content.WriteString("\n\n")

	// Description field
	descStyle := m.styles.Input
	if m.focusedField == EditFieldDescription {
		descStyle = m.styles.InputFocus
	}
	content.WriteString(descStyle.Width(40).Render(m.descInput.View()))
	content.WriteString("\n\n")

	// Actions list
	content.WriteString(m.styles.Label.Render(fmt.Sprintf("Actions (%d):", len(m.actions))))
	content.WriteString("\n")
	content.WriteString(m.renderActionsList())
	content.WriteString("\n\n")

	// Buttons
	saveStyle := m.styles.Button
	cancelStyle := m.styles.Button
	if m.focusedField == EditFieldSave {
		saveStyle = m.styles.ButtonFocus
	}
	if m.focusedField == EditFieldCancel {
		cancelStyle = m.styles.ButtonFocus
	}
	content.WriteString(saveStyle.Render("Save") + "  " + cancelStyle.Render("Cancel"))
	content.WriteString("\n\n")
	content.WriteString(m.styles.Muted.Render("Tab: next field | Shift+Tab: prev | Enter: select | Esc: cancel"))

	box := m.styles.Box.
		Width(min(60, m.width-4)).
		Render(content.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m EditModel) renderActionsList() string {
	if len(m.actions) == 0 {
		return m.styles.Muted.Render("  No actions defined")
	}

	var lines []string
	for i, action := range m.actions {
		line := fmt.Sprintf("  %s: %s", action.Device, action.Method)
		if m.focusedField == EditFieldActions && i == m.actionCursor {
			lines = append(lines, m.styles.ActionFocus.Render("> "+output.Truncate(line, 50)))
		} else {
			lines = append(lines, m.styles.Action.Render("  "+output.Truncate(line, 50)))
		}
	}

	lines = m.paginateActionLines(lines)
	return strings.Join(lines, "\n")
}

func (m EditModel) paginateActionLines(lines []string) []string {
	const maxLines = 5
	if len(lines) <= maxLines {
		return lines
	}

	start, end := calculatePaginationRange(m.actionCursor, len(lines), maxLines)
	return lines[start:end]
}

func calculatePaginationRange(cursor, total, maxVisible int) (start, end int) {
	start = cursor - maxVisible/2
	start = max(start, 0)
	end = start + maxVisible
	if end > total {
		end = total
		start = max(end-maxVisible, 0)
	}
	return start, end
}

// AddAction adds an action to the scene being edited.
func (m EditModel) AddAction(action config.SceneAction) EditModel {
	m.actions = append(m.actions, action)
	return m
}

// Package fleet provides TUI components for Shelly Cloud Fleet management.
package fleet

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// GroupEditMode represents the mode of the group edit modal.
type GroupEditMode int

// Group edit mode constants.
const (
	GroupEditModeCreate GroupEditMode = iota
	GroupEditModeEdit
	GroupEditModeDelete
)

// GroupEditField represents a field in the group edit form.
type GroupEditField int

// Group edit field constants.
const (
	GroupEditFieldName GroupEditField = iota
	GroupEditFieldDevices
	GroupEditFieldCount
)

// GroupEditSaveResultMsg signals a save operation completed.
type GroupEditSaveResultMsg struct {
	GroupID string
	Mode    GroupEditMode
	Err     error
}

// GroupEditOpenedMsg signals the edit modal was opened.
type GroupEditOpenedMsg struct{}

// GroupEditClosedMsg signals the edit modal was closed.
type GroupEditClosedMsg struct {
	Saved bool
}

// GroupEditModel represents the group edit modal.
type GroupEditModel struct {
	fleet   *integrator.FleetManager
	visible bool
	mode    GroupEditMode
	cursor  GroupEditField
	saving  bool
	err     error
	width   int
	height  int
	styles  GroupEditStyles

	// Group being edited
	groupID  string
	original *integrator.DeviceGroup

	// Form inputs
	nameInput form.TextInput

	// Device selection
	allDevices     []integrator.AccountDevice
	selectedIDs    map[string]bool
	deviceScroller *panel.Scroller
}

// GroupEditStyles holds styles for the group edit modal.
type GroupEditStyles struct {
	Modal        lipgloss.Style
	Title        lipgloss.Style
	Label        lipgloss.Style
	LabelFocus   lipgloss.Style
	Error        lipgloss.Style
	Help         lipgloss.Style
	Selector     lipgloss.Style
	DeviceLine   lipgloss.Style
	DeviceSelect lipgloss.Style
	DeviceName   lipgloss.Style
	Checkbox     lipgloss.Style
	CheckboxOn   lipgloss.Style
	CheckboxOff  lipgloss.Style
	Warning      lipgloss.Style
	Info         lipgloss.Style
}

// DefaultGroupEditStyles returns the default edit modal styles.
func DefaultGroupEditStyles() GroupEditStyles {
	colors := theme.GetSemanticColors()
	return GroupEditStyles{
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
		DeviceLine: lipgloss.NewStyle().
			Foreground(colors.Text),
		DeviceSelect: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight),
		DeviceName: lipgloss.NewStyle().
			Foreground(colors.Text),
		Checkbox: lipgloss.NewStyle().
			Foreground(colors.Muted),
		CheckboxOn: lipgloss.NewStyle().
			Foreground(colors.Online),
		CheckboxOff: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Info: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// NewGroupEditModel creates a new group edit modal.
func NewGroupEditModel() GroupEditModel {
	nameInput := form.NewTextInput(
		form.WithPlaceholder("Group Name"),
		form.WithCharLimit(64),
		form.WithWidth(30),
		form.WithHelp("Enter a descriptive name for this group"),
	)

	return GroupEditModel{
		styles:         DefaultGroupEditStyles(),
		nameInput:      nameInput,
		selectedIDs:    make(map[string]bool),
		deviceScroller: panel.NewScroller(0, 8),
	}
}

// ShowCreate displays the modal for creating a new group.
func (m GroupEditModel) ShowCreate(fleet *integrator.FleetManager) GroupEditModel {
	m.fleet = fleet
	m.visible = true
	m.mode = GroupEditModeCreate
	m.cursor = GroupEditFieldName
	m.saving = false
	m.err = nil
	m.groupID = uuid.New().String()
	m.original = nil

	// Reset form
	m.nameInput = m.nameInput.SetValue("")
	m.nameInput, _ = m.nameInput.Focus()

	// Load all devices
	m.loadDevicesInto(&m)
	m.selectedIDs = make(map[string]bool)

	return m
}

// ShowEdit displays the modal for editing an existing group.
func (m GroupEditModel) ShowEdit(fleet *integrator.FleetManager, group *integrator.DeviceGroup) GroupEditModel {
	m.fleet = fleet
	m.visible = true
	m.mode = GroupEditModeEdit
	m.cursor = GroupEditFieldName
	m.saving = false
	m.err = nil
	m.groupID = group.ID
	m.original = group

	// Set form values
	m.nameInput = m.nameInput.SetValue(group.Name)
	m.nameInput, _ = m.nameInput.Focus()

	// Load all devices and mark selected
	m.loadDevicesInto(&m)
	m.selectedIDs = make(map[string]bool)
	for _, id := range group.DeviceIDs {
		m.selectedIDs[id] = true
	}

	return m
}

// ShowDelete displays the delete confirmation.
func (m GroupEditModel) ShowDelete(fleet *integrator.FleetManager, group *integrator.DeviceGroup) GroupEditModel {
	m.fleet = fleet
	m.visible = true
	m.mode = GroupEditModeDelete
	m.cursor = GroupEditFieldName
	m.saving = false
	m.err = nil
	m.groupID = group.ID
	m.original = group

	return m
}

func (m GroupEditModel) loadDevicesInto(target *GroupEditModel) {
	if m.fleet == nil {
		target.allDevices = nil
		target.deviceScroller.SetItemCount(0)
		return
	}

	target.allDevices = m.fleet.AccountManager().ListDevices()
	target.deviceScroller.SetItemCount(len(target.allDevices))
	target.deviceScroller.CursorToStart()
}

// Hide hides the edit modal.
func (m GroupEditModel) Hide() GroupEditModel {
	m.visible = false
	m.nameInput = m.nameInput.Blur()
	return m
}

// Visible returns whether the modal is visible.
func (m GroupEditModel) Visible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m GroupEditModel) SetSize(width, height int) GroupEditModel {
	m.width = width
	m.height = height
	// Calculate device list height based on modal size
	deviceListHeight := min(10, height/3)
	if deviceListHeight < 3 {
		deviceListHeight = 3
	}
	m.deviceScroller.SetVisibleRows(deviceListHeight)
	return m
}

// Init returns the initial command.
func (m GroupEditModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m GroupEditModel) Update(msg tea.Msg) (GroupEditModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case GroupEditSaveResultMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - close modal
		m = m.Hide()
		return m, func() tea.Msg { return GroupEditClosedMsg{Saved: true} }

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input (only for name field)
	if m.cursor == GroupEditFieldName && m.mode != GroupEditModeDelete {
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m GroupEditModel) handleKey(msg tea.KeyPressMsg) (GroupEditModel, tea.Cmd) {
	key := msg.String()

	// Delete mode has different keybindings
	if m.mode == GroupEditModeDelete {
		return m.handleDeleteKey(key)
	}

	switch key {
	case "esc", "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return GroupEditClosedMsg{Saved: false} }

	case "ctrl+s", keyconst.KeyEnter:
		if m.cursor == GroupEditFieldName {
			// Move to devices instead of saving
			return m.nextField(), nil
		}
		return m.save()

	case "tab", "down":
		return m.nextField(), nil

	case "shift+tab", "up":
		return m.prevField(), nil

	case " ":
		// Toggle device selection if in devices field
		if m.cursor == GroupEditFieldDevices && len(m.allDevices) > 0 {
			return m.toggleCurrentDevice(), nil
		}

	case "j":
		// Scroll down in device list
		if m.cursor == GroupEditFieldDevices {
			m.deviceScroller.CursorDown()
			return m, nil
		}

	case "k":
		// Scroll up in device list
		if m.cursor == GroupEditFieldDevices {
			m.deviceScroller.CursorUp()
			return m, nil
		}
	}

	// Forward to name input
	if m.cursor == GroupEditFieldName {
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m GroupEditModel) handleDeleteKey(key string) (GroupEditModel, tea.Cmd) {
	switch key {
	case "esc", "n", "N":
		m = m.Hide()
		return m, func() tea.Msg { return GroupEditClosedMsg{Saved: false} }

	case "y", "Y", keyconst.KeyEnter:
		return m.doDelete()
	}

	return m, nil
}

func (m GroupEditModel) toggleCurrentDevice() GroupEditModel {
	if len(m.allDevices) == 0 {
		return m
	}

	cursor := m.deviceScroller.Cursor()
	if cursor < 0 || cursor >= len(m.allDevices) {
		return m
	}

	deviceID := m.allDevices[cursor].DeviceID
	if m.selectedIDs[deviceID] {
		delete(m.selectedIDs, deviceID)
	} else {
		m.selectedIDs[deviceID] = true
	}

	return m
}

func (m GroupEditModel) nextField() GroupEditModel {
	m = m.blurCurrentField()
	if m.cursor < GroupEditFieldCount-1 {
		m.cursor++
	}
	m = m.focusCurrentField()
	return m
}

func (m GroupEditModel) prevField() GroupEditModel {
	m = m.blurCurrentField()
	if m.cursor > 0 {
		m.cursor--
	}
	m = m.focusCurrentField()
	return m
}

func (m GroupEditModel) blurCurrentField() GroupEditModel {
	if m.cursor == GroupEditFieldName {
		m.nameInput = m.nameInput.Blur()
	}
	return m
}

func (m GroupEditModel) focusCurrentField() GroupEditModel {
	if m.cursor == GroupEditFieldName {
		m.nameInput, _ = m.nameInput.Focus()
	}
	return m
}

func (m GroupEditModel) save() (GroupEditModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	name := strings.TrimSpace(m.nameInput.Value())
	if name == "" {
		m.err = fmt.Errorf("group name is required")
		return m, nil
	}

	m.saving = true
	m.err = nil

	// Collect selected device IDs
	deviceIDs := make([]string, 0, len(m.selectedIDs))
	for id := range m.selectedIDs {
		deviceIDs = append(deviceIDs, id)
	}

	return m, m.createSaveCmd(name, deviceIDs)
}

func (m GroupEditModel) createSaveCmd(name string, deviceIDs []string) tea.Cmd {
	return func() tea.Msg {
		if m.fleet == nil {
			return GroupEditSaveResultMsg{Err: fmt.Errorf("not connected to fleet")}
		}

		switch m.mode {
		case GroupEditModeCreate:
			group := m.fleet.CreateGroup(m.groupID, name, deviceIDs)
			if group == nil {
				return GroupEditSaveResultMsg{Err: fmt.Errorf("failed to create group")}
			}
			return GroupEditSaveResultMsg{GroupID: group.ID, Mode: m.mode}

		case GroupEditModeEdit:
			// For edit, we need to update name and devices
			// First get the current group
			group, ok := m.fleet.GetGroup(m.groupID)
			if !ok {
				return GroupEditSaveResultMsg{Err: fmt.Errorf("group not found")}
			}

			// Update name by recreating with new name but same devices temporarily
			// Then update device membership
			// Since FleetManager doesn't have UpdateGroup, we'll delete and recreate
			m.fleet.DeleteGroup(m.groupID)
			newGroup := m.fleet.CreateGroup(m.groupID, name, deviceIDs)
			if newGroup == nil {
				// Try to restore original
				m.fleet.CreateGroup(group.ID, group.Name, group.DeviceIDs)
				return GroupEditSaveResultMsg{Err: fmt.Errorf("failed to update group")}
			}
			return GroupEditSaveResultMsg{GroupID: m.groupID, Mode: m.mode}

		case GroupEditModeDelete:
			// Delete mode is handled by doDelete(), not createSaveCmd
			return GroupEditSaveResultMsg{Err: fmt.Errorf("delete mode should use doDelete()")}
		}

		return GroupEditSaveResultMsg{Err: fmt.Errorf("invalid mode")}
	}
}

func (m GroupEditModel) doDelete() (GroupEditModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	m.saving = true
	m.err = nil

	return m, func() tea.Msg {
		if m.fleet == nil {
			return GroupEditSaveResultMsg{Err: fmt.Errorf("not connected to fleet")}
		}

		ok := m.fleet.DeleteGroup(m.groupID)
		if !ok {
			return GroupEditSaveResultMsg{Err: fmt.Errorf("failed to delete group")}
		}

		return GroupEditSaveResultMsg{GroupID: m.groupID, Mode: GroupEditModeDelete}
	}
}

// View renders the edit modal.
func (m GroupEditModel) View() string {
	if !m.visible {
		return ""
	}

	// Build title and footer based on mode
	var title, footer string
	switch m.mode {
	case GroupEditModeDelete:
		title = "Delete Group"
		if m.saving {
			footer = "Deleting..."
		} else {
			footer = "y: Delete | n/Esc: Cancel"
		}
	case GroupEditModeCreate:
		title = "Create Group"
		footer = m.buildEditFooter()
	default:
		title = "Edit Group"
		footer = m.buildEditFooter()
	}

	// Use common modal helper
	r := rendering.NewModal(m.width, m.height, title, footer)

	// Build content based on mode
	var content string
	if m.mode == GroupEditModeDelete {
		content = m.renderDeleteContent()
	} else {
		content = m.renderEditContent()
	}

	return r.SetContent(content).Render()
}

func (m GroupEditModel) buildEditFooter() string {
	if m.saving {
		return "Saving..."
	}
	if m.cursor == GroupEditFieldDevices {
		return "j/k: Navigate | Space: Toggle | Ctrl+S/Enter: Save | Esc: Cancel"
	}
	return "Tab: Next field | Ctrl+S: Save | Esc: Cancel"
}

func (m GroupEditModel) renderDeleteContent() string {
	var content strings.Builder

	if m.original != nil {
		content.WriteString(m.styles.Warning.Render("Are you sure you want to delete this group?"))
		content.WriteString("\n\n")
		content.WriteString(m.styles.Label.Render("Name: "))
		content.WriteString(m.styles.DeviceName.Render(m.original.Name))
		content.WriteString("\n")
		content.WriteString(m.styles.Label.Render("Devices: "))
		content.WriteString(m.styles.DeviceName.Render(fmt.Sprintf("%d", len(m.original.DeviceIDs))))
	}

	if m.err != nil {
		content.WriteString("\n\n")
		content.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
	}

	return content.String()
}

func (m GroupEditModel) renderEditContent() string {
	var content strings.Builder

	// Name field
	content.WriteString(m.renderField(GroupEditFieldName, "Name:", m.nameInput.View()))
	content.WriteString("\n\n")

	// Devices field
	content.WriteString(m.renderDevicesField())

	// Error display
	if m.err != nil {
		content.WriteString("\n\n")
		content.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
	}

	return content.String()
}

func (m GroupEditModel) renderField(field GroupEditField, label, input string) string {
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

func (m GroupEditModel) renderDevicesField() string {
	var content strings.Builder

	// Label
	if m.cursor == GroupEditFieldDevices {
		content.WriteString(m.styles.Selector.Render("> "))
		content.WriteString(m.styles.LabelFocus.Render("Devices:"))
	} else {
		content.WriteString("  ")
		content.WriteString(m.styles.Label.Render("Devices:"))
	}

	// Selection count
	content.WriteString(" ")
	content.WriteString(m.styles.Info.Render(fmt.Sprintf("(%d selected)", len(m.selectedIDs))))
	content.WriteString("\n")

	if len(m.allDevices) == 0 {
		content.WriteString("    ")
		content.WriteString(m.styles.Info.Render("No devices available"))
		return content.String()
	}

	// Device list with scrolling
	start, end := m.deviceScroller.VisibleRange()
	for i := start; i < end; i++ {
		device := m.allDevices[i]
		isSelected := m.selectedIDs[device.DeviceID]
		isCursor := m.deviceScroller.IsCursorAt(i) && m.cursor == GroupEditFieldDevices

		content.WriteString(m.renderDeviceLine(device, isSelected, isCursor))
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(m.allDevices) > m.deviceScroller.VisibleRows() {
		content.WriteString("\n    ")
		content.WriteString(m.styles.Info.Render(m.deviceScroller.ScrollInfo()))
	}

	return content.String()
}

func (m GroupEditModel) renderDeviceLine(device integrator.AccountDevice, isSelected, isCursor bool) string {
	var line strings.Builder

	// Indentation
	line.WriteString("    ")

	// Cursor indicator
	if isCursor {
		line.WriteString(m.styles.Selector.Render("> "))
	} else {
		line.WriteString("  ")
	}

	// Checkbox
	if isSelected {
		line.WriteString(m.styles.CheckboxOn.Render("[✓] "))
	} else {
		line.WriteString(m.styles.CheckboxOff.Render("[ ] "))
	}

	// Device name
	name := device.Name
	if name == "" {
		name = device.DeviceID
	}
	if len(name) > 25 {
		name = name[:22] + "..."
	}

	if isCursor {
		line.WriteString(m.styles.DeviceSelect.Render(name))
	} else {
		line.WriteString(m.styles.DeviceName.Render(name))
	}

	return line.String()
}

// Mode returns the current edit mode.
func (m GroupEditModel) Mode() GroupEditMode {
	return m.mode
}

// GroupID returns the current group ID.
func (m GroupEditModel) GroupID() string {
	return m.groupID
}

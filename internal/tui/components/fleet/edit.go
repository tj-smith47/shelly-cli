// Package fleet provides TUI components for Shelly Cloud Fleet management.
package fleet

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
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

// GroupEditSaveResultMsg is an alias for the shared save result message.
type GroupEditSaveResultMsg = messages.SaveResultMsg

// GroupEditOpenedMsg is an alias for the shared edit opened message.
type GroupEditOpenedMsg = messages.EditOpenedMsg

// GroupEditClosedMsg is an alias for the shared edit closed message.
type GroupEditClosedMsg = messages.EditClosedMsg

// GroupEditModel represents the group edit modal.
type GroupEditModel struct {
	editmodal.Base

	fleet *integrator.FleetManager
	mode  GroupEditMode

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

// NewGroupEditModel creates a new group edit modal.
func NewGroupEditModel() GroupEditModel {
	nameInput := form.NewTextInput(
		form.WithPlaceholder("Group Name"),
		form.WithCharLimit(64),
		form.WithWidth(30),
		form.WithHelp("Enter a descriptive name for this group"),
	)

	return GroupEditModel{
		Base:           editmodal.Base{Styles: editmodal.DefaultStyles().WithLabelWidth(10)},
		nameInput:      nameInput,
		selectedIDs:    make(map[string]bool),
		deviceScroller: panel.NewScroller(0, 8),
	}
}

// ShowCreate displays the modal for creating a new group.
func (m GroupEditModel) ShowCreate(fleet *integrator.FleetManager) GroupEditModel {
	m.fleet = fleet
	m.Show("", int(GroupEditFieldCount))
	m.mode = GroupEditModeCreate
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
	m.Show("", int(GroupEditFieldCount))
	m.mode = GroupEditModeEdit
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
	m.Show("", int(GroupEditFieldCount))
	m.mode = GroupEditModeDelete
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
	m.Base.Hide()
	m.nameInput = m.nameInput.Blur()
	return m
}

// Visible returns whether the modal is visible.
func (m GroupEditModel) Visible() bool {
	return m.Base.Visible()
}

// SetSize sets the modal dimensions.
func (m GroupEditModel) SetSize(width, height int) GroupEditModel {
	m.Base.SetSize(width, height)
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
	if !m.Visible() {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m GroupEditModel) handleMessage(msg tea.Msg) (GroupEditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SaveResultMsg:
		saved, cmd := m.HandleSaveResult(msg)
		if saved {
			return m, cmd
		}
		return m, nil

	// Action messages from context system
	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case messages.ToggleEnableRequestMsg:
		// Toggle device selection if in devices field
		if GroupEditField(m.Cursor) == GroupEditFieldDevices && len(m.allDevices) > 0 {
			return m.toggleCurrentDevice(), nil
		}
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input (only for name field)
	if GroupEditField(m.Cursor) == GroupEditFieldName && m.mode != GroupEditModeDelete {
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m GroupEditModel) handleNavigation(msg messages.NavigationMsg) (GroupEditModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		if GroupEditField(m.Cursor) == GroupEditFieldDevices {
			m.deviceScroller.CursorUp()
			return m, nil
		}
		return m.prevField(), nil
	case messages.NavDown:
		if GroupEditField(m.Cursor) == GroupEditFieldDevices {
			m.deviceScroller.CursorDown()
			return m, nil
		}
		return m.nextField(), nil
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this component
	}
	return m, nil
}

func (m GroupEditModel) handleKey(msg tea.KeyPressMsg) (GroupEditModel, tea.Cmd) {
	key := msg.String()

	// Delete mode has different keybindings
	if m.mode == GroupEditModeDelete {
		return m.handleDeleteKey(key)
	}

	// Modal-specific keys not covered by action messages
	switch key {
	case keyconst.KeyEsc, "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return GroupEditClosedMsg{Saved: false} }
	case keyconst.KeyCtrlS:
		if GroupEditField(m.Cursor) == GroupEditFieldName {
			// Move to devices instead of saving
			return m.nextField(), nil
		}
		return m.save()
	case keyconst.KeyTab:
		return m.nextField(), nil
	case keyconst.KeyShiftTab:
		return m.prevField(), nil
	}

	// Forward to name input when focused
	if GroupEditField(m.Cursor) == GroupEditFieldName {
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
	case "y", "Y", "enter":
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
	if m.Cursor < int(GroupEditFieldCount)-1 {
		m.Cursor++
	}
	m = m.focusCurrentField()
	return m
}

func (m GroupEditModel) prevField() GroupEditModel {
	m = m.blurCurrentField()
	if m.Cursor > 0 {
		m.Cursor--
	}
	m = m.focusCurrentField()
	return m
}

func (m GroupEditModel) blurCurrentField() GroupEditModel {
	if GroupEditField(m.Cursor) == GroupEditFieldName {
		m.nameInput = m.nameInput.Blur()
	}
	return m
}

func (m GroupEditModel) focusCurrentField() GroupEditModel {
	if GroupEditField(m.Cursor) == GroupEditFieldName {
		m.nameInput, _ = m.nameInput.Focus()
	}
	return m
}

func (m GroupEditModel) save() (GroupEditModel, tea.Cmd) {
	if m.Saving {
		return m, nil
	}

	name := strings.TrimSpace(m.nameInput.Value())
	if name == "" {
		m.Err = fmt.Errorf("group name is required")
		return m, nil
	}

	m.StartSave()

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
			return messages.NewSaveError(nil, fmt.Errorf("not connected to fleet"))
		}

		switch m.mode {
		case GroupEditModeCreate:
			group := m.fleet.CreateGroup(m.groupID, name, deviceIDs)
			if group == nil {
				return messages.NewSaveError(nil, fmt.Errorf("failed to create group"))
			}
			return messages.NewSaveResult(group.ID)

		case GroupEditModeEdit:
			// For edit, we need to update name and devices
			// First get the current group
			group, ok := m.fleet.GetGroup(m.groupID)
			if !ok {
				return messages.NewSaveError(m.groupID, fmt.Errorf("group not found"))
			}

			// Update name by recreating with new name but same devices temporarily
			// Then update device membership
			// Since FleetManager doesn't have UpdateGroup, we'll delete and recreate
			m.fleet.DeleteGroup(m.groupID)
			newGroup := m.fleet.CreateGroup(m.groupID, name, deviceIDs)
			if newGroup == nil {
				// Try to restore original
				m.fleet.CreateGroup(group.ID, group.Name, group.DeviceIDs)
				return messages.NewSaveError(m.groupID, fmt.Errorf("failed to update group"))
			}
			return messages.NewSaveResult(m.groupID)

		case GroupEditModeDelete:
			// Delete mode is handled by doDelete(), not createSaveCmd
			return messages.NewSaveError(nil, fmt.Errorf("delete mode should use doDelete()"))
		}

		return messages.NewSaveError(nil, fmt.Errorf("invalid mode"))
	}
}

func (m GroupEditModel) doDelete() (GroupEditModel, tea.Cmd) {
	if m.Saving {
		return m, nil
	}

	m.StartSave()

	return m, func() tea.Msg {
		if m.fleet == nil {
			return messages.NewSaveError(nil, fmt.Errorf("not connected to fleet"))
		}

		ok := m.fleet.DeleteGroup(m.groupID)
		if !ok {
			return messages.NewSaveError(m.groupID, fmt.Errorf("failed to delete group"))
		}

		return messages.NewSaveResult(m.groupID)
	}
}

// View renders the edit modal.
func (m GroupEditModel) View() string {
	if !m.Visible() {
		return ""
	}

	// Build title and footer based on mode
	var title, footer string
	switch m.mode {
	case GroupEditModeDelete:
		title = "Delete Group"
		if m.Saving {
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

	// Build content based on mode
	var content string
	if m.mode == GroupEditModeDelete {
		content = m.renderDeleteContent()
	} else {
		content = m.renderEditContent()
	}

	return m.RenderModal(title, content, footer)
}

func (m GroupEditModel) buildEditFooter() string {
	if m.Saving {
		return "Saving..."
	}
	if GroupEditField(m.Cursor) == GroupEditFieldDevices {
		return "j/k: Navigate | Space: Toggle | Ctrl+S/Enter: Save | Esc: Cancel"
	}
	return "Tab: Next field | Ctrl+S: Save | Esc: Cancel"
}

func (m GroupEditModel) renderDeleteContent() string {
	var content strings.Builder

	if m.original != nil {
		content.WriteString(m.Styles.Warning.Render("Are you sure you want to delete this group?"))
		content.WriteString("\n\n")
		content.WriteString(m.Styles.Label.Render("Name: "))
		content.WriteString(m.Styles.Value.Render(m.original.Name))
		content.WriteString("\n")
		content.WriteString(m.Styles.Label.Render("Devices: "))
		content.WriteString(m.Styles.Value.Render(fmt.Sprintf("%d", len(m.original.DeviceIDs))))
	}

	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n\n")
		content.WriteString(errStr)
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
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n\n")
		content.WriteString(errStr)
	}

	return content.String()
}

func (m GroupEditModel) renderField(field GroupEditField, label, input string) string {
	var selector, labelStr string

	if GroupEditField(m.Cursor) == field {
		selector = m.Styles.Selector.Render("> ")
		labelStr = m.Styles.LabelFocus.Render(label)
	} else {
		selector = "  "
		labelStr = m.Styles.Label.Render(label)
	}

	return selector + labelStr + " " + input
}

func (m GroupEditModel) renderDevicesField() string {
	var content strings.Builder

	// Label
	if GroupEditField(m.Cursor) == GroupEditFieldDevices {
		content.WriteString(m.Styles.Selector.Render("> "))
		content.WriteString(m.Styles.LabelFocus.Render("Devices:"))
	} else {
		content.WriteString("  ")
		content.WriteString(m.Styles.Label.Render("Devices:"))
	}

	// Selection count
	content.WriteString(" ")
	content.WriteString(m.Styles.Info.Render(fmt.Sprintf("(%d selected)", len(m.selectedIDs))))
	content.WriteString("\n")

	if len(m.allDevices) == 0 {
		content.WriteString("    ")
		content.WriteString(m.Styles.Info.Render("No devices available"))
		return content.String()
	}

	// Device list with scrolling using generic helper
	fieldFocused := GroupEditField(m.Cursor) == GroupEditFieldDevices
	content.WriteString(generics.RenderScrollableItems(m.allDevices, m.deviceScroller,
		func(device integrator.AccountDevice, _ int, scrollerCursor bool) string {
			isSelected := m.selectedIDs[device.DeviceID]
			isCursor := scrollerCursor && fieldFocused
			return m.renderDeviceLine(device, isSelected, isCursor)
		}))

	// Scroll indicator (with custom indentation for modal alignment)
	if m.deviceScroller.HasMore() || m.deviceScroller.HasPrevious() {
		content.WriteString("\n    ")
		content.WriteString(m.Styles.Info.Render(m.deviceScroller.ScrollInfo()))
	}

	return content.String()
}

func (m GroupEditModel) renderDeviceLine(device integrator.AccountDevice, isSelected, isCursor bool) string {
	var line strings.Builder

	// Indentation
	line.WriteString("    ")

	// Cursor indicator
	if isCursor {
		line.WriteString(m.Styles.Selector.Render("> "))
	} else {
		line.WriteString("  ")
	}

	// Checkbox
	if isSelected {
		line.WriteString(m.Styles.StatusOn.Render("[✓] "))
	} else {
		line.WriteString(m.Styles.Muted.Render("[ ] "))
	}

	// Device name
	name := device.Name
	if name == "" {
		name = device.DeviceID
	}
	name = output.Truncate(name, 25)

	if isCursor {
		line.WriteString(m.Styles.Selected.Render(name))
	} else {
		line.WriteString(m.Styles.Value.Render(name))
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

package fleet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// GroupsDeps holds the dependencies for the Groups component.
type GroupsDeps struct {
	Ctx context.Context
}

// Validate ensures all required dependencies are set.
func (d GroupsDeps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	return nil
}

// GroupsLoadedMsg signals that groups were loaded.
type GroupsLoadedMsg struct {
	Groups []*integrator.DeviceGroup
	Err    error
}

// GroupsModel displays and manages device groups.
type GroupsModel struct {
	ctx        context.Context
	fleet      *integrator.FleetManager
	groups     []*integrator.DeviceGroup
	scroller   *panel.Scroller
	loading    bool
	editing    bool
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int
	styles     GroupsStyles
	loader     loading.Model
	editModal  GroupEditModel
}

// GroupsStyles holds styles for the Groups component.
type GroupsStyles struct {
	Name     lipgloss.Style
	Count    lipgloss.Style
	Cursor   lipgloss.Style
	Muted    lipgloss.Style
	Error    lipgloss.Style
	Title    lipgloss.Style
	Selected lipgloss.Style
}

// DefaultGroupsStyles returns the default styles for the Groups component.
func DefaultGroupsStyles() GroupsStyles {
	colors := theme.GetSemanticColors()
	return GroupsStyles{
		Name: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		Count: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Cursor: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Selected: lipgloss.NewStyle().
			Background(colors.Highlight).
			Foreground(colors.Primary),
	}
}

// NewGroups creates a new Groups model.
func NewGroups(deps GroupsDeps) GroupsModel {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("fleet/groups component init", err)
		panic(fmt.Sprintf("fleet/groups: invalid deps: %v", err))
	}

	return GroupsModel{
		ctx:      deps.Ctx,
		scroller: panel.NewScroller(0, 10),
		styles:   DefaultGroupsStyles(),
		loader: loading.New(
			loading.WithMessage("Loading groups..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
		editModal: NewGroupEditModel(),
	}
}

// Init returns the initial command.
func (m GroupsModel) Init() tea.Cmd {
	return nil
}

// SetFleetManager sets the fleet manager.
func (m GroupsModel) SetFleetManager(fm *integrator.FleetManager) (GroupsModel, tea.Cmd) {
	m.fleet = fm
	if fm == nil {
		m.groups = nil
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.loadGroups())
}

func (m GroupsModel) loadGroups() tea.Cmd {
	return func() tea.Msg {
		if m.fleet == nil {
			return GroupsLoadedMsg{Err: fmt.Errorf("not connected to fleet")}
		}
		groups := m.fleet.ListGroups()
		return GroupsLoadedMsg{Groups: groups}
	}
}

// SetSize sets the component dimensions.
func (m GroupsModel) SetSize(width, height int) GroupsModel {
	m.width = width
	m.height = height
	visibleRows := height - 5 // Account for borders, title, stats
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
	// Update loader size for proper centering
	m.loader = m.loader.SetSize(width-4, height-4)
	return m
}

// SetFocused sets the focus state.
func (m GroupsModel) SetFocused(focused bool) GroupsModel {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m GroupsModel) SetPanelIndex(index int) GroupsModel {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m GroupsModel) Update(msg tea.Msg) (GroupsModel, tea.Cmd) {
	// Handle edit modal if visible
	if m.editing {
		return m.handleEditModalUpdate(msg)
	}

	// Forward tick messages to loader when loading
	if m.loading {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		// Continue processing GroupsLoadedMsg even during loading
		if _, ok := msg.(GroupsLoadedMsg); !ok {
			if cmd != nil {
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case GroupsLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.groups = msg.Groups
		m.err = nil
		m.scroller.SetItemCount(len(m.groups))
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m GroupsModel) handleEditModalUpdate(msg tea.Msg) (GroupsModel, tea.Cmd) {
	var cmd tea.Cmd
	m.editModal, cmd = m.editModal.Update(msg)

	// Check if modal was closed
	if !m.editModal.Visible() {
		m.editing = false
		// Refresh data after edit
		m.loading = true
		return m, tea.Batch(cmd, m.loader.Tick(), m.loadGroups())
	}

	// Handle save result message
	if saveMsg, ok := msg.(GroupEditSaveResultMsg); ok {
		if saveMsg.Err == nil {
			m.editing = false
			m.editModal = m.editModal.Hide()
			// Refresh data after successful save
			m.loading = true
			return m, tea.Batch(m.loader.Tick(), m.loadGroups(), func() tea.Msg {
				return GroupEditClosedMsg{Saved: true}
			})
		}
	}

	return m, cmd
}

// canModify returns true if we can perform group operations.
func (m GroupsModel) canModify() bool {
	return m.fleet != nil && !m.loading
}

// canActOnSelected returns true if we can act on the selected group.
func (m GroupsModel) canActOnSelected() bool {
	return m.canModify() && len(m.groups) > 0
}

func (m GroupsModel) handleKey(msg tea.KeyPressMsg) (GroupsModel, tea.Cmd) {
	key := msg.String()

	// Navigation keys
	switch key {
	case "j", keyconst.KeyDown:
		m.scroller.CursorDown()
		return m, nil
	case "k", keyconst.KeyUp:
		m.scroller.CursorUp()
		return m, nil
	case "g":
		m.scroller.CursorToStart()
		return m, nil
	case "G":
		m.scroller.CursorToEnd()
		return m, nil
	case "ctrl+d", keyconst.KeyPgDown:
		m.scroller.PageDown()
		return m, nil
	case "ctrl+u", keyconst.KeyPgUp:
		m.scroller.PageUp()
		return m, nil
	case "r":
		return m.handleRefresh()
	case "n":
		return m.handleCreate()
	case "e", "enter":
		return m.handleEdit()
	case "d":
		return m.handleDelete()
	case "o":
		return m.handleGroupOn()
	case "f":
		return m.handleGroupOff()
	}

	return m, nil
}

func (m GroupsModel) handleRefresh() (GroupsModel, tea.Cmd) {
	if m.loading {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.loadGroups())
}

func (m GroupsModel) handleCreate() (GroupsModel, tea.Cmd) {
	if !m.canModify() {
		return m, nil
	}
	m.editing = true
	m.editModal = m.editModal.SetSize(m.width, m.height)
	m.editModal = m.editModal.ShowCreate(m.fleet)
	return m, func() tea.Msg { return GroupEditOpenedMsg{} }
}

func (m GroupsModel) handleEdit() (GroupsModel, tea.Cmd) {
	if !m.canActOnSelected() {
		return m, nil
	}
	group := m.SelectedGroup()
	if group == nil {
		return m, nil
	}
	m.editing = true
	m.editModal = m.editModal.SetSize(m.width, m.height)
	m.editModal = m.editModal.ShowEdit(m.fleet, group)
	return m, func() tea.Msg { return GroupEditOpenedMsg{} }
}

func (m GroupsModel) handleDelete() (GroupsModel, tea.Cmd) {
	if !m.canActOnSelected() {
		return m, nil
	}
	group := m.SelectedGroup()
	if group == nil {
		return m, nil
	}
	m.editing = true
	m.editModal = m.editModal.SetSize(m.width, m.height)
	m.editModal = m.editModal.ShowDelete(m.fleet, group)
	return m, func() tea.Msg { return GroupEditOpenedMsg{} }
}

func (m GroupsModel) handleGroupOn() (GroupsModel, tea.Cmd) {
	if !m.canActOnSelected() {
		return m, nil
	}
	group := m.SelectedGroup()
	if group == nil {
		return m, nil
	}
	return m, m.sendGroupCommand(group.ID, true)
}

func (m GroupsModel) handleGroupOff() (GroupsModel, tea.Cmd) {
	if !m.canActOnSelected() {
		return m, nil
	}
	group := m.SelectedGroup()
	if group == nil {
		return m, nil
	}
	return m, m.sendGroupCommand(group.ID, false)
}

// GroupCommandResultMsg signals a group command completed.
type GroupCommandResultMsg struct {
	GroupID string
	On      bool
	Err     error
}

func (m GroupsModel) sendGroupCommand(groupID string, on bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		var results []integrator.BatchResult
		if on {
			results = m.fleet.GroupRelaysOn(ctx, groupID)
		} else {
			results = m.fleet.GroupRelaysOff(ctx, groupID)
		}

		// Check for errors
		for _, r := range results {
			if !r.Success {
				return GroupCommandResultMsg{GroupID: groupID, On: on, Err: errors.New(r.Error)}
			}
		}

		return GroupCommandResultMsg{GroupID: groupID, On: on}
	}
}

// View renders the Groups component.
func (m GroupsModel) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Device Groups").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused {
		r.SetFooter("n:new e:edit d:del o:on f:off r:refresh")
	}

	// Calculate content area for centering (accounting for panel borders)
	contentWidth := m.width - 4
	contentHeight := m.height - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	if m.fleet == nil {
		msg := m.styles.Muted.Render("Not connected to Shelly Cloud")
		r.SetContent(lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, msg))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.loader.View())
		return r.Render()
	}

	if m.err != nil {
		msg := m.styles.Error.Render("Error: " + m.err.Error())
		r.SetContent(lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, msg))
		return r.Render()
	}

	if len(m.groups) == 0 {
		msg := m.styles.Muted.Render("No device groups defined")
		r.SetContent(lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, msg))
		return r.Render()
	}

	var content strings.Builder

	// Stats line
	content.WriteString(m.styles.Muted.Render(fmt.Sprintf("%d groups", len(m.groups))))
	content.WriteString("\n\n")

	// Group list
	start, end := m.scroller.VisibleRange()
	for i := start; i < end; i++ {
		group := m.groups[i]
		isSelected := m.scroller.IsCursorAt(i)

		cursor := "  "
		if isSelected && m.focused {
			cursor = m.styles.Cursor.Render("> ")
		}

		deviceCount := len(group.DeviceIDs)
		countStr := fmt.Sprintf("(%d devices)", deviceCount)

		// Calculate available width for name
		// Fixed: cursor(2) + space(1) + countStr length
		available := output.ContentWidth(m.width, 4+3+len(countStr))
		name := output.Truncate(group.Name, max(available, 10))

		line := fmt.Sprintf("%s%s %s",
			cursor,
			m.styles.Name.Render(name),
			m.styles.Count.Render(countStr),
		)

		if isSelected && m.focused {
			line = m.styles.Selected.Render(line)
		}

		content.WriteString(line)
		content.WriteString("\n")
	}

	// Scroll indicator
	content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))

	r.SetContent(content.String())
	return r.Render()
}

// SelectedGroup returns the currently selected group.
func (m GroupsModel) SelectedGroup() *integrator.DeviceGroup {
	cursor := m.scroller.Cursor()
	if cursor >= 0 && cursor < len(m.groups) {
		return m.groups[cursor]
	}
	return nil
}

// Cursor returns the current cursor position.
func (m GroupsModel) Cursor() int {
	return m.scroller.Cursor()
}

// Groups returns all groups.
func (m GroupsModel) Groups() []*integrator.DeviceGroup {
	return m.groups
}

// GroupCount returns the number of groups.
func (m GroupsModel) GroupCount() int {
	return len(m.groups)
}

// Loading returns whether groups are being loaded.
func (m GroupsModel) Loading() bool {
	return m.loading
}

// Error returns any error that occurred.
func (m GroupsModel) Error() error {
	return m.err
}

// Refresh triggers a group reload.
func (m GroupsModel) Refresh() (GroupsModel, tea.Cmd) {
	if m.fleet == nil {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.loadGroups())
}

// FooterText returns keybinding hints for the footer.
func (m GroupsModel) FooterText() string {
	return "n:new e:edit d:del o:on f:off r:refresh"
}

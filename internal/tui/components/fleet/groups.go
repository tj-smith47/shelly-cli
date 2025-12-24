package fleet

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/theme"
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
	ctx     context.Context
	fleet   *integrator.FleetManager
	groups  []*integrator.DeviceGroup
	cursor  int
	loading bool
	err     error
	width   int
	height  int
	focused bool
	styles  GroupsStyles
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
			Foreground(colors.Background),
	}
}

// NewGroups creates a new Groups model.
func NewGroups(deps GroupsDeps) GroupsModel {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("fleet/groups: invalid deps: %v", err))
	}

	return GroupsModel{
		ctx:    deps.Ctx,
		styles: DefaultGroupsStyles(),
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
	return m, m.loadGroups()
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
	return m
}

// SetFocused sets the focus state.
func (m GroupsModel) SetFocused(focused bool) GroupsModel {
	m.focused = focused
	return m
}

// Update handles messages.
func (m GroupsModel) Update(msg tea.Msg) (GroupsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case GroupsLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.groups = msg.Groups
		m.err = nil
		if m.cursor >= len(m.groups) {
			m.cursor = max(0, len(m.groups)-1)
		}
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m GroupsModel) handleKey(msg tea.KeyPressMsg) (GroupsModel, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(m.groups)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "r":
		if !m.loading {
			m.loading = true
			return m, m.loadGroups()
		}
	}

	return m, nil
}

// View renders the Groups component.
func (m GroupsModel) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Device Groups").
		SetFocused(m.focused)

	// Add footer with keybindings when focused
	if m.focused {
		r.SetFooter("j/k:nav r:refresh")
	}

	if m.fleet == nil {
		r.SetContent(m.styles.Muted.Render("Not connected to Shelly Cloud"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.styles.Muted.Render("Loading groups..."))
		return r.Render()
	}

	if m.err != nil {
		r.SetContent(m.styles.Error.Render("Error: " + m.err.Error()))
		return r.Render()
	}

	if len(m.groups) == 0 {
		r.SetContent(m.styles.Muted.Render("No device groups defined"))
		return r.Render()
	}

	var content strings.Builder

	// Stats line
	content.WriteString(m.styles.Muted.Render(fmt.Sprintf("%d groups", len(m.groups))))
	content.WriteString("\n\n")

	// Group list
	visibleHeight := m.height - 5
	if visibleHeight < 1 {
		visibleHeight = 5
	}

	startIdx := 0
	if m.cursor >= visibleHeight {
		startIdx = m.cursor - visibleHeight + 1
	}
	endIdx := startIdx + visibleHeight
	if endIdx > len(m.groups) {
		endIdx = len(m.groups)
	}

	for i := startIdx; i < endIdx; i++ {
		group := m.groups[i]

		cursor := "  "
		if i == m.cursor && m.focused {
			cursor = m.styles.Cursor.Render("> ")
		}

		name := group.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		deviceCount := len(group.DeviceIDs)
		countStr := fmt.Sprintf("(%d devices)", deviceCount)

		line := fmt.Sprintf("%s%s %s",
			cursor,
			m.styles.Name.Render(name),
			m.styles.Count.Render(countStr),
		)

		if i == m.cursor && m.focused {
			line = m.styles.Selected.Render(line)
		}

		content.WriteString(line)
		content.WriteString("\n")
	}

	r.SetContent(content.String())
	return r.Render()
}

// SelectedGroup returns the currently selected group.
func (m GroupsModel) SelectedGroup() *integrator.DeviceGroup {
	if m.cursor >= 0 && m.cursor < len(m.groups) {
		return m.groups[m.cursor]
	}
	return nil
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
	return m, m.loadGroups()
}

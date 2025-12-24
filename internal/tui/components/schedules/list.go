// Package schedules provides TUI components for managing device schedules.
package schedules

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Schedule represents a schedule job on a device.
type Schedule struct {
	ID       int
	Enable   bool
	Timespec string
	Calls    []shelly.ScheduleCall
}

// ListDeps holds the dependencies for the schedules list component.
type ListDeps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d ListDeps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	return nil
}

// LoadedMsg signals that schedules were loaded.
type LoadedMsg struct {
	Schedules []Schedule
	Err       error
}

// ActionMsg signals a schedule action result.
type ActionMsg struct {
	Action     string // "enable", "disable", "delete"
	ScheduleID int
	Err        error
}

// SelectScheduleMsg signals that a schedule was selected for viewing.
type SelectScheduleMsg struct {
	Schedule Schedule
}

// ListModel displays schedules for a device.
type ListModel struct {
	ctx       context.Context
	svc       *shelly.Service
	device    string
	schedules []Schedule
	cursor    int
	scroll    int
	loading   bool
	err       error
	width     int
	height    int
	focused   bool
	styles    ListStyles
}

// ListStyles holds styles for the list component.
type ListStyles struct {
	Enabled  lipgloss.Style
	Disabled lipgloss.Style
	Method   lipgloss.Style
	Timespec lipgloss.Style
	Selected lipgloss.Style
	Error    lipgloss.Style
	Muted    lipgloss.Style
}

// DefaultListStyles returns the default styles for the schedule list.
func DefaultListStyles() ListStyles {
	colors := theme.GetSemanticColors()
	return ListStyles{
		Enabled: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Method: lipgloss.NewStyle().
			Foreground(colors.Info),
		Timespec: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// NewList creates a new schedules list model.
func NewList(deps ListDeps) ListModel {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("schedules: invalid deps: %v", err))
	}

	return ListModel{
		ctx:     deps.Ctx,
		svc:     deps.Svc,
		loading: false,
		styles:  DefaultListStyles(),
	}
}

// Init returns the initial command.
func (m ListModel) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to list schedules for and triggers a fetch.
func (m ListModel) SetDevice(device string) (ListModel, tea.Cmd) {
	m.device = device
	m.schedules = nil
	m.cursor = 0
	m.scroll = 0
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, m.fetchSchedules()
}

// fetchSchedules creates a command to fetch schedules from the device.
func (m ListModel) fetchSchedules() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		jobs, err := m.svc.ListSchedules(ctx, m.device)
		if err != nil {
			return LoadedMsg{Err: err}
		}

		result := make([]Schedule, len(jobs))
		for i, job := range jobs {
			result[i] = Schedule{
				ID:       job.ID,
				Enable:   job.Enable,
				Timespec: job.Timespec,
				Calls:    job.Calls,
			}
		}

		return LoadedMsg{Schedules: result}
	}
}

// SetSize sets the component dimensions.
func (m ListModel) SetSize(width, height int) ListModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focus state.
func (m ListModel) SetFocused(focused bool) ListModel {
	m.focused = focused
	return m
}

// Update handles messages.
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.schedules = msg.Schedules
		m.cursor = 0
		m.scroll = 0
		return m, nil

	case ActionMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, m.fetchSchedules()

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m ListModel) handleKey(msg tea.KeyPressMsg) (ListModel, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m = m.cursorDown()
	case "k", "up":
		m = m.cursorUp()
	case "g":
		m.cursor = 0
		m.scroll = 0
	case "G":
		m = m.cursorToEnd()
	case "enter":
		return m, m.selectSchedule()
	case "e":
		return m, m.enableSchedule()
	case "E":
		return m, m.disableSchedule()
	case "d":
		return m, m.deleteSchedule()
	case "r":
		m.loading = true
		return m, m.fetchSchedules()
	}

	return m, nil
}

func (m ListModel) cursorDown() ListModel {
	if m.cursor < len(m.schedules)-1 {
		m.cursor++
		m = m.ensureVisible()
	}
	return m
}

func (m ListModel) cursorUp() ListModel {
	if m.cursor > 0 {
		m.cursor--
		m = m.ensureVisible()
	}
	return m
}

func (m ListModel) cursorToEnd() ListModel {
	if len(m.schedules) > 0 {
		m.cursor = len(m.schedules) - 1
		m = m.ensureVisible()
	}
	return m
}

func (m ListModel) ensureVisible() ListModel {
	visible := m.visibleRows()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	} else if m.cursor >= m.scroll+visible {
		m.scroll = m.cursor - visible + 1
	}
	return m
}

func (m ListModel) visibleRows() int {
	rows := m.height - 4
	if rows < 1 {
		return 1
	}
	return rows
}

func (m ListModel) selectSchedule() tea.Cmd {
	if len(m.schedules) == 0 || m.cursor >= len(m.schedules) {
		return nil
	}
	schedule := m.schedules[m.cursor]
	return func() tea.Msg {
		return SelectScheduleMsg{Schedule: schedule}
	}
}

func (m ListModel) enableSchedule() tea.Cmd {
	if len(m.schedules) == 0 || m.cursor >= len(m.schedules) {
		return nil
	}
	schedule := m.schedules[m.cursor]
	if schedule.Enable {
		return nil // Already enabled
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.EnableSchedule(ctx, m.device, schedule.ID)
		return ActionMsg{Action: "enable", ScheduleID: schedule.ID, Err: err}
	}
}

func (m ListModel) disableSchedule() tea.Cmd {
	if len(m.schedules) == 0 || m.cursor >= len(m.schedules) {
		return nil
	}
	schedule := m.schedules[m.cursor]
	if !schedule.Enable {
		return nil // Already disabled
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.DisableSchedule(ctx, m.device, schedule.ID)
		return ActionMsg{Action: "disable", ScheduleID: schedule.ID, Err: err}
	}
}

func (m ListModel) deleteSchedule() tea.Cmd {
	if len(m.schedules) == 0 || m.cursor >= len(m.schedules) {
		return nil
	}
	schedule := m.schedules[m.cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.DeleteSchedule(ctx, m.device, schedule.ID)
		return ActionMsg{Action: "delete", ScheduleID: schedule.ID, Err: err}
	}
}

// View renders the schedules list.
func (m ListModel) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Schedules").
		SetFocused(m.focused)

	if m.device == "" {
		r.SetContent(m.styles.Muted.Render("No device selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.styles.Muted.Render("Loading schedules..."))
		return r.Render()
	}

	if m.err != nil {
		r.SetContent(m.styles.Error.Render("Error: " + m.err.Error()))
		return r.Render()
	}

	if len(m.schedules) == 0 {
		r.SetContent(m.styles.Muted.Render("No schedules on device"))
		return r.Render()
	}

	var content strings.Builder
	visible := m.visibleRows()
	endIdx := m.scroll + visible
	if endIdx > len(m.schedules) {
		endIdx = len(m.schedules)
	}

	for i := m.scroll; i < endIdx; i++ {
		schedule := m.schedules[i]
		isSelected := i == m.cursor

		line := m.renderScheduleLine(schedule, isSelected)
		content.WriteString(line)
		if i < endIdx-1 {
			content.WriteString("\n")
		}
	}

	if len(m.schedules) > visible {
		content.WriteString(m.styles.Muted.Render(
			fmt.Sprintf("\n[%d/%d]", m.cursor+1, len(m.schedules)),
		))
	}

	r.SetContent(content.String())
	return r.Render()
}

func (m ListModel) renderScheduleLine(schedule Schedule, isSelected bool) string {
	// Status icon
	var icon string
	if schedule.Enable {
		icon = m.styles.Enabled.Render("●")
	} else {
		icon = m.styles.Disabled.Render("○")
	}

	// Selection indicator
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Timespec (truncate if too long)
	timespec := schedule.Timespec
	if len(timespec) > 20 {
		timespec = timespec[:17] + "..."
	}
	timespecStr := m.styles.Timespec.Render(timespec)

	// Primary method
	methodStr := ""
	if len(schedule.Calls) > 0 {
		method := schedule.Calls[0].Method
		if len(schedule.Calls) > 1 {
			method = fmt.Sprintf("%s +%d", method, len(schedule.Calls)-1)
		}
		methodStr = m.styles.Method.Render(method)
	}

	line := fmt.Sprintf("%s%s %s %s", selector, icon, timespecStr, methodStr)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

// SelectedSchedule returns the currently selected schedule, if any.
func (m ListModel) SelectedSchedule() *Schedule {
	if len(m.schedules) == 0 || m.cursor >= len(m.schedules) {
		return nil
	}
	return &m.schedules[m.cursor]
}

// ScheduleCount returns the number of schedules.
func (m ListModel) ScheduleCount() int {
	return len(m.schedules)
}

// Device returns the current device address.
func (m ListModel) Device() string {
	return m.device
}

// Loading returns whether the component is loading.
func (m ListModel) Loading() bool {
	return m.loading
}

// Error returns any error that occurred.
func (m ListModel) Error() error {
	return m.err
}

// Refresh triggers a refresh of the schedule list.
func (m ListModel) Refresh() (ListModel, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, m.fetchSchedules()
}

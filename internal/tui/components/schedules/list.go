// Package schedules provides TUI components for managing device schedules.
package schedules

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Schedule represents a schedule job on a device.
type Schedule struct {
	ID       int
	Enable   bool
	Timespec string
	Calls    []automation.ScheduleCall
}

// ListDeps holds the dependencies for the schedules list component.
type ListDeps struct {
	Ctx context.Context
	Svc *automation.Service
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

// CreateScheduleMsg signals that a new schedule should be created.
type CreateScheduleMsg struct {
	Device string
}

// ListModel displays schedules for a device.
type ListModel struct {
	ctx        context.Context
	svc        *automation.Service
	device     string
	schedules  []Schedule
	scroller   *panel.Scroller
	loading    bool
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int // 1-based panel index for Shift+N hotkey hint
	styles     ListStyles
	loader     loading.Model
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
		iostreams.DebugErr("schedules list component init", err)
		panic(fmt.Sprintf("schedules: invalid deps: %v", err))
	}

	return ListModel{
		ctx:      deps.Ctx,
		svc:      deps.Svc,
		scroller: panel.NewScroller(0, 10),
		loading:  false,
		styles:   DefaultListStyles(),
		loader: loading.New(
			loading.WithMessage("Loading schedules..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
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
	m.scroller.SetItemCount(0)
	m.scroller.CursorToStart()
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchSchedules())
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
	visibleRows := height - 4
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
	// Update loader size for proper centering
	m.loader = m.loader.SetSize(width-4, height-4)
	return m
}

// SetFocused sets the focus state.
func (m ListModel) SetFocused(focused bool) ListModel {
	m.focused = focused
	return m
}

// SetPanelIndex sets the 1-based panel index for Shift+N hotkey hint.
func (m ListModel) SetPanelIndex(index int) ListModel {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	// Forward tick messages to loader when loading
	if m.loading {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		// Continue processing LoadedMsg/ActionMsg even during loading
		switch msg.(type) {
		case LoadedMsg, ActionMsg:
			// Pass through to main switch below
		default:
			if cmd != nil {
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case LoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.schedules = msg.Schedules
		m.scroller.SetItemCount(len(m.schedules))
		m.scroller.CursorToStart()
		return m, nil

	case ActionMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, tea.Batch(m.loader.Tick(), m.fetchSchedules())

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m ListModel) handleKey(msg tea.KeyPressMsg) (ListModel, tea.Cmd) {
	// Handle navigation keys first
	if keys.HandleScrollNavigation(msg.String(), m.scroller) {
		return m, nil
	}

	// Handle action keys
	switch msg.String() {
	case "enter", "e":
		// Edit schedule (open in editor)
		return m, m.selectSchedule()
	case "t":
		// Toggle enable/disable
		return m, m.toggleSchedule()
	case "d":
		// Delete schedule
		return m, m.deleteSchedule()
	case "n":
		// New schedule - will be handled by parent
		return m, m.createSchedule()
	case "R":
		// Refresh list
		m.loading = true
		return m, tea.Batch(m.loader.Tick(), m.fetchSchedules())
	}

	return m, nil
}

func (m ListModel) selectSchedule() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.schedules) == 0 || cursor >= len(m.schedules) {
		return nil
	}
	schedule := m.schedules[cursor]
	return func() tea.Msg {
		return SelectScheduleMsg{Schedule: schedule}
	}
}

func (m ListModel) deleteSchedule() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.schedules) == 0 || cursor >= len(m.schedules) {
		return nil
	}
	schedule := m.schedules[cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.DeleteSchedule(ctx, m.device, schedule.ID)
		return ActionMsg{Action: "delete", ScheduleID: schedule.ID, Err: err}
	}
}

func (m ListModel) toggleSchedule() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.schedules) == 0 || cursor >= len(m.schedules) {
		return nil
	}
	schedule := m.schedules[cursor]

	// Toggle based on current state
	if schedule.Enable {
		return func() tea.Msg {
			ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
			defer cancel()

			err := m.svc.DisableSchedule(ctx, m.device, schedule.ID)
			return ActionMsg{Action: "disable", ScheduleID: schedule.ID, Err: err}
		}
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.EnableSchedule(ctx, m.device, schedule.ID)
		return ActionMsg{Action: "enable", ScheduleID: schedule.ID, Err: err}
	}
}

func (m ListModel) createSchedule() tea.Cmd {
	if m.device == "" {
		return nil
	}
	return func() tea.Msg {
		return CreateScheduleMsg{Device: m.device}
	}
}

// View renders the schedules list.
func (m ListModel) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Schedules").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused && m.device != "" && len(m.schedules) > 0 {
		r.SetFooter("e:edit t:toggle d:del n:new")
	}

	if m.device == "" {
		r.SetContent(m.styles.Muted.Render("No device selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.loader.View())
		return r.Render()
	}

	if m.err != nil {
		errMsg := m.err.Error()
		// Detect Gen1 or unsupported device errors and show a friendly message
		if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "unknown method") ||
			strings.Contains(errMsg, "not found") {
			r.SetContent(m.styles.Muted.Render("Schedules not supported on this device"))
		} else {
			r.SetContent(m.styles.Error.Render("Error: " + errMsg))
		}
		return r.Render()
	}

	if len(m.schedules) == 0 {
		r.SetContent(m.styles.Muted.Render("No schedules on device"))
		return r.Render()
	}

	var content strings.Builder

	start, end := m.scroller.VisibleRange()
	for i := start; i < end; i++ {
		schedule := m.schedules[i]
		isSelected := m.scroller.IsCursorAt(i)

		line := m.renderScheduleLine(schedule, isSelected)
		content.WriteString(line)
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))

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
	timespec := output.Truncate(schedule.Timespec, 20)
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
	cursor := m.scroller.Cursor()
	if len(m.schedules) == 0 || cursor >= len(m.schedules) {
		return nil
	}
	return &m.schedules[cursor]
}

// Cursor returns the current cursor position.
func (m ListModel) Cursor() int {
	return m.scroller.Cursor()
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
	return m, tea.Batch(m.loader.Tick(), m.fetchSchedules())
}

// FooterText returns keybinding hints for the footer.
func (m ListModel) FooterText() string {
	return "j/k:scroll g/G:top/bottom space:toggle enter:edit"
}

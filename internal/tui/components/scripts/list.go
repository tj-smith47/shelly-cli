// Package scripts provides TUI components for managing device scripts.
package scripts

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Script represents a script on a device.
type Script struct {
	ID      int
	Name    string
	Enabled bool
	Running bool
}

// ListDeps holds the dependencies for the scripts list component.
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

// LoadedMsg signals that scripts were loaded.
type LoadedMsg struct {
	Scripts []Script
	Err     error
}

// ActionMsg signals a script action result.
type ActionMsg struct {
	Action   string // "start", "stop", "delete"
	ScriptID int
	Err      error
}

// SelectScriptMsg signals that a script was selected for viewing.
type SelectScriptMsg struct {
	Script Script
}

// EditScriptMsg signals that a script should be edited in external editor.
type EditScriptMsg struct {
	Script Script
}

// CreateScriptMsg signals that a new script should be created.
type CreateScriptMsg struct {
	Device string
}

// ListModel displays scripts for a device.
type ListModel struct {
	ctx        context.Context
	svc        *shelly.Service
	device     string
	scripts    []Script
	scroller   *panel.Scroller
	loading    bool
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int // 1-based panel index for Shift+N hotkey hint
	styles     ListStyles
}

// ListStyles holds styles for the list component.
type ListStyles struct {
	Running  lipgloss.Style
	Stopped  lipgloss.Style
	Disabled lipgloss.Style
	Name     lipgloss.Style
	Selected lipgloss.Style
	Status   lipgloss.Style
	Error    lipgloss.Style
	Muted    lipgloss.Style
}

// DefaultListStyles returns the default styles for the script list.
func DefaultListStyles() ListStyles {
	colors := theme.GetSemanticColors()
	return ListStyles{
		Running: lipgloss.NewStyle().
			Foreground(colors.Online),
		Stopped: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Name: lipgloss.NewStyle().
			Foreground(colors.Text),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Status: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// NewList creates a new scripts list model.
func NewList(deps ListDeps) ListModel {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("scripts: invalid deps: %v", err))
	}

	return ListModel{
		ctx:      deps.Ctx,
		svc:      deps.Svc,
		scroller: panel.NewScroller(0, 1),
		loading:  false,
		styles:   DefaultListStyles(),
	}
}

// Init returns the initial command.
func (m ListModel) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to list scripts for and triggers a fetch.
func (m ListModel) SetDevice(device string) (ListModel, tea.Cmd) {
	m.device = device
	m.scripts = nil
	m.scroller.SetItemCount(0)
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, m.fetchScripts()
}

// fetchScripts creates a command to fetch scripts from the device.
func (m ListModel) fetchScripts() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		scripts, err := m.svc.ListScripts(ctx, m.device)
		if err != nil {
			return LoadedMsg{Err: err}
		}

		result := make([]Script, len(scripts))
		for i, s := range scripts {
			result[i] = Script{
				ID:      s.ID,
				Name:    s.Name,
				Enabled: s.Enable,
				Running: s.Running,
			}
		}

		return LoadedMsg{Scripts: result}
	}
}

// SetSize sets the component dimensions.
func (m ListModel) SetSize(width, height int) ListModel {
	m.width = width
	m.height = height
	// Calculate visible rows: height - borders (2) - title (1) - footer (1)
	visibleRows := height - 4
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
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
	switch msg := msg.(type) {
	case LoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.scripts = msg.Scripts
		m.scroller.SetItemCount(len(m.scripts))
		m.scroller.CursorToStart()
		return m, nil

	case ActionMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Refresh scripts after action
		m.loading = true
		return m, m.fetchScripts()

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
		m.scroller.CursorDown()
	case "k", "up":
		m.scroller.CursorUp()
	case "g":
		m.scroller.CursorToStart()
	case "G":
		m.scroller.CursorToEnd()
	case "ctrl+d", "pgdown":
		m.scroller.PageDown()
	case "ctrl+u", "pgup":
		m.scroller.PageUp()
	case "enter":
		// View script (open in viewer)
		return m, m.selectScript()
	case "e":
		// Edit script (open in external editor)
		return m, m.editScript()
	case "r":
		// Run/start script
		return m, m.startScript()
	case "s":
		// Stop script
		return m, m.stopScript()
	case "d":
		// Delete script
		return m, m.deleteScript()
	case "n":
		// New script - will be handled by parent
		return m, m.createScript()
	case "R":
		// Refresh list
		m.loading = true
		return m, m.fetchScripts()
	}

	return m, nil
}

func (m ListModel) selectScript() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	script := m.scripts[cursor]
	return func() tea.Msg {
		return SelectScriptMsg{Script: script}
	}
}

func (m ListModel) editScript() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	script := m.scripts[cursor]
	return func() tea.Msg {
		return EditScriptMsg{Script: script}
	}
}

func (m ListModel) startScript() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	script := m.scripts[cursor]
	if script.Running {
		return nil // Already running
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.StartScript(ctx, m.device, script.ID)
		return ActionMsg{Action: "start", ScriptID: script.ID, Err: err}
	}
}

func (m ListModel) stopScript() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	script := m.scripts[cursor]
	if !script.Running {
		return nil // Not running
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.StopScript(ctx, m.device, script.ID)
		return ActionMsg{Action: "stop", ScriptID: script.ID, Err: err}
	}
}

func (m ListModel) deleteScript() tea.Cmd {
	cursor := m.scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	script := m.scripts[cursor]

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.DeleteScript(ctx, m.device, script.ID)
		return ActionMsg{Action: "delete", ScriptID: script.ID, Err: err}
	}
}

func (m ListModel) createScript() tea.Cmd {
	if m.device == "" {
		return nil
	}
	return func() tea.Msg {
		return CreateScriptMsg{Device: m.device}
	}
}

// View renders the scripts list.
func (m ListModel) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Scripts").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused && m.device != "" && len(m.scripts) > 0 {
		r.SetFooter("e:edit r:run s:stop d:del n:new")
	}

	if m.device == "" {
		r.SetContent(m.styles.Muted.Render("No device selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.styles.Muted.Render("Loading scripts..."))
		return r.Render()
	}

	if m.err != nil {
		errMsg := m.err.Error()
		// Detect Gen1 or unsupported device errors and show a friendly message
		if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "unknown method") ||
			strings.Contains(errMsg, "not found") {
			r.SetContent(m.styles.Muted.Render("Scripts not supported on this device"))
		} else {
			r.SetContent(m.styles.Error.Render("Error: " + errMsg))
		}
		return r.Render()
	}

	if len(m.scripts) == 0 {
		r.SetContent(m.styles.Muted.Render("No scripts on device"))
		return r.Render()
	}

	var content strings.Builder
	start, end := m.scroller.VisibleRange()

	for i := start; i < end; i++ {
		script := m.scripts[i]
		isSelected := m.scroller.IsCursorAt(i)

		line := m.renderScriptLine(script, isSelected)
		content.WriteString(line)
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Add scroll indicator if needed
	if m.scroller.HasMore() || m.scroller.HasPrevious() {
		content.WriteString(m.styles.Muted.Render("\n" + m.scroller.ScrollInfo()))
	}

	r.SetContent(content.String())
	return r.Render()
}

func (m ListModel) renderScriptLine(script Script, isSelected bool) string {
	// Status icon
	var icon, status string
	switch {
	case !script.Enabled:
		icon = m.styles.Disabled.Render("-")
		status = m.styles.Status.Render("(disabled)")
	case script.Running:
		icon = m.styles.Running.Render("●")
		status = m.styles.Status.Render("(running)")
	default:
		icon = m.styles.Stopped.Render("○")
		status = m.styles.Status.Render("(stopped)")
	}

	// Name
	name := script.Name
	if name == "" {
		name = fmt.Sprintf("script_%d", script.ID)
	}

	// Selection indicator
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// Build line
	line := fmt.Sprintf("%s%s %s %s", selector, icon, name, status)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

// SelectedScript returns the currently selected script, if any.
func (m ListModel) SelectedScript() *Script {
	cursor := m.scroller.Cursor()
	if len(m.scripts) == 0 || cursor >= len(m.scripts) {
		return nil
	}
	return &m.scripts[cursor]
}

// ScriptCount returns the number of scripts.
func (m ListModel) ScriptCount() int {
	return len(m.scripts)
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

// Refresh triggers a refresh of the script list.
func (m ListModel) Refresh() (ListModel, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, m.fetchScripts()
}

// FooterText returns keybinding hints for the footer.
func (m ListModel) FooterText() string {
	return "j/k:scroll g/G:top/bottom space:toggle enter:edit"
}

// Package alerts provides alert management for the TUI.
package alerts

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// AlertItem represents an alert with display state.
type AlertItem struct {
	config.Alert
	LastTriggered *time.Time
	LastValue     string
	IsTriggered   bool
}

// actionNotify is the default alert action.
const actionNotify = "notify"

// LoadedMsg is sent when alerts are loaded.
type LoadedMsg struct {
	Alerts []AlertItem
	Err    error
}

// AlertToggleMsg is sent when an alert's enabled state is toggled.
type AlertToggleMsg struct {
	Name    string
	Enabled bool
}

// AlertDeleteMsg is sent to request alert deletion (with confirmation).
type AlertDeleteMsg struct {
	Name string
}

// AlertSnoozeMsg is sent to snooze an alert.
type AlertSnoozeMsg struct {
	Name     string
	Duration time.Duration
}

// AlertTestMsg is sent to test an alert.
type AlertTestMsg struct {
	Name string
}

// AlertCreateMsg is sent to request alert creation dialog.
type AlertCreateMsg struct{}

// AlertEditMsg is sent to request alert edit dialog.
type AlertEditMsg struct {
	Name string
}

// AlertActionResultMsg is sent when an alert action completes.
type AlertActionResultMsg struct {
	Action string // "toggle", "delete", "snooze", "test", "create"
	Name   string
	Err    error
}

// Deps holds dependencies for the alerts component.
type Deps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Model holds the alerts panel state.
type Model struct {
	helpers.Sizable
	ctx    context.Context
	svc    *shelly.Service
	styles Styles

	// State
	alerts   []AlertItem
	cursor   int
	loading  bool
	err      error
	focused  bool
	panelIdx int
}

// Styles for the alerts component.
type Styles struct {
	Header    lipgloss.Style
	AlertName lipgloss.Style
	Condition lipgloss.Style
	Device    lipgloss.Style
	Enabled   lipgloss.Style
	Disabled  lipgloss.Style
	Triggered lipgloss.Style
	Snoozed   lipgloss.Style
	Muted     lipgloss.Style
	Selected  lipgloss.Style
}

// DefaultStyles returns default styles for alerts.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Header: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		AlertName: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		Condition: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Device: lipgloss.NewStyle().
			Foreground(colors.Primary),
		Enabled: lipgloss.NewStyle().
			Foreground(colors.Success),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Triggered: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
		Snoozed: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Italic(true),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground),
	}
}

// New creates a new alerts component.
func New(deps Deps) Model {
	m := Model{
		Sizable: helpers.NewSizableLoaderOnly(),
		ctx:     deps.Ctx,
		svc:     deps.Svc,
		styles:  DefaultStyles(),
		loading: true,
	}
	m.Loader = m.Loader.SetMessage("Loading alerts...")
	return m
}

// Init initializes the alerts component.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.Loader.Tick(),
		m.loadAlerts(),
	)
}

// Update handles messages for the alerts component.
func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m.handleMessage(msg)
}

func (m *Model) handleMessage(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return *m, nil
		}
		m.alerts = msg.Alerts
		return *m, nil

	case AlertActionResultMsg:
		// Reload alerts after action
		if msg.Err == nil {
			return *m, m.loadAlerts()
		}
		return *m, nil

	case messages.NavigationMsg:
		return m.handleNavigationMsg(msg)
	case messages.ToggleEnableRequestMsg, messages.NewRequestMsg, messages.EditRequestMsg,
		messages.DeleteRequestMsg, messages.SnoozeRequestMsg, messages.TestRequestMsg,
		messages.RefreshRequestMsg:
		return m.handleActionMsg(msg)
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	// Update loader while loading
	if m.loading {
		var cmd tea.Cmd
		m.Loader, cmd = m.Loader.Update(msg)
		return *m, cmd
	}

	return *m, nil
}

func (m *Model) handleNavigationMsg(msg messages.NavigationMsg) (Model, tea.Cmd) {
	if !m.focused {
		return *m, nil
	}
	return m.handleNavigation(msg)
}

func (m *Model) handleActionMsg(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return *m, nil
	}
	switch msg := msg.(type) {
	case messages.ToggleEnableRequestMsg:
		return *m, m.handleToggle()
	case messages.NewRequestMsg:
		return *m, func() tea.Msg { return AlertCreateMsg{} }
	case messages.EditRequestMsg:
		return *m, m.handleEdit()
	case messages.DeleteRequestMsg:
		return *m, m.handleDelete()
	case messages.SnoozeRequestMsg:
		return m.handleSnoozeMsg(msg)
	case messages.TestRequestMsg:
		return *m, m.handleTest()
	case messages.RefreshRequestMsg:
		m.loading = true
		return *m, tea.Batch(m.Loader.Tick(), m.loadAlerts())
	}
	return *m, nil
}

func (m *Model) handleNavigation(msg messages.NavigationMsg) (Model, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		m.cursorUp()
	case messages.NavDown:
		m.cursorDown()
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this component
	}
	return *m, nil
}

func (m *Model) handleSnoozeMsg(msg messages.SnoozeRequestMsg) (Model, tea.Cmd) {
	var duration time.Duration
	switch msg.Duration {
	case "1h":
		duration = 1 * time.Hour
	case "24h":
		duration = 24 * time.Hour
	default:
		return *m, nil
	}
	return *m, m.handleSnooze(duration)
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	// Component-specific keys not covered by action messages
	// (none currently - all keys migrated to action messages)
	_ = msg
	if !m.focused {
		return *m, nil
	}
	return *m, nil
}

func (m *Model) cursorDown() {
	if m.cursor < len(m.alerts)-1 {
		m.cursor++
	}
}

func (m *Model) cursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *Model) hasSelection() bool {
	return len(m.alerts) > 0 && m.cursor < len(m.alerts)
}

func (m *Model) handleToggle() tea.Cmd {
	if !m.hasSelection() {
		return nil
	}
	alert := m.alerts[m.cursor]
	return m.toggleAlert(alert.Name, !alert.Enabled)
}

func (m *Model) handleEdit() tea.Cmd {
	if !m.hasSelection() {
		return nil
	}
	name := m.alerts[m.cursor].Name
	return func() tea.Msg { return AlertEditMsg{Name: name} }
}

func (m *Model) handleDelete() tea.Cmd {
	if !m.hasSelection() {
		return nil
	}
	name := m.alerts[m.cursor].Name
	return func() tea.Msg { return AlertDeleteMsg{Name: name} }
}

func (m *Model) handleSnooze(duration time.Duration) tea.Cmd {
	if !m.hasSelection() {
		return nil
	}
	return m.snoozeAlert(m.alerts[m.cursor].Name, duration)
}

func (m *Model) handleTest() tea.Cmd {
	if !m.hasSelection() {
		return nil
	}
	name := m.alerts[m.cursor].Name
	return func() tea.Msg { return AlertTestMsg{Name: name} }
}

func (m *Model) loadAlerts() tea.Cmd {
	return func() tea.Msg {
		alertsMap := config.ListAlerts()

		items := make([]AlertItem, 0, len(alertsMap))
		for _, alert := range alertsMap {
			items = append(items, AlertItem{Alert: alert})
		}

		// Sort by name
		sort.Slice(items, func(i, j int) bool {
			return items[i].Name < items[j].Name
		})

		return LoadedMsg{Alerts: items}
	}
}

func (m *Model) toggleAlert(name string, enabled bool) tea.Cmd {
	return func() tea.Msg {
		if err := config.UpdateAlert(name, &enabled, ""); err != nil {
			return AlertActionResultMsg{Action: "toggle", Name: name, Err: err}
		}
		return AlertActionResultMsg{Action: "toggle", Name: name}
	}
}

func (m *Model) snoozeAlert(name string, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		snoozedUntil := time.Now().Add(duration).Format(time.RFC3339)
		if err := config.UpdateAlert(name, nil, snoozedUntil); err != nil {
			return AlertActionResultMsg{Action: "snooze", Name: name, Err: err}
		}
		return AlertActionResultMsg{Action: "snooze", Name: name}
	}
}

// DeleteAlert deletes an alert (called after confirmation).
func (m *Model) DeleteAlert(name string) tea.Cmd {
	return func() tea.Msg {
		if err := config.DeleteAlert(name); err != nil {
			return AlertActionResultMsg{Action: "delete", Name: name, Err: err}
		}
		return AlertActionResultMsg{Action: "delete", Name: name}
	}
}

// TestAlert tests an alert by checking its condition.
func (m *Model) TestAlert(name string) tea.Cmd {
	return func() tea.Msg {
		alert, ok := config.GetAlert(name)
		if !ok {
			return AlertActionResultMsg{Action: "test", Name: name, Err: fmt.Errorf("alert not found")}
		}

		if m.svc == nil {
			return AlertActionResultMsg{Action: "test", Name: name, Err: fmt.Errorf("service not available")}
		}

		// Check the alert condition
		state := &shelly.AlertState{}
		result := m.svc.CheckAlert(m.ctx, alert, state)

		// Return result with value
		return AlertTestResultMsg{
			Name:      name,
			Triggered: result.Action == shelly.AlertActionTriggered,
			Value:     result.Value,
		}
	}
}

// AlertTestResultMsg is sent when an alert test completes.
type AlertTestResultMsg struct {
	Name      string
	Triggered bool
	Value     string
	Err       error
}

// View renders the alerts component.
func (m *Model) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	content := m.renderContent()

	r := rendering.New(m.Width, m.Height).
		SetTitle("Alerts").
		SetBadge(fmt.Sprintf("%d", len(m.alerts))).
		SetFocused(m.focused).
		SetPanelIndex(m.panelIdx)

	// Only show footer when focused
	if m.focused {
		r.SetFooter(theme.StyledKeybindings("e:toggle n:new d:del s/S:snooze t:test"))
	}

	return r.SetContent(content).Render()
}

func (m *Model) renderContent() string {
	if len(m.alerts) == 0 {
		return m.styles.Muted.Render("No alerts configured\n\nPress 'n' to create a new alert")
	}

	var b strings.Builder

	for i, alert := range m.alerts {
		line := m.renderAlertLine(alert, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m *Model) renderAlertLine(alert AlertItem, selected bool) string {
	var parts []string

	// Status indicator
	var status string
	switch {
	case alert.IsTriggered:
		status = m.styles.Triggered.Render("●")
	case alert.IsSnoozed():
		status = m.styles.Snoozed.Render("◐")
	case alert.Enabled:
		status = m.styles.Enabled.Render("●")
	default:
		status = m.styles.Disabled.Render("○")
	}
	parts = append(parts, status)

	// Name
	nameStyle := m.styles.AlertName
	if selected {
		nameStyle = m.styles.Selected.Inherit(nameStyle)
	}
	action := abbreviateAction(alert.Action)
	parts = append(parts,
		nameStyle.Render(truncate(alert.Name, 16)),
		m.styles.Device.Render(truncate(alert.Device, 12)),
		m.styles.Condition.Render(truncate(alert.Condition, 12)),
		m.styles.Muted.Render(action),
	)

	// Snooze status
	if alert.IsSnoozed() {
		snoozedUntil, err := time.Parse(time.RFC3339, alert.SnoozedUntil)
		if err == nil {
			remaining := time.Until(snoozedUntil).Round(time.Minute)
			parts = append(parts, m.styles.Snoozed.Render(fmt.Sprintf("(%s)", formatDuration(remaining))))
		}
	}

	return strings.Join(parts, "  ")
}

func (m *Model) renderLoading() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Alerts").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIdx)

	return r.SetContent(m.Loader.View()).Render()
}

func (m *Model) renderError() string {
	r := rendering.New(m.Width, m.Height).
		SetTitle("Alerts").
		SetBadge("Error").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIdx)

	content := theme.StatusError().Render("Error: " + m.err.Error())
	return r.SetContent(content).Render()
}

// SetSize sets the component dimensions.
func (m *Model) SetSize(width, height int) Model {
	m.ApplySize(width, height)
	return *m
}

// SetFocused sets whether this panel has focus.
func (m *Model) SetFocused(focused bool) Model {
	m.focused = focused
	return *m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m *Model) SetPanelIndex(index int) Model {
	m.panelIdx = index
	return *m
}

// SelectedAlert returns the currently selected alert name.
func (m *Model) SelectedAlert() string {
	if len(m.alerts) == 0 || m.cursor >= len(m.alerts) {
		return ""
	}
	return m.alerts[m.cursor].Name
}

// Helpers

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func abbreviateAction(action string) string {
	switch {
	case action == actionNotify || action == "":
		return actionNotify
	case strings.HasPrefix(action, "webhook:"):
		return "webhook"
	case strings.HasPrefix(action, "command:"):
		return "cmd"
	default:
		return truncate(action, 8)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "<1m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

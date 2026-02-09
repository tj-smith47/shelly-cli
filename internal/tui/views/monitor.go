package views

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/monitor"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/toast"
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/layout"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// MonitorDeps holds dependencies for the monitor view.
type MonitorDeps struct {
	Ctx         context.Context
	Svc         *shelly.Service
	IOS         *iostreams.IOStreams
	EventStream *automation.EventStream // Shared event stream
	FocusState  *focus.State
}

// Validate ensures all required dependencies are set.
func (d MonitorDeps) Validate() error {
	if d.Ctx == nil {
		return errNilContext
	}
	if d.Svc == nil {
		return errNilService
	}
	if d.IOS == nil {
		return errNilIOStreams
	}
	if d.EventStream == nil {
		return errNilEventStream
	}
	if d.FocusState == nil {
		return errNilFocusState
	}
	return nil
}

// monitorCols holds the left/right column assignments for panel cycle order.
type monitorCols struct {
	left  []focus.GlobalPanelID
	right []focus.GlobalPanelID
}

// Monitor is the multi-panel monitoring view.
type Monitor struct {
	ctx context.Context
	id  ViewID

	// The existing monitor model handles data fetching, WebSocket events, and export.
	// We keep it as the data source and delegate data to the new panel components.
	dataSource monitor.Model

	// Panel components
	summary      monitor.SummaryModel
	powerRanking monitor.PowerRankingModel
	environment  monitor.EnvironmentModel

	// Focus management
	focusState *focus.State
	cols       monitorCols

	// Layout
	layout *layout.TwoColumnLayout
	width  int
	height int
}

// summaryBarHeight is the fixed height of the summary bar.
const summaryBarHeight = 3

// NewMonitor creates a new monitor view.
func NewMonitor(deps MonitorDeps) *Monitor {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("monitor view init", err)
		panic("monitor: " + err.Error())
	}

	// Load energy config for cost calculation
	energyCfg := config.DefaultEnergyConfig()
	if cfg, err := config.Load(); err == nil {
		energyCfg = cfg.GetEnergyConfig()
	}

	// Create the data source (existing monitor model)
	dataSource := monitor.New(monitor.Deps{
		Ctx:         deps.Ctx,
		Svc:         deps.Svc,
		IOS:         deps.IOS,
		EventStream: deps.EventStream,
	})

	// Create panel components
	summary := monitor.NewSummary()
	summary = summary.SetData(monitor.SummaryData{
		CostRate: energyCfg.CostRate,
		Currency: energyCfg.Currency,
	})

	powerRanking := monitor.NewPowerRanking()
	environment := monitor.NewEnvironment()

	// Create 2-column layout (50/50 split)
	layoutCalc := layout.NewTwoColumnLayout(0.5, 1)
	layoutCalc.LeftColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(focus.PanelMonitorPowerRanking), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelMonitorAlerts), MinHeight: 5, ExpandOnFocus: true},
	}
	layoutCalc.RightColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(focus.PanelMonitorEnvironment), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelMonitorEventFeed), MinHeight: 5, ExpandOnFocus: true},
	}

	m := &Monitor{
		ctx:          deps.Ctx,
		id:           tabs.TabMonitor,
		dataSource:   dataSource,
		summary:      summary,
		powerRanking: powerRanking,
		environment:  environment,
		focusState:   deps.FocusState,
		layout:       layoutCalc,
		cols: monitorCols{
			left:  []focus.GlobalPanelID{focus.PanelMonitorPowerRanking, focus.PanelMonitorAlerts},
			right: []focus.GlobalPanelID{focus.PanelMonitorEnvironment, focus.PanelMonitorEventFeed},
		},
	}

	m.updateFocusStates()
	return m
}

// ID returns the view ID.
func (m *Monitor) ID() ViewID {
	return m.id
}

// Init returns the initial command for the monitor view.
func (m *Monitor) Init() tea.Cmd {
	return tea.Batch(
		m.dataSource.Init(),
		m.summary.Init(),
		m.powerRanking.Init(),
		m.environment.Init(),
	)
}

// Update handles messages for the monitor view.
func (m *Monitor) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle focus changes
	if _, ok := msg.(focus.ChangedMsg); ok {
		m.updateFocusStates()
		return m, nil
	}

	// Handle keyboard input - Tab/Shift+Tab cycling and Shift+N hotkeys
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if cmd := m.handleKeyPress(keyMsg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Update data source (handles StatusUpdateMsg, DeviceEventMsg, etc.)
	var dataCmd tea.Cmd
	m.dataSource, dataCmd = m.dataSource.Update(msg)
	if dataCmd != nil {
		cmds = append(cmds, dataCmd)
	}

	// Sync data source state to panel components
	m.syncDataToComponents()

	// Update summary spinner
	var summaryCmd tea.Cmd
	m.summary, summaryCmd = m.summary.Update(msg)
	if summaryCmd != nil {
		cmds = append(cmds, summaryCmd)
	}

	// Route key messages to focused component only
	if _, ok := msg.(tea.KeyPressMsg); ok {
		cmds = append(cmds, m.updateFocusedComponent(msg))
	} else if messages.IsActionRequest(msg) {
		cmds = append(cmds, m.updateFocusedComponent(msg))
	}

	// Handle export results
	if exportMsg, ok := msg.(monitor.ExportResultMsg); ok {
		cmds = append(cmds, handleMonitorExportResult(exportMsg))
	}

	return m, tea.Batch(cmds...)
}

// syncDataToComponents updates panel components with latest data from the data source.
func (m *Monitor) syncDataToComponents() {
	statuses := m.dataSource.Statuses()
	if len(statuses) == 0 && m.dataSource.IsLoading() {
		return
	}

	// Update summary data
	var totalPower, totalEnergy float64
	var onlineCount int
	for _, s := range statuses {
		if s.Online {
			onlineCount++
			totalPower += s.Power
			totalEnergy += s.TotalEnergy
		}
	}

	m.summary = m.summary.SetData(monitor.SummaryData{
		TotalPower:  totalPower,
		TotalEnergy: totalEnergy,
		OnlineCount: onlineCount,
		TotalCount:  len(statuses),
		CostRate:    m.summary.Data().CostRate,
		Currency:    m.summary.Data().Currency,
	})

	// Update refresh state
	if m.dataSource.IsRefreshing() && !m.summary.IsRefreshing() {
		var cmd tea.Cmd
		m.summary, cmd = m.summary.StartRefresh()
		_ = cmd // Spinner tick handled by Update
	} else if !m.dataSource.IsRefreshing() && !m.dataSource.IsLoading() && m.summary.IsRefreshing() {
		m.summary = m.summary.StopRefresh()
	}

	// Update power ranking
	m.powerRanking = m.powerRanking.SetDevices(statuses)

	// Update environment panel
	m.environment = m.environment.SetDevices(statuses)
}

// handleKeyPress handles Shift+N panel jumping.
// Tab/Shift+Tab are handled at the app level, so not duplicated here.
func (m *Monitor) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	prevPanel := m.focusState.ActivePanel()

	switch msg.String() {
	case keyconst.Shift1:
		m.focusState.JumpToPanel(1) // PowerRanking
		m.updateFocusStates()
		return m.emitFocusChanged(prevPanel)
	case keyconst.Shift2:
		m.focusState.JumpToPanel(2) // Environment
		m.updateFocusStates()
		return m.emitFocusChanged(prevPanel)
	case keyconst.Shift3:
		m.focusState.JumpToPanel(3) // Alerts
		m.updateFocusStates()
		return m.emitFocusChanged(prevPanel)
	case keyconst.Shift4:
		m.focusState.JumpToPanel(4) // EventFeed
		m.updateFocusStates()
		return m.emitFocusChanged(prevPanel)
	}
	return nil
}

// emitFocusChanged returns a command that emits a FocusChangedMsg if panel actually changed.
func (m *Monitor) emitFocusChanged(prevPanel focus.GlobalPanelID) tea.Cmd {
	newPanel := m.focusState.ActivePanel()
	if newPanel == prevPanel {
		return nil
	}
	return func() tea.Msg {
		return m.focusState.NewChangedMsg(
			m.focusState.ActiveTab(),
			prevPanel,
			false, // tab didn't change
			true,  // panel changed
			false, // overlay didn't change
		)
	}
}

// updateFocusStates propagates focus state to all panel components.
func (m *Monitor) updateFocusStates() {
	m.powerRanking = m.powerRanking.SetFocused(m.focusState.IsPanelFocused(focus.PanelMonitorPowerRanking)).
		SetPanelIndex(focus.PanelMonitorPowerRanking.PanelIndex())
	m.environment = m.environment.SetFocused(m.focusState.IsPanelFocused(focus.PanelMonitorEnvironment)).
		SetPanelIndex(focus.PanelMonitorEnvironment.PanelIndex())

	// Recalculate layout with new focus
	if m.layout != nil && m.width > 0 && m.height > 0 {
		m.SetSize(m.width, m.height)
	}
}

// updateFocusedComponent routes messages to the currently focused panel.
func (m *Monitor) updateFocusedComponent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch m.focusState.ActivePanel() {
	case focus.PanelMonitorPowerRanking:
		m.powerRanking, cmd = m.powerRanking.Update(msg)
	case focus.PanelMonitorEnvironment:
		m.environment, cmd = m.environment.Update(msg)
	default:
		// Alerts, EventFeed panels will be wired in Session 29
	}
	return cmd
}

// View renders the multi-panel monitor layout.
func (m *Monitor) View() string {
	if m.isNarrow() {
		return m.renderNarrowLayout()
	}
	return m.renderStandardLayout()
}

// isNarrow returns true if the view should use narrow/vertical layout.
func (m *Monitor) isNarrow() bool {
	return m.width < 80
}

// renderNarrowLayout shows only the focused panel at full width.
func (m *Monitor) renderNarrowLayout() string {
	// Summary always at top
	summaryView := m.summary.View()

	// Show only focused panel below
	var panelView string
	switch m.focusState.ActivePanel() {
	case focus.PanelMonitorPowerRanking:
		panelView = m.powerRanking.View()
	case focus.PanelMonitorEnvironment:
		panelView = m.environment.View()
	case focus.PanelMonitorAlerts:
		panelView = m.renderPendingPanel("Alerts", m.width, m.height-summaryBarHeight)
	case focus.PanelMonitorEventFeed:
		panelView = m.renderPendingPanel("Event Feed", m.width, m.height-summaryBarHeight)
	default:
		panelView = m.powerRanking.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, summaryView, panelView)
}

// renderStandardLayout renders the multi-panel 2-column layout.
func (m *Monitor) renderStandardLayout() string {
	// Summary bar at top (fixed height)
	summaryView := m.summary.View()

	// Calculate panel dimensions from layout
	dims := m.layout.Calculate()

	// Left column: Power Ranking (top) + Alerts (bottom)
	var leftPanels []string

	if d, ok := dims[layout.PanelID(focus.PanelMonitorPowerRanking)]; ok {
		m.powerRanking = m.powerRanking.SetSize(d.Width, d.Height)
		leftPanels = append(leftPanels, m.powerRanking.View())
	}
	if d, ok := dims[layout.PanelID(focus.PanelMonitorAlerts)]; ok {
		leftPanels = append(leftPanels, m.renderPendingPanel("Alerts",
			d.Width, d.Height))
	}

	// Right column: Environment (top) + Event Feed (bottom)
	var rightPanels []string

	if d, ok := dims[layout.PanelID(focus.PanelMonitorEnvironment)]; ok {
		m.environment = m.environment.SetSize(d.Width, d.Height)
		rightPanels = append(rightPanels, m.environment.View())
	}
	if d, ok := dims[layout.PanelID(focus.PanelMonitorEventFeed)]; ok {
		rightPanels = append(rightPanels, m.renderPendingPanel("Event Feed",
			d.Width, d.Height))
	}

	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftPanels...)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightPanels...)
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, " ", rightCol)

	return lipgloss.JoinVertical(lipgloss.Left, summaryView, columns)
}

// renderPendingPanel renders a stub panel for components not yet built.
// These will be replaced in Sessions 28-29 with actual components.
func (m *Monitor) renderPendingPanel(title string, width, height int) string {
	panelID := m.titleToPanelID(title)
	focused := m.focusState.IsPanelFocused(panelID)

	colors := theme.GetSemanticColors()
	content := lipgloss.NewStyle().
		Foreground(colors.Muted).
		Render("Coming in next session")

	r := rendering.New(width, height).
		SetTitle(title).
		SetFocused(focused).
		SetPanelIndex(panelID.PanelIndex()).
		SetContent(content)

	return r.Render()
}

// titleToPanelID maps a panel title to its GlobalPanelID.
func (m *Monitor) titleToPanelID(title string) focus.GlobalPanelID {
	switch title {
	case "Environment":
		return focus.PanelMonitorEnvironment
	case "Alerts":
		return focus.PanelMonitorAlerts
	case "Event Feed":
		return focus.PanelMonitorEventFeed
	default:
		return focus.PanelMonitorPowerRanking
	}
}

// SetSize sets the view dimensions and recalculates layout.
func (m *Monitor) SetSize(width, height int) View {
	m.width = width
	m.height = height

	// Summary bar is fixed at top
	m.summary = m.summary.SetSize(width, summaryBarHeight)

	// Remaining height for 2-column panels
	panelHeight := height - summaryBarHeight
	if panelHeight < 10 {
		panelHeight = 10
	}

	if m.isNarrow() {
		// Narrow mode: full width, remaining height for focused panel
		m.powerRanking = m.powerRanking.SetSize(width, panelHeight)
		m.environment = m.environment.SetSize(width, panelHeight)
		return m
	}

	// Update layout dimensions
	m.layout.SetSize(width, panelHeight)
	activePanel := m.focusState.ActivePanel()
	if activePanel.TabFor() == tabs.TabMonitor {
		m.layout.SetFocus(layout.PanelID(activePanel))
	} else {
		m.layout.SetFocus(-1)
	}

	// Apply calculated dimensions to panel components
	dims := m.layout.Calculate()
	if d, ok := dims[layout.PanelID(focus.PanelMonitorPowerRanking)]; ok {
		m.powerRanking = m.powerRanking.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelMonitorEnvironment)]; ok {
		m.environment = m.environment.SetSize(d.Width, d.Height)
	}

	return m
}

// HasActiveModal returns true if any component has a modal overlay visible.
func (m *Monitor) HasActiveModal() bool {
	return false // No modals yet
}

// RenderModal returns the active modal content for full-screen rendering.
func (m *Monitor) RenderModal() string {
	return ""
}

// SelectedDevice returns the currently selected device from the power ranking.
// Returns nil if no device is selected or the power ranking is empty.
func (m *Monitor) SelectedDevice() *monitor.DeviceStatus {
	ranked := m.powerRanking.SelectedDevice()
	if ranked == nil {
		return nil
	}

	// Find the matching DeviceStatus from the data source
	for i := range m.dataSource.Statuses() {
		if m.dataSource.Statuses()[i].Name == ranked.Name {
			return &m.dataSource.Statuses()[i]
		}
	}
	return nil
}

// handleMonitorExportResult handles export result messages with toast notifications.
func handleMonitorExportResult(msg monitor.ExportResultMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Export failed: " + msg.Err.Error())
	}
	format := "CSV"
	if msg.Format == monitor.ExportJSON {
		format = "JSON"
	}
	return toast.Success(format + " exported to " + msg.Path)
}

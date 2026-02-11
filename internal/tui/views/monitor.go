package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/shelly/monitoring"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/alerts"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/events"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/monitor"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/toast"
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
	"github.com/tj-smith47/shelly-cli/internal/tui/layout"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// EnergyHistoryRequestMsg requests opening the energy history overlay.
type EnergyHistoryRequestMsg struct {
	DeviceName string
	Address    string
	Type       string
}

// energyHistoryDataMsg returns fetched historical data.
type energyHistoryDataMsg struct {
	DeviceName string
	Energy     float64 // kWh
	AvgPower   float64 // W
	PeakPower  float64 // W
	DataPoints int
	PowerData  []float64 // time series power values for sparkline
	Err        error
}

// PhaseDetailRequestMsg requests opening the 3-phase detail overlay.
type PhaseDetailRequestMsg struct {
	DeviceName string
	Address    string
}

// phaseDetailDataMsg returns fetched 3-phase snapshot data.
type phaseDetailDataMsg struct {
	DeviceName string
	EM         *model.EMStatus
	Err        error
}

// overlayBase holds common state for all overlay modals.
type overlayBase struct {
	deviceName string
	loading    bool
	err        error
}

// energyHistoryOverlay holds state for the energy history modal.
type energyHistoryOverlay struct {
	overlayBase
	energy     float64 // kWh
	avgPower   float64 // W
	peakPower  float64 // W
	dataPoints int
	powerData  []float64 // time series for sparkline
}

// phaseDetailOverlay holds state for the 3-phase detail modal.
type phaseDetailOverlay struct {
	overlayBase
	em *model.EMStatus
}

// overlayRenderConfig describes how to render an overlay modal.
type overlayRenderConfig struct {
	title     string
	base      *overlayBase     // nil if overlay not initialized
	contentFn func(int) string // renders content given width
}

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
	alerts       alerts.Model
	eventFeed    events.Model

	// Service for data fetching
	svc *shelly.Service

	// Alert form modal
	alertForm     alerts.AlertForm
	alertFormOpen bool

	// Energy history overlay
	energyHistory     *energyHistoryOverlay
	energyHistoryOpen bool

	// 3-phase detail overlay
	phaseDetail     *phaseDetailOverlay
	phaseDetailOpen bool

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

	// Create alerts panel (reuse existing component)
	alertsModel := alerts.New(alerts.Deps{
		Ctx: deps.Ctx,
		Svc: deps.Svc,
	})

	// Create event feed panel (reuse existing component)
	eventsConfig := config.DefaultTUIEventsConfig()
	if cfg, err := config.Load(); err == nil {
		eventsConfig = cfg.GetTUIEventsConfig()
	}
	eventFeed := events.New(events.Deps{
		Ctx:                deps.Ctx,
		Svc:                deps.Svc,
		IOS:                deps.IOS,
		EventStream:        deps.EventStream,
		FilteredEvents:     eventsConfig.FilteredEvents,
		FilteredComponents: eventsConfig.FilteredComponents,
		MaxItems:           eventsConfig.MaxItems,
	})

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
		svc:          deps.Svc,
		summary:      summary,
		powerRanking: powerRanking,
		environment:  environment,
		alerts:       alertsModel,
		eventFeed:    eventFeed,
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
		m.alerts.Init(),
		m.eventFeed.Init(),
	)
}

// Update handles messages for the monitor view.
func (m *Monitor) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle modal overlays first (capture all input when open)
	if m.alertFormOpen {
		var formCmd tea.Cmd
		m.alertForm, formCmd = m.alertForm.Update(msg)
		if formCmd != nil {
			cmds = append(cmds, formCmd)
		}
		if cmd := m.handleAlertFormMessages(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	if m.energyHistoryOpen || m.phaseDetailOpen {
		return m.updateOverlay(msg)
	}

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

	// Update all child components and sync data
	cmds = append(cmds, m.updateComponents(msg)...)

	// Handle typed messages (alerts, exports, overlays, panel-specific keys)
	cmds = append(cmds, m.handleTypedMessages(msg)...)

	return m, tea.Batch(cmds...)
}

// updateComponents updates the data source and all child components.
func (m *Monitor) updateComponents(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd

	// Update data source (handles StatusUpdateMsg, DeviceEventMsg, etc.)
	var dataCmd tea.Cmd
	m.dataSource, dataCmd = m.dataSource.Update(msg)
	if dataCmd != nil {
		cmds = append(cmds, dataCmd)
	}

	m.syncDataToComponents()

	// Update summary, alerts, event feed
	var summaryCmd tea.Cmd
	m.summary, summaryCmd = m.summary.Update(msg)
	if summaryCmd != nil {
		cmds = append(cmds, summaryCmd)
	}

	var alertsCmd tea.Cmd
	m.alerts, alertsCmd = m.alerts.Update(msg)
	if alertsCmd != nil {
		cmds = append(cmds, alertsCmd)
	}

	var eventCmd tea.Cmd
	m.eventFeed, eventCmd = m.eventFeed.Update(msg)
	if eventCmd != nil {
		cmds = append(cmds, eventCmd)
	}

	// Route key/action messages to focused component
	if _, isKey := msg.(tea.KeyPressMsg); isKey || messages.IsActionRequest(msg) {
		cmds = append(cmds, m.updateFocusedComponent(msg))
	}

	return cmds
}

// handleTypedMessages processes alert, export, overlay, and panel-specific messages.
func (m *Monitor) handleTypedMessages(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd

	if cmd := m.handleAlertMessages(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if exportMsg, ok := msg.(monitor.ExportResultMsg); ok {
		cmds = append(cmds, handleMonitorExportResult(exportMsg))
	}

	if cmd := m.handleOverlayMessages(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if cmd := m.handlePanelSpecificKeys(keyMsg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return cmds
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
	m.alerts = m.alerts.SetFocused(m.focusState.IsPanelFocused(focus.PanelMonitorAlerts))
	m.alerts = m.alerts.SetPanelIndex(focus.PanelMonitorAlerts.PanelIndex())
	m.eventFeed = m.eventFeed.SetFocused(m.focusState.IsPanelFocused(focus.PanelMonitorEventFeed))

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
	case focus.PanelMonitorAlerts:
		m.alerts, cmd = m.alerts.Update(msg)
	case focus.PanelMonitorEventFeed:
		m.eventFeed, cmd = m.eventFeed.Update(msg)
	default:
		// Other panels not on this tab
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
		panelView = m.alerts.View()
	case focus.PanelMonitorEventFeed:
		panelView = m.renderEventFeedPanel(m.width, m.height-summaryBarHeight)
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
		m.alerts = m.alerts.SetSize(d.Width, d.Height)
		leftPanels = append(leftPanels, m.alerts.View())
	}

	// Right column: Environment (top) + Event Feed (bottom)
	var rightPanels []string

	if d, ok := dims[layout.PanelID(focus.PanelMonitorEnvironment)]; ok {
		m.environment = m.environment.SetSize(d.Width, d.Height)
		rightPanels = append(rightPanels, m.environment.View())
	}
	if d, ok := dims[layout.PanelID(focus.PanelMonitorEventFeed)]; ok {
		rightPanels = append(rightPanels, m.renderEventFeedPanel(d.Width, d.Height))
	}

	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftPanels...)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightPanels...)
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, " ", rightCol)

	return lipgloss.JoinVertical(lipgloss.Left, summaryView, columns)
}

// renderEventFeedPanel renders the event feed panel with a wrapper (title/badge/footer).
// The events component renders content only; the wrapper handles the border/chrome.
func (m *Monitor) renderEventFeedPanel(width, height int) string {
	focused := m.focusState.IsPanelFocused(focus.PanelMonitorEventFeed)

	badge := m.eventFeed.StatusBadge()
	footer := m.eventFeed.FooterText()
	scrollInfo := m.eventFeed.ScrollInfo()

	r := rendering.New(width, height).
		SetTitle("Event Feed").
		SetBadge(badge).
		SetFooter(footer).
		SetFooterBadge(scrollInfo).
		SetFocused(focused).
		SetPanelIndex(focus.PanelMonitorEventFeed.PanelIndex())

	m.eventFeed = m.eventFeed.SetSize(r.ContentWidth(), r.ContentHeight()).SetFocused(focused)
	return r.SetContent(m.eventFeed.View()).Render()
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
		m.alerts = m.alerts.SetSize(width, panelHeight)
		m.eventFeed = m.eventFeed.SetSize(width, panelHeight)
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
	if d, ok := dims[layout.PanelID(focus.PanelMonitorAlerts)]; ok {
		m.alerts = m.alerts.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelMonitorEventFeed)]; ok {
		m.eventFeed = m.eventFeed.SetSize(d.Width, d.Height)
	}

	return m
}

// HasActiveModal returns true if any component has a modal overlay visible.
func (m *Monitor) HasActiveModal() bool {
	return m.alertFormOpen || m.energyHistoryOpen || m.phaseDetailOpen
}

// RenderModal returns the active modal content for full-screen rendering.
func (m *Monitor) RenderModal() string {
	if m.alertFormOpen {
		return m.alertForm.View()
	}
	if m.energyHistoryOpen {
		return m.renderEnergyHistoryOverlay()
	}
	if m.phaseDetailOpen {
		return m.renderPhaseDetailOverlay()
	}
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

// handleAlertMessages handles alert-specific messages (create, edit, delete, test results).
func (m *Monitor) handleAlertMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case alerts.AlertCreateMsg:
		m.alertForm = alerts.NewAlertForm(alerts.FormModeCreate, nil)
		m.alertForm = m.alertForm.SetSize(m.width*2/3, m.height*2/3)
		m.alertFormOpen = true
		return nil
	case alerts.AlertEditMsg:
		alert, ok := config.GetAlert(msg.Name)
		if !ok {
			return toast.Error("Alert not found: " + msg.Name)
		}
		m.alertForm = alerts.NewAlertForm(alerts.FormModeEdit, &alert)
		m.alertForm = m.alertForm.SetSize(m.width*2/3, m.height*2/3)
		m.alertFormOpen = true
		return nil
	case alerts.AlertDeleteMsg:
		return m.alerts.DeleteAlert(msg.Name)
	case alerts.AlertTestMsg:
		return m.alerts.TestAlert(msg.Name)
	case alerts.AlertActionResultMsg:
		return handleMonitorAlertActionResult(msg)
	case alerts.AlertTestResultMsg:
		return handleMonitorAlertTestResult(msg)
	}
	return nil
}

// handleAlertFormMessages handles alert form modal submit/cancel.
func (m *Monitor) handleAlertFormMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case alerts.AlertFormSubmitMsg:
		m.alertFormOpen = false
		return m.saveAlert(msg)
	case alerts.AlertFormCancelMsg:
		m.alertFormOpen = false
		return nil
	}
	return nil
}

// saveAlert persists an alert to config.
func (m *Monitor) saveAlert(msg alerts.AlertFormSubmitMsg) tea.Cmd {
	return func() tea.Msg {
		err := config.CreateAlert(msg.Name, msg.Description, msg.Device, msg.Condition, msg.Action, msg.Enabled)
		if err != nil {
			return alerts.AlertActionResultMsg{Action: "save", Name: msg.Name, Err: err}
		}
		return alerts.AlertActionResultMsg{Action: "save", Name: msg.Name}
	}
}

// TriggeredAlertCount returns the number of currently triggered alerts (for Monitor tab badge).
func (m *Monitor) TriggeredAlertCount() int {
	return m.alerts.TriggeredCount()
}

func handleMonitorAlertActionResult(msg alerts.AlertActionResultMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Alert " + msg.Action + " failed: " + msg.Err.Error())
	}
	switch msg.Action {
	case actionToggle:
		return toast.Success("Alert toggled")
	case actionDelete:
		return toast.Success("Alert deleted")
	case actionSnooze:
		return toast.Success("Alert snoozed")
	case actionSave:
		return toast.Success("Alert saved")
	}
	return nil
}

func handleMonitorAlertTestResult(msg alerts.AlertTestResultMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Test failed: " + msg.Err.Error())
	}
	if msg.Triggered {
		return toast.Warning(fmt.Sprintf("Alert would trigger (value: %s)", msg.Value))
	}
	return toast.Success(fmt.Sprintf("Alert OK (value: %s)", msg.Value))
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

// updateOverlay handles input when an overlay modal is open.
func (m *Monitor) updateOverlay(msg tea.Msg) (View, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case keyconst.KeyEsc, "q":
			m.energyHistoryOpen = false
			m.phaseDetailOpen = false
			return m, nil
		}
	}
	// Still process data messages while overlay is open
	cmd := m.handleOverlayDataMessages(msg)
	return m, cmd
}

// handlePanelSpecificKeys translates raw key presses into panel-specific messages
// for the alerts and events panels (d=delete, t=test, s=snooze when alerts focused).
func (m *Monitor) handlePanelSpecificKeys(msg tea.KeyPressMsg) tea.Cmd {
	if m.focusState.ActivePanel() != focus.PanelMonitorAlerts {
		return nil
	}
	switch msg.String() {
	case "d":
		return m.updateFocusedComponent(messages.DeleteRequestMsg{})
	case "t":
		return m.updateFocusedComponent(messages.TestRequestMsg{})
	case "s":
		return m.updateFocusedComponent(messages.SnoozeRequestMsg{})
	case "S":
		return m.updateFocusedComponent(messages.SnoozeRequestMsg{Duration: "24h"})
	}
	return nil
}

// handleOverlayMessages handles energy history and phase detail request messages.
func (m *Monitor) handleOverlayMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case EnergyHistoryRequestMsg:
		return m.openEnergyHistory(msg)
	case PhaseDetailRequestMsg:
		return m.openPhaseDetail(msg)
	}
	return m.handleOverlayDataMessages(msg)
}

// handleOverlayDataMessages processes data responses for overlay modals.
func (m *Monitor) handleOverlayDataMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case energyHistoryDataMsg:
		if m.energyHistory != nil && m.energyHistory.deviceName == msg.DeviceName {
			m.energyHistory.loading = false
			m.energyHistory.energy = msg.Energy
			m.energyHistory.avgPower = msg.AvgPower
			m.energyHistory.peakPower = msg.PeakPower
			m.energyHistory.dataPoints = msg.DataPoints
			m.energyHistory.powerData = msg.PowerData
			m.energyHistory.err = msg.Err
		}
	case phaseDetailDataMsg:
		if m.phaseDetail != nil && m.phaseDetail.deviceName == msg.DeviceName {
			m.phaseDetail.loading = false
			m.phaseDetail.em = msg.EM
			m.phaseDetail.err = msg.Err
		}
	}
	return nil
}

// openEnergyHistory opens the energy history overlay and starts fetching data.
func (m *Monitor) openEnergyHistory(msg EnergyHistoryRequestMsg) tea.Cmd {
	m.energyHistory = &energyHistoryOverlay{
		overlayBase: overlayBase{deviceName: msg.DeviceName, loading: true},
	}
	m.energyHistoryOpen = true

	svc := m.svc
	ctx := m.ctx
	deviceName := msg.DeviceName
	address := msg.Address

	return func() tea.Msg {
		monSvc := svc.Monitoring()

		// First get snapshot to determine EM/EM1 type
		snapshot, err := svc.GetMonitoringSnapshotAuto(ctx, address)
		if err != nil {
			return energyHistoryDataMsg{DeviceName: deviceName, Err: fmt.Errorf("get snapshot: %w", err)}
		}

		// Fetch last 24 hours of data
		endTS := time.Now().Unix()
		startTS := endTS - 86400 // 24 hours

		switch {
		case len(snapshot.EM) > 0:
			return fetchEMHistory(ctx, monSvc, address, deviceName, snapshot.EM[0].ID, &startTS, &endTS)
		case len(snapshot.EM1) > 0:
			return fetchEM1History(ctx, monSvc, address, deviceName, snapshot.EM1[0].ID, &startTS, &endTS)
		default:
			return energyHistoryDataMsg{
				DeviceName: deviceName,
				Err:        fmt.Errorf("no historical data — only EM/EM1 devices store history"),
			}
		}
	}
}

func fetchEMHistory(ctx context.Context, monSvc *monitoring.Service, address, deviceName string, id int, startTS, endTS *int64) energyHistoryDataMsg {
	data, err := monSvc.GetEMDataHistory(ctx, address, id, startTS, endTS)
	if err != nil {
		return energyHistoryDataMsg{DeviceName: deviceName, Err: fmt.Errorf("fetch EM data: %w", err)}
	}
	energy, avgPower, peakPower, dataPoints := monitoring.CalculateEMMetrics(data)
	powerData := extractPowerSeries(data.Data,
		func(b components.EMDataBlock) []components.EMDataValues { return b.Values },
		func(v components.EMDataValues) float64 { return v.TotalActivePower },
	)
	return energyHistoryDataMsg{
		DeviceName: deviceName, Energy: energy, AvgPower: avgPower,
		PeakPower: peakPower, DataPoints: dataPoints, PowerData: powerData,
	}
}

func fetchEM1History(ctx context.Context, monSvc *monitoring.Service, address, deviceName string, id int, startTS, endTS *int64) energyHistoryDataMsg {
	data, err := monSvc.GetEM1DataHistory(ctx, address, id, startTS, endTS)
	if err != nil {
		return energyHistoryDataMsg{DeviceName: deviceName, Err: fmt.Errorf("fetch EM1 data: %w", err)}
	}
	energy, avgPower, peakPower, dataPoints := monitoring.CalculateEM1Metrics(data)
	powerData := extractPowerSeries(data.Data,
		func(b components.EM1DataBlock) []components.EM1DataValues { return b.Values },
		func(v components.EM1DataValues) float64 { return v.ActivePower },
	)
	return energyHistoryDataMsg{
		DeviceName: deviceName, Energy: energy, AvgPower: avgPower,
		PeakPower: peakPower, DataPoints: dataPoints, PowerData: powerData,
	}
}

// extractPowerSeries extracts power values from data blocks using a generic accessor.
// Replaces the duplicated extractEMPowerSeries/extractEM1PowerSeries functions.
func extractPowerSeries[B any, V any](blocks []B, getValues func(B) []V, getPower func(V) float64) []float64 {
	var total int
	for _, block := range blocks {
		total += len(getValues(block))
	}
	powers := make([]float64, 0, total)
	for _, block := range blocks {
		for _, v := range getValues(block) {
			powers = append(powers, getPower(v))
		}
	}
	return powers
}

// openPhaseDetail opens the 3-phase detail overlay and starts fetching data.
func (m *Monitor) openPhaseDetail(msg PhaseDetailRequestMsg) tea.Cmd {
	m.phaseDetail = &phaseDetailOverlay{
		overlayBase: overlayBase{deviceName: msg.DeviceName, loading: true},
	}
	m.phaseDetailOpen = true

	svc := m.svc
	ctx := m.ctx
	deviceName := msg.DeviceName
	address := msg.Address

	return func() tea.Msg {
		snapshot, err := svc.GetMonitoringSnapshotAuto(ctx, address)
		if err != nil {
			return phaseDetailDataMsg{DeviceName: deviceName, Err: fmt.Errorf("get snapshot: %w", err)}
		}
		if len(snapshot.EM) == 0 {
			return phaseDetailDataMsg{DeviceName: deviceName, Err: fmt.Errorf("no 3-phase EM data — device is not a 3-phase energy meter")}
		}
		return phaseDetailDataMsg{DeviceName: deviceName, EM: &snapshot.EM[0]}
	}
}

// Sparkline characters for history overlay.
var overlaySparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// renderOverlayModal renders any overlay modal using a shared layout pattern.
// Handles nil/loading/error states uniformly; delegates content to cfg.contentFn.
func (m *Monitor) renderOverlayModal(cfg overlayRenderConfig) string {
	w := m.width * 2 / 3
	h := m.height * 2 / 3
	colors := theme.GetSemanticColors()

	errorStyle := lipgloss.NewStyle().Foreground(colors.Error)
	centered := lipgloss.NewStyle().Width(w-4).Height(h-4).
		Align(lipgloss.Center, lipgloss.Center)

	var content string

	switch {
	case cfg.base == nil:
		content = "No data"
	case cfg.base.loading:
		content = centered.Render("Loading...")
	case cfg.base.err != nil:
		content = centered.Render(errorStyle.Render(cfg.base.err.Error()))
	default:
		content = cfg.contentFn(w)
	}

	footer := keys.FormatHints([]keys.Hint{
		{Key: "esc", Desc: "close"},
	}, keys.FooterHintWidth(w))

	return rendering.New(w, h).
		SetTitle(cfg.title).
		SetFocused(true).
		SetFooter(footer).
		SetContent(content).
		Render()
}

// renderEnergyHistoryOverlay renders the energy history modal content.
func (m *Monitor) renderEnergyHistoryOverlay() string {
	var base *overlayBase
	if m.energyHistory != nil {
		base = &m.energyHistory.overlayBase
	}
	return m.renderOverlayModal(overlayRenderConfig{
		title: "Energy History",
		base:  base,
		contentFn: func(w int) string {
			return m.renderEnergyHistoryContent(w)
		},
	})
}

func (m *Monitor) renderEnergyHistoryContent(w int) string {
	colors := theme.GetSemanticColors()
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Highlight)
	labelStyle := lipgloss.NewStyle().Foreground(colors.Muted)
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Warning)
	eh := m.energyHistory
	lines := []string{
		titleStyle.Render("Energy History — " + eh.deviceName),
		"",
		labelStyle.Render("Average Power: ") + valueStyle.Render(formatOverlayPower(eh.avgPower)),
		labelStyle.Render("Peak Power:    ") + valueStyle.Render(formatOverlayPower(eh.peakPower)),
		labelStyle.Render("Total Energy:  ") + valueStyle.Render(fmt.Sprintf("%.3f kWh", eh.energy)),
		labelStyle.Render("Data Points:   ") + valueStyle.Render(fmt.Sprintf("%d", eh.dataPoints)),
		"",
	}

	// Sparkline chart
	if len(eh.powerData) > 0 {
		sparkWidth := max(20, w-8)
		lines = append(lines,
			labelStyle.Render("Power (last 24h):"),
			renderOverlaySparkline(eh.powerData, sparkWidth),
			labelStyle.Render("24h ago")+
				strings.Repeat(" ", max(0, sparkWidth-10))+
				labelStyle.Render("now"),
		)
	}

	return strings.Join(lines, "\n")
}

// renderPhaseDetailOverlay renders the 3-phase detail modal content.
func (m *Monitor) renderPhaseDetailOverlay() string {
	var base *overlayBase
	if m.phaseDetail != nil {
		base = &m.phaseDetail.overlayBase
	}
	return m.renderOverlayModal(overlayRenderConfig{
		title: "3-Phase Detail",
		base:  base,
		contentFn: func(w int) string {
			colors := theme.GetSemanticColors()
			return m.renderPhaseDetailContent(colors, w)
		},
	})
}

func (m *Monitor) renderPhaseDetailContent(colors theme.SemanticColors, w int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Highlight)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Text)
	labelStyle := lipgloss.NewStyle().Foreground(colors.Muted)
	valueStyle := lipgloss.NewStyle().Foreground(colors.Text)
	totalStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Warning)

	em := m.phaseDetail.em
	// Adapt column width to available overlay width (label=12 + 3 data columns)
	colW := max(12, (w-16)/3)
	tableW := min(12+colW*3, w-4)

	lines := []string{
		titleStyle.Render("3-Phase Detail — " + m.phaseDetail.deviceName),
		"",
		headerStyle.Render(padRight("", 12)) +
			headerStyle.Render(padRight("Phase A", colW)) +
			headerStyle.Render(padRight("Phase B", colW)) +
			headerStyle.Render(padRight("Phase C", colW)),
		strings.Repeat("─", tableW),
		labelStyle.Render(padRight("Voltage", 12)) +
			valueStyle.Render(padRight(fmt.Sprintf("%.1f V", em.AVoltage), colW)) +
			valueStyle.Render(padRight(fmt.Sprintf("%.1f V", em.BVoltage), colW)) +
			valueStyle.Render(padRight(fmt.Sprintf("%.1f V", em.CVoltage), colW)),
		labelStyle.Render(padRight("Current", 12)) +
			valueStyle.Render(padRight(fmt.Sprintf("%.3f A", em.ACurrent), colW)) +
			valueStyle.Render(padRight(fmt.Sprintf("%.3f A", em.BCurrent), colW)) +
			valueStyle.Render(padRight(fmt.Sprintf("%.3f A", em.CCurrent), colW)),
		labelStyle.Render(padRight("Power", 12)) +
			valueStyle.Render(padRight(formatOverlayPower(em.AActivePower), colW)) +
			valueStyle.Render(padRight(formatOverlayPower(em.BActivePower), colW)) +
			valueStyle.Render(padRight(formatOverlayPower(em.CActivePower), colW)),
		labelStyle.Render(padRight("Apparent", 12)) +
			valueStyle.Render(padRight(formatOverlayPower(em.AApparentPower), colW)) +
			valueStyle.Render(padRight(formatOverlayPower(em.BApparentPower), colW)) +
			valueStyle.Render(padRight(formatOverlayPower(em.CApparentPower), colW)),
		labelStyle.Render(padRight("PF", 12)) +
			valueStyle.Render(padRight(formatOptionalFloat(em.APowerFactor, "%.3f"), colW)) +
			valueStyle.Render(padRight(formatOptionalFloat(em.BPowerFactor, "%.3f"), colW)) +
			valueStyle.Render(padRight(formatOptionalFloat(em.CPowerFactor, "%.3f"), colW)),
		labelStyle.Render(padRight("Frequency", 12)) +
			valueStyle.Render(padRight(formatOptionalFloat(em.AFreq, "%.1f Hz"), colW)) +
			valueStyle.Render(padRight(formatOptionalFloat(em.BFreq, "%.1f Hz"), colW)) +
			valueStyle.Render(padRight(formatOptionalFloat(em.CFreq, "%.1f Hz"), colW)),
		strings.Repeat("─", tableW),
		totalStyle.Render(padRight("Total", 12)) +
			totalStyle.Render(padRight(fmt.Sprintf("%.3f A", em.TotalCurrent), colW)) +
			totalStyle.Render(padRight(formatOverlayPower(em.TotalActivePower), colW)) +
			totalStyle.Render(padRight(formatOverlayPower(em.TotalAprtPower), colW)),
	}

	// Neutral current if available
	if em.NCurrent != nil {
		lines = append(lines, "",
			labelStyle.Render("Neutral Current: ")+valueStyle.Render(fmt.Sprintf("%.3f A", *em.NCurrent)),
		)
	}

	return strings.Join(lines, "\n")
}

// renderOverlaySparkline renders a sparkline from power data scaled to width.
func renderOverlaySparkline(data []float64, width int) string {
	if len(data) == 0 {
		return ""
	}

	// Scale data to fit width
	scaled := scaleFloatData(data, width)

	// Find min/max for normalization
	minVal, maxVal := scaled[0], scaled[0]
	for _, v := range scaled {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	valRange := maxVal - minVal
	gradient := theme.SemanticGradientStyles()

	var spark strings.Builder
	for _, v := range scaled {
		var idx int
		if valRange < 0.001 {
			if maxVal < 1.0 {
				idx = 0
			} else {
				idx = 4
			}
		} else {
			normalized := (v - minVal) / valRange * 7
			idx = max(0, min(7, int(normalized)))
		}
		spark.WriteString(gradient[idx].Render(string(overlaySparkChars[idx])))
	}
	return spark.String()
}

// scaleFloatData scales a float slice to target width by averaging or interpolating.
func scaleFloatData(data []float64, width int) []float64 {
	n := len(data)
	if n == 0 || width <= 0 {
		return data
	}
	if n == width {
		return data
	}
	if n > width {
		// Compress by averaging
		result := make([]float64, width)
		ratio := float64(n) / float64(width)
		for i := range width {
			startIdx := int(float64(i) * ratio)
			endIdx := int(float64(i+1) * ratio)
			if endIdx > n {
				endIdx = n
			}
			if startIdx >= endIdx {
				startIdx = endIdx - 1
			}
			var sum float64
			count := 0
			for j := startIdx; j < endIdx; j++ {
				sum += data[j]
				count++
			}
			result[i] = sum / float64(count)
		}
		return result
	}
	// Stretch by interpolation
	result := make([]float64, width)
	for i := range width {
		srcPos := float64(i) * float64(n-1) / float64(width-1)
		lowIdx := int(srcPos)
		highIdx := lowIdx + 1
		if highIdx >= n {
			result[i] = data[n-1]
			continue
		}
		frac := srcPos - float64(lowIdx)
		result[i] = data[lowIdx]*(1-frac) + data[highIdx]*frac
	}
	return result
}

// formatOverlayPower delegates to the shared output.FormatPower formatter.
func formatOverlayPower(watts float64) string {
	return output.FormatPower(watts)
}

func formatOptionalFloat(v *float64, format string) string {
	if v == nil {
		return "—"
	}
	return fmt.Sprintf(format, *v)
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// Package views provides view management for the TUI.
package views

import (
	"context"
	"errors"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	shellykvs "github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/alerts"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/kvs"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/schedules"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/scripts"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/toast"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/virtuals"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/webhooks"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/layout"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// Error variables for validation.
var (
	errNilContext     = errors.New("context is required")
	errNilService     = errors.New("service is required")
	errNilIOStreams   = errors.New("iostreams is required")
	errNilEventStream = errors.New("event stream is required")
)

// AutomationPanel identifies which panel is focused.
// Layout is 3x3: Left(Scripts, Schedules, Webhooks) Right(Virtuals, KVS, Alerts)
// ScriptEditor and ScheduleEditor are modal overlays, not panels.
type AutomationPanel int

const (
	// PanelScripts is the scripts list panel.
	PanelScripts AutomationPanel = iota
	// PanelSchedules is the schedules list panel.
	PanelSchedules
	// PanelWebhooks is the webhooks panel.
	PanelWebhooks
	// PanelVirtuals is the virtual components panel.
	PanelVirtuals
	// PanelKVS is the KVS browser panel.
	PanelKVS
	// PanelAlerts is the alerts panel.
	PanelAlerts
)

// automationLoadPhase tracks which component is being loaded.
type automationLoadPhase int

const (
	automationLoadIdle automationLoadPhase = iota
	automationLoadScripts
	automationLoadSchedules
	automationLoadWebhooks
	automationLoadVirtuals
	automationLoadKVS
)

// automationLoadNextMsg triggers loading the next component in sequence.
type automationLoadNextMsg struct {
	phase automationLoadPhase
}

// Key string constants for key handling.
const (
	keyTab      = "tab"
	keyShiftTab = "shift+tab"
)

// Action string constants for component messages.
const (
	actionDelete = "delete"
)

// AutomationDeps holds dependencies for the automation view.
type AutomationDeps struct {
	Ctx         context.Context
	Svc         *shelly.Service
	AutoSvc     *automation.Service
	KVSSvc      *shellykvs.Service
	EventStream *automation.EventStream
}

// Validate ensures all required dependencies are set.
func (d AutomationDeps) Validate() error {
	if d.Ctx == nil {
		return errNilContext
	}
	if d.Svc == nil {
		return errNilService
	}
	if d.AutoSvc == nil {
		return errors.New("automation service is required")
	}
	if d.KVSSvc == nil {
		return errors.New("kvs service is required")
	}
	// EventStream is optional - console won't work without it but other features will
	return nil
}

// Automation is the automation view that composes all automation components.
type Automation struct {
	ctx     context.Context
	svc     *shelly.Service
	autoSvc *automation.Service
	id      ViewID
	cols    automationCols

	// Component models
	scripts        scripts.ListModel
	scriptCreate   scripts.CreateModel
	scriptEditor   scripts.EditorModel
	scriptConsole  scripts.ConsoleModel
	scriptTemplate scripts.TemplateModel
	scriptEval     scripts.EvalModel
	schedules      schedules.ListModel
	scheduleCreate schedules.CreateModel
	scheduleEditor schedules.EditorModel
	webhooks       webhooks.Model
	virtuals       virtuals.Model
	kvs            kvs.Model
	alerts         alerts.Model
	alertForm      alerts.AlertForm
	alertFormOpen  bool
	templateOpen   bool
	evalOpen       bool

	// State
	device       string
	focusedPanel AutomationPanel
	viewFocused  bool // Whether the view content has focus (vs device list)
	width        int
	height       int
	styles       AutomationStyles
	loadPhase    automationLoadPhase // Tracks sequential loading progress
	pendingEdit  bool                // Flag to launch editor after code loads

	// Modal state (script and schedule editors are modals, not panels)
	scriptEditorOpen   bool
	scheduleEditorOpen bool
	consoleOpen        bool

	// Layout calculator for flexible panel sizing
	layout *layout.TwoColumnLayout
}

// automationCols holds the left/right column assignments.
type automationCols struct {
	left  []AutomationPanel
	right []AutomationPanel
}

// AutomationStyles holds styles for the automation view.
type AutomationStyles struct {
	Panel       lipgloss.Style
	PanelActive lipgloss.Style
	Title       lipgloss.Style
	Muted       lipgloss.Style
}

// DefaultAutomationStyles returns default styles for the automation view.
func DefaultAutomationStyles() AutomationStyles {
	colors := theme.GetSemanticColors()
	return AutomationStyles{
		Panel:       styles.PanelBorder(),
		PanelActive: styles.PanelBorderActive(),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// NewAutomation creates a new automation view.
func NewAutomation(deps AutomationDeps) *Automation {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("automation view init", err)
		panic("automation: " + err.Error())
	}

	scriptListDeps := scripts.ListDeps{Ctx: deps.Ctx, Svc: deps.AutoSvc}
	scriptEditorDeps := scripts.EditorDeps{Ctx: deps.Ctx, Svc: deps.AutoSvc}
	schedulesListDeps := schedules.ListDeps{Ctx: deps.Ctx, Svc: deps.AutoSvc}
	webhooksDeps := webhooks.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	virtualsDeps := virtuals.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	kvsDeps := kvs.Deps{Ctx: deps.Ctx, Svc: deps.KVSSvc}

	// Create flexible layout with 50/50 column split
	layoutCalc := layout.NewTwoColumnLayout(0.5, 1)

	// Configure 3x3 layout: Left(Scripts, Schedules, Webhooks) Right(Virtuals, KVS, Alerts)
	// ScriptEditor and ScheduleEditor are modal overlays, not panels
	// MinHeight includes borders (2) + top padding (1) + content (1) + bottom padding (1) = 5 minimum
	layoutCalc.LeftColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(PanelScripts), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelSchedules), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelWebhooks), MinHeight: 5, ExpandOnFocus: true},
	}

	// Configure right column panels with expansion on focus
	layoutCalc.RightColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(PanelVirtuals), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelKVS), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelAlerts), MinHeight: 5, ExpandOnFocus: true},
	}

	alertsDeps := alerts.Deps{Ctx: deps.Ctx, Svc: deps.Svc}

	// Initialize console model (may be nil if no event stream)
	consoleModel := scripts.NewConsoleModel(deps.EventStream)

	a := &Automation{
		ctx:            deps.Ctx,
		svc:            deps.Svc,
		autoSvc:        deps.AutoSvc,
		id:             tabs.TabAutomation,
		scripts:        scripts.NewList(scriptListDeps),
		scriptCreate:   scripts.NewCreateModel(deps.Ctx, deps.AutoSvc),
		scriptEditor:   scripts.NewEditor(scriptEditorDeps),
		scriptConsole:  consoleModel,
		scriptTemplate: scripts.NewTemplateModel(),
		scriptEval:     scripts.NewEvalModel(deps.Ctx, deps.AutoSvc),
		schedules:      schedules.NewList(schedulesListDeps),
		scheduleCreate: schedules.NewCreateModel(deps.Ctx, deps.AutoSvc),
		scheduleEditor: schedules.NewEditor(),
		webhooks:       webhooks.New(webhooksDeps),
		virtuals:       virtuals.New(virtualsDeps),
		kvs:            kvs.New(kvsDeps),
		alerts:         alerts.New(alertsDeps),
		focusedPanel:   PanelScripts,
		styles:         DefaultAutomationStyles(),
		layout:         layoutCalc,
		cols: automationCols{
			left:  []AutomationPanel{PanelScripts, PanelSchedules, PanelWebhooks},
			right: []AutomationPanel{PanelVirtuals, PanelKVS, PanelAlerts},
		},
	}

	// Initialize focus states so the default focused panel (Scripts) receives key events
	a.updateFocusStates()

	return a
}

// Init returns the initial command.
func (a *Automation) Init() tea.Cmd {
	return tea.Batch(
		a.scripts.Init(),
		a.scriptCreate.Init(),
		a.scriptEditor.Init(),
		a.scriptConsole.Init(),
		a.scriptTemplate.Init(),
		a.scriptEval.Init(),
		a.schedules.Init(),
		a.scheduleCreate.Init(),
		a.scheduleEditor.Init(),
		a.webhooks.Init(),
		a.virtuals.Init(),
		a.kvs.Init(),
		a.alerts.Init(),
	)
}

// ID returns the view ID.
func (a *Automation) ID() ViewID {
	return a.id
}

// SetDevice sets the device for all components.
// Components are loaded sequentially to avoid overwhelming the device with concurrent RPC calls.
func (a *Automation) SetDevice(device string) tea.Cmd {
	if device == a.device {
		return nil
	}
	a.device = device

	// Reset all components by clearing their device (ignore cmds - no loading yet)
	a.scripts, _ = a.scripts.SetDevice("")
	a.schedules, _ = a.schedules.SetDevice("")
	a.webhooks, _ = a.webhooks.SetDevice("")
	a.virtuals, _ = a.virtuals.SetDevice("")
	a.kvs, _ = a.kvs.SetDevice("")

	// Clear editor states
	a.scriptEditor = a.scriptEditor.Clear()

	// Ensure focus states are propagated after device change
	a.updateFocusStates()

	// Start sequential loading with first component
	a.loadPhase = automationLoadScripts
	return func() tea.Msg {
		return automationLoadNextMsg{phase: automationLoadScripts}
	}
}

// loadNextComponent triggers loading for the current phase.
func (a *Automation) loadNextComponent() tea.Cmd {
	switch a.loadPhase {
	case automationLoadScripts:
		newScripts, cmd := a.scripts.SetDevice(a.device)
		a.scripts = newScripts
		return cmd
	case automationLoadSchedules:
		newSchedules, cmd := a.schedules.SetDevice(a.device)
		a.schedules = newSchedules
		return cmd
	case automationLoadWebhooks:
		newWebhooks, cmd := a.webhooks.SetDevice(a.device)
		a.webhooks = newWebhooks
		return cmd
	case automationLoadVirtuals:
		newVirtuals, cmd := a.virtuals.SetDevice(a.device)
		a.virtuals = newVirtuals
		return cmd
	case automationLoadKVS:
		newKVS, cmd := a.kvs.SetDevice(a.device)
		a.kvs = newKVS
		return cmd
	default:
		return nil
	}
}

// advanceLoadPhase moves to the next loading phase and returns command to trigger it.
func (a *Automation) advanceLoadPhase() tea.Cmd {
	switch a.loadPhase {
	case automationLoadScripts:
		a.loadPhase = automationLoadSchedules
	case automationLoadSchedules:
		a.loadPhase = automationLoadWebhooks
	case automationLoadWebhooks:
		a.loadPhase = automationLoadVirtuals
	case automationLoadVirtuals:
		a.loadPhase = automationLoadKVS
	case automationLoadKVS:
		a.loadPhase = automationLoadIdle
		return nil // All done
	default:
		return nil
	}
	return func() tea.Msg {
		return automationLoadNextMsg{phase: a.loadPhase}
	}
}

// Update handles messages.
func (a *Automation) Update(msg tea.Msg) (View, tea.Cmd) {
	// Handle view focus changes from app.go
	if focusMsg, ok := msg.(ViewFocusChangedMsg); ok {
		// When regaining focus, reset to first panel so Tab cycling starts fresh
		if focusMsg.Focused && !a.viewFocused {
			a.focusedPanel = PanelScripts
		}
		a.viewFocused = focusMsg.Focused
		a.updateFocusStates()
		return a, nil
	}

	// Handle sequential loading messages
	if loadMsg, ok := msg.(automationLoadNextMsg); ok {
		if loadMsg.phase == a.loadPhase {
			cmd := a.loadNextComponent()
			return a, cmd
		}
		return a, nil
	}

	// Route to active modal if any
	if cmd, handled := a.routeToActiveModal(msg); handled {
		return a, cmd
	}

	var cmds []tea.Cmd

	// Check for component completion to advance sequential loading
	if cmd := a.handleComponentLoaded(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Handle keyboard input - only update focused component for key messages
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if cmd := a.handleKeyPress(keyMsg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, a.updateFocusedComponent(msg))
	} else {
		// For non-key messages (async results), update ALL components
		cmds = append(cmds, a.updateAllComponents(msg))
	}

	// Handle cross-component messages
	cmds = append(cmds, a.handleComponentMessages(msg))

	return a, tea.Batch(cmds...)
}

// handleComponentLoaded checks for component completion messages and advances loading.
func (a *Automation) handleComponentLoaded(msg tea.Msg) tea.Cmd {
	expectedPhase := a.phaseForMessage(msg)
	if expectedPhase != automationLoadIdle && a.loadPhase == expectedPhase {
		return a.advanceLoadPhase()
	}
	return nil
}

// phaseForMessage returns the load phase that corresponds to a message type.
func (a *Automation) phaseForMessage(msg tea.Msg) automationLoadPhase {
	switch msg.(type) {
	case scripts.LoadedMsg:
		return automationLoadScripts
	case schedules.LoadedMsg:
		return automationLoadSchedules
	case webhooks.LoadedMsg:
		return automationLoadWebhooks
	case virtuals.LoadedMsg:
		return automationLoadVirtuals
	case kvs.LoadedMsg:
		return automationLoadKVS
	default:
		return automationLoadIdle
	}
}

// routeToActiveModal routes messages to the currently active modal if any.
// Returns the command and true if a modal handled the message.
func (a *Automation) routeToActiveModal(msg tea.Msg) (tea.Cmd, bool) {
	var cmds []tea.Cmd

	if a.scriptCreate.IsVisible() {
		var cmd tea.Cmd
		a.scriptCreate, cmd = a.scriptCreate.Update(msg)
		cmds = append(cmds, cmd, a.handleComponentMessages(msg))
		return tea.Batch(cmds...), true
	}

	if a.scheduleCreate.IsVisible() {
		var cmd tea.Cmd
		a.scheduleCreate, cmd = a.scheduleCreate.Update(msg)
		cmds = append(cmds, cmd, a.handleComponentMessages(msg))
		return tea.Batch(cmds...), true
	}

	if a.alertFormOpen {
		var cmd tea.Cmd
		a.alertForm, cmd = a.alertForm.Update(msg)
		cmds = append(cmds, cmd, a.handleAlertFormMessages(msg))
		return tea.Batch(cmds...), true
	}

	if a.scriptEditorOpen {
		var cmd tea.Cmd
		a.scriptEditor, cmd = a.scriptEditor.Update(msg)
		cmds = append(cmds, cmd, a.handleScriptEditorModalMessages(msg))
		return tea.Batch(cmds...), true
	}

	if a.scheduleEditorOpen {
		var cmd tea.Cmd
		a.scheduleEditor, cmd = a.scheduleEditor.Update(msg)
		cmds = append(cmds, cmd, a.handleScheduleEditorModalMessages(msg))
		return tea.Batch(cmds...), true
	}

	if a.consoleOpen {
		var cmd tea.Cmd
		a.scriptConsole, cmd = a.scriptConsole.Update(msg)
		cmds = append(cmds, cmd, a.handleConsoleModalMessages(msg))
		return tea.Batch(cmds...), true
	}

	if a.templateOpen {
		var cmd tea.Cmd
		a.scriptTemplate, cmd = a.scriptTemplate.Update(msg)
		cmds = append(cmds, cmd, a.handleTemplateModalMessages(msg))
		return tea.Batch(cmds...), true
	}

	if a.evalOpen {
		var cmd tea.Cmd
		a.scriptEval, cmd = a.scriptEval.Update(msg)
		cmds = append(cmds, cmd, a.handleEvalModalMessages(msg))
		return tea.Batch(cmds...), true
	}

	return nil, false
}

func (a *Automation) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case keyTab:
		// If on last panel, return focus to device list
		if a.focusedPanel == PanelAlerts {
			a.viewFocused = false
			a.updateFocusStates()
			return func() tea.Msg { return ReturnFocusMsg{} }
		}
		a.viewFocused = true // View has focus when cycling panels
		a.focusNext()
	case keyShiftTab:
		// If on first panel, return focus to device list
		if a.focusedPanel == PanelScripts {
			a.viewFocused = false
			a.updateFocusStates()
			return func() tea.Msg { return ReturnFocusMsg{} }
		}
		a.viewFocused = true // View has focus when cycling panels
		a.focusPrev()
	// Shift+N hotkeys: 3x3 layout - left(2-4) right(5-7)
	case keyconst.Shift2:
		a.viewFocused = true
		a.focusedPanel = PanelScripts
		a.updateFocusStates()
	case keyconst.Shift3:
		a.viewFocused = true
		a.focusedPanel = PanelSchedules
		a.updateFocusStates()
	case keyconst.Shift4:
		a.viewFocused = true
		a.focusedPanel = PanelWebhooks
		a.updateFocusStates()
	case keyconst.Shift5:
		a.viewFocused = true
		a.focusedPanel = PanelVirtuals
		a.updateFocusStates()
	case keyconst.Shift6:
		a.viewFocused = true
		a.focusedPanel = PanelKVS
		a.updateFocusStates()
	case keyconst.Shift7:
		a.viewFocused = true
		a.focusedPanel = PanelAlerts
		a.updateFocusStates()
	case "c":
		// Open console modal (only from Scripts panel)
		if a.focusedPanel == PanelScripts && a.viewFocused {
			return a.openConsoleModal()
		}
	}
	return nil
}

func (a *Automation) focusNext() {
	// 3x3 layout: left column then right column
	panels := []AutomationPanel{
		PanelScripts, PanelSchedules, PanelWebhooks, // Left column
		PanelVirtuals, PanelKVS, PanelAlerts, // Right column
	}
	for i, p := range panels {
		if p == a.focusedPanel {
			a.focusedPanel = panels[(i+1)%len(panels)]
			break
		}
	}
	a.updateFocusStates()
}

func (a *Automation) focusPrev() {
	// 3x3 layout: left column then right column
	panels := []AutomationPanel{
		PanelScripts, PanelSchedules, PanelWebhooks, // Left column
		PanelVirtuals, PanelKVS, PanelAlerts, // Right column
	}
	for i, p := range panels {
		if p == a.focusedPanel {
			prevIdx := (i - 1 + len(panels)) % len(panels)
			a.focusedPanel = panels[prevIdx]
			break
		}
	}
	a.updateFocusStates()
}

func (a *Automation) updateFocusStates() {
	// Panels only show focused when the view has overall focus AND it's the active panel
	// 3x3 layout: left column (2-4), right column (5-7)
	a.scripts = a.scripts.SetFocused(a.viewFocused && a.focusedPanel == PanelScripts).SetPanelIndex(2)
	a.schedules = a.schedules.SetFocused(a.viewFocused && a.focusedPanel == PanelSchedules).SetPanelIndex(3)
	a.webhooks = a.webhooks.SetFocused(a.viewFocused && a.focusedPanel == PanelWebhooks).SetPanelIndex(4)
	a.virtuals = a.virtuals.SetFocused(a.viewFocused && a.focusedPanel == PanelVirtuals).SetPanelIndex(5)
	a.kvs = a.kvs.SetFocused(a.viewFocused && a.focusedPanel == PanelKVS).SetPanelIndex(6)
	// alerts uses pointer receivers, so we need to split the chain
	a.alerts = a.alerts.SetFocused(a.viewFocused && a.focusedPanel == PanelAlerts)
	a.alerts = a.alerts.SetPanelIndex(7)

	// Recalculate layout with new focus (panels resize on focus change)
	if a.layout != nil && a.width > 0 && a.height > 0 {
		a.SetSize(a.width, a.height)
	}
}

func (a *Automation) updateFocusedComponent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch a.focusedPanel {
	case PanelScripts:
		a.scripts, cmd = a.scripts.Update(msg)
	case PanelSchedules:
		a.schedules, cmd = a.schedules.Update(msg)
	case PanelWebhooks:
		a.webhooks, cmd = a.webhooks.Update(msg)
	case PanelVirtuals:
		a.virtuals, cmd = a.virtuals.Update(msg)
	case PanelKVS:
		a.kvs, cmd = a.kvs.Update(msg)
	case PanelAlerts:
		a.alerts, cmd = a.alerts.Update(msg)
	}
	return cmd
}

func (a *Automation) updateAllComponents(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 0, 9)

	var scriptsCmd, scriptEditorCmd, scriptConsoleCmd, schedulesCmd, scheduleEditorCmd tea.Cmd
	var webhooksCmd, virtualsCmd, kvsCmd, alertsCmd tea.Cmd

	a.scripts, scriptsCmd = a.scripts.Update(msg)
	a.scriptEditor, scriptEditorCmd = a.scriptEditor.Update(msg)
	a.scriptConsole, scriptConsoleCmd = a.scriptConsole.Update(msg)
	a.schedules, schedulesCmd = a.schedules.Update(msg)
	a.scheduleEditor, scheduleEditorCmd = a.scheduleEditor.Update(msg)
	a.webhooks, webhooksCmd = a.webhooks.Update(msg)
	a.virtuals, virtualsCmd = a.virtuals.Update(msg)
	a.kvs, kvsCmd = a.kvs.Update(msg)
	a.alerts, alertsCmd = a.alerts.Update(msg)

	cmds = append(cmds, scriptsCmd, scriptEditorCmd, scriptConsoleCmd, schedulesCmd, scheduleEditorCmd, webhooksCmd, virtualsCmd, kvsCmd, alertsCmd)
	return tea.Batch(cmds...)
}

func (a *Automation) handleComponentMessages(msg tea.Msg) tea.Cmd {
	// Handle script-related messages
	if cmd := a.handleScriptMessages(msg); cmd != nil {
		return cmd
	}
	// Handle schedule-related messages
	if cmd := a.handleScheduleMessages(msg); cmd != nil {
		return cmd
	}
	// Handle alert-related messages
	if cmd := a.handleAlertMessages(msg); cmd != nil {
		return cmd
	}
	// Handle action messages
	return handleActionMessages(msg)
}

func (a *Automation) handleScriptMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case scripts.SelectScriptMsg:
		return a.handleScriptSelect(msg)
	case scripts.CreateScriptMsg:
		return a.handleScriptCreate(msg)
	case scripts.CreatedMsg:
		return a.handleScriptCreated(msg)
	case scripts.EditScriptMsg:
		return a.handleScriptEdit(msg)
	case scripts.CodeLoadedMsg:
		return a.handleCodeLoaded()
	case scripts.EditorFinishedMsg:
		return a.handleEditorFinished(msg)
	case scripts.CodeUploadedMsg:
		return a.handleCodeUploaded(msg)
	case scripts.ScriptDownloadedMsg:
		return a.handleScriptDownloaded(msg)
	case scripts.InsertTemplateMsg:
		return a.openTemplateModal(msg.ScriptID)
	case scripts.OpenEvalMsg:
		return a.openEvalModal(msg.ScriptID)
	}
	return nil
}

func (a *Automation) handleScheduleMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case schedules.SelectScheduleMsg:
		return a.handleScheduleSelect(msg)
	case schedules.CreateScheduleMsg:
		return a.handleScheduleCreate(msg)
	case schedules.CreatedMsg:
		return a.handleScheduleCreated(msg)
	case messages.EditClosedMsg:
		if msg.Saved {
			return toast.Success("Changes saved")
		}
	}
	return nil
}

func (a *Automation) handleAlertMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case alerts.AlertCreateMsg:
		a.alertForm = alerts.NewAlertForm(alerts.FormModeCreate, nil)
		a.alertForm = a.alertForm.SetSize(a.width, a.height)
		a.alertFormOpen = true
		return nil
	case alerts.AlertEditMsg:
		alert, ok := config.GetAlert(msg.Name)
		if !ok {
			return toast.Error("Alert not found: " + msg.Name)
		}
		a.alertForm = alerts.NewAlertForm(alerts.FormModeEdit, &alert)
		a.alertForm = a.alertForm.SetSize(a.width, a.height)
		a.alertFormOpen = true
		return nil
	case alerts.AlertDeleteMsg:
		return a.alerts.DeleteAlert(msg.Name)
	case alerts.AlertActionResultMsg:
		return handleAlertActionResult(msg)
	case alerts.AlertTestResultMsg:
		return handleAlertTestResult(msg)
	}
	return nil
}

func (a *Automation) handleAlertFormMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case alerts.AlertFormSubmitMsg:
		a.alertFormOpen = false
		return a.saveAlert(msg)
	case alerts.AlertFormCancelMsg:
		a.alertFormOpen = false
		return nil
	}
	return nil
}

func (a *Automation) handleScriptEditorModalMessages(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		// Close modal on Escape - unfocus the editor
		if keyMsg.String() == keyconst.KeyEsc {
			a.scriptEditorOpen = false
			a.scriptEditor = a.scriptEditor.SetFocused(false)
			// Return a command that resolves to nil (needed for type signature)
			return func() tea.Msg { return nil }
		}
	}
	// Note: EditorFinishedMsg and CodeUploadedMsg are handled by handleScriptMessages
	// because the external editor (nano/vim) takes over the terminal - no modal needed
	return nil
}

func (a *Automation) handleScheduleEditorModalMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Close modal on Escape
		if msg.String() == keyconst.KeyEsc {
			a.scheduleEditorOpen = false
			return nil
		}
	case messages.EditClosedMsg:
		a.scheduleEditorOpen = false
		if msg.Saved {
			// Refresh schedules list
			var cmd tea.Cmd
			a.schedules, cmd = a.schedules.Refresh()
			return tea.Batch(cmd, toast.Success("Schedule saved"))
		}
		return nil
	}
	return nil
}

func (a *Automation) handleConsoleModalMessages(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		// Close modal on Escape
		if keyMsg.String() == keyconst.KeyEsc {
			a.consoleOpen = false
			a.scriptConsole = a.scriptConsole.SetFocused(false)
			// Return a command that resolves to nil (needed for type signature)
			return func() tea.Msg { return nil }
		}
	}
	return nil
}

func (a *Automation) openConsoleModal() tea.Cmd {
	a.consoleOpen = true
	a.scriptConsole = a.scriptConsole.SetDevice(a.device)
	a.scriptConsole = a.scriptConsole.SetFocused(true)
	// Size the console to fill most of the screen
	modalWidth := a.width - 8
	modalHeight := a.height - 4
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalHeight < 10 {
		modalHeight = 10
	}
	a.scriptConsole = a.scriptConsole.SetSize(modalWidth, modalHeight)
	return nil
}

func (a *Automation) handleTemplateModalMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case scripts.TemplateSelectedMsg:
		a.templateOpen = false
		return a.insertTemplate(msg)
	case messages.EditClosedMsg:
		a.templateOpen = false
		return func() tea.Msg { return nil }
	}
	return nil
}

func (a *Automation) openTemplateModal(scriptID int) tea.Cmd {
	a.templateOpen = true
	a.scriptTemplate = a.scriptTemplate.Show(a.device, scriptID)
	modalWidth := a.width - 8
	modalHeight := a.height - 4
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalHeight < 10 {
		modalHeight = 10
	}
	a.scriptTemplate = a.scriptTemplate.SetSize(modalWidth, modalHeight)
	return nil
}

func (a *Automation) insertTemplate(msg scripts.TemplateSelectedMsg) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
		defer cancel()

		var code string
		if msg.Append {
			// Get existing code and append
			existing, err := a.autoSvc.GetScriptCode(ctx, msg.Device, msg.ScriptID)
			if err != nil {
				return scripts.CodeUploadedMsg{Device: msg.Device, ScriptID: msg.ScriptID, Err: err}
			}
			code = existing + "\n\n" + msg.Code
		} else {
			code = msg.Code
		}

		err := a.autoSvc.UpdateScriptCode(ctx, msg.Device, msg.ScriptID, code, false)
		return scripts.CodeUploadedMsg{Device: msg.Device, ScriptID: msg.ScriptID, Err: err}
	}
}

func (a *Automation) handleEvalModalMessages(msg tea.Msg) tea.Cmd {
	if _, ok := msg.(messages.EditClosedMsg); ok {
		a.evalOpen = false
		return func() tea.Msg { return nil }
	}
	return nil
}

func (a *Automation) openEvalModal(scriptID int) tea.Cmd {
	a.evalOpen = true
	a.scriptEval = a.scriptEval.Show(a.device, scriptID)
	modalWidth := a.width - 8
	modalHeight := a.height - 4
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalHeight < 10 {
		modalHeight = 10
	}
	a.scriptEval = a.scriptEval.SetSize(modalWidth, modalHeight)
	return nil
}

func (a *Automation) saveAlert(msg alerts.AlertFormSubmitMsg) tea.Cmd {
	return func() tea.Msg {
		// CreateAlert handles both create and update
		err := config.CreateAlert(msg.Name, msg.Description, msg.Device, msg.Condition, msg.Action, msg.Enabled)
		if err != nil {
			return alerts.AlertActionResultMsg{Action: "save", Name: msg.Name, Err: err}
		}
		return alerts.AlertActionResultMsg{Action: "save", Name: msg.Name}
	}
}

func handleAlertActionResult(msg alerts.AlertActionResultMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Alert " + msg.Action + " failed: " + msg.Err.Error())
	}
	switch msg.Action {
	case "toggle":
		return toast.Success("Alert toggled")
	case "delete":
		return toast.Success("Alert deleted")
	case "snooze":
		return toast.Success("Alert snoozed")
	case "save":
		return toast.Success("Alert saved")
	}
	return nil
}

func handleAlertTestResult(msg alerts.AlertTestResultMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Test failed: " + msg.Err.Error())
	}
	if msg.Triggered {
		return toast.Warning(fmt.Sprintf("Alert would trigger (value: %s)", msg.Value))
	}
	return toast.Success(fmt.Sprintf("Alert OK (value: %s)", msg.Value))
}

func handleActionMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case scripts.ActionMsg:
		return handleScriptAction(msg)
	case schedules.ActionMsg:
		return handleScheduleAction(msg)
	case kvs.ActionMsg:
		return handleKVSAction(msg)
	case kvs.ExportedMsg:
		return handleKVSExported(msg)
	case kvs.ImportedMsg:
		return handleKVSImported(msg)
	case webhooks.ActionMsg:
		return handleWebhookAction(msg)
	case webhooks.TestResultMsg:
		return handleWebhookTestResult(msg)
	case virtuals.ActionMsg:
		return handleVirtualAction(msg)
	}
	return nil
}

func handleScriptAction(msg scripts.ActionMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Script " + msg.Action + " failed: " + msg.Err.Error())
	}
	switch msg.Action {
	case "start":
		return toast.Success("Script started")
	case "stop":
		return toast.Success("Script stopped")
	case actionDelete:
		return toast.Success("Script deleted")
	}
	return nil
}

func handleScheduleAction(msg schedules.ActionMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Schedule " + msg.Action + " failed: " + msg.Err.Error())
	}
	switch msg.Action {
	case "enable":
		return toast.Success("Schedule enabled")
	case "disable":
		return toast.Success("Schedule disabled")
	case actionDelete:
		return toast.Success("Schedule deleted")
	}
	return nil
}

func (a *Automation) handleScriptSelect(msg scripts.SelectScriptMsg) tea.Cmd {
	var cmd tea.Cmd
	a.scriptEditor, cmd = a.scriptEditor.SetScript(a.device, msg.Script)
	// Open script editor as modal overlay
	a.scriptEditorOpen = true
	// Set focused so key events are handled
	a.scriptEditor = a.scriptEditor.SetFocused(true)
	// Size the editor to fill most of the screen
	modalWidth := a.width - 8
	modalHeight := a.height - 4
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalHeight < 10 {
		modalHeight = 10
	}
	a.scriptEditor = a.scriptEditor.SetSize(modalWidth, modalHeight)
	return cmd
}

func (a *Automation) handleScriptCreate(_ scripts.CreateScriptMsg) tea.Cmd {
	a.scriptCreate = a.scriptCreate.Show(a.device)
	a.scriptCreate = a.scriptCreate.SetSize(a.width, a.height)
	return nil
}

func (a *Automation) handleScriptCreated(msg scripts.CreatedMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Failed to create script: " + msg.Err.Error())
	}
	// Refresh scripts list
	var cmd tea.Cmd
	a.scripts, cmd = a.scripts.Refresh()
	return tea.Batch(cmd, toast.Success("Script '"+msg.Name+"' created"))
}

func (a *Automation) handleScriptEdit(msg scripts.EditScriptMsg) tea.Cmd {
	a.pendingEdit = true
	var loadCmd tea.Cmd
	a.scriptEditor, loadCmd = a.scriptEditor.SetScript(a.device, msg.Script)
	return loadCmd
}

func (a *Automation) handleCodeLoaded() tea.Cmd {
	if a.pendingEdit {
		a.pendingEdit = false
		return a.scriptEditor.Edit()
	}
	return nil
}

func (a *Automation) handleEditorFinished(msg scripts.EditorFinishedMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Editor error: " + msg.Err.Error())
	}
	return a.uploadScriptCode(msg.Device, msg.ScriptID, msg.Code)
}

func (a *Automation) handleCodeUploaded(msg scripts.CodeUploadedMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Failed to save script: " + msg.Err.Error())
	}
	cmds := make([]tea.Cmd, 0, 3)
	var scriptsCmd tea.Cmd
	a.scripts, scriptsCmd = a.scripts.Refresh()
	cmds = append(cmds, scriptsCmd, toast.Success("Script saved to device"))
	var editorCmd tea.Cmd
	a.scriptEditor, editorCmd = a.scriptEditor.Refresh()
	cmds = append(cmds, editorCmd)
	return tea.Batch(cmds...)
}

func (a *Automation) handleScriptDownloaded(msg scripts.ScriptDownloadedMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Failed to save script: " + msg.Err.Error())
	}
	return toast.Success("Script saved to " + msg.Path)
}

func (a *Automation) handleScheduleSelect(msg schedules.SelectScheduleMsg) tea.Cmd {
	a.scheduleEditor = a.scheduleEditor.SetSchedule(&msg.Schedule)
	// Open schedule editor as modal overlay
	a.scheduleEditorOpen = true
	// Set focused so key events are handled
	a.scheduleEditor = a.scheduleEditor.SetFocused(true)
	// Size the editor to fill most of the screen
	modalWidth := a.width - 8
	modalHeight := a.height - 4
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalHeight < 10 {
		modalHeight = 10
	}
	a.scheduleEditor = a.scheduleEditor.SetSize(modalWidth, modalHeight)
	return nil
}

func (a *Automation) handleScheduleCreate(_ schedules.CreateScheduleMsg) tea.Cmd {
	a.scheduleCreate = a.scheduleCreate.Show(a.device)
	a.scheduleCreate = a.scheduleCreate.SetSize(a.width, a.height)
	return nil
}

func (a *Automation) handleScheduleCreated(msg schedules.CreatedMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Failed to create schedule: " + msg.Err.Error())
	}
	// Refresh schedules list
	var cmd tea.Cmd
	a.schedules, cmd = a.schedules.Refresh()
	return tea.Batch(cmd, toast.Success("Schedule created"))
}

func handleKVSAction(msg kvs.ActionMsg) tea.Cmd {
	if msg.Action != actionDelete {
		return nil
	}
	if msg.Err != nil {
		return toast.Error("Failed to delete: " + msg.Err.Error())
	}
	return toast.Success("KVS entry deleted")
}

func handleKVSExported(msg kvs.ExportedMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Export failed: " + msg.Err.Error())
	}
	return toast.Success("KVS exported to " + msg.Path)
}

func handleKVSImported(msg kvs.ImportedMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Import failed: " + msg.Err.Error())
	}
	return toast.Success(fmt.Sprintf("Imported %d KVS entries", msg.Count))
}

func handleWebhookAction(msg webhooks.ActionMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Webhook action failed: " + msg.Err.Error())
	}
	switch msg.Action {
	case "enable":
		return toast.Success("Webhook enabled")
	case "disable":
		return toast.Success("Webhook disabled")
	case actionDelete:
		return toast.Success("Webhook deleted")
	}
	return nil
}

func handleWebhookTestResult(msg webhooks.TestResultMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Test failed: " + msg.Err.Error())
	}
	if msg.StatusCode >= 200 && msg.StatusCode < 300 {
		return toast.Success(fmt.Sprintf("Test OK: HTTP %d", msg.StatusCode))
	}
	return toast.Warning(fmt.Sprintf("Test returned HTTP %d", msg.StatusCode))
}

func handleVirtualAction(msg virtuals.ActionMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Action failed: " + msg.Err.Error())
	}
	switch msg.Action {
	case "toggle":
		return toast.Success("Value toggled")
	case "trigger":
		return toast.Success("Button triggered")
	case "set":
		return toast.Success("Value updated")
	case actionDelete:
		return toast.Success("Virtual component deleted")
	}
	return nil
}

// uploadScriptCode uploads the modified script code to the device.
func (a *Automation) uploadScriptCode(device string, scriptID int, code string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
		defer cancel()

		err := a.autoSvc.UpdateScriptCode(ctx, device, scriptID, code, false)
		return scripts.CodeUploadedMsg{Device: device, ScriptID: scriptID, Err: err}
	}
}

// isNarrow returns true if the view should use narrow/vertical layout.
func (a *Automation) isNarrow() bool {
	return a.width < 80
}

// View renders the automation view.
func (a *Automation) View() string {
	if a.device == "" {
		return a.styles.Muted.Render("No device selected. Select a device from the Devices tab.")
	}

	if a.isNarrow() {
		return a.renderNarrowLayout()
	}

	return a.renderStandardLayout()
}

func (a *Automation) renderNarrowLayout() string {
	// In narrow mode, show only the focused panel at full width
	// 3x3 layout: Scripts, Schedules, Webhooks, Virtuals, KVS, Alerts
	// Components already have embedded titles from rendering.New()
	switch a.focusedPanel {
	case PanelScripts:
		return a.scripts.View()
	case PanelSchedules:
		return a.schedules.View()
	case PanelWebhooks:
		return a.webhooks.View()
	case PanelVirtuals:
		return a.virtuals.View()
	case PanelKVS:
		return a.kvs.View()
	case PanelAlerts:
		return a.alerts.View()
	default:
		return a.scripts.View()
	}
}

func (a *Automation) renderStandardLayout() string {
	// Render 3x3 layout (components already have embedded titles from rendering.New())
	// Left column: Scripts, Schedules, Webhooks
	// Right column: Virtuals, KVS, Alerts
	leftPanels := []string{
		a.scripts.View(),
		a.schedules.View(),
		a.webhooks.View(),
	}

	rightPanels := []string{
		a.virtuals.View(),
		a.kvs.View(),
		a.alerts.View(),
	}

	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftPanels...)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightPanels...)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, " ", rightCol)
}

// SetSize sets the view dimensions.
func (a *Automation) SetSize(width, height int) View {
	a.width = width
	a.height = height

	if a.isNarrow() {
		// Narrow mode: 3x3 panels get full width, full height
		contentWidth := width - 4
		contentHeight := height - 6

		a.scripts = a.scripts.SetSize(contentWidth, contentHeight)
		a.schedules = a.schedules.SetSize(contentWidth, contentHeight)
		a.webhooks = a.webhooks.SetSize(contentWidth, contentHeight)
		a.virtuals = a.virtuals.SetSize(contentWidth, contentHeight)
		a.kvs = a.kvs.SetSize(contentWidth, contentHeight)
		a.alerts = a.alerts.SetSize(contentWidth, contentHeight)
		return a
	}

	// Update layout with new dimensions and focus
	a.layout.SetSize(width, height)
	// Only expand panels when view has focus, otherwise distribute evenly
	if a.viewFocused {
		a.layout.SetFocus(layout.PanelID(a.focusedPanel))
	} else {
		a.layout.SetFocus(-1) // No expansion when device list is focused
	}

	// Calculate panel dimensions using flexible layout
	dims := a.layout.Calculate()

	// Apply sizes to left column components
	// Pass full panel dimensions - components handle their own borders via rendering.New()
	if d, ok := dims[layout.PanelID(PanelScripts)]; ok {
		a.scripts = a.scripts.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelSchedules)]; ok {
		a.schedules = a.schedules.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelWebhooks)]; ok {
		a.webhooks = a.webhooks.SetSize(d.Width, d.Height)
	}

	// Apply sizes to right column components (3x3 layout)
	if d, ok := dims[layout.PanelID(PanelVirtuals)]; ok {
		a.virtuals = a.virtuals.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelKVS)]; ok {
		a.kvs = a.kvs.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelAlerts)]; ok {
		a.alerts = a.alerts.SetSize(d.Width, d.Height)
	}

	return a
}

// Device returns the current device.
func (a *Automation) Device() string {
	return a.device
}

// FocusedPanel returns the currently focused panel.
func (a *Automation) FocusedPanel() AutomationPanel {
	return a.focusedPanel
}

// SetFocusedPanel sets the focused panel.
func (a *Automation) SetFocusedPanel(panel AutomationPanel) *Automation {
	a.focusedPanel = panel
	a.updateFocusStates()
	return a
}

// SetViewFocused sets whether the view has overall focus (vs device list).
// When false, all panels show as unfocused.
func (a *Automation) SetViewFocused(focused bool) *Automation {
	a.viewFocused = focused
	a.updateFocusStates()
	return a
}

// HasActiveModal returns true if any component has a modal overlay visible.
// Implements ModalProvider interface.
func (a *Automation) HasActiveModal() bool {
	return a.scriptCreate.IsVisible() || a.scheduleCreate.IsVisible() ||
		a.webhooks.IsEditing() || a.virtuals.IsEditing() || a.kvs.IsEditing() ||
		a.alertFormOpen || a.scriptEditorOpen || a.scheduleEditorOpen || a.consoleOpen ||
		a.templateOpen || a.evalOpen
}

// RenderModal returns the active modal content for full-screen rendering.
// Implements ModalProvider interface.
func (a *Automation) RenderModal() string {
	if a.scriptCreate.IsVisible() {
		return a.scriptCreate.View()
	}
	if a.scheduleCreate.IsVisible() {
		return a.scheduleCreate.View()
	}
	if a.webhooks.IsEditing() {
		return a.webhooks.RenderEditModal()
	}
	if a.virtuals.IsEditing() {
		return a.virtuals.RenderEditModal()
	}
	if a.kvs.IsEditing() {
		return a.kvs.RenderEditModal()
	}
	if a.alertFormOpen {
		return a.alertForm.View()
	}
	if a.scriptEditorOpen {
		return a.scriptEditor.View()
	}
	if a.scheduleEditorOpen {
		return a.scheduleEditor.View()
	}
	if a.consoleOpen {
		return a.scriptConsole.View()
	}
	if a.templateOpen {
		return a.scriptTemplate.View()
	}
	if a.evalOpen {
		return a.scriptEval.View()
	}
	return ""
}

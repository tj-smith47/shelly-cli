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
	"github.com/tj-smith47/shelly-cli/internal/tui/focus"
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
	errNilFocusState  = errors.New("focus state is required")
)

// automationLoadPhase tracks which component is being loaded.
type automationLoadPhase int

// scriptStoppedForEditMsg signals that a running script was stopped for editing.
type scriptStoppedForEditMsg struct {
	script scripts.Script
	err    error
}

// scriptRestartedMsg signals that a script was restarted after editing.
type scriptRestartedMsg struct {
	err error
}

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
	FocusState  *focus.State
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
	if d.FocusState == nil {
		return errNilFocusState
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
	scheduleEdit   schedules.EditModel
	webhooks       webhooks.Model
	virtuals       virtuals.Model
	kvs            kvs.Model
	alerts         alerts.Model
	alertForm      alerts.AlertForm
	alertFormOpen  bool
	templateOpen   bool
	evalOpen       bool

	// State
	device               string
	focusState           *focus.State // Unified focus state (single source of truth)
	width                int
	height               int
	styles               AutomationStyles
	loadPhase            automationLoadPhase // Tracks sequential loading progress
	pendingEdit          bool                // Flag to launch editor after code loads
	editScriptID         int                 // Script ID being edited (for restart after save)
	editScriptWasRunning bool                // Whether script was running before edit (to restart after save)

	// Modal state (script and schedule editors are modals, not panels)
	scriptEditorOpen   bool
	scheduleEditorOpen bool // Read-only view
	scheduleEditOpen   bool // Editable form
	consoleOpen        bool

	// Layout calculator for flexible panel sizing
	layout *layout.TwoColumnLayout
}

// automationCols holds the left/right column assignments for panel cycle order.
type automationCols struct {
	left  []focus.GlobalPanelID
	right []focus.GlobalPanelID
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
		{ID: layout.PanelID(focus.PanelAutoScripts), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelAutoSchedules), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelAutoWebhooks), MinHeight: 5, ExpandOnFocus: true},
	}

	// Configure right column panels with expansion on focus
	layoutCalc.RightColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(focus.PanelAutoVirtuals), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelAutoKVS), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(focus.PanelAutoAlerts), MinHeight: 5, ExpandOnFocus: true},
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
		scheduleEdit:   schedules.NewEditModel(deps.Ctx, deps.AutoSvc),
		webhooks:       webhooks.New(webhooksDeps),
		virtuals:       virtuals.New(virtualsDeps),
		kvs:            kvs.New(kvsDeps),
		alerts:         alerts.New(alertsDeps),
		focusState:     deps.FocusState,
		styles:         DefaultAutomationStyles(),
		layout:         layoutCalc,
		cols: automationCols{
			left:  []focus.GlobalPanelID{focus.PanelAutoScripts, focus.PanelAutoSchedules, focus.PanelAutoWebhooks},
			right: []focus.GlobalPanelID{focus.PanelAutoVirtuals, focus.PanelAutoKVS, focus.PanelAutoAlerts},
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
		a.scheduleEdit.Init(),
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
	// Handle focus changes from the unified focus state
	if _, ok := msg.(focus.ChangedMsg); ok {
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

	// Handle edit modal coordination messages - notify app.go of modal state changes
	if _, ok := msg.(messages.EditOpenedMsg); ok {
		cmds = append(cmds, func() tea.Msg {
			return messages.ModalOpenedMsg{ID: focus.OverlayEditModal, Mode: focus.ModeModal}
		})
	}
	if _, ok := msg.(messages.EditClosedMsg); ok {
		cmds = append(cmds, func() tea.Msg {
			return messages.ModalClosedMsg{ID: focus.OverlayEditModal}
		})
	}

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
	} else if messages.IsActionRequest(msg) {
		// Action request messages go only to the focused component
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

	if a.scheduleEditOpen {
		var cmd tea.Cmd
		a.scheduleEdit, cmd = a.scheduleEdit.Update(msg)
		cmds = append(cmds, cmd, a.handleScheduleEditModalMessages(msg))
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
	prevPanel := a.focusState.ActivePanel()

	switch msg.String() {
	case keyconst.KeyTab:
		a.focusState.NextPanel()
		a.updateFocusStates()
		return a.emitFocusChanged(prevPanel)
	case keyconst.KeyShiftTab:
		a.focusState.PrevPanel()
		a.updateFocusStates()
		return a.emitFocusChanged(prevPanel)
	// Shift+N hotkeys: 3x3 layout - left(2-4) right(5-7)
	case keyconst.Shift2:
		a.focusState.JumpToPanel(2) // Scripts
		a.updateFocusStates()
		return a.emitFocusChanged(prevPanel)
	case keyconst.Shift3:
		a.focusState.JumpToPanel(3) // Schedules
		a.updateFocusStates()
		return a.emitFocusChanged(prevPanel)
	case keyconst.Shift4:
		a.focusState.JumpToPanel(4) // Webhooks
		a.updateFocusStates()
		return a.emitFocusChanged(prevPanel)
	case keyconst.Shift5:
		a.focusState.JumpToPanel(5) // Virtuals
		a.updateFocusStates()
		return a.emitFocusChanged(prevPanel)
	case keyconst.Shift6:
		a.focusState.JumpToPanel(6) // KVS
		a.updateFocusStates()
		return a.emitFocusChanged(prevPanel)
	case keyconst.Shift7:
		a.focusState.JumpToPanel(7) // Alerts
		a.updateFocusStates()
		return a.emitFocusChanged(prevPanel)
	case "c":
		// Open console modal (only from Scripts panel)
		if a.focusState.IsPanelFocused(focus.PanelAutoScripts) {
			return a.openConsoleModal()
		}
	}
	return nil
}

// emitFocusChanged returns a command that emits a FocusChangedMsg if panel actually changed.
func (a *Automation) emitFocusChanged(prevPanel focus.GlobalPanelID) tea.Cmd {
	newPanel := a.focusState.ActivePanel()
	if newPanel == prevPanel {
		return nil
	}
	return func() tea.Msg {
		return a.focusState.NewChangedMsg(
			a.focusState.ActiveTab(),
			prevPanel,
			false, // tab didn't change
			true,  // panel changed
			false, // overlay didn't change
		)
	}
}

func (a *Automation) updateFocusStates() {
	// Query focusState for panel focus (single source of truth)
	// 3x3 layout: left column (2-4), right column (5-7)
	a.scripts = a.scripts.SetFocused(a.focusState.IsPanelFocused(focus.PanelAutoScripts)).SetPanelIndex(2)
	a.schedules = a.schedules.SetFocused(a.focusState.IsPanelFocused(focus.PanelAutoSchedules)).SetPanelIndex(3)
	a.webhooks = a.webhooks.SetFocused(a.focusState.IsPanelFocused(focus.PanelAutoWebhooks)).SetPanelIndex(4)
	a.virtuals = a.virtuals.SetFocused(a.focusState.IsPanelFocused(focus.PanelAutoVirtuals)).SetPanelIndex(5)
	a.kvs = a.kvs.SetFocused(a.focusState.IsPanelFocused(focus.PanelAutoKVS)).SetPanelIndex(6)
	// alerts uses pointer receivers, so we need to split the chain
	a.alerts = a.alerts.SetFocused(a.focusState.IsPanelFocused(focus.PanelAutoAlerts))
	a.alerts = a.alerts.SetPanelIndex(7)

	// Recalculate layout with new focus (panels resize on focus change)
	if a.layout != nil && a.width > 0 && a.height > 0 {
		a.SetSize(a.width, a.height)
	}
}

func (a *Automation) updateFocusedComponent(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch a.focusState.ActivePanel() {
	case focus.PanelAutoScripts:
		a.scripts, cmd = a.scripts.Update(msg)
	case focus.PanelAutoSchedules:
		a.schedules, cmd = a.schedules.Update(msg)
	case focus.PanelAutoWebhooks:
		a.webhooks, cmd = a.webhooks.Update(msg)
	case focus.PanelAutoVirtuals:
		a.virtuals, cmd = a.virtuals.Update(msg)
	case focus.PanelAutoKVS:
		a.kvs, cmd = a.kvs.Update(msg)
	case focus.PanelAutoAlerts:
		a.alerts, cmd = a.alerts.Update(msg)
	default:
		// Panels from other tabs - no action needed
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
	case scriptStoppedForEditMsg:
		return a.handleScriptStoppedForEdit(msg)
	case scripts.CodeLoadedMsg:
		return a.handleCodeLoaded()
	case scripts.EditorFinishedMsg:
		return a.handleEditorFinished(msg)
	case scripts.CodeUploadedMsg:
		return a.handleCodeUploaded(msg)
	case scriptRestartedMsg:
		return a.handleScriptRestarted(msg)
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
	case schedules.EditScheduleMsg:
		return a.handleScheduleEdit(msg)
	case schedules.CreateScheduleMsg:
		return a.handleScheduleCreate(msg)
	case schedules.CreatedMsg:
		return a.handleScheduleCreated(msg)
	case schedules.UpdatedMsg:
		return a.handleScheduleUpdated(msg)
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

func (a *Automation) handleScheduleEditModalMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Close modal on Escape
		if msg.String() == keyconst.KeyEsc {
			a.scheduleEditOpen = false
			return nil
		}
	case messages.EditClosedMsg:
		a.scheduleEditOpen = false
		if msg.Saved {
			// Refresh schedules list
			var cmd tea.Cmd
			a.schedules, cmd = a.schedules.Refresh()
			return tea.Batch(cmd, toast.Success("Schedule updated"))
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
	a.editScriptID = msg.Script.ID
	a.editScriptWasRunning = msg.Script.Running

	// If script is running, stop it first (Shelly API doesn't allow editing running scripts)
	if msg.Script.Running {
		// Store script for later use in handler
		script := msg.Script
		return func() tea.Msg {
			ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
			defer cancel()

			err := a.autoSvc.StopScript(ctx, a.device, script.ID)
			return scriptStoppedForEditMsg{script: script, err: err}
		}
	}

	var loadCmd tea.Cmd
	a.scriptEditor, loadCmd = a.scriptEditor.SetScript(a.device, msg.Script)
	return loadCmd
}

func (a *Automation) handleScriptStoppedForEdit(msg scriptStoppedForEditMsg) tea.Cmd {
	if msg.err != nil {
		a.pendingEdit = false
		a.editScriptWasRunning = false
		return toast.Error("Failed to stop script for editing: " + msg.err.Error())
	}

	// Script stopped, now load it for editing
	var loadCmd tea.Cmd
	a.scriptEditor, loadCmd = a.scriptEditor.SetScript(a.device, msg.script)
	return tea.Batch(loadCmd, toast.Info("Script stopped for editing"))
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
		// Reset edit state on error
		a.editScriptWasRunning = false
		a.editScriptID = 0
		return toast.Error("Failed to save script: " + msg.Err.Error())
	}

	cmds := make([]tea.Cmd, 0, 4)
	var scriptsCmd tea.Cmd
	a.scripts, scriptsCmd = a.scripts.Refresh()
	cmds = append(cmds, scriptsCmd, toast.Success("Script saved to device"))
	var editorCmd tea.Cmd
	a.scriptEditor, editorCmd = a.scriptEditor.Refresh()
	cmds = append(cmds, editorCmd)

	// Restart the script if it was running before editing
	if a.editScriptWasRunning && a.editScriptID == msg.ScriptID {
		cmds = append(cmds, a.restartScriptAfterEdit(msg.Device, msg.ScriptID))
	}

	return tea.Batch(cmds...)
}

func (a *Automation) handleScriptRestarted(msg scriptRestartedMsg) tea.Cmd {
	// Reset edit state
	a.editScriptWasRunning = false
	a.editScriptID = 0

	if msg.err != nil {
		return toast.Warning("Script saved but failed to restart: " + msg.err.Error())
	}

	// Refresh the scripts list to show updated running status
	var cmd tea.Cmd
	a.scripts, cmd = a.scripts.Refresh()
	return tea.Batch(cmd, toast.Success("Script restarted"))
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

func (a *Automation) handleScheduleEdit(msg schedules.EditScheduleMsg) tea.Cmd {
	a.scheduleEdit = a.scheduleEdit.Show(msg.Device, msg.Schedule)
	a.scheduleEdit = a.scheduleEdit.SetSize(a.width, a.height)
	a.scheduleEditOpen = true
	return nil
}

func (a *Automation) handleScheduleCreate(_ schedules.CreateScheduleMsg) tea.Cmd {
	a.scheduleCreate = a.scheduleCreate.Show(a.device)
	a.scheduleCreate = a.scheduleCreate.SetSize(a.width, a.height)
	return nil
}

func (a *Automation) handleScheduleUpdated(msg schedules.UpdatedMsg) tea.Cmd {
	if msg.Err != nil {
		return toast.Error("Failed to update schedule: " + msg.Err.Error())
	}
	// Refresh schedules list
	var cmd tea.Cmd
	a.schedules, cmd = a.schedules.Refresh()
	return tea.Batch(cmd, toast.Success("Schedule updated"))
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

// restartScriptAfterEdit restarts a script that was stopped for editing.
func (a *Automation) restartScriptAfterEdit(device string, scriptID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
		defer cancel()

		err := a.autoSvc.StartScript(ctx, device, scriptID)
		return scriptRestartedMsg{err: err}
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
	switch a.focusState.ActivePanel() {
	case focus.PanelAutoScripts:
		return a.scripts.View()
	case focus.PanelAutoSchedules:
		return a.schedules.View()
	case focus.PanelAutoWebhooks:
		return a.webhooks.View()
	case focus.PanelAutoVirtuals:
		return a.virtuals.View()
	case focus.PanelAutoKVS:
		return a.kvs.View()
	case focus.PanelAutoAlerts:
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
	// Only expand panels when an automation panel has focus, otherwise distribute evenly
	activePanel := a.focusState.ActivePanel()
	if activePanel.TabFor() == tabs.TabAutomation && activePanel != focus.PanelDeviceList {
		a.layout.SetFocus(layout.PanelID(activePanel))
	} else {
		a.layout.SetFocus(-1) // No expansion when device list is focused
	}

	// Calculate panel dimensions using flexible layout
	dims := a.layout.Calculate()

	// Apply sizes to left column components
	// Pass full panel dimensions - components handle their own borders via rendering.New()
	if d, ok := dims[layout.PanelID(focus.PanelAutoScripts)]; ok {
		a.scripts = a.scripts.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelAutoSchedules)]; ok {
		a.schedules = a.schedules.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelAutoWebhooks)]; ok {
		a.webhooks = a.webhooks.SetSize(d.Width, d.Height)
	}

	// Apply sizes to right column components (3x3 layout)
	if d, ok := dims[layout.PanelID(focus.PanelAutoVirtuals)]; ok {
		a.virtuals = a.virtuals.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelAutoKVS)]; ok {
		a.kvs = a.kvs.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(focus.PanelAutoAlerts)]; ok {
		a.alerts = a.alerts.SetSize(d.Width, d.Height)
	}

	return a
}

// Device returns the current device.
func (a *Automation) Device() string {
	return a.device
}

// HasActiveModal returns true if any component has a modal overlay visible.
// Implements ModalProvider interface.
func (a *Automation) HasActiveModal() bool {
	return a.scriptCreate.IsVisible() || a.scheduleCreate.IsVisible() ||
		a.webhooks.IsEditing() || a.virtuals.IsEditing() || a.kvs.IsEditing() ||
		a.alertFormOpen || a.scriptEditorOpen || a.scheduleEditorOpen || a.scheduleEditOpen ||
		a.consoleOpen || a.templateOpen || a.evalOpen
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
	if a.scheduleEditOpen {
		return a.scheduleEdit.View()
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

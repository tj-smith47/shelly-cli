// Package views provides view management for the TUI.
package views

import (
	"context"
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/kvs"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/schedules"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/scripts"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/virtuals"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/webhooks"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/layout"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// Error variables for validation.
var (
	errNilContext   = errors.New("context is required")
	errNilService   = errors.New("service is required")
	errNilIOStreams = errors.New("iostreams is required")
)

// AutomationPanel identifies which panel is focused.
type AutomationPanel int

const (
	// PanelScripts is the scripts list panel.
	PanelScripts AutomationPanel = iota
	// PanelScriptEditor is the script editor panel.
	PanelScriptEditor
	// PanelSchedules is the schedules list panel.
	PanelSchedules
	// PanelScheduleEditor is the schedule editor panel.
	PanelScheduleEditor
	// PanelWebhooks is the webhooks panel.
	PanelWebhooks
	// PanelVirtuals is the virtual components panel.
	PanelVirtuals
	// PanelKVS is the KVS browser panel.
	PanelKVS
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

// AutomationDeps holds dependencies for the automation view.
type AutomationDeps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d AutomationDeps) Validate() error {
	if d.Ctx == nil {
		return errNilContext
	}
	if d.Svc == nil {
		return errNilService
	}
	return nil
}

// Automation is the automation view that composes all automation components.
type Automation struct {
	ctx  context.Context
	svc  *shelly.Service
	id   ViewID
	cols automationCols

	// Component models
	scripts        scripts.ListModel
	scriptEditor   scripts.EditorModel
	schedules      schedules.ListModel
	scheduleEditor schedules.EditorModel
	webhooks       webhooks.Model
	virtuals       virtuals.Model
	kvs            kvs.Model

	// State
	device       string
	focusedPanel AutomationPanel
	width        int
	height       int
	styles       AutomationStyles
	loadPhase    automationLoadPhase // Tracks sequential loading progress
	pendingEdit  bool                // Flag to launch editor after code loads

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
		Panel: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.TableBorder),
		PanelActive: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight),
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
		panic("automation: " + err.Error())
	}

	scriptListDeps := scripts.ListDeps{Ctx: deps.Ctx, Svc: deps.Svc}
	scriptEditorDeps := scripts.EditorDeps{Ctx: deps.Ctx, Svc: deps.Svc}
	schedulesListDeps := schedules.ListDeps{Ctx: deps.Ctx, Svc: deps.Svc}
	webhooksDeps := webhooks.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	virtualsDeps := virtuals.Deps{Ctx: deps.Ctx, Svc: deps.Svc}
	kvsDeps := kvs.Deps{Ctx: deps.Ctx, Svc: deps.Svc}

	// Create flexible layout with 50/50 column split
	layoutCalc := layout.NewTwoColumnLayout(0.5, 1)

	// Configure left column panels with expansion on focus
	layoutCalc.LeftColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(PanelScripts), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelSchedules), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelWebhooks), MinHeight: 4, ExpandOnFocus: true},
	}

	// Configure right column panels with expansion on focus
	layoutCalc.RightColumn.Panels = []layout.PanelConfig{
		{ID: layout.PanelID(PanelScriptEditor), MinHeight: 6, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelScheduleEditor), MinHeight: 5, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelVirtuals), MinHeight: 4, ExpandOnFocus: true},
		{ID: layout.PanelID(PanelKVS), MinHeight: 4, ExpandOnFocus: true},
	}

	a := &Automation{
		ctx:            deps.Ctx,
		svc:            deps.Svc,
		id:             tabs.TabAutomation,
		scripts:        scripts.NewList(scriptListDeps),
		scriptEditor:   scripts.NewEditor(scriptEditorDeps),
		schedules:      schedules.NewList(schedulesListDeps),
		scheduleEditor: schedules.NewEditor(),
		webhooks:       webhooks.New(webhooksDeps),
		virtuals:       virtuals.New(virtualsDeps),
		kvs:            kvs.New(kvsDeps),
		focusedPanel:   PanelScripts,
		styles:         DefaultAutomationStyles(),
		layout:         layoutCalc,
		cols: automationCols{
			left:  []AutomationPanel{PanelScripts, PanelSchedules, PanelWebhooks},
			right: []AutomationPanel{PanelScriptEditor, PanelScheduleEditor, PanelVirtuals, PanelKVS},
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
		a.scriptEditor.Init(),
		a.schedules.Init(),
		a.scheduleEditor.Init(),
		a.webhooks.Init(),
		a.virtuals.Init(),
		a.kvs.Init(),
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
	var cmds []tea.Cmd

	// Handle sequential loading messages
	if loadMsg, ok := msg.(automationLoadNextMsg); ok {
		if loadMsg.phase == a.loadPhase {
			cmd := a.loadNextComponent()
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)
	}

	// Check for component completion to advance sequential loading
	switch msg.(type) {
	case scripts.LoadedMsg:
		if a.loadPhase == automationLoadScripts {
			cmds = append(cmds, a.advanceLoadPhase())
		}
	case schedules.LoadedMsg:
		if a.loadPhase == automationLoadSchedules {
			cmds = append(cmds, a.advanceLoadPhase())
		}
	case webhooks.LoadedMsg:
		if a.loadPhase == automationLoadWebhooks {
			cmds = append(cmds, a.advanceLoadPhase())
		}
	case virtuals.LoadedMsg:
		if a.loadPhase == automationLoadVirtuals {
			cmds = append(cmds, a.advanceLoadPhase())
		}
	case kvs.LoadedMsg:
		if a.loadPhase == automationLoadKVS {
			cmds = append(cmds, a.advanceLoadPhase())
		}
	}

	// Handle keyboard input - only update focused component for key messages
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		a.handleKeyPress(keyMsg)
		cmd := a.updateFocusedComponent(msg)
		cmds = append(cmds, cmd)
	} else {
		// For non-key messages (async results), update ALL components
		cmd := a.updateAllComponents(msg)
		cmds = append(cmds, cmd)
	}

	// Handle cross-component messages
	cmd := a.handleComponentMessages(msg)
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

func (a *Automation) handleKeyPress(msg tea.KeyPressMsg) {
	switch msg.String() {
	case keyTab:
		a.focusNext()
	case keyShiftTab:
		a.focusPrev()
	case keyconst.Shift2:
		a.focusedPanel = PanelScripts
		a.updateFocusStates()
	case keyconst.Shift3:
		a.focusedPanel = PanelScriptEditor
		a.updateFocusStates()
	case keyconst.Shift4:
		a.focusedPanel = PanelSchedules
		a.updateFocusStates()
	case keyconst.Shift5:
		a.focusedPanel = PanelScheduleEditor
		a.updateFocusStates()
	case keyconst.Shift6:
		a.focusedPanel = PanelWebhooks
		a.updateFocusStates()
	case keyconst.Shift7:
		a.focusedPanel = PanelVirtuals
		a.updateFocusStates()
	case keyconst.Shift8:
		a.focusedPanel = PanelKVS
		a.updateFocusStates()
	}
}

func (a *Automation) focusNext() {
	panels := []AutomationPanel{
		PanelScripts, PanelScriptEditor,
		PanelSchedules, PanelScheduleEditor,
		PanelWebhooks,
		PanelVirtuals, PanelKVS,
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
	panels := []AutomationPanel{
		PanelScripts, PanelScriptEditor,
		PanelSchedules, PanelScheduleEditor,
		PanelWebhooks,
		PanelVirtuals, PanelKVS,
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
	a.scripts = a.scripts.SetFocused(a.focusedPanel == PanelScripts).SetPanelIndex(1)
	a.scriptEditor = a.scriptEditor.SetFocused(a.focusedPanel == PanelScriptEditor).SetPanelIndex(2)
	a.schedules = a.schedules.SetFocused(a.focusedPanel == PanelSchedules).SetPanelIndex(3)
	a.scheduleEditor = a.scheduleEditor.SetFocused(a.focusedPanel == PanelScheduleEditor).SetPanelIndex(4)
	a.webhooks = a.webhooks.SetFocused(a.focusedPanel == PanelWebhooks).SetPanelIndex(5)
	a.virtuals = a.virtuals.SetFocused(a.focusedPanel == PanelVirtuals).SetPanelIndex(6)
	a.kvs = a.kvs.SetFocused(a.focusedPanel == PanelKVS).SetPanelIndex(7)

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
	case PanelScriptEditor:
		a.scriptEditor, cmd = a.scriptEditor.Update(msg)
	case PanelSchedules:
		a.schedules, cmd = a.schedules.Update(msg)
	case PanelScheduleEditor:
		a.scheduleEditor, cmd = a.scheduleEditor.Update(msg)
	case PanelWebhooks:
		a.webhooks, cmd = a.webhooks.Update(msg)
	case PanelVirtuals:
		a.virtuals, cmd = a.virtuals.Update(msg)
	case PanelKVS:
		a.kvs, cmd = a.kvs.Update(msg)
	}
	return cmd
}

func (a *Automation) updateAllComponents(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	var scriptsCmd, scriptEditorCmd, schedulesCmd, scheduleEditorCmd tea.Cmd
	var webhooksCmd, virtualsCmd, kvsCmd tea.Cmd

	a.scripts, scriptsCmd = a.scripts.Update(msg)
	a.scriptEditor, scriptEditorCmd = a.scriptEditor.Update(msg)
	a.schedules, schedulesCmd = a.schedules.Update(msg)
	a.scheduleEditor, scheduleEditorCmd = a.scheduleEditor.Update(msg)
	a.webhooks, webhooksCmd = a.webhooks.Update(msg)
	a.virtuals, virtualsCmd = a.virtuals.Update(msg)
	a.kvs, kvsCmd = a.kvs.Update(msg)

	cmds = append(cmds, scriptsCmd, scriptEditorCmd, schedulesCmd, scheduleEditorCmd, webhooksCmd, virtualsCmd, kvsCmd)
	return tea.Batch(cmds...)
}

func (a *Automation) handleComponentMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case scripts.SelectScriptMsg:
		// When a script is selected, load it in the viewer
		var cmd tea.Cmd
		a.scriptEditor, cmd = a.scriptEditor.SetScript(a.device, msg.Script)
		a.focusedPanel = PanelScriptEditor
		a.updateFocusStates()
		return cmd

	case scripts.EditScriptMsg:
		// When edit is requested, set pending flag and load the script
		a.pendingEdit = true
		var loadCmd tea.Cmd
		a.scriptEditor, loadCmd = a.scriptEditor.SetScript(a.device, msg.Script)
		return loadCmd

	case scripts.CodeLoadedMsg:
		// Code loaded - if we have a pending edit, launch the external editor
		if a.pendingEdit {
			a.pendingEdit = false
			return a.scriptEditor.Edit()
		}
		return nil

	case scripts.EditorFinishedMsg:
		// External editor closed, upload the modified code
		if msg.Err != nil {
			// Editor failed - could show error toast here
			return nil
		}
		// Upload the modified code to the device
		return a.uploadScriptCode(msg.Device, msg.ScriptID, msg.Code)

	case scripts.CodeUploadedMsg:
		// Code upload completed
		if msg.Err != nil {
			// Upload failed - could show error toast here
			return nil
		}
		// Refresh the script list and editor
		var cmds []tea.Cmd
		var scriptsCmd tea.Cmd
		a.scripts, scriptsCmd = a.scripts.Refresh()
		cmds = append(cmds, scriptsCmd)
		var editorCmd tea.Cmd
		a.scriptEditor, editorCmd = a.scriptEditor.Refresh()
		cmds = append(cmds, editorCmd)
		return tea.Batch(cmds...)

	case schedules.SelectScheduleMsg:
		// When a schedule is selected, load it in the editor
		a.scheduleEditor = a.scheduleEditor.SetSchedule(&msg.Schedule)
		a.focusedPanel = PanelScheduleEditor
		a.updateFocusStates()
		return nil
	}
	return nil
}

// uploadScriptCode uploads the modified script code to the device.
func (a *Automation) uploadScriptCode(device string, scriptID int, code string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
		defer cancel()

		err := a.svc.UpdateScriptCode(ctx, device, scriptID, code, false)
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
	// Components already have embedded titles from rendering.New()
	switch a.focusedPanel {
	case PanelScripts:
		return a.scripts.View()
	case PanelScriptEditor:
		return a.scriptEditor.View()
	case PanelSchedules:
		return a.schedules.View()
	case PanelScheduleEditor:
		return a.scheduleEditor.View()
	case PanelWebhooks:
		return a.webhooks.View()
	case PanelVirtuals:
		return a.virtuals.View()
	case PanelKVS:
		return a.kvs.View()
	default:
		return a.scripts.View()
	}
}

func (a *Automation) renderStandardLayout() string {
	// Render panels (components already have embedded titles)
	leftPanels := []string{
		a.scripts.View(),
		a.schedules.View(),
		a.webhooks.View(),
	}

	rightPanels := []string{
		a.scriptEditor.View(),
		a.scheduleEditor.View(),
		a.virtuals.View(),
		a.kvs.View(),
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
		// Narrow mode: all components get full width, full height
		contentWidth := width - 4
		contentHeight := height - 6

		a.scripts = a.scripts.SetSize(contentWidth, contentHeight)
		a.schedules = a.schedules.SetSize(contentWidth, contentHeight)
		a.webhooks = a.webhooks.SetSize(contentWidth, contentHeight)
		a.scriptEditor = a.scriptEditor.SetSize(contentWidth, contentHeight)
		a.scheduleEditor = a.scheduleEditor.SetSize(contentWidth, contentHeight)
		a.virtuals = a.virtuals.SetSize(contentWidth, contentHeight)
		a.kvs = a.kvs.SetSize(contentWidth, contentHeight)
		return a
	}

	// Update layout with new dimensions and focus
	a.layout.SetSize(width, height)
	a.layout.SetFocus(layout.PanelID(a.focusedPanel))

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

	// Apply sizes to right column components
	if d, ok := dims[layout.PanelID(PanelScriptEditor)]; ok {
		a.scriptEditor = a.scriptEditor.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelScheduleEditor)]; ok {
		a.scheduleEditor = a.scheduleEditor.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelVirtuals)]; ok {
		a.virtuals = a.virtuals.SetSize(d.Width, d.Height)
	}
	if d, ok := dims[layout.PanelID(PanelKVS)]; ok {
		a.kvs = a.kvs.SetSize(d.Width, d.Height)
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

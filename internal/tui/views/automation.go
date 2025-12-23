// Package views provides view management for the TUI.
package views

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/kvs"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/schedules"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/scripts"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/virtuals"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/webhooks"
	"github.com/tj-smith47/shelly-cli/internal/tui/tabs"
)

// Error variables for validation.
var (
	errNilContext = errors.New("context is required")
	errNilService = errors.New("service is required")
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

// TabAutomation is the automation tab ID.
const TabAutomation tabs.TabID = 10

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

	return &Automation{
		ctx:            deps.Ctx,
		svc:            deps.Svc,
		id:             TabAutomation,
		scripts:        scripts.NewList(scriptListDeps),
		scriptEditor:   scripts.NewEditor(scriptEditorDeps),
		schedules:      schedules.NewList(schedulesListDeps),
		scheduleEditor: schedules.NewEditor(),
		webhooks:       webhooks.New(webhooksDeps),
		virtuals:       virtuals.New(virtualsDeps),
		kvs:            kvs.New(kvsDeps),
		focusedPanel:   PanelScripts,
		styles:         DefaultAutomationStyles(),
		cols: automationCols{
			left:  []AutomationPanel{PanelScripts, PanelSchedules, PanelWebhooks},
			right: []AutomationPanel{PanelScriptEditor, PanelScheduleEditor, PanelVirtuals, PanelKVS},
		},
	}
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
func (a *Automation) SetDevice(device string) tea.Cmd {
	if device == a.device {
		return nil
	}
	a.device = device

	var cmds []tea.Cmd

	// Set device on list components
	newScripts, scriptsCmd := a.scripts.SetDevice(device)
	a.scripts = newScripts
	cmds = append(cmds, scriptsCmd)

	newSchedules, schedulesCmd := a.schedules.SetDevice(device)
	a.schedules = newSchedules
	cmds = append(cmds, schedulesCmd)

	newWebhooks, webhooksCmd := a.webhooks.SetDevice(device)
	a.webhooks = newWebhooks
	cmds = append(cmds, webhooksCmd)

	newVirtuals, virtualsCmd := a.virtuals.SetDevice(device)
	a.virtuals = newVirtuals
	cmds = append(cmds, virtualsCmd)

	newKVS, kvsCmd := a.kvs.SetDevice(device)
	a.kvs = newKVS
	cmds = append(cmds, kvsCmd)

	return tea.Batch(cmds...)
}

// Update handles messages.
func (a *Automation) Update(msg tea.Msg) (View, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle keyboard input
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		a.handleKeyPress(keyMsg)
	}

	// Update focused component
	cmd := a.updateFocusedComponent(msg)
	cmds = append(cmds, cmd)

	// Handle cross-component messages
	cmd = a.handleComponentMessages(msg)
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

func (a *Automation) handleKeyPress(msg tea.KeyPressMsg) {
	switch msg.String() {
	case keyTab:
		a.focusNext()
	case keyShiftTab:
		a.focusPrev()
	case "1":
		a.focusedPanel = PanelScripts
		a.updateFocusStates()
	case "2":
		a.focusedPanel = PanelSchedules
		a.updateFocusStates()
	case "3":
		a.focusedPanel = PanelWebhooks
		a.updateFocusStates()
	case "4":
		a.focusedPanel = PanelVirtuals
		a.updateFocusStates()
	case "5":
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
	a.scripts = a.scripts.SetFocused(a.focusedPanel == PanelScripts)
	a.scriptEditor = a.scriptEditor.SetFocused(a.focusedPanel == PanelScriptEditor)
	a.schedules = a.schedules.SetFocused(a.focusedPanel == PanelSchedules)
	a.scheduleEditor = a.scheduleEditor.SetFocused(a.focusedPanel == PanelScheduleEditor)
	a.webhooks = a.webhooks.SetFocused(a.focusedPanel == PanelWebhooks)
	a.virtuals = a.virtuals.SetFocused(a.focusedPanel == PanelVirtuals)
	a.kvs = a.kvs.SetFocused(a.focusedPanel == PanelKVS)
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

func (a *Automation) handleComponentMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case scripts.SelectScriptMsg:
		// When a script is selected, load it in the editor
		var cmd tea.Cmd
		a.scriptEditor, cmd = a.scriptEditor.SetScript(a.device, msg.Script)
		a.focusedPanel = PanelScriptEditor
		a.updateFocusStates()
		return cmd

	case schedules.SelectScheduleMsg:
		// When a schedule is selected, load it in the editor
		a.scheduleEditor = a.scheduleEditor.SetSchedule(&msg.Schedule)
		a.focusedPanel = PanelScheduleEditor
		a.updateFocusStates()
		return nil
	}
	return nil
}

// View renders the automation view.
func (a *Automation) View() string {
	if a.device == "" {
		return a.styles.Muted.Render("No device selected. Select a device from the Devices tab.")
	}

	// Calculate column widths (50/50 split)
	leftWidth := a.width / 2
	rightWidth := a.width - leftWidth - 1 // -1 for gap

	// Calculate panel heights
	leftPanelCount := 3  // Scripts, Schedules, Webhooks
	rightPanelCount := 4 // ScriptEditor, ScheduleEditor, Virtuals, KVS

	leftPanelHeight := a.height / leftPanelCount
	rightPanelHeight := a.height / rightPanelCount

	// Render left column panels
	leftPanels := []string{
		a.renderPanel("Scripts", a.scripts.View(), leftWidth, leftPanelHeight, a.focusedPanel == PanelScripts),
		a.renderPanel("Schedules", a.schedules.View(), leftWidth, leftPanelHeight, a.focusedPanel == PanelSchedules),
		a.renderPanel("Webhooks", a.webhooks.View(), leftWidth, a.height-2*leftPanelHeight, a.focusedPanel == PanelWebhooks),
	}

	// Render right column panels
	rightPanels := []string{
		a.renderPanel("Script Editor", a.scriptEditor.View(), rightWidth, rightPanelHeight, a.focusedPanel == PanelScriptEditor),
		a.renderPanel("Schedule Details", a.scheduleEditor.View(), rightWidth, rightPanelHeight, a.focusedPanel == PanelScheduleEditor),
		a.renderPanel("Virtual Components", a.virtuals.View(), rightWidth, rightPanelHeight, a.focusedPanel == PanelVirtuals),
		a.renderPanel("Key-Value Store", a.kvs.View(), rightWidth, a.height-3*rightPanelHeight, a.focusedPanel == PanelKVS),
	}

	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftPanels...)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightPanels...)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, " ", rightCol)
}

func (a *Automation) renderPanel(title, content string, width, height int, focused bool) string {
	style := a.styles.Panel
	if focused {
		style = a.styles.PanelActive
	}

	style = style.Width(width - 2).Height(height - 2)

	titleStr := a.styles.Title.Render(title)
	if content == "" {
		content = a.styles.Muted.Render("(empty)")
	}

	inner := lipgloss.JoinVertical(lipgloss.Left, titleStr, "", content)
	return style.Render(inner)
}

// SetSize sets the view dimensions.
func (a *Automation) SetSize(width, height int) View {
	a.width = width
	a.height = height

	// Calculate component sizes
	leftWidth := width / 2
	rightWidth := width - leftWidth - 1

	leftPanelCount := 3
	rightPanelCount := 4

	leftPanelHeight := height / leftPanelCount
	rightPanelHeight := height / rightPanelCount

	// Set sizes for left column components
	contentHeight := leftPanelHeight - 4 // Account for border and title
	a.scripts = a.scripts.SetSize(leftWidth-4, contentHeight)
	a.schedules = a.schedules.SetSize(leftWidth-4, contentHeight)
	a.webhooks = a.webhooks.SetSize(leftWidth-4, height-2*leftPanelHeight-4)

	// Set sizes for right column components
	contentHeight = rightPanelHeight - 4
	a.scriptEditor = a.scriptEditor.SetSize(rightWidth-4, contentHeight)
	a.scheduleEditor = a.scheduleEditor.SetSize(rightWidth-4, contentHeight)
	a.virtuals = a.virtuals.SetSize(rightWidth-4, contentHeight)
	a.kvs = a.kvs.SetSize(rightWidth-4, height-3*rightPanelHeight-4)

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

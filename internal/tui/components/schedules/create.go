// Package schedules provides TUI components for managing device schedules.
package schedules

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// CreateField represents a field in the schedule create form.
type CreateField int

// Create field constants.
const (
	CreateFieldTimePattern CreateField = iota
	CreateFieldHour
	CreateFieldMinute
	CreateFieldSecond
	CreateFieldMethod
	CreateFieldParams
	CreateFieldCount
)

// Time pattern options.
const (
	patternDaily    = "Daily"
	patternWeekdays = "Weekdays (Mon-Fri)"
	patternWeekends = "Weekends (Sat-Sun)"
	patternCustom   = "Custom Days"
)

// Common RPC methods for schedules.
const (
	methodCustom      = "(Custom method)"
	methodSwitchOn    = "Switch.Set → on"
	methodSwitchOff   = "Switch.Set → off"
	methodSwitchTog   = "Switch.Toggle"
	methodCoverOpen   = "Cover.Open"
	methodCoverClose  = "Cover.Close"
	methodCoverStop   = "Cover.Stop"
	methodLightOn     = "Light.Set → on"
	methodLightOff    = "Light.Set → off"
	methodScriptStart = "Script.Start"
	methodScriptStop  = "Script.Stop"
)

// rpcMethodInfo maps display names to actual RPC method and default params.
type rpcMethodInfo struct {
	Method string
	Params map[string]any
}

// getRPCMethodInfo returns the RPC method details for a display name.
func getRPCMethodInfo(displayName string) rpcMethodInfo {
	methods := map[string]rpcMethodInfo{
		methodSwitchOn:    {Method: "Switch.Set", Params: map[string]any{"id": 0, "on": true}},
		methodSwitchOff:   {Method: "Switch.Set", Params: map[string]any{"id": 0, "on": false}},
		methodSwitchTog:   {Method: "Switch.Toggle", Params: map[string]any{"id": 0}},
		methodCoverOpen:   {Method: "Cover.Open", Params: map[string]any{"id": 0}},
		methodCoverClose:  {Method: "Cover.Close", Params: map[string]any{"id": 0}},
		methodCoverStop:   {Method: "Cover.Stop", Params: map[string]any{"id": 0}},
		methodLightOn:     {Method: "Light.Set", Params: map[string]any{"id": 0, "on": true}},
		methodLightOff:    {Method: "Light.Set", Params: map[string]any{"id": 0, "on": false}},
		methodScriptStart: {Method: "Script.Start", Params: map[string]any{"id": 1}},
		methodScriptStop:  {Method: "Script.Stop", Params: map[string]any{"id": 1}},
	}
	if info, ok := methods[displayName]; ok {
		return info
	}
	return rpcMethodInfo{}
}

// CreatedMsg signals that a new schedule was created.
type CreatedMsg struct {
	Device     string
	ScheduleID int
	Err        error
}

// CreateModel represents the schedule create modal.
type CreateModel struct {
	ctx     context.Context
	svc     *automation.Service
	device  string
	visible bool
	cursor  CreateField
	saving  bool
	err     error
	width   int
	height  int
	styles  editmodal.Styles

	// Form inputs
	patternDropdown   form.Select
	hourInput         form.TextInput
	minuteInput       form.TextInput
	secondInput       form.TextInput
	methodDropdown    form.Select
	customMethodInput form.TextInput // Used when methodCustom is selected
	paramsInput       form.TextInput

	// Custom day selection (for patternCustom)
	selectedDays [7]bool // Sun(0) through Sat(6)
}

// NewCreateModel creates a new schedule create modal.
func NewCreateModel(ctx context.Context, svc *automation.Service) CreateModel {
	patternDropdown := form.NewSelect(
		form.WithSelectOptions([]string{
			patternDaily,
			patternWeekdays,
			patternWeekends,
			patternCustom,
		}),
		form.WithSelectHelp("When to run"),
	)

	hourInput := form.NewTextInput(
		form.WithPlaceholder("12"),
		form.WithCharLimit(2),
		form.WithWidth(4),
		form.WithHelp("Hour (0-23)"),
	)

	minuteInput := form.NewTextInput(
		form.WithPlaceholder("00"),
		form.WithCharLimit(2),
		form.WithWidth(4),
		form.WithHelp("Minute (0-59)"),
	)

	secondInput := form.NewTextInput(
		form.WithPlaceholder("0"),
		form.WithCharLimit(2),
		form.WithWidth(4),
		form.WithHelp("Second (0-59)"),
	)

	methodDropdown := form.NewSelect(
		form.WithSelectOptions([]string{
			methodSwitchOn,
			methodSwitchOff,
			methodSwitchTog,
			methodCoverOpen,
			methodCoverClose,
			methodCoverStop,
			methodLightOn,
			methodLightOff,
			methodScriptStart,
			methodScriptStop,
			methodCustom,
		}),
		form.WithSelectHelp("Action to perform"),
	)

	customMethodInput := form.NewTextInput(
		form.WithPlaceholder("Switch.Set"),
		form.WithCharLimit(64),
		form.WithWidth(30),
		form.WithHelp("Custom RPC method name"),
	)

	paramsInput := form.NewTextInput(
		form.WithPlaceholder("{\"id\":0,\"on\":true}"),
		form.WithCharLimit(256),
		form.WithWidth(40),
		form.WithHelp("JSON parameters (optional override)"),
	)

	return CreateModel{
		ctx:               ctx,
		svc:               svc,
		styles:            editmodal.DefaultStyles().WithLabelWidth(10),
		patternDropdown:   patternDropdown,
		hourInput:         hourInput,
		minuteInput:       minuteInput,
		secondInput:       secondInput,
		methodDropdown:    methodDropdown,
		customMethodInput: customMethodInput,
		paramsInput:       paramsInput,
	}
}

// Show displays the create modal.
func (m CreateModel) Show(device string) CreateModel {
	m.device = device
	m.visible = true
	m.cursor = CreateFieldTimePattern
	m.saving = false
	m.err = nil

	// Reset inputs to defaults
	m.patternDropdown = m.patternDropdown.SetSelected(0)
	m.hourInput = m.hourInput.SetValue("12")
	m.minuteInput = m.minuteInput.SetValue("00")
	m.secondInput = m.secondInput.SetValue("0")
	m.methodDropdown = m.methodDropdown.SetSelected(0)
	m.customMethodInput = m.customMethodInput.SetValue("")
	m.paramsInput = m.paramsInput.SetValue("")
	m.selectedDays = [7]bool{} // Reset days

	// Focus pattern dropdown
	m.patternDropdown = m.patternDropdown.Focus()

	return m
}

// Hide hides the create modal.
func (m CreateModel) Hide() CreateModel {
	m.visible = false
	m.patternDropdown = m.patternDropdown.Blur()
	m.hourInput = m.hourInput.Blur()
	m.minuteInput = m.minuteInput.Blur()
	m.secondInput = m.secondInput.Blur()
	m.methodDropdown = m.methodDropdown.Blur()
	m.customMethodInput = m.customMethodInput.Blur()
	m.paramsInput = m.paramsInput.Blur()
	return m
}

// IsVisible returns whether the modal is visible.
func (m CreateModel) IsVisible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m CreateModel) SetSize(width, height int) CreateModel {
	m.width = width
	m.height = height
	return m
}

// Init returns the initial command.
func (m CreateModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m CreateModel) Update(msg tea.Msg) (CreateModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m CreateModel) handleMessage(msg tea.Msg) (CreateModel, tea.Cmd) {
	switch msg := msg.(type) {
	case CreatedMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - close modal
		m = m.Hide()
		return m, func() tea.Msg { return messages.EditClosedMsg{Saved: true} }

	// Action messages from context system
	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case messages.ToggleEnableRequestMsg:
		if handled, result := m.handleSpace(); handled {
			return result, nil
		}
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m CreateModel) handleNavigation(msg messages.NavigationMsg) (CreateModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		return m.prevField(), nil
	case messages.NavDown:
		return m.nextField(), nil
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this form
	}
	return m, nil
}

func (m CreateModel) handleKey(msg tea.KeyPressMsg) (CreateModel, tea.Cmd) {
	// Modal-specific keys not covered by action messages
	switch msg.String() {
	case keyconst.KeyEsc, "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return messages.EditClosedMsg{Saved: false} }

	case keyconst.KeyCtrlS:
		return m.save()

	case keyconst.KeyEnter:
		return m.handleEnter()

	case keyconst.KeyTab:
		return m.nextField(), nil

	case keyconst.KeyShiftTab:
		return m.prevField(), nil
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m CreateModel) handleEnter() (CreateModel, tea.Cmd) {
	// If on dropdown and expanded, collapse it
	if m.cursor == CreateFieldTimePattern && m.patternDropdown.IsExpanded() {
		m.patternDropdown = m.patternDropdown.Collapse()
		return m, nil
	}
	if m.cursor == CreateFieldMethod && m.methodDropdown.IsExpanded() {
		m.methodDropdown = m.methodDropdown.Collapse()
		return m, nil
	}
	// Otherwise save
	return m.save()
}

func (m CreateModel) handleSpace() (bool, CreateModel) {
	switch m.cursor {
	case CreateFieldTimePattern:
		if m.patternDropdown.IsExpanded() {
			m.patternDropdown = m.patternDropdown.Collapse()
		} else {
			m.patternDropdown = m.patternDropdown.Expand()
		}
		return true, m
	case CreateFieldMethod:
		if m.methodDropdown.IsExpanded() {
			m.methodDropdown = m.methodDropdown.Collapse()
		} else {
			m.methodDropdown = m.methodDropdown.Expand()
		}
		return true, m
	case CreateFieldHour, CreateFieldMinute, CreateFieldSecond, CreateFieldParams, CreateFieldCount:
		// Space handled by text inputs
	}
	return false, m
}

func (m CreateModel) updateFocusedInput(msg tea.Msg) (CreateModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.cursor {
	case CreateFieldTimePattern:
		m.patternDropdown, cmd = m.patternDropdown.Update(msg)
	case CreateFieldHour:
		m.hourInput, cmd = m.hourInput.Update(msg)
	case CreateFieldMinute:
		m.minuteInput, cmd = m.minuteInput.Update(msg)
	case CreateFieldSecond:
		m.secondInput, cmd = m.secondInput.Update(msg)
	case CreateFieldMethod:
		// Update both dropdown and custom input based on selection
		m.methodDropdown, cmd = m.methodDropdown.Update(msg)
		if m.methodDropdown.SelectedValue() == methodCustom {
			m.customMethodInput, cmd = m.customMethodInput.Update(msg)
		}
	case CreateFieldParams:
		m.paramsInput, cmd = m.paramsInput.Update(msg)
	case CreateFieldCount:
		// No-op
	}

	return m, cmd
}

func (m CreateModel) nextField() CreateModel {
	m = m.blurCurrentField()
	if m.cursor < CreateFieldCount-1 {
		m.cursor++
	}
	m = m.focusCurrentField()
	return m
}

func (m CreateModel) prevField() CreateModel {
	m = m.blurCurrentField()
	if m.cursor > 0 {
		m.cursor--
	}
	m = m.focusCurrentField()
	return m
}

func (m CreateModel) blurCurrentField() CreateModel {
	switch m.cursor {
	case CreateFieldTimePattern:
		m.patternDropdown = m.patternDropdown.Blur()
	case CreateFieldHour:
		m.hourInput = m.hourInput.Blur()
	case CreateFieldMinute:
		m.minuteInput = m.minuteInput.Blur()
	case CreateFieldSecond:
		m.secondInput = m.secondInput.Blur()
	case CreateFieldMethod:
		m.methodDropdown = m.methodDropdown.Blur()
		m.customMethodInput = m.customMethodInput.Blur()
	case CreateFieldParams:
		m.paramsInput = m.paramsInput.Blur()
	case CreateFieldCount:
		// No-op
	}
	return m
}

func (m CreateModel) focusCurrentField() CreateModel {
	switch m.cursor {
	case CreateFieldTimePattern:
		m.patternDropdown = m.patternDropdown.Focus()
	case CreateFieldHour:
		m.hourInput, _ = m.hourInput.Focus()
	case CreateFieldMinute:
		m.minuteInput, _ = m.minuteInput.Focus()
	case CreateFieldSecond:
		m.secondInput, _ = m.secondInput.Focus()
	case CreateFieldMethod:
		m.methodDropdown = m.methodDropdown.Focus()
		// Also focus custom input if custom method selected
		if m.methodDropdown.SelectedValue() == methodCustom {
			m.customMethodInput, _ = m.customMethodInput.Focus()
		}
	case CreateFieldParams:
		m.paramsInput, _ = m.paramsInput.Focus()
	case CreateFieldCount:
		// No-op
	}
	return m
}

func (m CreateModel) save() (CreateModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	// Get method and params from selection
	selectedMethod := m.methodDropdown.SelectedValue()
	var method string
	var params map[string]any

	if selectedMethod == methodCustom {
		// Use custom method input
		method = strings.TrimSpace(m.customMethodInput.Value())
		if method == "" {
			m.err = fmt.Errorf("custom RPC method is required")
			return m, nil
		}
	} else {
		// Use predefined method info
		info := getRPCMethodInfo(selectedMethod)
		method = info.Method
		params = info.Params
	}

	// Override params if user provided custom JSON
	paramsStr := strings.TrimSpace(m.paramsInput.Value())
	if paramsStr != "" {
		if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
			m.err = fmt.Errorf("invalid JSON params: %w", err)
			return m, nil
		}
	}

	// Validate and parse hour/minute/second
	hour := strings.TrimSpace(m.hourInput.Value())
	minute := strings.TrimSpace(m.minuteInput.Value())
	second := strings.TrimSpace(m.secondInput.Value())

	if hour == "" {
		hour = "0"
	}
	if minute == "" {
		minute = "0"
	}
	if second == "" {
		second = "0"
	}

	// Validate numeric values
	if err := validateTimeField(hour, 0, 23, "hour"); err != nil {
		m.err = err
		return m, nil
	}
	if err := validateTimeField(minute, 0, 59, "minute"); err != nil {
		m.err = err
		return m, nil
	}
	if err := validateTimeField(second, 0, 59, "second"); err != nil {
		m.err = err
		return m, nil
	}

	// Build timespec
	timespec, err := m.buildTimespec(second, minute, hour)
	if err != nil {
		m.err = err
		return m, nil
	}

	m.saving = true
	m.err = nil

	return m, m.createSaveCmd(timespec, method, params)
}

func validateTimeField(value string, minVal, maxVal int, name string) error {
	var num int
	if _, err := fmt.Sscanf(value, "%d", &num); err != nil {
		return fmt.Errorf("%s must be a number", name)
	}
	if num < minVal || num > maxVal {
		return fmt.Errorf("%s must be between %d and %d", name, minVal, maxVal)
	}
	return nil
}

func (m CreateModel) buildTimespec(second, minute, hour string) (string, error) {
	// Shelly timespec: ss mm hh DD MM WW
	// DD = day of month (1-31 or *)
	// MM = month (1-12 or *)
	// WW = weekday (0-6, SUN-SAT, or MON-FRI, etc.)

	pattern := m.patternDropdown.SelectedValue()
	var weekday string

	switch pattern {
	case patternDaily:
		weekday = "*"
	case patternWeekdays:
		weekday = "MON,TUE,WED,THU,FRI"
	case patternWeekends:
		weekday = "SAT,SUN"
	case patternCustom:
		weekday = m.buildCustomDays()
		if weekday == "" {
			return "", fmt.Errorf("select at least one day")
		}
	default:
		weekday = "*"
	}

	// Format: ss mm hh DD MM WW
	return fmt.Sprintf("%s %s %s * * %s", second, minute, hour, weekday), nil
}

func (m CreateModel) buildCustomDays() string {
	dayNames := []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	var selected []string
	for i, sel := range m.selectedDays {
		if sel {
			selected = append(selected, dayNames[i])
		}
	}
	return strings.Join(selected, ",")
}

func (m CreateModel) createSaveCmd(timespec, method string, params map[string]any) tea.Cmd {
	device := m.device

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		calls := []automation.ScheduleCall{
			{
				Method: method,
				Params: params,
			},
		}

		scheduleID, err := m.svc.CreateSchedule(ctx, device, true, timespec, calls)
		if err != nil {
			return CreatedMsg{Device: device, Err: err}
		}

		return CreatedMsg{Device: device, ScheduleID: scheduleID}
	}
}

// View renders the create modal.
func (m CreateModel) View() string {
	if !m.visible {
		return ""
	}

	// Build footer
	footer := "Tab: Next | Ctrl+S: Save | Esc: Cancel"
	if m.saving {
		footer = "Creating..."
	}

	// Use common modal helper
	r := rendering.NewModal(m.width, m.height, "New Schedule", footer)

	// Build content
	return r.SetContent(m.renderFormFields()).Render()
}

func (m CreateModel) renderFormFields() string {
	var content strings.Builder

	// Time Pattern field
	content.WriteString(m.renderField(CreateFieldTimePattern, "Pattern:", m.patternDropdown.View()))
	content.WriteString("\n\n")

	// Time fields (hour:minute:second)
	content.WriteString(m.renderField(CreateFieldHour, "Hour:", m.hourInput.View()))
	content.WriteString("  ")
	content.WriteString(m.renderField(CreateFieldMinute, "Min:", m.minuteInput.View()))
	content.WriteString("  ")
	content.WriteString(m.renderField(CreateFieldSecond, "Sec:", m.secondInput.View()))
	content.WriteString("\n\n")

	// Action/Method field
	content.WriteString(m.renderField(CreateFieldMethod, "Action:", m.methodDropdown.View()))
	content.WriteString("\n")

	// Show custom method input only when custom is selected
	if m.methodDropdown.SelectedValue() == methodCustom {
		content.WriteString("  ")
		content.WriteString(m.styles.Label.Render("Custom:"))
		content.WriteString(" ")
		content.WriteString(m.customMethodInput.View())
		content.WriteString("\n")
	}
	content.WriteString("\n")

	// Params field (optional override)
	content.WriteString(m.renderField(CreateFieldParams, "Params:", m.paramsInput.View()))
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("  (optional JSON override)"))

	// Error display
	if m.err != nil {
		content.WriteString("\n\n")
		msg, _ := tuierrors.FormatError(m.err)
		content.WriteString(m.styles.Error.Render(msg))
	}

	return content.String()
}

func (m CreateModel) renderField(field CreateField, label, input string) string {
	var selector, labelStr string

	if m.cursor == field {
		selector = m.styles.Selector.Render("▶ ")
		labelStr = m.styles.LabelFocus.Render(label)
	} else {
		selector = "  "
		labelStr = m.styles.Label.Render(label)
	}

	return selector + labelStr + " " + input
}

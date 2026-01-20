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

// EditField represents a field in the schedule edit form.
type EditField int

// Edit field constants.
const (
	EditFieldEnabled EditField = iota
	EditFieldTimePattern
	EditFieldHour
	EditFieldMinute
	EditFieldSecond
	EditFieldMethod
	EditFieldParams
	EditFieldCount
)

// UpdatedMsg signals that a schedule was updated.
type UpdatedMsg struct {
	Device     string
	ScheduleID int
	Err        error
}

// EditModel represents the schedule edit modal.
type EditModel struct {
	ctx        context.Context
	svc        *automation.Service
	device     string
	scheduleID int
	visible    bool
	cursor     EditField
	saving     bool
	err        error
	width      int
	height     int
	styles     editmodal.Styles

	// Form inputs
	enableToggle      bool // Simple toggle instead of dropdown
	patternDropdown   form.Select
	hourInput         form.TextInput
	minuteInput       form.TextInput
	secondInput       form.TextInput
	methodDropdown    form.Select
	customMethodInput form.TextInput
	paramsInput       form.TextInput

	// Custom day selection (for patternCustom)
	selectedDays [7]bool
}

// NewEditModel creates a new schedule edit modal.
func NewEditModel(ctx context.Context, svc *automation.Service) EditModel {
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
		form.WithHelp("JSON parameters"),
	)

	return EditModel{
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

// Show displays the edit modal with the schedule data pre-populated.
func (m EditModel) Show(device string, schedule Schedule) EditModel {
	m.device = device
	m.scheduleID = schedule.ID
	m.visible = true
	m.cursor = EditFieldEnabled
	m.saving = false
	m.err = nil

	// Populate from schedule
	m.enableToggle = schedule.Enable

	// Parse timespec to populate form fields
	m = m.populateFromTimespec(schedule.Timespec)

	// Populate method and params from first call
	if len(schedule.Calls) > 0 {
		call := schedule.Calls[0]
		m = m.populateFromCall(call)
	} else {
		// Default to first method
		m.methodDropdown = m.methodDropdown.SetSelected(0)
		m.customMethodInput = m.customMethodInput.SetValue("")
		m.paramsInput = m.paramsInput.SetValue("")
	}

	return m
}

// populateFromTimespec parses a timespec and sets form fields.
// Format: ss mm hh DD MM WW.
func (m EditModel) populateFromTimespec(timespec string) EditModel {
	parts := strings.Fields(timespec)
	if len(parts) < 6 {
		// Default values if timespec can't be parsed
		m.secondInput = m.secondInput.SetValue("0")
		m.minuteInput = m.minuteInput.SetValue("0")
		m.hourInput = m.hourInput.SetValue("12")
		m.patternDropdown = m.patternDropdown.SetSelected(0)
		return m
	}

	m.secondInput = m.secondInput.SetValue(parts[0])
	m.minuteInput = m.minuteInput.SetValue(parts[1])
	m.hourInput = m.hourInput.SetValue(parts[2])

	// Parse weekday pattern
	weekday := parts[5]
	switch {
	case weekday == "*":
		m.patternDropdown = m.patternDropdown.SetSelected(0) // Daily
	case isWeekdaysPattern(weekday):
		m.patternDropdown = m.patternDropdown.SetSelected(1) // Weekdays
	case isWeekendsPattern(weekday):
		m.patternDropdown = m.patternDropdown.SetSelected(2) // Weekends
	default:
		m.patternDropdown = m.patternDropdown.SetSelected(3) // Custom
		m = m.parseCustomDays(weekday)
	}
	return m
}

func isWeekdaysPattern(pattern string) bool {
	upper := strings.ToUpper(pattern)
	return strings.Contains(upper, "MON") && strings.Contains(upper, "FRI") && !strings.Contains(upper, "SAT")
}

func isWeekendsPattern(pattern string) bool {
	upper := strings.ToUpper(pattern)
	return (strings.Contains(upper, "SAT") || strings.Contains(upper, "SUN")) &&
		!strings.Contains(upper, "MON") && !strings.Contains(upper, "TUE")
}

func (m EditModel) parseCustomDays(pattern string) EditModel {
	m.selectedDays = [7]bool{} // Reset
	upper := strings.ToUpper(pattern)

	dayMappings := map[string]int{
		"SUN": 0, "MON": 1, "TUE": 2, "WED": 3, "THU": 4, "FRI": 5, "SAT": 6,
		"0": 0, "1": 1, "2": 2, "3": 3, "4": 4, "5": 5, "6": 6,
	}

	for day, idx := range dayMappings {
		if strings.Contains(upper, day) {
			m.selectedDays[idx] = true
		}
	}
	return m
}

// populateFromCall populates method and params fields from a schedule call.
func (m EditModel) populateFromCall(call automation.ScheduleCall) EditModel {
	// Try to match to a predefined method
	methodIndex := m.findMethodIndex(call.Method, call.Params)
	if methodIndex >= 0 {
		m.methodDropdown = m.methodDropdown.SetSelected(methodIndex)
		m.customMethodInput = m.customMethodInput.SetValue("")
	} else {
		// Custom method
		m.methodDropdown = m.methodDropdown.SetSelected(10) // methodCustom index
		m.customMethodInput = m.customMethodInput.SetValue(call.Method)
	}

	// Populate params
	if len(call.Params) > 0 {
		paramsJSON, err := json.Marshal(call.Params)
		if err == nil {
			m.paramsInput = m.paramsInput.SetValue(string(paramsJSON))
		}
	} else {
		m.paramsInput = m.paramsInput.SetValue("")
	}
	return m
}

// findMethodIndex finds the index of a predefined method matching the call.
func (m EditModel) findMethodIndex(method string, params map[string]any) int {
	// Map of method+key params to display option index
	methods := []struct {
		Method string
		Params map[string]any
		Index  int
	}{
		{"Switch.Set", map[string]any{"on": true}, 0},
		{"Switch.Set", map[string]any{"on": false}, 1},
		{"Switch.Toggle", nil, 2},
		{"Cover.Open", nil, 3},
		{"Cover.Close", nil, 4},
		{"Cover.Stop", nil, 5},
		{"Light.Set", map[string]any{"on": true}, 6},
		{"Light.Set", map[string]any{"on": false}, 7},
		{"Script.Start", nil, 8},
		{"Script.Stop", nil, 9},
	}

	for _, m := range methods {
		if m.Method == method {
			if m.Params == nil {
				return m.Index
			}
			// Check key params match
			if paramsMatch(params, m.Params) {
				return m.Index
			}
		}
	}
	return -1 // Custom
}

// paramsMatch checks if expected params are present in actual params.
func paramsMatch(actual, expected map[string]any) bool {
	for k, v := range expected {
		if av, ok := actual[k]; !ok || !valuesEqual(av, v) {
			return false
		}
	}
	return true
}

func valuesEqual(a, b any) bool {
	// Simple comparison for bool and numeric types
	switch av := a.(type) {
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
		if bv, ok := b.(int); ok {
			return av == float64(bv)
		}
	case int:
		if bv, ok := b.(int); ok {
			return av == bv
		}
		if bv, ok := b.(float64); ok {
			return float64(av) == bv
		}
	}
	return false
}

// Hide hides the edit modal.
func (m EditModel) Hide() EditModel {
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
func (m EditModel) IsVisible() bool {
	return m.visible
}

// SetSize sets the modal dimensions.
func (m EditModel) SetSize(width, height int) EditModel {
	m.width = width
	m.height = height
	return m
}

// Init returns the initial command.
func (m EditModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m EditModel) handleMessage(msg tea.Msg) (EditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case UpdatedMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m = m.Hide()
		return m, func() tea.Msg { return messages.EditClosedMsg{Saved: true} }

	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case messages.ToggleEnableRequestMsg:
		if handled, result := m.handleSpace(); handled {
			return result, nil
		}
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m.updateFocusedInput(msg)
}

func (m EditModel) handleNavigation(msg messages.NavigationMsg) (EditModel, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		return m.prevField(), nil
	case messages.NavDown:
		return m.nextField(), nil
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable
	}
	return m, nil
}

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
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

	return m.updateFocusedInput(msg)
}

func (m EditModel) handleEnter() (EditModel, tea.Cmd) {
	if m.cursor == EditFieldEnabled {
		m.enableToggle = !m.enableToggle
		return m, nil
	}
	if m.cursor == EditFieldTimePattern && m.patternDropdown.IsExpanded() {
		m.patternDropdown = m.patternDropdown.Collapse()
		return m, nil
	}
	if m.cursor == EditFieldMethod && m.methodDropdown.IsExpanded() {
		m.methodDropdown = m.methodDropdown.Collapse()
		return m, nil
	}
	return m.save()
}

func (m EditModel) handleSpace() (bool, EditModel) {
	switch m.cursor {
	case EditFieldEnabled:
		m.enableToggle = !m.enableToggle
		return true, m
	case EditFieldTimePattern:
		if m.patternDropdown.IsExpanded() {
			m.patternDropdown = m.patternDropdown.Collapse()
		} else {
			m.patternDropdown = m.patternDropdown.Expand()
		}
		return true, m
	case EditFieldMethod:
		if m.methodDropdown.IsExpanded() {
			m.methodDropdown = m.methodDropdown.Collapse()
		} else {
			m.methodDropdown = m.methodDropdown.Expand()
		}
		return true, m
	case EditFieldHour, EditFieldMinute, EditFieldSecond, EditFieldParams, EditFieldCount:
		// Space handled by text inputs
	}
	return false, m
}

func (m EditModel) updateFocusedInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.cursor {
	case EditFieldEnabled:
		// No input for toggle
	case EditFieldTimePattern:
		m.patternDropdown, cmd = m.patternDropdown.Update(msg)
	case EditFieldHour:
		m.hourInput, cmd = m.hourInput.Update(msg)
	case EditFieldMinute:
		m.minuteInput, cmd = m.minuteInput.Update(msg)
	case EditFieldSecond:
		m.secondInput, cmd = m.secondInput.Update(msg)
	case EditFieldMethod:
		m.methodDropdown, cmd = m.methodDropdown.Update(msg)
		if m.methodDropdown.SelectedValue() == methodCustom {
			m.customMethodInput, cmd = m.customMethodInput.Update(msg)
		}
	case EditFieldParams:
		m.paramsInput, cmd = m.paramsInput.Update(msg)
	case EditFieldCount:
		// No-op
	}

	return m, cmd
}

func (m EditModel) nextField() EditModel {
	m = m.blurCurrentField()
	if m.cursor < EditFieldCount-1 {
		m.cursor++
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) prevField() EditModel {
	m = m.blurCurrentField()
	if m.cursor > 0 {
		m.cursor--
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) blurCurrentField() EditModel {
	switch m.cursor {
	case EditFieldEnabled:
		// No input to blur
	case EditFieldTimePattern:
		m.patternDropdown = m.patternDropdown.Blur()
	case EditFieldHour:
		m.hourInput = m.hourInput.Blur()
	case EditFieldMinute:
		m.minuteInput = m.minuteInput.Blur()
	case EditFieldSecond:
		m.secondInput = m.secondInput.Blur()
	case EditFieldMethod:
		m.methodDropdown = m.methodDropdown.Blur()
		m.customMethodInput = m.customMethodInput.Blur()
	case EditFieldParams:
		m.paramsInput = m.paramsInput.Blur()
	case EditFieldCount:
		// No-op
	}
	return m
}

func (m EditModel) focusCurrentField() EditModel {
	switch m.cursor {
	case EditFieldEnabled:
		// No input to focus
	case EditFieldTimePattern:
		m.patternDropdown = m.patternDropdown.Focus()
	case EditFieldHour:
		m.hourInput, _ = m.hourInput.Focus()
	case EditFieldMinute:
		m.minuteInput, _ = m.minuteInput.Focus()
	case EditFieldSecond:
		m.secondInput, _ = m.secondInput.Focus()
	case EditFieldMethod:
		m.methodDropdown = m.methodDropdown.Focus()
		if m.methodDropdown.SelectedValue() == methodCustom {
			m.customMethodInput, _ = m.customMethodInput.Focus()
		}
	case EditFieldParams:
		m.paramsInput, _ = m.paramsInput.Focus()
	case EditFieldCount:
		// No-op
	}
	return m
}

func (m EditModel) save() (EditModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	// Get method and params
	selectedMethod := m.methodDropdown.SelectedValue()
	var method string
	var params map[string]any

	if selectedMethod == methodCustom {
		method = strings.TrimSpace(m.customMethodInput.Value())
		if method == "" {
			m.err = fmt.Errorf("custom RPC method is required")
			return m, nil
		}
	} else {
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

	// Validate time fields
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

	if err := validateTimeField(hour, 23, "hour"); err != nil {
		m.err = err
		return m, nil
	}
	if err := validateTimeField(minute, 59, "minute"); err != nil {
		m.err = err
		return m, nil
	}
	if err := validateTimeField(second, 59, "second"); err != nil {
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

	return m, m.createUpdateCmd(timespec, method, params)
}

func (m EditModel) buildTimespec(second, minute, hour string) (string, error) {
	pattern := m.patternDropdown.SelectedValue()
	var weekday string

	switch pattern {
	case patternDaily:
		weekday = "*"
	case patternWeekdays:
		weekday = cronWeekdays
	case patternWeekends:
		weekday = cronWeekends
	case patternCustom:
		weekday = m.buildCustomDays()
		if weekday == "" {
			return "", fmt.Errorf("select at least one day")
		}
	default:
		weekday = "*"
	}

	return fmt.Sprintf("%s %s %s * * %s", second, minute, hour, weekday), nil
}

func (m EditModel) buildCustomDays() string {
	dayNames := []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	var selected []string
	for i, sel := range m.selectedDays {
		if sel {
			selected = append(selected, dayNames[i])
		}
	}
	return strings.Join(selected, ",")
}

func (m EditModel) createUpdateCmd(timespec, method string, params map[string]any) tea.Cmd {
	device := m.device
	scheduleID := m.scheduleID
	enable := m.enableToggle

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		calls := []automation.ScheduleCall{
			{
				Method: method,
				Params: params,
			},
		}

		err := m.svc.UpdateSchedule(ctx, device, scheduleID, &enable, &timespec, calls)
		if err != nil {
			return UpdatedMsg{Device: device, ScheduleID: scheduleID, Err: err}
		}

		return UpdatedMsg{Device: device, ScheduleID: scheduleID}
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	footer := "Tab: Next | Ctrl+S: Save | Esc: Cancel"
	if m.saving {
		footer = "Saving..."
	}

	r := rendering.NewModal(m.width, m.height, fmt.Sprintf("Edit Schedule #%d", m.scheduleID), footer)
	return r.SetContent(m.renderFormFields()).Render()
}

func (m EditModel) renderFormFields() string {
	var content strings.Builder

	// Enabled toggle
	enabledValue := "Disabled"
	if m.enableToggle {
		enabledValue = "Enabled"
	}
	content.WriteString(m.renderField(EditFieldEnabled, "Enabled:", enabledValue))
	content.WriteString("\n\n")

	// Time Pattern
	content.WriteString(m.renderField(EditFieldTimePattern, "Pattern:", m.patternDropdown.View()))
	content.WriteString("\n\n")

	// Time fields
	content.WriteString(m.renderField(EditFieldHour, "Hour:", m.hourInput.View()))
	content.WriteString("  ")
	content.WriteString(m.renderField(EditFieldMinute, "Min:", m.minuteInput.View()))
	content.WriteString("  ")
	content.WriteString(m.renderField(EditFieldSecond, "Sec:", m.secondInput.View()))
	content.WriteString("\n\n")

	// Action/Method
	content.WriteString(m.renderField(EditFieldMethod, "Action:", m.methodDropdown.View()))
	content.WriteString("\n")

	if m.methodDropdown.SelectedValue() == methodCustom {
		content.WriteString("  ")
		content.WriteString(m.styles.Label.Render("Custom:"))
		content.WriteString(" ")
		content.WriteString(m.customMethodInput.View())
		content.WriteString("\n")
	}
	content.WriteString("\n")

	// Params
	content.WriteString(m.renderField(EditFieldParams, "Params:", m.paramsInput.View()))
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("  (JSON parameters)"))

	// Error display
	if m.err != nil {
		content.WriteString("\n\n")
		msg, _ := tuierrors.FormatError(m.err)
		content.WriteString(m.styles.Error.Render(msg))
	}

	return content.String()
}

func (m EditModel) renderField(field EditField, label, input string) string {
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

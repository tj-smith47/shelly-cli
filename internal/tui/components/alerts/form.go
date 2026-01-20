package alerts

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// FormMode indicates whether the form is creating or editing.
type FormMode int

const (
	// FormModeCreate is for creating a new alert.
	FormModeCreate FormMode = iota
	// FormModeEdit is for editing an existing alert.
	FormModeEdit
)

// FormField indicates which field is focused.
type FormField int

// Form field constants.
const (
	FormFieldName FormField = iota
	FormFieldDevice
	FormFieldCondition
	FormFieldAction
	FormFieldEnabled
	FormFieldCount // sentinel
)

// AlertFormSubmitMsg is sent when the form is submitted.
type AlertFormSubmitMsg struct {
	Mode        FormMode
	Name        string
	Description string
	Device      string
	Condition   string
	Action      string
	Enabled     bool
}

// AlertFormCancelMsg is sent when the form is cancelled.
type AlertFormCancelMsg struct{}

// AlertForm is a form for creating/editing alerts.
type AlertForm struct {
	mode          FormMode
	originalName  string // For edit mode, to track renames
	focused       FormField
	width, height int

	// Form fields
	nameInput      form.TextInput
	deviceInput    form.TextInput
	conditionInput form.TextInput
	actionInput    form.TextInput
	enabled        bool

	styles FormStyles
}

// FormStyles holds styles for the form.
type FormStyles struct {
	Title     lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Help      lipgloss.Style
	Error     lipgloss.Style
	Button    lipgloss.Style
	ButtonSel lipgloss.Style
	Toggle    lipgloss.Style
}

// DefaultFormStyles returns default form styles.
func DefaultFormStyles() FormStyles {
	colors := theme.GetSemanticColors()
	return FormStyles{
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			MarginBottom(1),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		Value: lipgloss.NewStyle().
			Foreground(colors.Primary),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Button: lipgloss.NewStyle().
			Foreground(colors.Text).
			Padding(0, 2),
		ButtonSel: lipgloss.NewStyle().
			Foreground(colors.Text).
			Background(colors.Highlight).
			Padding(0, 2),
		Toggle: lipgloss.NewStyle().
			Foreground(colors.Success),
	}
}

// NewAlertForm creates a new alert form.
func NewAlertForm(mode FormMode, alert *config.Alert) AlertForm {
	f := AlertForm{
		mode:    mode,
		styles:  DefaultFormStyles(),
		enabled: true, // Default to enabled for new alerts
	}

	// Initialize form fields
	f.nameInput = form.NewTextInput(
		form.WithLabel("Name"),
		form.WithPlaceholder("my-alert"),
		form.WithCharLimit(64),
		form.WithValidation(validateRequired),
		form.WithHelp("Unique identifier for this alert"),
	)

	f.deviceInput = form.NewTextInput(
		form.WithLabel("Device"),
		form.WithPlaceholder("device-name or IP"),
		form.WithValidation(validateRequired),
		form.WithHelp("Device to monitor"),
	)

	f.conditionInput = form.NewTextInput(
		form.WithLabel("Condition"),
		form.WithPlaceholder("power>100, offline, temperature>30"),
		form.WithValidation(validateCondition),
		form.WithHelp("Trigger condition: offline, online, power>N, temperature>N"),
	)

	f.actionInput = form.NewTextInput(
		form.WithLabel("Action"),
		form.WithPlaceholder("notify, webhook:URL, command:CMD"),
		form.WithHelp("Action when triggered: notify, webhook:URL, command:CMD"),
	)

	// Populate with existing alert data for edit mode
	if mode == FormModeEdit && alert != nil {
		f.originalName = alert.Name
		f.nameInput = f.nameInput.SetValue(alert.Name)
		f.deviceInput = f.deviceInput.SetValue(alert.Device)
		f.conditionInput = f.conditionInput.SetValue(alert.Condition)
		f.actionInput = f.actionInput.SetValue(alert.Action)
		f.enabled = alert.Enabled
	}

	// Focus first field
	f.nameInput, _ = f.nameInput.Focus()

	return f
}

// Init returns the initial command.
func (f AlertForm) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (f AlertForm) Update(msg tea.Msg) (AlertForm, tea.Cmd) {
	return f.handleMessage(msg)
}

func (f AlertForm) handleMessage(msg tea.Msg) (AlertForm, tea.Cmd) {
	switch msg := msg.(type) {
	// Action messages from context system
	case messages.NavigationMsg:
		return f.handleNavigation(msg)
	case messages.ToggleEnableRequestMsg:
		return f.handleSpace()
	case tea.KeyPressMsg:
		return f.handleKeyPress(msg)
	}

	// Forward to focused field
	var cmd tea.Cmd
	switch f.focused {
	case FormFieldName:
		f.nameInput, cmd = f.nameInput.Update(msg)
	case FormFieldDevice:
		f.deviceInput, cmd = f.deviceInput.Update(msg)
	case FormFieldCondition:
		f.conditionInput, cmd = f.conditionInput.Update(msg)
	case FormFieldAction:
		f.actionInput, cmd = f.actionInput.Update(msg)
	case FormFieldEnabled, FormFieldCount:
		// No text input to update for these fields
	}

	return f, cmd
}

func (f AlertForm) handleNavigation(msg messages.NavigationMsg) (AlertForm, tea.Cmd) {
	switch msg.Direction {
	case messages.NavUp:
		return f.prevField()
	case messages.NavDown:
		return f.nextField()
	case messages.NavLeft, messages.NavRight, messages.NavPageUp, messages.NavPageDown, messages.NavHome, messages.NavEnd:
		// Not applicable for this form
	}
	return f, nil
}

func (f AlertForm) handleKeyPress(msg tea.KeyPressMsg) (AlertForm, tea.Cmd) {
	// Modal-specific keys not covered by action messages
	switch msg.String() {
	case keyconst.KeyTab:
		return f.nextField()
	case keyconst.KeyShiftTab:
		return f.prevField()
	case keyconst.KeyEnter:
		return f.handleEnter()
	case keyconst.KeyCtrlS:
		return f.trySubmit()
	case keyconst.KeyEsc:
		return f, func() tea.Msg { return AlertFormCancelMsg{} }
	}

	return f.forwardToFocusedInput(msg)
}

func (f AlertForm) handleEnter() (AlertForm, tea.Cmd) {
	if f.focused == FormFieldEnabled {
		return f.trySubmit()
	}
	return f.nextField()
}

func (f AlertForm) handleSpace() (AlertForm, tea.Cmd) {
	if f.focused == FormFieldEnabled {
		f.enabled = !f.enabled
	}
	return f, nil
}

func (f AlertForm) trySubmit() (AlertForm, tea.Cmd) {
	if !f.validate() {
		return f, nil
	}
	return f, f.buildSubmitCmd()
}

func (f AlertForm) buildSubmitCmd() tea.Cmd {
	action := f.actionInput.Value()
	if action == "" {
		action = "notify"
	}
	return func() tea.Msg {
		return AlertFormSubmitMsg{
			Mode:      f.mode,
			Name:      f.nameInput.Value(),
			Device:    f.deviceInput.Value(),
			Condition: f.conditionInput.Value(),
			Action:    action,
			Enabled:   f.enabled,
		}
	}
}

func (f AlertForm) forwardToFocusedInput(msg tea.KeyPressMsg) (AlertForm, tea.Cmd) {
	var cmd tea.Cmd
	switch f.focused {
	case FormFieldName:
		f.nameInput, cmd = f.nameInput.Update(msg)
	case FormFieldDevice:
		f.deviceInput, cmd = f.deviceInput.Update(msg)
	case FormFieldCondition:
		f.conditionInput, cmd = f.conditionInput.Update(msg)
	case FormFieldAction:
		f.actionInput, cmd = f.actionInput.Update(msg)
	case FormFieldEnabled, FormFieldCount:
		// No text input to update for these fields
	}
	return f, cmd
}

func (f AlertForm) nextField() (AlertForm, tea.Cmd) {
	// Blur current
	switch f.focused {
	case FormFieldName:
		f.nameInput = f.nameInput.Blur()
	case FormFieldDevice:
		f.deviceInput = f.deviceInput.Blur()
	case FormFieldCondition:
		f.conditionInput = f.conditionInput.Blur()
	case FormFieldAction:
		f.actionInput = f.actionInput.Blur()
	case FormFieldEnabled, FormFieldCount:
		// No text input to blur for these fields
	}

	// Move to next
	f.focused = (f.focused + 1) % FormFieldCount

	// Focus new
	var cmd tea.Cmd
	switch f.focused {
	case FormFieldName:
		f.nameInput, cmd = f.nameInput.Focus()
	case FormFieldDevice:
		f.deviceInput, cmd = f.deviceInput.Focus()
	case FormFieldCondition:
		f.conditionInput, cmd = f.conditionInput.Focus()
	case FormFieldAction:
		f.actionInput, cmd = f.actionInput.Focus()
	case FormFieldEnabled, FormFieldCount:
		// No text input to focus for these fields
	}

	return f, cmd
}

func (f AlertForm) prevField() (AlertForm, tea.Cmd) {
	// Blur current
	switch f.focused {
	case FormFieldName:
		f.nameInput = f.nameInput.Blur()
	case FormFieldDevice:
		f.deviceInput = f.deviceInput.Blur()
	case FormFieldCondition:
		f.conditionInput = f.conditionInput.Blur()
	case FormFieldAction:
		f.actionInput = f.actionInput.Blur()
	case FormFieldEnabled, FormFieldCount:
		// No text input to blur for these fields
	}

	// Move to previous
	if f.focused == 0 {
		f.focused = FormFieldCount - 1
	} else {
		f.focused--
	}

	// Focus new
	var cmd tea.Cmd
	switch f.focused {
	case FormFieldName:
		f.nameInput, cmd = f.nameInput.Focus()
	case FormFieldDevice:
		f.deviceInput, cmd = f.deviceInput.Focus()
	case FormFieldCondition:
		f.conditionInput, cmd = f.conditionInput.Focus()
	case FormFieldAction:
		f.actionInput, cmd = f.actionInput.Focus()
	case FormFieldEnabled, FormFieldCount:
		// No text input to focus for these fields
	}

	return f, cmd
}

func (f AlertForm) validate() bool {
	return f.nameInput.Valid() &&
		f.deviceInput.Valid() &&
		f.conditionInput.Valid()
}

// View renders the form.
func (f AlertForm) View() string {
	title := "Create Alert"
	if f.mode == FormModeEdit {
		title = "Edit Alert"
	}

	var b strings.Builder

	b.WriteString(f.styles.Title.Render(title))
	b.WriteString("\n\n")

	// Name
	b.WriteString(f.nameInput.View())
	b.WriteString("\n\n")

	// Device
	b.WriteString(f.deviceInput.View())
	b.WriteString("\n\n")

	// Condition
	b.WriteString(f.conditionInput.View())
	b.WriteString("\n\n")

	// Action
	b.WriteString(f.actionInput.View())
	b.WriteString("\n\n")

	// Enabled toggle
	enabledLabel := f.styles.Label.Render("Enabled")
	enabledValue := "[ ]"
	if f.enabled {
		enabledValue = f.styles.Toggle.Render("[✓]")
	}
	if f.focused == FormFieldEnabled {
		enabledLabel = f.styles.ButtonSel.Render("Enabled")
	}
	b.WriteString(fmt.Sprintf("%s  %s\n", enabledLabel, enabledValue))
	b.WriteString(f.styles.Help.Render("Press Space to toggle"))
	b.WriteString("\n\n")

	// Help
	b.WriteString(f.styles.Help.Render("Tab/↓:next  Shift+Tab/↑:prev  Ctrl+S:save  Esc:cancel"))

	r := rendering.New(f.width, f.height).
		SetTitle(title).
		SetFocused(true).
		SetContent(b.String())

	return r.Render()
}

// SetSize sets the form dimensions.
func (f AlertForm) SetSize(width, height int) AlertForm {
	f.width = width
	f.height = height

	// Update input widths
	inputWidth := width - 4
	if inputWidth < 20 {
		inputWidth = 20
	}
	f.nameInput = f.nameInput.SetWidth(inputWidth)
	f.deviceInput = f.deviceInput.SetWidth(inputWidth)
	f.conditionInput = f.conditionInput.SetWidth(inputWidth)
	f.actionInput = f.actionInput.SetWidth(inputWidth)

	return f
}

// Validators

func validateRequired(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("required")
	}
	return nil
}

func validateCondition(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("required")
	}

	// Valid patterns: offline, online, metric>N, metric<N
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "offline" || s == "online" {
		return nil
	}

	if strings.Contains(s, ">") || strings.Contains(s, "<") {
		return nil
	}

	return fmt.Errorf("use: offline, online, power>N, temperature>N")
}

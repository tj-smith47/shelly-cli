// Package security provides TUI components for displaying device security settings.
package security

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/form"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// EditField represents a field in the edit form.
type EditField int

// Edit field constants.
const (
	EditFieldPassword EditField = iota
	EditFieldConfirm
	EditFieldCount
)

// PasswordStrength represents password strength levels.
type PasswordStrength int

// Password strength constants.
const (
	StrengthNone PasswordStrength = iota
	StrengthWeak
	StrengthFair
	StrengthGood
	StrengthStrong
)

// EditSaveResultMsg is an alias for the shared save result message.
type EditSaveResultMsg = messages.SaveResultMsg

// EditOpenedMsg is an alias for the shared edit opened message.
type EditOpenedMsg = messages.EditOpenedMsg

// EditClosedMsg is an alias for the shared edit closed message.
type EditClosedMsg = messages.EditClosedMsg

// EditModel represents the auth configuration edit modal.
type EditModel struct {
	ctx     context.Context
	svc     *shelly.Service
	device  string
	visible bool
	cursor  EditField
	saving  bool
	err     error
	width   int
	height  int
	styles  editmodal.Styles

	// Password strength styles (component-specific)
	strengthStyles StrengthStyles

	// Auth state
	authEnabled bool // Current auth state from device

	// Form inputs
	passwordInput form.Password
	confirmInput  form.Password

	// UI state
	disableMode bool // True when disabling auth (no password required)
}

// StrengthStyles holds password strength indicator styles.
type StrengthStyles struct {
	Weak   lipgloss.Style
	Fair   lipgloss.Style
	Good   lipgloss.Style
	Strong lipgloss.Style
}

// DefaultStrengthStyles returns default password strength styles.
func DefaultStrengthStyles() StrengthStyles {
	colors := theme.GetSemanticColors()
	return StrengthStyles{
		Weak:   lipgloss.NewStyle().Foreground(colors.Error),
		Fair:   lipgloss.NewStyle().Foreground(colors.Warning),
		Good:   lipgloss.NewStyle().Foreground(colors.Success),
		Strong: lipgloss.NewStyle().Foreground(colors.Online),
	}
}

// NewEditModel creates a new auth configuration edit modal.
func NewEditModel(ctx context.Context, svc *shelly.Service) EditModel {
	passwordInput := form.NewPassword(
		form.WithPasswordPlaceholder("Enter new password"),
		form.WithPasswordCharLimit(64),
		form.WithPasswordWidth(30),
		form.WithPasswordHelp("Min 8 characters, Ctrl+T to show/hide"),
	)

	confirmInput := form.NewPassword(
		form.WithPasswordPlaceholder("Confirm password"),
		form.WithPasswordCharLimit(64),
		form.WithPasswordWidth(30),
	)

	return EditModel{
		ctx:            ctx,
		svc:            svc,
		styles:         editmodal.DefaultStyles(),
		strengthStyles: DefaultStrengthStyles(),
		passwordInput:  passwordInput,
		confirmInput:   confirmInput,
	}
}

// Show displays the edit modal with the given device and auth status.
func (m EditModel) Show(device string, authEnabled bool) EditModel {
	m.device = device
	m.visible = true
	m.cursor = EditFieldPassword
	m.saving = false
	m.err = nil
	m.authEnabled = authEnabled
	m.disableMode = false

	// Reset inputs
	m.passwordInput = m.passwordInput.Reset()
	m.confirmInput = m.confirmInput.Reset()

	// Focus password input
	m.passwordInput, _ = m.passwordInput.Focus()
	m.confirmInput = m.confirmInput.Blur()

	return m
}

// Hide hides the edit modal.
func (m EditModel) Hide() EditModel {
	m.visible = false
	m.passwordInput = m.passwordInput.Blur()
	m.confirmInput = m.confirmInput.Blur()
	return m
}

// Visible returns whether the modal is visible.
func (m EditModel) Visible() bool {
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
	case EditSaveResultMsg:
		m.saving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Success - close modal
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: true} }

	// Action messages from context system
	case messages.NavigationMsg:
		return m.handleNavigation(msg)
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleNavigation(msg messages.NavigationMsg) (EditModel, tea.Cmd) {
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

func (m EditModel) handleKey(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	// Modal-specific keys not covered by action messages
	switch msg.String() {
	case keyconst.KeyEsc, "ctrl+[":
		m = m.Hide()
		return m, func() tea.Msg { return EditClosedMsg{Saved: false} }

	case keyconst.KeyEnter:
		return m.handleEnter()

	case keyconst.KeyTab:
		return m.nextField(), nil

	case keyconst.KeyShiftTab:
		return m.prevField(), nil

	case "d":
		// Toggle disable mode when auth is currently enabled
		if m.authEnabled && !m.saving {
			m.disableMode = !m.disableMode
			if m.disableMode {
				// Clear password fields in disable mode
				m.passwordInput = m.passwordInput.Reset()
				m.confirmInput = m.confirmInput.Reset()
			}
			return m, nil
		}
	}

	// Forward to focused input
	return m.updateFocusedInput(msg)
}

func (m EditModel) handleEnter() (EditModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	if m.disableMode {
		return m.disableAuth()
	}

	return m.save()
}

func (m EditModel) updateFocusedInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd

	switch m.cursor {
	case EditFieldPassword:
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	case EditFieldConfirm:
		m.confirmInput, cmd = m.confirmInput.Update(msg)
	case EditFieldCount:
		// No input to update
	}

	return m, cmd
}

func (m EditModel) nextField() EditModel {
	// Blur current field
	m = m.blurCurrentField()

	// Move to next
	if m.cursor < EditFieldCount-1 {
		m.cursor++
	}

	// Focus new field
	m = m.focusCurrentField()

	return m
}

func (m EditModel) prevField() EditModel {
	// Blur current field
	m = m.blurCurrentField()

	// Move to previous
	if m.cursor > 0 {
		m.cursor--
	}

	// Focus new field
	m = m.focusCurrentField()

	return m
}

func (m EditModel) blurCurrentField() EditModel {
	switch m.cursor {
	case EditFieldPassword:
		m.passwordInput = m.passwordInput.Blur()
	case EditFieldConfirm:
		m.confirmInput = m.confirmInput.Blur()
	case EditFieldCount:
		// No input
	}
	return m
}

func (m EditModel) focusCurrentField() EditModel {
	switch m.cursor {
	case EditFieldPassword:
		m.passwordInput, _ = m.passwordInput.Focus()
	case EditFieldConfirm:
		m.confirmInput, _ = m.confirmInput.Focus()
	case EditFieldCount:
		// No input
	}
	return m
}

func (m EditModel) save() (EditModel, tea.Cmd) {
	password := m.passwordInput.Value()
	confirm := m.confirmInput.Value()

	// Validate password
	if password == "" {
		m.err = fmt.Errorf("password is required")
		return m, nil
	}

	if len(password) < 8 {
		m.err = fmt.Errorf("password must be at least 8 characters")
		return m, nil
	}

	if password != confirm {
		m.err = fmt.Errorf("passwords do not match")
		return m, nil
	}

	m.saving = true
	m.err = nil

	return m, m.createSaveCmd(password)
}

func (m EditModel) disableAuth() (EditModel, tea.Cmd) {
	m.saving = true
	m.err = nil

	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.DisableAuth(ctx, m.device)
		if err != nil {
			return messages.NewSaveError(nil, err)
		}
		return messages.NewSaveResult(nil)
	}
}

func (m EditModel) createSaveCmd(password string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		// Use standard admin user and device realm
		err := m.svc.SetAuth(ctx, m.device, "admin", m.device, password)
		if err != nil {
			return messages.NewSaveError(nil, err)
		}
		return messages.NewSaveResult(nil)
	}
}

// passwordCharTypes holds character type flags for password analysis.
type passwordCharTypes struct {
	hasLower   bool
	hasUpper   bool
	hasDigit   bool
	hasSpecial bool
}

// analyzePasswordChars checks what character types are present in a password.
func analyzePasswordChars(password string) passwordCharTypes {
	var types passwordCharTypes
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			types.hasLower = true
		case unicode.IsUpper(r):
			types.hasUpper = true
		case unicode.IsDigit(r):
			types.hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			types.hasSpecial = true
		}
	}
	return types
}

// scoreFromLength returns a score based on password length.
func scoreFromLength(length int) int {
	score := 0
	if length >= 8 {
		score++
	}
	if length >= 12 {
		score++
	}
	if length >= 16 {
		score++
	}
	return score
}

// scoreFromTypes returns a score based on character type variety.
func scoreFromTypes(types passwordCharTypes) int {
	score := 0
	if types.hasLower {
		score++
	}
	if types.hasUpper {
		score++
	}
	if types.hasDigit {
		score++
	}
	if types.hasSpecial {
		score++
	}
	return score
}

// strengthFromScore converts a numeric score to PasswordStrength.
func strengthFromScore(score int) PasswordStrength {
	switch {
	case score >= 6:
		return StrengthStrong
	case score >= 4:
		return StrengthGood
	case score >= 2:
		return StrengthFair
	default:
		return StrengthWeak
	}
}

// calculateStrength calculates password strength.
func calculateStrength(password string) PasswordStrength {
	if password == "" {
		return StrengthNone
	}

	types := analyzePasswordChars(password)
	score := scoreFromLength(len(password)) + scoreFromTypes(types)
	return strengthFromScore(score)
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	// Build footer based on mode
	var footer string
	switch {
	case m.saving:
		footer = "Saving..."
	case m.disableMode:
		footer = "Enter: Confirm | Esc: Cancel"
	default:
		footer = "Enter: Save | Tab: Next | Esc: Cancel"
		if m.authEnabled {
			footer += " | d: Disable auth"
		}
	}

	// Use common modal helper
	r := rendering.NewModal(m.width, m.height, "Authentication Settings", footer)

	// Build content
	var content strings.Builder

	// Current status
	content.WriteString(m.renderStatus())
	content.WriteString("\n\n")

	if m.disableMode {
		content.WriteString(m.renderDisableConfirmation())
	} else {
		content.WriteString(m.renderPasswordForm())
	}

	return r.SetContent(content.String()).Render()
}

func (m EditModel) renderDisableConfirmation() string {
	var content strings.Builder
	content.WriteString(m.styles.ButtonDanger.Render("⚠ Disable authentication?"))
	content.WriteString("\n")
	content.WriteString(m.styles.Help.Render("This will allow anyone to control your device"))
	content.WriteString("\n\n")
	content.WriteString(m.styles.Help.Render("Press Enter to confirm, Esc to cancel"))
	return content.String()
}

func (m EditModel) renderPasswordForm() string {
	var content strings.Builder

	// Password field
	content.WriteString(m.renderField(EditFieldPassword, "Password:", m.passwordInput.View()))
	content.WriteString("\n")

	// Password strength indicator
	strength := calculateStrength(m.passwordInput.Value())
	content.WriteString(m.renderStrength(strength))
	content.WriteString("\n\n")

	// Confirm field
	content.WriteString(m.renderField(EditFieldConfirm, "Confirm:", m.confirmInput.View()))
	content.WriteString("\n\n")

	// Error display
	if m.err != nil {
		msg, _ := tuierrors.FormatError(m.err)
		content.WriteString(m.styles.Error.Render(msg))
	}

	return content.String()
}

func (m EditModel) renderStatus() string {
	var content strings.Builder

	content.WriteString(m.styles.Label.Render("Current status: "))
	if m.authEnabled {
		content.WriteString(m.styles.StatusOn.Render("● Protected"))
	} else {
		content.WriteString(m.styles.StatusOff.Render("○ UNPROTECTED"))
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

func (m EditModel) renderStrength(strength PasswordStrength) string {
	// Indent to align with input
	indent := "                  " // 2 (selector) + 14 (label) + 2 (space)

	switch strength {
	case StrengthNone:
		return indent + m.styles.Help.Render("Enter a password")
	case StrengthWeak:
		return indent + m.strengthStyles.Weak.Render("█░░░ Weak")
	case StrengthFair:
		return indent + m.strengthStyles.Fair.Render("██░░ Fair")
	case StrengthGood:
		return indent + m.strengthStyles.Good.Render("███░ Good")
	case StrengthStrong:
		return indent + m.strengthStyles.Strong.Render("████ Strong")
	default:
		return ""
	}
}

// Package wifi provides TUI components for managing device WiFi settings.
package wifi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
	"github.com/tj-smith47/shelly-cli/internal/tui/keyconst"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

// EditMode represents which section is being edited.
type EditMode int

// Edit mode constants.
const (
	EditModeStation EditMode = iota
	EditModeAP
)

// Field identifies which field is focused.
type Field int

// Field constants.
const (
	FieldSSID Field = iota
	FieldPassword
	FieldEnabled
	FieldCount
)

// SaveResultMsg is an alias for the shared save result message.
type SaveResultMsg = messages.SaveResultMsg

// EditModel provides a modal form for editing WiFi settings.
type EditModel struct {
	editmodal.Base

	mode     EditMode
	networks []network.WiFiNetworkFull

	// Station fields
	staSSID     textinput.Model
	staPassword textinput.Model
	staEnabled  bool

	// AP fields
	apSSID     textinput.Model
	apPassword textinput.Model
	apEnabled  bool

	// Original values for cancel
	origConfig *network.WiFiConfigFull
}

// NewEditModel creates a new WiFi edit modal.
func NewEditModel(ctx context.Context, svc *shelly.Service) EditModel {
	colors := theme.GetSemanticColors()

	// Create text inputs with proper styling
	inputStyles := textinput.DefaultStyles(true)
	inputStyles.Focused.Prompt = inputStyles.Focused.Prompt.Foreground(colors.Highlight)
	inputStyles.Focused.Text = inputStyles.Focused.Text.Foreground(colors.Text)
	inputStyles.Focused.Placeholder = inputStyles.Focused.Placeholder.Foreground(colors.Muted)
	inputStyles.Blurred.Prompt = inputStyles.Blurred.Prompt.Foreground(colors.Muted)
	inputStyles.Blurred.Text = inputStyles.Blurred.Text.Foreground(colors.Text)
	inputStyles.Blurred.Placeholder = inputStyles.Blurred.Placeholder.Foreground(colors.Muted)

	staSSID := textinput.New()
	staSSID.Placeholder = "Network name"
	staSSID.CharLimit = 32
	staSSID.SetWidth(30)
	staSSID.SetStyles(inputStyles)

	staPassword := textinput.New()
	staPassword.Placeholder = "Password"
	staPassword.EchoMode = textinput.EchoPassword
	staPassword.CharLimit = 64
	staPassword.SetWidth(30)
	staPassword.SetStyles(inputStyles)

	apSSID := textinput.New()
	apSSID.Placeholder = "AP name"
	apSSID.CharLimit = 32
	apSSID.SetWidth(30)
	apSSID.SetStyles(inputStyles)

	apPassword := textinput.New()
	apPassword.Placeholder = "Password"
	apPassword.EchoMode = textinput.EchoPassword
	apPassword.CharLimit = 64
	apPassword.SetWidth(30)
	apPassword.SetStyles(inputStyles)

	return EditModel{
		Base: editmodal.Base{
			Ctx:    ctx,
			Svc:    svc,
			Styles: editmodal.DefaultStyles().WithLabelWidth(12),
		},
		mode:        EditModeStation,
		staSSID:     staSSID,
		staPassword: staPassword,
		apSSID:      apSSID,
		apPassword:  apPassword,
	}
}

// Init returns the initial command.
func (m EditModel) Init() tea.Cmd {
	return nil
}

// Show makes the modal visible and loads current config.
func (m EditModel) Show(device string, config *network.WiFiConfigFull, networks []network.WiFiNetworkFull) EditModel {
	m.Base.Show(device, int(FieldCount))
	m.origConfig = config
	m.networks = networks
	m.mode = EditModeStation

	// Populate fields from config
	if config != nil {
		if config.STA != nil {
			m.staSSID.SetValue(config.STA.SSID)
			m.staEnabled = config.STA.Enabled
		}
		if config.AP != nil {
			m.apSSID.SetValue(config.AP.SSID)
			m.apEnabled = config.AP.Enabled
		}
	}

	// Clear passwords (never show existing)
	m.staPassword.SetValue("")
	m.apPassword.SetValue("")

	// Focus first field
	m.staSSID.Focus()

	return m
}

// Hide hides the modal.
func (m EditModel) Hide() EditModel {
	m.Base.Hide()
	m.staSSID.Blur()
	m.staPassword.Blur()
	m.apSSID.Blur()
	m.apPassword.Blur()
	return m
}

// IsVisible returns whether the modal is visible.
func (m EditModel) IsVisible() bool {
	return m.Visible()
}

// SetSize sets the modal dimensions.
func (m EditModel) SetSize(width, height int) EditModel {
	m.Base.SetSize(width, height)
	return m
}

// Update handles messages.
func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	if !m.Visible() {
		return m, nil
	}

	return m.handleMessage(msg)
}

func (m EditModel) handleMessage(msg tea.Msg) (EditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.SaveResultMsg:
		// Use Base.HandleSaveResult for state management but ignore the
		// returned cmd since the parent model handles close detection
		// and EditClosedMsg sending via the IsVisible() check.
		m.HandleSaveResult(msg)
		return m, nil

	// Action messages from context system
	case messages.NavigationMsg:
		if m.Base.Saving {
			return m, nil
		}
		return m.handleNavigation(msg)
	case messages.ToggleEnableRequestMsg:
		if m.Base.Saving {
			return m, nil
		}
		if Field(m.Cursor) == FieldEnabled {
			if m.mode == EditModeStation {
				m.staEnabled = !m.staEnabled
			} else {
				m.apEnabled = !m.apEnabled
			}
		}
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	// Update focused text input
	m, cmd := m.updateFocusedInput(msg)
	return m, cmd
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

func (m EditModel) updateFocusedInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd
	switch {
	case Field(m.Cursor) == FieldSSID && m.mode == EditModeStation:
		m.staSSID, cmd = m.staSSID.Update(msg)
	case Field(m.Cursor) == FieldPassword && m.mode == EditModeStation:
		m.staPassword, cmd = m.staPassword.Update(msg)
	case Field(m.Cursor) == FieldSSID && m.mode == EditModeAP:
		m.apSSID, cmd = m.apSSID.Update(msg)
	case Field(m.Cursor) == FieldPassword && m.mode == EditModeAP:
		m.apPassword, cmd = m.apPassword.Update(msg)
	}
	return m, cmd
}

func (m EditModel) handleKeyPress(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	if m.Base.Saving {
		return m, nil
	}

	// WiFi uses Tab/Shift+Tab for mode switching (STA/AP), so handle keys
	// manually instead of using HandleKey which maps Tab to ActionNext.
	switch msg.String() {
	case "esc", "ctrl+[":
		return m.Hide(), nil

	case keyconst.KeyTab, keyconst.KeyShiftTab:
		// Switch between Station and AP modes
		if m.mode == EditModeStation {
			m.mode = EditModeAP
			m = m.blurAllInputs()
			m.apSSID.Focus()
		} else {
			m.mode = EditModeStation
			m = m.blurAllInputs()
			m.staSSID.Focus()
		}
		m.Cursor = int(FieldSSID)
		return m, nil

	case keyconst.KeyEnter:
		if Field(m.Cursor) == FieldEnabled {
			// Toggle enabled state
			if m.mode == EditModeStation {
				m.staEnabled = !m.staEnabled
			} else {
				m.apEnabled = !m.apEnabled
			}
			return m, nil
		}
		// Move to next field
		return m.nextField(), nil

	case keyconst.KeyCtrlS:
		// Save
		return m, m.save()
	}

	return m, nil
}

func (m EditModel) blurAllInputs() EditModel {
	m.staSSID.Blur()
	m.staPassword.Blur()
	m.apSSID.Blur()
	m.apPassword.Blur()
	return m
}

func (m EditModel) nextField() EditModel {
	m = m.blurAllInputs()
	m.Cursor = (m.Cursor + 1) % int(FieldCount)
	m = m.focusCurrentField()
	return m
}

func (m EditModel) prevField() EditModel {
	m = m.blurAllInputs()
	if m.Cursor == 0 {
		m.Cursor = int(FieldCount) - 1
	} else {
		m.Cursor--
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) focusCurrentField() EditModel {
	switch Field(m.Cursor) {
	case FieldSSID:
		if m.mode == EditModeStation {
			m.staSSID.Focus()
		} else {
			m.apSSID.Focus()
		}
	case FieldPassword:
		if m.mode == EditModeStation {
			m.staPassword.Focus()
		} else {
			m.apPassword.Focus()
		}
	case FieldEnabled, FieldCount:
		// Toggle fields don't need text input focus
	}
	return m
}

func (m EditModel) save() tea.Cmd {
	m.StartSave()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.Ctx, 30*time.Second)
		defer cancel()

		// Save station config
		if m.staSSID.Value() != "" {
			err := m.Svc.SetWiFiStation(ctx, m.Base.Device, m.staSSID.Value(), m.staPassword.Value(), m.staEnabled)
			if err != nil {
				return messages.NewSaveError(nil, fmt.Errorf("station: %w", err))
			}
		}

		// Save AP config
		if m.apSSID.Value() != "" {
			err := m.Svc.SetWiFiAP(ctx, m.Base.Device, m.apSSID.Value(), m.apPassword.Value(), m.apEnabled)
			if err != nil {
				return messages.NewSaveError(nil, fmt.Errorf("AP: %w", err))
			}
		}

		return messages.NewSaveResultWithMessage(nil, "WiFi settings saved")
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.Visible() {
		return ""
	}

	// Build footer with keybindings
	footer := m.RenderSavingFooter("Tab: Mode | ↑↓: Navigate | Space: Toggle | Ctrl+S: Save | Esc: Cancel")

	// Build content
	var content strings.Builder

	// Tab bar
	content.WriteString(m.renderTabs())
	content.WriteString("\n\n")

	// Current mode fields
	if m.mode == EditModeStation {
		content.WriteString(m.renderStationFields())
	} else {
		content.WriteString(m.renderAPFields())
	}

	// Error
	if errStr := m.RenderError(); errStr != "" {
		content.WriteString("\n\n")
		content.WriteString(errStr)
	}

	return m.RenderModal("Edit WiFi Settings", content.String(), footer)
}

func (m EditModel) renderTabs() string {
	var staTab, apTab string
	if m.mode == EditModeStation {
		staTab = m.Styles.TabActive.Render("Station (STA)")
		apTab = m.Styles.Tab.Render("Access Point (AP)")
	} else {
		staTab = m.Styles.Tab.Render("Station (STA)")
		apTab = m.Styles.TabActive.Render("Access Point (AP)")
	}
	return staTab + " " + apTab
}

func (m EditModel) renderStationFields() string {
	var content strings.Builder

	// SSID
	content.WriteString(m.renderFieldRow("SSID", m.staSSID.View()))
	content.WriteString("\n")

	// Password
	content.WriteString(m.renderFieldRow("Password", m.staPassword.View()))
	content.WriteString("\n")

	// Enabled toggle
	content.WriteString(m.renderToggleRow("Enabled", m.staEnabled, Field(m.Cursor) == FieldEnabled))

	// Show available networks if any
	if len(m.networks) > 0 && Field(m.Cursor) == FieldSSID {
		content.WriteString("\n\n")
		content.WriteString(m.Styles.Help.Render(fmt.Sprintf("Available: %s", m.networksHint())))
	}

	return content.String()
}

func (m EditModel) renderAPFields() string {
	var content strings.Builder

	// SSID
	content.WriteString(m.renderFieldRow("SSID", m.apSSID.View()))
	content.WriteString("\n")

	// Password
	content.WriteString(m.renderFieldRow("Password", m.apPassword.View()))
	content.WriteString("\n")

	// Enabled toggle
	content.WriteString(m.renderToggleRow("Enabled", m.apEnabled, Field(m.Cursor) == FieldEnabled))

	return content.String()
}

func (m EditModel) renderFieldRow(label, inputView string) string {
	labelStr := m.Styles.Label.Render(label + ":")
	return labelStr + inputView
}

func (m EditModel) renderToggleRow(label string, enabled, focused bool) string {
	labelStr := m.Styles.Label.Render(label + ":")

	var toggle string
	if enabled {
		toggle = m.Styles.StatusOn.Render("[●] On")
	} else {
		toggle = m.Styles.Muted.Render("[○] Off")
	}

	if focused {
		toggle = m.Styles.Selected.Render(toggle)
	}

	return labelStr + toggle
}

func (m EditModel) networksHint() string {
	if len(m.networks) == 0 {
		return "none found"
	}
	names := make([]string, 0, min(3, len(m.networks)))
	for i := 0; i < len(m.networks) && i < 3; i++ {
		names = append(names, m.networks[i].SSID)
	}
	hint := strings.Join(names, ", ")
	if len(m.networks) > 3 {
		hint += fmt.Sprintf(" (+%d more)", len(m.networks)-3)
	}
	return hint
}

// Saving returns whether a save is in progress.
func (m EditModel) Saving() bool {
	return m.Base.Saving
}

// Device returns the current device.
func (m EditModel) Device() string {
	return m.Base.Device
}

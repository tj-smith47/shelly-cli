// Package wifi provides TUI components for managing device WiFi settings.
package wifi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
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

// SaveResultMsg signals that a save operation completed.
type SaveResultMsg struct {
	Success bool
	Error   error
	Message string
}

// EditModel provides a modal form for editing WiFi settings.
type EditModel struct {
	ctx      context.Context
	svc      *shelly.Service
	device   string
	mode     EditMode
	field    Field
	visible  bool
	saving   bool
	err      error
	width    int
	height   int
	styles   EditStyles
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

// EditStyles holds styles for the edit modal.
type EditStyles struct {
	Container    lipgloss.Style
	Title        lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	Focused      lipgloss.Style
	Unfocused    lipgloss.Style
	Toggle       lipgloss.Style
	ToggleActive lipgloss.Style
	Error        lipgloss.Style
	Help         lipgloss.Style
	Selected     lipgloss.Style
	Tab          lipgloss.Style
	TabActive    lipgloss.Style
}

// DefaultEditStyles returns the default styles for the edit modal.
func DefaultEditStyles() EditStyles {
	colors := theme.GetSemanticColors()
	return EditStyles{
		Container: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Highlight).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight).
			MarginBottom(1),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Width(12),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Focused: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.Highlight).
			Padding(0, 1),
		Unfocused: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.TableBorder).
			Padding(0, 1),
		Toggle: lipgloss.NewStyle().
			Foreground(colors.Muted),
		ToggleActive: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight),
		Tab: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Padding(0, 2),
		TabActive: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Padding(0, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colors.Highlight),
	}
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
		ctx:         ctx,
		svc:         svc,
		mode:        EditModeStation,
		field:       FieldSSID,
		styles:      DefaultEditStyles(),
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
	m.device = device
	m.visible = true
	m.err = nil
	m.saving = false
	m.origConfig = config
	m.networks = networks
	m.mode = EditModeStation
	m.field = FieldSSID

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
	m.visible = false
	m.staSSID.Blur()
	m.staPassword.Blur()
	m.apSSID.Blur()
	m.apPassword.Blur()
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

// Update handles messages.
func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case SaveResultMsg:
		m.saving = false
		if msg.Success {
			m = m.Hide()
		} else {
			m.err = msg.Error
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	// Update focused text input
	m, cmd := m.updateFocusedInput(msg)
	return m, cmd
}

func (m EditModel) updateFocusedInput(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmd tea.Cmd
	switch {
	case m.field == FieldSSID && m.mode == EditModeStation:
		m.staSSID, cmd = m.staSSID.Update(msg)
	case m.field == FieldPassword && m.mode == EditModeStation:
		m.staPassword, cmd = m.staPassword.Update(msg)
	case m.field == FieldSSID && m.mode == EditModeAP:
		m.apSSID, cmd = m.apSSID.Update(msg)
	case m.field == FieldPassword && m.mode == EditModeAP:
		m.apPassword, cmd = m.apPassword.Update(msg)
	}
	return m, cmd
}

func (m EditModel) handleKeyPress(msg tea.KeyPressMsg) (EditModel, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	switch msg.String() {
	case "esc", "ctrl+[":
		return m.Hide(), nil

	case "tab", "shift+tab":
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
		m.field = FieldSSID
		return m, nil

	case "down", "j":
		return m.nextField(), nil

	case "up", "k":
		return m.prevField(), nil

	case "enter":
		if m.field == FieldEnabled {
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

	case " ":
		if m.field == FieldEnabled {
			// Toggle enabled state
			if m.mode == EditModeStation {
				m.staEnabled = !m.staEnabled
			} else {
				m.apEnabled = !m.apEnabled
			}
			return m, nil
		}

	case "ctrl+s":
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
	m.field = (m.field + 1) % FieldCount
	m = m.focusCurrentField()
	return m
}

func (m EditModel) prevField() EditModel {
	m = m.blurAllInputs()
	if m.field == 0 {
		m.field = FieldCount - 1
	} else {
		m.field--
	}
	m = m.focusCurrentField()
	return m
}

func (m EditModel) focusCurrentField() EditModel {
	switch m.field {
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
	m.saving = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		// Save station config
		if m.staSSID.Value() != "" {
			err := m.svc.SetWiFiStation(ctx, m.device, m.staSSID.Value(), m.staPassword.Value(), m.staEnabled)
			if err != nil {
				return SaveResultMsg{Success: false, Error: fmt.Errorf("station: %w", err)}
			}
		}

		// Save AP config
		if m.apSSID.Value() != "" {
			err := m.svc.SetWiFiAP(ctx, m.device, m.apSSID.Value(), m.apPassword.Value(), m.apEnabled)
			if err != nil {
				return SaveResultMsg{Success: false, Error: fmt.Errorf("AP: %w", err)}
			}
		}

		return SaveResultMsg{Success: true, Message: "WiFi settings saved"}
	}
}

// View renders the edit modal.
func (m EditModel) View() string {
	if !m.visible {
		return ""
	}

	// Build footer with keybindings
	footer := "Tab: Mode | ↑↓: Navigate | Space: Toggle | Ctrl+S: Save | Esc: Cancel"
	if m.saving {
		footer = "Saving..."
	}

	// Use common modal helper
	r := rendering.NewModal(m.width, m.height, "Edit WiFi Settings", footer)

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
	if m.err != nil {
		content.WriteString("\n\n")
		content.WriteString(m.styles.Error.Render("Error: " + m.err.Error()))
	}

	return r.SetContent(content.String()).Render()
}

func (m EditModel) renderTabs() string {
	var staTab, apTab string
	if m.mode == EditModeStation {
		staTab = m.styles.TabActive.Render("Station (STA)")
		apTab = m.styles.Tab.Render("Access Point (AP)")
	} else {
		staTab = m.styles.Tab.Render("Station (STA)")
		apTab = m.styles.TabActive.Render("Access Point (AP)")
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
	content.WriteString(m.renderToggleRow("Enabled", m.staEnabled, m.field == FieldEnabled))

	// Show available networks if any
	if len(m.networks) > 0 && m.field == FieldSSID {
		content.WriteString("\n\n")
		content.WriteString(m.styles.Help.Render(fmt.Sprintf("Available: %s", m.networksHint())))
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
	content.WriteString(m.renderToggleRow("Enabled", m.apEnabled, m.field == FieldEnabled))

	return content.String()
}

func (m EditModel) renderFieldRow(label, inputView string) string {
	labelStr := m.styles.Label.Render(label + ":")
	return labelStr + inputView
}

func (m EditModel) renderToggleRow(label string, enabled, focused bool) string {
	labelStr := m.styles.Label.Render(label + ":")

	var toggle string
	if enabled {
		toggle = m.styles.ToggleActive.Render("[●] On")
	} else {
		toggle = m.styles.Toggle.Render("[○] Off")
	}

	if focused {
		toggle = m.styles.Selected.Render(toggle)
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
	return m.saving
}

// Device returns the current device.
func (m EditModel) Device() string {
	return m.device
}

// Package security provides TUI components for displaying device security settings.
// This includes authentication status, debug mode, and device visibility.
package security

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Deps holds the dependencies for the Security component.
type Deps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d Deps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	return nil
}

// StatusLoadedMsg signals that security status was loaded.
type StatusLoadedMsg struct {
	Status *shelly.TUISecurityStatus
	Err    error
}

// Model displays security settings for a device.
type Model struct {
	ctx        context.Context
	svc        *shelly.Service
	device     string
	status     *shelly.TUISecurityStatus
	loading    bool
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int // 1-based panel index for Shift+N hotkey hint
	styles     Styles
	loader     loading.Model
}

// Styles holds styles for the Security component.
type Styles struct {
	Enabled   lipgloss.Style
	Disabled  lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Highlight lipgloss.Style
	Error     lipgloss.Style
	Muted     lipgloss.Style
	Section   lipgloss.Style
	Warning   lipgloss.Style
	Danger    lipgloss.Style
}

// DefaultStyles returns the default styles for the Security component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Enabled: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Highlight: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Section: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Danger: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
	}
}

// New creates a new Security model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("security: invalid deps: %v", err))
	}

	return Model{
		ctx:    deps.Ctx,
		svc:    deps.Svc,
		styles: DefaultStyles(),
		loader: loading.New(
			loading.WithMessage("Loading security settings..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display security settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.status = nil
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		status, err := m.svc.GetTUISecurityStatus(ctx, m.device)
		return StatusLoadedMsg{
			Status: status,
			Err:    err,
		}
	}
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	// Update loader size for proper centering
	m.loader = m.loader.SetSize(width-4, height-4)
	return m
}

// SetFocused sets the focus state.
func (m Model) SetFocused(focused bool) Model {
	m.focused = focused
	return m
}

// SetPanelIndex sets the 1-based panel index for Shift+N hotkey hint.
func (m Model) SetPanelIndex(index int) Model {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Forward tick messages to loader when loading
	if m.loading {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		// Continue processing StatusLoadedMsg even during loading
		if _, ok := msg.(StatusLoadedMsg); !ok {
			if cmd != nil {
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case StatusLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.status = msg.Status
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if msg.String() == "r" && !m.loading && m.device != "" {
		m.loading = true
		return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
	}

	return m, nil
}

// View renders the Security component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Security").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	if m.device == "" {
		r.SetContent(m.styles.Muted.Render("No device selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.loader.View())
		return r.Render()
	}

	if m.err != nil {
		r.SetContent(m.styles.Error.Render("Error: " + m.err.Error()))
		return r.Render()
	}

	if m.status == nil {
		r.SetContent(m.styles.Muted.Render("No security data available"))
		return r.Render()
	}

	var content strings.Builder

	// Authentication Section
	content.WriteString(m.renderAuth())
	content.WriteString("\n\n")

	// Device Visibility Section
	content.WriteString(m.renderVisibility())
	content.WriteString("\n\n")

	// Debug Mode Section
	content.WriteString(m.renderDebug())

	// Help text
	content.WriteString("\n\n")
	content.WriteString(m.styles.Muted.Render("r: refresh"))

	r.SetContent(content.String())
	return r.Render()
}

func (m Model) renderAuth() string {
	var content strings.Builder

	content.WriteString(m.styles.Section.Render("Authentication"))
	content.WriteString("\n")

	if m.status.AuthEnabled {
		content.WriteString("  " + m.styles.Enabled.Render("● Protected"))
		content.WriteString("\n")
		content.WriteString("  " + m.styles.Muted.Render("Device requires password for access"))
	} else {
		content.WriteString("  " + m.styles.Danger.Render("○ UNPROTECTED"))
		content.WriteString("\n")
		content.WriteString("  " + m.styles.Warning.Render("⚠ No password set - anyone can control this device"))
	}

	return content.String()
}

func (m Model) renderVisibility() string {
	var content strings.Builder

	content.WriteString(m.styles.Section.Render("Device Visibility"))
	content.WriteString("\n")

	// Discoverable
	content.WriteString("  " + m.styles.Label.Render("Discoverable: "))
	if m.status.Discoverable {
		content.WriteString(m.styles.Enabled.Render("Yes"))
		content.WriteString(m.styles.Muted.Render(" (visible on network)"))
	} else {
		content.WriteString(m.styles.Disabled.Render("No"))
		content.WriteString(m.styles.Muted.Render(" (hidden)"))
	}
	content.WriteString("\n")

	// Eco Mode
	content.WriteString("  " + m.styles.Label.Render("Eco Mode:     "))
	if m.status.EcoMode {
		content.WriteString(m.styles.Enabled.Render("Enabled"))
		content.WriteString(m.styles.Muted.Render(" (reduced power)"))
	} else {
		content.WriteString(m.styles.Disabled.Render("Disabled"))
	}

	return content.String()
}

func (m Model) renderDebug() string {
	var content strings.Builder

	content.WriteString(m.styles.Section.Render("Debug Logging"))
	content.WriteString("\n")

	hasDebug := m.status.DebugMQTT || m.status.DebugWS || m.status.DebugUDP

	if !hasDebug {
		content.WriteString("  " + m.styles.Muted.Render("○ No debug logging enabled"))
		return content.String()
	}

	content.WriteString("  " + m.styles.Warning.Render("● Debug logging active"))
	content.WriteString("\n")

	if m.status.DebugMQTT {
		content.WriteString("  " + m.styles.Label.Render("  MQTT:      "))
		content.WriteString(m.styles.Enabled.Render("Enabled"))
		content.WriteString("\n")
	}

	if m.status.DebugWS {
		content.WriteString("  " + m.styles.Label.Render("  WebSocket: "))
		content.WriteString(m.styles.Enabled.Render("Enabled"))
		content.WriteString("\n")
	}

	if m.status.DebugUDP {
		content.WriteString("  " + m.styles.Label.Render("  UDP:       "))
		content.WriteString(m.styles.Enabled.Render("Enabled"))
		if m.status.DebugUDPAddr != "" {
			content.WriteString(m.styles.Muted.Render(" → " + m.status.DebugUDPAddr))
		}
	}

	return strings.TrimSuffix(content.String(), "\n")
}

// Status returns the current security status.
func (m Model) Status() *shelly.TUISecurityStatus {
	return m.status
}

// Device returns the current device address.
func (m Model) Device() string {
	return m.device
}

// Loading returns whether the component is loading.
func (m Model) Loading() bool {
	return m.loading
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Refresh triggers a refresh of the security data.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
}

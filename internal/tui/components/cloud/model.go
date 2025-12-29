// Package cloud provides TUI components for managing device cloud settings.
package cloud

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

// Deps holds the dependencies for the Cloud component.
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

// StatusLoadedMsg signals that cloud status was loaded.
type StatusLoadedMsg struct {
	Status  *shelly.CloudStatus
	Config  map[string]any
	Enabled bool
	Server  string
	Err     error
}

// ToggleResultMsg signals the result of a toggle operation.
type ToggleResultMsg struct {
	Enabled bool
	Err     error
}

// Model displays cloud settings for a device.
type Model struct {
	ctx        context.Context
	svc        *shelly.Service
	device     string
	connected  bool
	enabled    bool
	server     string
	loading    bool
	toggling   bool
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int // 1-based panel index for Shift+N hotkey hint
	styles     Styles
	loader     loading.Model

	// Edit modal
	editModel EditModel
}

// Styles holds styles for the Cloud component.
type Styles struct {
	Connected    lipgloss.Style
	Disconnected lipgloss.Style
	Enabled      lipgloss.Style
	Disabled     lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	Error        lipgloss.Style
	Muted        lipgloss.Style
	Title        lipgloss.Style
}

// DefaultStyles returns the default styles for the Cloud component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Connected: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disconnected: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Enabled: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// New creates a new Cloud model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("cloud: invalid deps: %v", err))
	}

	return Model{
		ctx:     deps.Ctx,
		svc:     deps.Svc,
		loading: false,
		styles:  DefaultStyles(),
		loader: loading.New(
			loading.WithMessage("Loading cloud status..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
		editModel: NewEditModel(deps.Ctx, deps.Svc),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display cloud settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.connected = false
	m.enabled = false
	m.server = ""
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
}

func (m Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetCloudStatus(ctx, m.device)
		if err != nil {
			return StatusLoadedMsg{Err: err}
		}

		config, configErr := m.svc.GetCloudConfig(ctx, m.device)
		if configErr != nil {
			return StatusLoadedMsg{Status: status, Err: configErr}
		}

		// Extract enabled and server from config
		var enabled bool
		var server string
		if e, ok := config["enable"].(*bool); ok && e != nil {
			enabled = *e
		} else if e, ok := config["enable"].(bool); ok {
			enabled = e
		}
		if s, ok := config["server"].(*string); ok && s != nil {
			server = *s
		} else if s, ok := config["server"].(string); ok {
			server = s
		}

		return StatusLoadedMsg{
			Status:  status,
			Config:  config,
			Enabled: enabled,
			Server:  server,
		}
	}
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	m.editModel = m.editModel.SetSize(width, height)
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
	// Forward messages to edit modal when visible
	if m.editModel.Visible() {
		return m.handleEditModalUpdate(msg)
	}

	// Forward tick messages to loader when loading
	if m.loading {
		return m.handleLoadingUpdate(msg)
	}

	switch msg := msg.(type) {
	case StatusLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		if msg.Status != nil {
			m.connected = msg.Status.Connected
		}
		m.enabled = msg.Enabled
		m.server = msg.Server
		return m, nil

	case ToggleResultMsg:
		m.toggling = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.enabled = msg.Enabled
		// Refresh to get updated connection status
		m.loading = true
		return m, tea.Batch(m.loader.Tick(), m.fetchStatus())

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleEditModalUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.editModel, cmd = m.editModel.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	// Check for EditClosedMsg to refresh data
	if closedMsg, ok := msg.(EditClosedMsg); ok && closedMsg.Saved && m.device != "" {
		m.loading = true
		cmds = append(cmds, m.loader.Tick(), m.fetchStatus())
	}
	return m, tea.Batch(cmds...)
}

func (m Model) handleLoadingUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.loader, cmd = m.loader.Update(msg)

	// Process StatusLoadedMsg even during loading
	if loadedMsg, ok := msg.(StatusLoadedMsg); ok {
		m.loading = false
		if loadedMsg.Err != nil {
			m.err = loadedMsg.Err
			return m, nil
		}
		if loadedMsg.Status != nil {
			m.connected = loadedMsg.Status.Connected
		}
		m.enabled = loadedMsg.Enabled
		m.server = loadedMsg.Server
		return m, nil
	}

	return m, cmd
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		// Open cloud configuration modal
		if m.device != "" && !m.loading && !m.toggling {
			m.editModel = m.editModel.Show(m.device, m.connected, m.enabled, m.server)
			return m, func() tea.Msg { return EditOpenedMsg{} }
		}
	case "t", "enter":
		if !m.toggling && !m.loading && m.device != "" {
			m.toggling = true
			m.err = nil
			return m, m.toggleCloud()
		}
	case "r":
		if !m.loading && m.device != "" {
			m.loading = true
			return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
		}
	}

	return m, nil
}

func (m Model) toggleCloud() tea.Cmd {
	newEnabled := !m.enabled
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		err := m.svc.SetCloudEnabled(ctx, m.device, newEnabled)
		if err != nil {
			return ToggleResultMsg{Err: err}
		}

		return ToggleResultMsg{Enabled: newEnabled}
	}
}

// View renders the Cloud component.
func (m Model) View() string {
	// If edit modal is visible, render it as overlay
	if m.editModel.Visible() {
		return m.editModel.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Cloud").
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

	var content strings.Builder

	// Connection status
	content.WriteString(m.styles.Label.Render("Status:  "))
	if m.connected {
		content.WriteString(m.styles.Connected.Render("● Connected"))
	} else {
		content.WriteString(m.styles.Disconnected.Render("○ Disconnected"))
	}
	content.WriteString("\n\n")

	// Enabled status
	content.WriteString(m.styles.Label.Render("Enabled: "))
	if m.enabled {
		content.WriteString(m.styles.Enabled.Render("Yes"))
	} else {
		content.WriteString(m.styles.Disabled.Render("No"))
	}
	content.WriteString("\n")

	// Server
	if m.server != "" {
		content.WriteString(m.styles.Label.Render("Server:  "))
		content.WriteString(m.styles.Value.Render(m.server))
		content.WriteString("\n")
	}

	// Toggling indicator
	if m.toggling {
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render("Updating..."))
	}

	// Help text
	content.WriteString("\n\n")
	content.WriteString(m.styles.Muted.Render("c: config | t: toggle | r: refresh"))

	r.SetContent(content.String())
	return r.Render()
}

// Connected returns whether the device is connected to cloud.
func (m Model) Connected() bool {
	return m.connected
}

// Enabled returns whether cloud is enabled.
func (m Model) Enabled() bool {
	return m.enabled
}

// Server returns the cloud server address.
func (m Model) Server() string {
	return m.server
}

// Device returns the current device address.
func (m Model) Device() string {
	return m.device
}

// Loading returns whether the component is loading.
func (m Model) Loading() bool {
	return m.loading
}

// Toggling returns whether a toggle is in progress.
func (m Model) Toggling() bool {
	return m.toggling
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Refresh triggers a refresh of the cloud status.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
}

// EditModelVisible returns whether the edit modal is currently visible.
func (m Model) EditModelVisible() bool {
	return m.editModel.Visible()
}

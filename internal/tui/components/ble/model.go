// Package ble provides TUI components for managing device Bluetooth settings.
// This includes BLE configuration and BTHome device management.
package ble

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

// Deps holds the dependencies for the BLE component.
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

// StatusLoadedMsg signals that BLE status was loaded.
type StatusLoadedMsg struct {
	BLE       *shelly.BLEConfig
	Discovery *shelly.BTHomeDiscovery
	Err       error
}

// DiscoveryStartedMsg signals that BTHome discovery was started.
type DiscoveryStartedMsg struct {
	Err error
}

// Model displays BLE and BTHome settings for a device.
type Model struct {
	ctx        context.Context
	svc        *shelly.Service
	device     string
	ble        *shelly.BLEConfig
	discovery  *shelly.BTHomeDiscovery
	loading    bool
	starting   bool
	editing    bool
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int // 1-based panel index for Shift+N hotkey hint
	styles     Styles
	loader     loading.Model
	editModal  EditModel
}

// Styles holds styles for the BLE component.
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
}

// DefaultStyles returns the default styles for the BLE component.
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
	}
}

// New creates a new BLE model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("ble: invalid deps: %v", err))
	}

	return Model{
		ctx:    deps.Ctx,
		svc:    deps.Svc,
		styles: DefaultStyles(),
		loader: loading.New(
			loading.WithMessage("Loading Bluetooth settings..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
		editModal: NewEditModel(deps.Ctx, deps.Svc),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display BLE settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.ble = nil
	m.discovery = nil
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

		var msg StatusLoadedMsg

		// Fetch BLE config
		bleConfig, err := m.svc.GetBLEConfig(ctx, m.device)
		if err == nil {
			msg.BLE = bleConfig
		}

		// Fetch BTHome status
		discovery, err := m.svc.GetBTHomeStatus(ctx, m.device)
		if err == nil {
			msg.Discovery = discovery
		}

		// If we got nothing, set an error
		if msg.BLE == nil && msg.Discovery == nil {
			msg.Err = fmt.Errorf("BLE not supported on this device")
		}

		return msg
	}
}

func (m Model) startDiscovery() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		err := m.svc.StartBTHomeDiscovery(ctx, m.device, 30)
		return DiscoveryStartedMsg{Err: err}
	}
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
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
	// Handle edit modal if visible
	if m.editing {
		return m.handleEditModalUpdate(msg)
	}

	// Forward tick messages to loader when loading
	if m.loading {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		// Continue processing other messages even during loading
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
		m.ble = msg.BLE
		m.discovery = msg.Discovery
		return m, nil

	case DiscoveryStartedMsg:
		m.starting = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		// Refresh to see discovery status
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
	var cmd tea.Cmd
	m.editModal, cmd = m.editModal.Update(msg)

	// Check if modal was closed
	if !m.editModal.Visible() {
		m.editing = false
		// Refresh data after edit
		m.loading = true
		return m, tea.Batch(cmd, m.loader.Tick(), m.fetchStatus())
	}

	// Handle save result message
	if saveMsg, ok := msg.(EditSaveResultMsg); ok {
		if saveMsg.Err == nil {
			m.editing = false
			m.editModal = m.editModal.Hide()
			// Refresh data after successful save
			m.loading = true
			return m, tea.Batch(m.loader.Tick(), m.fetchStatus(), func() tea.Msg {
				return EditClosedMsg{Saved: true}
			})
		}
	}

	return m, cmd
}

func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		if !m.loading && m.device != "" {
			m.loading = true
			return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
		}
	case "d":
		if !m.starting && !m.loading && m.device != "" && m.ble != nil && m.ble.Enable {
			m.starting = true
			m.err = nil
			return m, m.startDiscovery()
		}
	case "e", "enter":
		// Open edit modal
		if m.device != "" && !m.loading && m.ble != nil {
			m.editing = true
			m.editModal = m.editModal.SetSize(m.width, m.height)
			var cmd tea.Cmd
			m.editModal, cmd = m.editModal.Show(m.device, m.ble)
			return m, cmd
		}
	}

	return m, nil
}

// View renders the BLE component.
func (m Model) View() string {
	// Render edit modal if editing
	if m.editing {
		return m.editModal.View()
	}

	r := rendering.New(m.width, m.height).
		SetTitle("Bluetooth").
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

	// BLE Configuration Section
	content.WriteString(m.renderBLEConfig())
	content.WriteString("\n\n")

	// BTHome Section
	content.WriteString(m.renderBTHome())

	r.SetContent(content.String())

	// Footer with keybindings (shown when focused)
	if m.focused {
		if m.ble != nil && m.ble.Enable {
			r.SetFooter("e:edit r:refresh d:discover")
		} else {
			r.SetFooter("e:edit r:refresh")
		}
	}
	return r.Render()
}

func (m Model) renderBLEConfig() string {
	var content strings.Builder

	content.WriteString(m.styles.Section.Render("BLE Configuration"))
	content.WriteString("\n")

	if m.ble == nil {
		content.WriteString(m.styles.Muted.Render("  Not supported"))
		return content.String()
	}

	// Bluetooth enabled status
	if m.ble.Enable {
		content.WriteString("  " + m.styles.Enabled.Render("● Bluetooth Enabled") + "\n")
	} else {
		content.WriteString("  " + m.styles.Disabled.Render("○ Bluetooth Disabled") + "\n")
	}

	if !m.ble.Enable {
		return content.String()
	}

	// RPC status
	content.WriteString("  " + m.styles.Label.Render("RPC:      "))
	if m.ble.RPCEnabled {
		content.WriteString(m.styles.Enabled.Render("Enabled"))
	} else {
		content.WriteString(m.styles.Disabled.Render("Disabled"))
	}
	content.WriteString("\n")

	// Observer mode
	content.WriteString("  " + m.styles.Label.Render("Observer: "))
	if m.ble.ObserverMode {
		content.WriteString(m.styles.Enabled.Render("Enabled"))
		content.WriteString(m.styles.Muted.Render(" (receives BLU broadcasts)"))
	} else {
		content.WriteString(m.styles.Disabled.Render("Disabled"))
	}

	return content.String()
}

func (m Model) renderBTHome() string {
	var content strings.Builder

	content.WriteString(m.styles.Section.Render("BTHome Devices"))
	content.WriteString("\n")

	if m.ble == nil || !m.ble.Enable {
		content.WriteString(m.styles.Muted.Render("  Enable Bluetooth to manage BTHome devices"))
		return content.String()
	}

	// Discovery status
	switch {
	case m.discovery != nil && m.discovery.Active:
		content.WriteString("  " + m.styles.Warning.Render("◐ Discovery in progress...") + "\n")
		content.WriteString(m.styles.Muted.Render(
			fmt.Sprintf("    Duration: %ds", m.discovery.Duration),
		))
	case m.starting:
		content.WriteString("  " + m.styles.Muted.Render("◐ Starting discovery..."))
	default:
		content.WriteString("  " + m.styles.Muted.Render("No active discovery"))
		content.WriteString("\n")
		content.WriteString("  " + m.styles.Muted.Render("Press 'd' to scan for BTHome devices"))
	}

	return content.String()
}

// BLE returns the current BLE configuration.
func (m Model) BLE() *shelly.BLEConfig {
	return m.ble
}

// Discovery returns the current BTHome discovery status.
func (m Model) Discovery() *shelly.BTHomeDiscovery {
	return m.discovery
}

// Device returns the current device address.
func (m Model) Device() string {
	return m.device
}

// Loading returns whether the component is loading.
func (m Model) Loading() bool {
	return m.loading
}

// Starting returns whether discovery is starting.
func (m Model) Starting() bool {
	return m.starting
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// Refresh triggers a refresh of the BLE data.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchStatus())
}

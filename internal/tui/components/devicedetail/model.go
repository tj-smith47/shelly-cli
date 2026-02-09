// Package devicedetail provides a device detail overlay component for the TUI.
package devicedetail

import (
	"context"
	"encoding/json"
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// Deps holds the dependencies for the device detail component.
type Deps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Msg signals that device details were loaded.
type Msg struct {
	Device model.Device
	Status *model.MonitoringSnapshot
	Config map[string]any
	Err    error
}

// ClosedMsg signals that the detail view was closed.
type ClosedMsg struct{}

// Model holds the device detail state.
type Model struct {
	panel.Sizable
	ctx      context.Context
	svc      *shelly.Service
	device   *model.Device
	status   *model.MonitoringSnapshot
	config   map[string]any
	viewport viewport.Model
	visible  bool
	loading  bool
	err      error
	styles   Styles
}

// Styles for the device detail component.
type Styles struct {
	Container lipgloss.Style
	Title     lipgloss.Style
	Section   lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Online    lipgloss.Style
	Offline   lipgloss.Style
	Error     lipgloss.Style
}

// DefaultStyles returns default styles for the device detail component.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: styles.ModalBorder().Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true).
			Underline(true).
			MarginBottom(1),
		Section: lipgloss.NewStyle().
			Foreground(colors.Warning).
			Bold(true).
			MarginTop(1),
		Label: lipgloss.NewStyle().
			Foreground(colors.Text).
			Width(15),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Online: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		Offline: lipgloss.NewStyle().
			Foreground(colors.Offline).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
	}
}

// New creates a new device detail model.
func New(deps Deps) Model {
	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))

	m := Model{
		Sizable:  panel.NewSizableLoaderOnly(),
		ctx:      deps.Ctx,
		svc:      deps.Svc,
		viewport: vp,
		styles:   DefaultStyles(),
	}
	m.Loader = m.Loader.SetMessage("Loading device details...")
	return m
}

// Init initializes the device detail component.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the device detail component.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	// Forward tick messages to loader when loading
	if m.loading {
		result := generics.UpdateLoader(m.Loader, msg, func(msg tea.Msg) bool {
			_, ok := msg.(Msg)
			return ok
		})
		m.Loader = result.Loader
		if result.Consumed {
			return m, result.Cmd
		}
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if key.Matches(keyMsg, key.NewBinding(key.WithKeys("escape", "q"))) {
			m.visible = false
			m.device = nil
			m.status = nil
			m.config = nil
			return m, func() tea.Msg { return ClosedMsg{} }
		}
	}

	if detailMsg, ok := msg.(Msg); ok {
		m.loading = false
		if detailMsg.Err != nil {
			m.err = detailMsg.Err
			return m, nil
		}
		m.device = &detailMsg.Device
		m.status = detailMsg.Status
		m.config = detailMsg.Config
		m.viewport.SetContent(m.renderContent())
		return m, nil
	}

	// Forward to viewport for scrolling
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the device detail overlay.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	var content string

	switch {
	case m.loading:
		content = m.Loader.View()
	case m.err != nil:
		msg, _ := tuierrors.FormatError(m.err)
		content = m.styles.Error.Render(msg)
	default:
		content = m.viewport.View()
	}

	footer := m.styles.Label.Render("j/k scroll") + " " +
		m.styles.Label.Render("q/Esc close")

	return m.styles.Container.
		Width(m.Width - 4).
		Height(m.Height - 2).
		Render(content + "\n" + footer)
}

// renderContent renders the device detail content.
func (m Model) renderContent() string {
	if m.device == nil {
		return "No device selected"
	}

	content := m.styles.Title.Render("Device: "+m.device.Name) + "\n\n"

	// Basic info section
	content += m.styles.Section.Render("Basic Information") + "\n"
	content += m.renderRow("Name", m.device.Name)
	content += m.renderRow("Address", m.device.Address)
	content += m.renderRow("Generation", fmt.Sprintf("%d", m.device.Generation))
	content += m.renderRow("Type", m.device.Type)
	content += m.renderRow("Model", m.device.Model)

	// Authentication
	if m.device.Auth != nil && m.device.Auth.Username != "" {
		content += m.renderRow("Auth", "Configured ("+m.device.Auth.Username+")")
	} else {
		content += m.renderRow("Auth", "None")
	}

	// Status section
	content += m.renderStatusSection()

	// Config section (JSON preview)
	if len(m.config) > 0 {
		content += m.styles.Section.Render("Configuration Keys") + "\n"
		for configKey := range m.config {
			content += m.renderRow("", configKey)
		}
	}

	return content
}

// renderRow renders a label-value row.
func (m Model) renderRow(label, value string) string {
	if label == "" {
		return "  " + m.styles.Value.Render(value) + "\n"
	}
	return m.styles.Label.Render(label+":") + " " + m.styles.Value.Render(value) + "\n"
}

// renderStatusSection renders the live status section.
func (m Model) renderStatusSection() string {
	if m.status == nil || !m.status.Online {
		content := m.styles.Section.Render("Live Status") + "\n"
		if m.status != nil && m.status.Error != "" {
			return content + m.renderRow("Status", m.styles.Offline.Render("Error: "+m.status.Error))
		}
		return content + m.renderRow("Status", m.styles.Offline.Render("Offline"))
	}

	content := m.styles.Section.Render("Live Status") + "\n"
	content += m.renderRow("Status", m.styles.Online.Render("Online"))
	content += m.renderRow("Timestamp", m.status.Timestamp.Format("15:04:05"))

	// Power monitoring (PM components)
	content += m.renderPowerMonitoring()

	return content
}

// renderPowerMonitoring renders power monitoring data.
func (m Model) renderPowerMonitoring() string {
	var content string

	if len(m.status.PM) > 0 {
		content += m.styles.Section.Render("Power Monitoring (PM)") + "\n"
		for i, pm := range m.status.PM {
			content += m.renderRow(fmt.Sprintf("PM[%d] Power", i), fmt.Sprintf("%.2f W", pm.APower))
			content += m.renderRow(fmt.Sprintf("PM[%d] Voltage", i), fmt.Sprintf("%.1f V", pm.Voltage))
			content += m.renderRow(fmt.Sprintf("PM[%d] Current", i), fmt.Sprintf("%.3f A", pm.Current))
		}
	}

	if len(m.status.EM) > 0 {
		content += m.styles.Section.Render("Energy Meter (EM)") + "\n"
		for i, em := range m.status.EM {
			content += m.renderRow(fmt.Sprintf("EM[%d] Power", i), fmt.Sprintf("%.2f W", em.TotalActivePower))
			if csvURL := m.getEMDataCSVURL(em.ID); csvURL != "" {
				content += m.renderRow(fmt.Sprintf("EM[%d] CSV URL", i), csvURL)
			}
		}
	}

	if len(m.status.EM1) > 0 {
		content += m.styles.Section.Render("Energy Meter (EM1)") + "\n"
		for i, em1 := range m.status.EM1 {
			content += m.renderRow(fmt.Sprintf("EM1[%d] Power", i), fmt.Sprintf("%.2f W", em1.ActPower))
			if csvURL := m.getEM1DataCSVURL(em1.ID); csvURL != "" {
				content += m.renderRow(fmt.Sprintf("EM1[%d] CSV URL", i), csvURL)
			}
		}
	}

	return content
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.ApplySize(width, height)
	// Account for container borders and padding
	m.viewport.SetWidth(width - 8)
	m.viewport.SetHeight(height - 6)
	return m
}

// Show shows the device detail overlay and starts loading data.
func (m Model) Show(device model.Device) (Model, tea.Cmd) {
	m.visible = true
	m.loading = true
	m.err = nil
	m.device = &device

	return m, tea.Batch(m.Loader.Tick(), m.fetchDeviceDetails(device))
}

// Hide hides the device detail overlay.
func (m Model) Hide() Model {
	m.visible = false
	m.device = nil
	m.status = nil
	m.config = nil
	return m
}

// Visible returns whether the detail overlay is visible.
func (m Model) Visible() bool {
	return m.visible
}

// fetchDeviceDetails returns a command that loads device details.
func (m Model) fetchDeviceDetails(device model.Device) tea.Cmd {
	return func() tea.Msg {
		// Get monitoring snapshot (auto-detects Gen1 vs Gen2)
		status, err := m.svc.GetMonitoringSnapshotAuto(m.ctx, device.Address)
		if err != nil {
			return Msg{
				Device: device,
				Err:    fmt.Errorf("failed to get device status: %w", err),
			}
		}

		// Get config
		deviceConfig, err := m.svc.GetConfig(m.ctx, device.Address)
		if err != nil {
			// Config fetch failed but status succeeded, still show what we have
			return Msg{
				Device: device,
				Status: status,
				Err:    nil, // Don't fail entirely
			}
		}

		return Msg{
			Device: device,
			Status: status,
			Config: deviceConfig,
		}
	}
}

// getEMDataCSVURL returns the CSV download URL for an EM component, or empty string on error.
func (m Model) getEMDataCSVURL(id int) string {
	if m.device == nil {
		return ""
	}
	monSvc := m.svc.Monitoring()
	url, err := monSvc.GetEMDataCSVURL(m.device.Address, id, nil, nil, true)
	if err != nil {
		return ""
	}
	return url
}

// getEM1DataCSVURL returns the CSV download URL for an EM1 component, or empty string on error.
func (m Model) getEM1DataCSVURL(id int) string {
	if m.device == nil {
		return ""
	}
	monSvc := m.svc.Monitoring()
	url, err := monSvc.GetEM1DataCSVURL(m.device.Address, id, nil, nil, true)
	if err != nil {
		return ""
	}
	return url
}

// FormatJSON formats any value as indented JSON.
func FormatJSON(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(data)
}

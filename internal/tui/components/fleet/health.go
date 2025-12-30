package fleet

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// HealthDeps holds the dependencies for the Health component.
type HealthDeps struct {
	Ctx context.Context
}

// Validate ensures all required dependencies are set.
func (d HealthDeps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	return nil
}

// HealthLoadedMsg signals that health data was loaded.
type HealthLoadedMsg struct {
	Stats *integrator.FleetStats
	Err   error
}

// HealthModel displays fleet health and statistics.
type HealthModel struct {
	ctx        context.Context
	fleet      *integrator.FleetManager
	stats      *integrator.FleetStats
	loading    bool
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int
	styles     HealthStyles
	lastFetch  time.Time
	loader     loading.Model
}

// HealthStyles holds styles for the Health component.
type HealthStyles struct {
	Online   lipgloss.Style
	Offline  lipgloss.Style
	Label    lipgloss.Style
	Value    lipgloss.Style
	Muted    lipgloss.Style
	Error    lipgloss.Style
	Title    lipgloss.Style
	Stat     lipgloss.Style
	StatGood lipgloss.Style
	StatBad  lipgloss.Style
}

// DefaultHealthStyles returns the default styles for the Health component.
func DefaultHealthStyles() HealthStyles {
	colors := theme.GetSemanticColors()
	return HealthStyles{
		Online: lipgloss.NewStyle().
			Foreground(colors.Online),
		Offline: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Title: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Stat: lipgloss.NewStyle().
			Foreground(colors.Text),
		StatGood: lipgloss.NewStyle().
			Foreground(colors.Online),
		StatBad: lipgloss.NewStyle().
			Foreground(colors.Offline),
	}
}

// NewHealth creates a new Health model.
func NewHealth(deps HealthDeps) HealthModel {
	if err := deps.Validate(); err != nil {
		iostreams.DebugErr("fleet/health component init", err)
		panic(fmt.Sprintf("fleet/health: invalid deps: %v", err))
	}

	return HealthModel{
		ctx:    deps.Ctx,
		styles: DefaultHealthStyles(),
		loader: loading.New(
			loading.WithMessage("Loading health data..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
	}
}

// Init returns the initial command.
func (m HealthModel) Init() tea.Cmd {
	return nil
}

// SetFleetManager sets the fleet manager.
func (m HealthModel) SetFleetManager(fm *integrator.FleetManager) (HealthModel, tea.Cmd) {
	m.fleet = fm
	if fm == nil {
		m.stats = nil
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.loadHealth())
}

func (m HealthModel) loadHealth() tea.Cmd {
	return func() tea.Msg {
		if m.fleet == nil {
			return HealthLoadedMsg{Err: fmt.Errorf("not connected to fleet")}
		}
		stats := m.fleet.GetStats()
		return HealthLoadedMsg{Stats: stats}
	}
}

// SetSize sets the component dimensions.
func (m HealthModel) SetSize(width, height int) HealthModel {
	m.width = width
	m.height = height
	// Update loader size for proper centering
	m.loader = m.loader.SetSize(width-4, height-4)
	return m
}

// SetFocused sets the focus state.
func (m HealthModel) SetFocused(focused bool) HealthModel {
	m.focused = focused
	return m
}

// SetPanelIndex sets the panel index for Shift+N hint.
func (m HealthModel) SetPanelIndex(index int) HealthModel {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m HealthModel) Update(msg tea.Msg) (HealthModel, tea.Cmd) {
	// Forward tick messages to loader when loading
	if m.loading {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		// Continue processing HealthLoadedMsg even during loading
		if _, ok := msg.(HealthLoadedMsg); !ok {
			if cmd != nil {
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case HealthLoadedMsg:
		m.loading = false
		m.lastFetch = time.Now()
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.stats = msg.Stats
		m.err = nil
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m HealthModel) handleKey(msg tea.KeyPressMsg) (HealthModel, tea.Cmd) {
	if msg.String() == "r" && !m.loading && m.fleet != nil {
		m.loading = true
		return m, tea.Batch(m.loader.Tick(), m.loadHealth())
	}

	return m, nil
}

// View renders the Health component.
func (m HealthModel) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Fleet Health").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	// Add footer with keybindings when focused
	if m.focused {
		r.SetFooter("r:refresh")
	}

	// Calculate content area for centering (accounting for panel borders)
	contentWidth := m.width - 4
	contentHeight := m.height - 4
	if contentWidth < 1 {
		contentWidth = 1
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	if m.fleet == nil {
		msg := m.styles.Muted.Render("Not connected to Shelly Cloud")
		r.SetContent(lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, msg))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.loader.View())
		return r.Render()
	}

	if m.err != nil {
		msg := m.styles.Error.Render("Error: " + m.err.Error())
		r.SetContent(lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, msg))
		return r.Render()
	}

	if m.stats == nil {
		msg := m.styles.Muted.Render("No health data available")
		r.SetContent(lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, msg))
		return r.Render()
	}

	var content strings.Builder

	// Connection stats
	content.WriteString(m.styles.Label.Render("Connections: "))
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.stats.TotalConnections)))
	content.WriteString("\n")

	// Device stats
	content.WriteString(m.styles.Label.Render("Total Devices: "))
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.stats.TotalDevices)))
	content.WriteString("\n")

	// Online devices
	content.WriteString(m.styles.Label.Render("Online: "))
	if m.stats.OnlineDevices > 0 {
		content.WriteString(m.styles.StatGood.Render(fmt.Sprintf("%d", m.stats.OnlineDevices)))
	} else {
		content.WriteString(m.styles.Value.Render("0"))
	}
	content.WriteString("\n")

	// Offline devices
	content.WriteString(m.styles.Label.Render("Offline: "))
	if m.stats.OfflineDevices > 0 {
		content.WriteString(m.styles.StatBad.Render(fmt.Sprintf("%d", m.stats.OfflineDevices)))
	} else {
		content.WriteString(m.styles.Value.Render("0"))
	}
	content.WriteString("\n")

	// Groups
	content.WriteString(m.styles.Label.Render("Groups: "))
	content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.stats.TotalGroups)))
	content.WriteString("\n")

	// Health percentage
	if m.stats.TotalDevices > 0 {
		healthPct := float64(m.stats.OnlineDevices) / float64(m.stats.TotalDevices) * 100
		content.WriteString("\n")
		content.WriteString(m.styles.Label.Render("Fleet Health: "))
		healthStyle := m.styles.StatBad
		if healthPct >= 90 {
			healthStyle = m.styles.StatGood
		} else if healthPct >= 50 {
			healthStyle = m.styles.Value
		}
		content.WriteString(healthStyle.Render(fmt.Sprintf("%.0f%%", healthPct)))
	}

	// Account stats if available
	if m.stats.AccountStats != nil {
		content.WriteString("\n\n")
		content.WriteString(m.styles.Label.Render("Accounts: "))
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.stats.AccountStats.TotalAccounts)))
		content.WriteString("\n")
		content.WriteString(m.styles.Label.Render("Controllable: "))
		content.WriteString(m.styles.Value.Render(fmt.Sprintf("%d", m.stats.AccountStats.ControllableDevices)))
	}

	r.SetContent(content.String())
	return r.Render()
}

// Stats returns the fleet stats.
func (m HealthModel) Stats() *integrator.FleetStats {
	return m.stats
}

// Loading returns whether health data is being loaded.
func (m HealthModel) Loading() bool {
	return m.loading
}

// Error returns any error that occurred.
func (m HealthModel) Error() error {
	return m.err
}

// Refresh triggers a health data reload.
func (m HealthModel) Refresh() (HealthModel, tea.Cmd) {
	if m.fleet == nil {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.loadHealth())
}

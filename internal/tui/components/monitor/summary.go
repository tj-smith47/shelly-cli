package monitor

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/cachestatus"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// SummaryData holds the aggregated data for the summary bar.
type SummaryData struct {
	TotalPower  float64
	TotalEnergy float64
	PeakPower   float64 // Highest power seen this session
	OnlineCount int
	TotalCount  int
	CostRate    float64
	Currency    string
}

// SummaryModel is a non-focusable summary bar that displays aggregated monitoring data.
type SummaryModel struct {
	data        SummaryData
	cacheStatus cachestatus.Model
	styles      SummaryStyles
	width       int
	height      int
}

// SummaryStyles holds styles for the summary bar.
type SummaryStyles struct {
	Label   lipgloss.Style
	Value   lipgloss.Style
	Warning lipgloss.Style
	Muted   lipgloss.Style
	Online  lipgloss.Style
	Offline lipgloss.Style
}

// defaultSummaryStyles returns default styles for the summary bar.
func defaultSummaryStyles() SummaryStyles {
	colors := theme.GetSemanticColors()
	return SummaryStyles{
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Value: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Highlight),
		Warning: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Warning),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Online: lipgloss.NewStyle().
			Foreground(colors.Online),
		Offline: lipgloss.NewStyle().
			Foreground(colors.Offline),
	}
}

// NewSummary creates a new summary bar model.
func NewSummary() SummaryModel {
	return SummaryModel{
		cacheStatus: cachestatus.New(),
		styles:      defaultSummaryStyles(),
		height:      3, // Fixed height: border top + content + border bottom
	}
}

// Init returns the initial command.
func (m SummaryModel) Init() tea.Cmd {
	return m.cacheStatus.Init()
}

// Update handles messages for the summary bar.
func (m SummaryModel) Update(msg tea.Msg) (SummaryModel, tea.Cmd) {
	var cmd tea.Cmd
	m.cacheStatus, cmd = m.cacheStatus.Update(msg)
	return m, cmd
}

// SetData updates the summary data.
func (m SummaryModel) SetData(data SummaryData) SummaryModel {
	// Track peak power across the session
	if data.TotalPower > m.data.PeakPower {
		data.PeakPower = data.TotalPower
	} else {
		data.PeakPower = m.data.PeakPower
	}
	m.data = data
	return m
}

// StartRefresh starts the cache status spinner.
func (m SummaryModel) StartRefresh() (SummaryModel, tea.Cmd) {
	var cmd tea.Cmd
	m.cacheStatus, cmd = m.cacheStatus.StartRefresh()
	return m, cmd
}

// StopRefresh stops the cache status spinner and marks updated.
func (m SummaryModel) StopRefresh() SummaryModel {
	m.cacheStatus = m.cacheStatus.StopRefresh()
	return m
}

// SetSize sets the width of the summary bar. Height is always fixed at 3.
func (m SummaryModel) SetSize(width, height int) SummaryModel {
	m.width = width
	m.height = height
	return m
}

// View renders the summary bar as a single-line bordered panel.
func (m SummaryModel) View() string {
	if m.width < 20 {
		return ""
	}

	content := m.renderContent()

	r := rendering.New(m.width, m.height).
		SetTitle("Monitor").
		SetContent(content).
		SetFocused(false)

	return r.Render()
}

// renderContent builds the summary line content.
func (m SummaryModel) renderContent() string {
	parts := make([]string, 0, 6)

	// Total Power and Energy
	parts = append(parts,
		m.styles.Label.Render("Power: ")+m.styles.Warning.Render(formatPower(m.data.TotalPower)),
		m.styles.Label.Render("Energy: ")+m.styles.Value.Render(formatEnergy(m.data.TotalEnergy)))

	// Cost (only if rate is set and energy > 0)
	if m.data.CostRate > 0 && m.data.TotalEnergy > 0 {
		cost := (m.data.TotalEnergy / 1000) * m.data.CostRate
		parts = append(parts,
			m.styles.Label.Render("Cost: ")+m.styles.Value.Render(fmt.Sprintf("%s%.2f", m.data.Currency, cost)))
	}

	// Online/Total devices
	deviceStr := fmt.Sprintf("%d/%d", m.data.OnlineCount, m.data.TotalCount)
	var deviceStyled string
	if offlineCount := m.data.TotalCount - m.data.OnlineCount; offlineCount > 0 {
		deviceStyled = m.styles.Offline.Render(deviceStr)
	} else {
		deviceStyled = m.styles.Online.Render(deviceStr)
	}
	parts = append(parts, m.styles.Label.Render("Online: ")+deviceStyled)

	// Peak power indicator
	if m.data.PeakPower > 0 {
		parts = append(parts,
			m.styles.Label.Render("Peak: ")+m.styles.Muted.Render(formatPower(m.data.PeakPower)))
	}

	// Cache status (Updated Xs ago / spinner)
	cacheView := m.cacheStatus.ViewCompact()
	if cacheView != "" {
		parts = append(parts, cacheView)
	}

	sep := m.styles.Muted.Render("  |  ")
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += sep
		}
		result += part
	}
	return result
}

// Data returns the current summary data.
func (m SummaryModel) Data() SummaryData {
	return m.data
}

// UpdatedAt returns the last update timestamp from cache status.
func (m SummaryModel) UpdatedAt() time.Time {
	return m.cacheStatus.UpdatedAt()
}

// IsRefreshing returns whether the summary is showing a refresh spinner.
func (m SummaryModel) IsRefreshing() bool {
	return m.cacheStatus.IsRefreshing()
}

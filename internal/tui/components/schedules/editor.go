package schedules

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// EditorModel displays schedule details.
type EditorModel struct {
	schedule   *Schedule
	scroll     int
	width      int
	height     int
	focused    bool
	panelIndex int // 1-based panel index for Shift+N hotkey hint
	styles     EditorStyles
}

// EditorStyles holds styles for the editor component.
type EditorStyles struct {
	Header    lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Enabled   lipgloss.Style
	Disabled  lipgloss.Style
	Method    lipgloss.Style
	Params    lipgloss.Style
	Separator lipgloss.Style
	Muted     lipgloss.Style
}

// DefaultEditorStyles returns the default styles for the schedule editor.
func DefaultEditorStyles() EditorStyles {
	colors := theme.GetSemanticColors()
	return EditorStyles{
		Header: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Width(12),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Enabled: lipgloss.NewStyle().
			Foreground(colors.Online),
		Disabled: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Method: lipgloss.NewStyle().
			Foreground(colors.Info).
			Bold(true),
		Params: lipgloss.NewStyle().
			Foreground(colors.Text),
		Separator: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// NewEditor creates a new schedule editor model.
func NewEditor() EditorModel {
	return EditorModel{
		styles: DefaultEditorStyles(),
	}
}

// Init returns the initial command.
func (m EditorModel) Init() tea.Cmd {
	return nil
}

// SetSchedule sets the schedule to display.
func (m EditorModel) SetSchedule(schedule *Schedule) EditorModel {
	m.schedule = schedule
	m.scroll = 0
	return m
}

// Clear clears the editor state.
func (m EditorModel) Clear() EditorModel {
	m.schedule = nil
	m.scroll = 0
	return m
}

// SetSize sets the component dimensions.
func (m EditorModel) SetSize(width, height int) EditorModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets the focus state.
func (m EditorModel) SetFocused(focused bool) EditorModel {
	m.focused = focused
	return m
}

// SetPanelIndex sets the 1-based panel index for Shift+N hotkey hint.
func (m EditorModel) SetPanelIndex(index int) EditorModel {
	m.panelIndex = index
	return m
}

// Update handles messages.
func (m EditorModel) Update(msg tea.Msg) (EditorModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if !m.focused {
			return m, nil
		}
		return m.handleKey(keyMsg)
	}

	return m, nil
}

func (m EditorModel) handleKey(msg tea.KeyPressMsg) (EditorModel, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m = m.scrollDown()
	case "k", "up":
		m = m.scrollUp()
	case "g":
		m.scroll = 0
	case "G":
		m = m.scrollToEnd()
	}

	return m, nil
}

func (m EditorModel) scrollDown() EditorModel {
	maxScroll := m.maxScroll()
	if m.scroll < maxScroll {
		m.scroll++
	}
	return m
}

func (m EditorModel) scrollUp() EditorModel {
	if m.scroll > 0 {
		m.scroll--
	}
	return m
}

func (m EditorModel) scrollToEnd() EditorModel {
	m.scroll = m.maxScroll()
	return m
}

func (m EditorModel) visibleLines() int {
	lines := m.height - 4
	if lines < 1 {
		return 1
	}
	return lines
}

func (m EditorModel) maxScroll() int {
	// Estimate content lines
	if m.schedule == nil {
		return 0
	}
	contentLines := 10 + len(m.schedule.Calls)*4 // Approx 4 lines per call
	maxLines := contentLines - m.visibleLines()
	if maxLines < 0 {
		return 0
	}
	return maxLines
}

// View renders the schedule editor.
func (m EditorModel) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Schedule Details").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	if m.schedule == nil {
		r.SetContent(m.styles.Muted.Render("No schedule selected"))
		return r.Render()
	}

	var content strings.Builder

	// Header with ID
	content.WriteString(m.styles.Header.Render(fmt.Sprintf("Schedule #%d", m.schedule.ID)))
	content.WriteString("\n\n")

	// Status
	if m.schedule.Enable {
		content.WriteString(m.styles.Label.Render("Status:"))
		content.WriteString(" ")
		content.WriteString(m.styles.Enabled.Render("Enabled"))
	} else {
		content.WriteString(m.styles.Label.Render("Status:"))
		content.WriteString(" ")
		content.WriteString(m.styles.Disabled.Render("Disabled"))
	}
	content.WriteString("\n")

	// Timespec
	content.WriteString(m.styles.Label.Render("Timespec:"))
	content.WriteString(" ")
	content.WriteString(m.styles.Value.Render(m.schedule.Timespec))
	content.WriteString("\n")

	// Timespec explanation
	if explanation := m.explainTimespec(m.schedule.Timespec); explanation != "" {
		content.WriteString(m.styles.Label.Render(""))
		content.WriteString(" ")
		content.WriteString(m.styles.Muted.Render("(" + explanation + ")"))
		content.WriteString("\n")
	}

	// Separator
	content.WriteString("\n")
	content.WriteString(m.styles.Separator.Render(strings.Repeat("─", m.width-6)))
	content.WriteString("\n\n")

	// Calls
	content.WriteString(m.styles.Header.Render("RPC Calls"))
	content.WriteString(m.styles.Muted.Render(fmt.Sprintf(" (%d)", len(m.schedule.Calls))))
	content.WriteString("\n\n")

	if len(m.schedule.Calls) == 0 {
		content.WriteString(m.styles.Muted.Render("No calls configured"))
	} else {
		for i, call := range m.schedule.Calls {
			content.WriteString(m.renderCall(i+1, call))
			if i < len(m.schedule.Calls)-1 {
				content.WriteString("\n")
			}
		}
	}

	r.SetContent(content.String())
	return r.Render()
}

func (m EditorModel) renderCall(index int, call automation.ScheduleCall) string {
	var sb strings.Builder

	// Method
	sb.WriteString(fmt.Sprintf("%d. ", index))
	sb.WriteString(m.styles.Method.Render(call.Method))
	sb.WriteString("\n")

	// Params
	if len(call.Params) > 0 {
		paramsJSON, err := json.MarshalIndent(call.Params, "   ", "  ")
		if err == nil {
			sb.WriteString("   ")
			sb.WriteString(m.styles.Params.Render(string(paramsJSON)))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// explainTimespec provides a human-readable explanation of the timespec.
func (m EditorModel) explainTimespec(spec string) string {
	// Handle sunrise/sunset
	if desc := m.explainSunEvent(spec); desc != "" {
		return desc
	}

	// Handle cron-like spec
	return m.explainCronSpec(spec)
}

// explainSunEvent explains sunrise/sunset timespecs.
func (m EditorModel) explainSunEvent(spec string) string {
	if strings.HasPrefix(spec, "@sunrise") {
		return m.describeSunOffset(spec, "sunrise")
	}
	if strings.HasPrefix(spec, "@sunset") {
		return m.describeSunOffset(spec, "sunset")
	}
	return ""
}

func (m EditorModel) describeSunOffset(spec, event string) string {
	switch {
	case strings.Contains(spec, "+"):
		return fmt.Sprintf("After %s", event)
	case strings.Contains(spec, "-"):
		return fmt.Sprintf("Before %s", event)
	default:
		return fmt.Sprintf("At %s", event)
	}
}

// explainCronSpec explains cron-like timespecs.
// Shelly uses 6-part cron format: ss mm hh DD MM WW.
func (m EditorModel) explainCronSpec(spec string) string {
	parts := strings.Fields(spec)
	if len(parts) < 6 {
		return ""
	}

	// Check weekday patterns (index 5 is day of week)
	switch parts[5] {
	case "MON-FRI", "1-5":
		return "Weekdays only"
	case "SAT,SUN", "0,6":
		return "Weekends only"
	}

	// Check daily pattern (month and weekday are wildcards)
	if parts[0] == "0" && parts[3] == "*" && parts[4] == "*" && parts[5] == "*" {
		if parts[2] != "*" && parts[1] != "*" {
			return fmt.Sprintf("Daily at %s:%s", parts[2], parts[1])
		}
	}

	return ""
}

// Schedule returns the current schedule.
func (m EditorModel) Schedule() *Schedule {
	return m.schedule
}

// HasSchedule returns whether a schedule is set.
func (m EditorModel) HasSchedule() bool {
	return m.schedule != nil
}

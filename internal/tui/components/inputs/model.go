// Package inputs provides TUI components for managing device input settings.
package inputs

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
)

// Deps holds the dependencies for the Inputs component.
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

// LoadedMsg signals that inputs were loaded.
type LoadedMsg struct {
	Inputs []shelly.InputInfo
	Err    error
}

// Model displays input settings for a device.
type Model struct {
	ctx     context.Context
	svc     *shelly.Service
	device  string
	inputs  []shelly.InputInfo
	cursor  int
	scroll  int
	loading bool
	err     error
	width   int
	height  int
	focused bool
	styles  Styles
}

// Styles holds styles for the Inputs component.
type Styles struct {
	StateOn  lipgloss.Style
	StateOff lipgloss.Style
	Type     lipgloss.Style
	Name     lipgloss.Style
	ID       lipgloss.Style
	Label    lipgloss.Style
	Selected lipgloss.Style
	Error    lipgloss.Style
	Muted    lipgloss.Style
}

// DefaultStyles returns the default styles for the Inputs component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		StateOn: lipgloss.NewStyle().
			Foreground(colors.Online),
		StateOff: lipgloss.NewStyle().
			Foreground(colors.Offline),
		Type: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Name: lipgloss.NewStyle().
			Foreground(colors.Text),
		ID: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Label: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Selected: lipgloss.NewStyle().
			Background(colors.AltBackground).
			Foreground(colors.Highlight).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
	}
}

// New creates a new Inputs model.
func New(deps Deps) Model {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("inputs: invalid deps: %v", err))
	}

	return Model{
		ctx:     deps.Ctx,
		svc:     deps.Svc,
		loading: false,
		styles:  DefaultStyles(),
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetDevice sets the device to display input settings for and triggers a fetch.
func (m Model) SetDevice(device string) (Model, tea.Cmd) {
	m.device = device
	m.inputs = nil
	m.cursor = 0
	m.scroll = 0
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, m.fetchInputs()
}

func (m Model) fetchInputs() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
		defer cancel()

		inputs, err := m.svc.InputList(ctx, m.device)
		return LoadedMsg{Inputs: inputs, Err: err}
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

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.inputs = msg.Inputs
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
	switch msg.String() {
	case "j", "down":
		m = m.cursorDown()
	case "k", "up":
		m = m.cursorUp()
	case "r":
		if !m.loading && m.device != "" {
			m.loading = true
			return m, m.fetchInputs()
		}
	}

	return m, nil
}

func (m Model) cursorDown() Model {
	if m.cursor < len(m.inputs)-1 {
		m.cursor++
		m = m.ensureVisible()
	}
	return m
}

func (m Model) cursorUp() Model {
	if m.cursor > 0 {
		m.cursor--
		m = m.ensureVisible()
	}
	return m
}

func (m Model) ensureVisible() Model {
	visible := m.visibleRows()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	} else if m.cursor >= m.scroll+visible {
		m.scroll = m.cursor - visible + 1
	}
	return m
}

func (m Model) visibleRows() int {
	rows := m.height - 6 // Reserve space for header and footer
	if rows < 1 {
		return 1
	}
	return rows
}

// View renders the Inputs component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Inputs").
		SetFocused(m.focused)

	if m.device == "" {
		r.SetContent(m.styles.Muted.Render("No device selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.styles.Muted.Render("Loading inputs..."))
		return r.Render()
	}

	if m.err != nil {
		r.SetContent(m.styles.Error.Render("Error: " + m.err.Error()))
		return r.Render()
	}

	if len(m.inputs) == 0 {
		r.SetContent(m.styles.Muted.Render("No inputs found"))
		return r.Render()
	}

	var content strings.Builder

	// Header
	content.WriteString(m.styles.Label.Render(fmt.Sprintf("Inputs (%d):\n\n", len(m.inputs))))

	// Input list
	visible := m.visibleRows()
	endIdx := m.scroll + visible
	if endIdx > len(m.inputs) {
		endIdx = len(m.inputs)
	}

	for i := m.scroll; i < endIdx; i++ {
		input := m.inputs[i]
		isSelected := i == m.cursor
		content.WriteString(m.renderInputLine(input, isSelected))
		if i < endIdx-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(m.inputs) > visible {
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render(
			fmt.Sprintf("[%d/%d]", m.cursor+1, len(m.inputs)),
		))
	}

	// Help text
	content.WriteString("\n\n")
	content.WriteString(m.styles.Muted.Render("r: refresh"))

	r.SetContent(content.String())
	return r.Render()
}

func (m Model) renderInputLine(input shelly.InputInfo, isSelected bool) string {
	selector := "  "
	if isSelected {
		selector = "▶ "
	}

	// State indicator
	var stateStr string
	if input.State {
		stateStr = m.styles.StateOn.Render("●")
	} else {
		stateStr = m.styles.StateOff.Render("○")
	}

	// ID
	idStr := m.styles.ID.Render(fmt.Sprintf("Input:%d", input.ID))

	// Name
	name := input.Name
	if name == "" {
		name = "(unnamed)"
	}
	nameStr := m.styles.Name.Render(name)

	// Type
	typeStr := m.styles.Type.Render(fmt.Sprintf("[%s]", input.Type))

	line := fmt.Sprintf("%s%s %s %s %s", selector, stateStr, idStr, nameStr, typeStr)

	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return line
}

// Inputs returns the current list of inputs.
func (m Model) Inputs() []shelly.InputInfo {
	return m.inputs
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

// Cursor returns the current cursor position.
func (m Model) Cursor() int {
	return m.cursor
}

// Refresh triggers a refresh of the inputs.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, m.fetchInputs()
}

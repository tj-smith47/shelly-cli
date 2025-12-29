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
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
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
	ctx        context.Context
	svc        *shelly.Service
	device     string
	inputs     []shelly.InputInfo
	scroller   *panel.Scroller
	loading    bool
	err        error
	width      int
	height     int
	focused    bool
	panelIndex int // 1-based panel index for Shift+N hotkey hint
	styles     Styles
	loader     loading.Model
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
		ctx:      deps.Ctx,
		svc:      deps.Svc,
		scroller: panel.NewScroller(0, 10),
		loading:  false,
		styles:   DefaultStyles(),
		loader: loading.New(
			loading.WithMessage("Loading inputs..."),
			loading.WithStyle(loading.StyleDot),
			loading.WithCentered(true, true),
		),
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
	m.scroller.SetItemCount(0)
	m.scroller.CursorToStart()
	m.err = nil

	if device == "" {
		return m, nil
	}

	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchInputs())
}

func (m Model) fetchInputs() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		inputs, err := m.svc.InputList(ctx, m.device)
		return LoadedMsg{Inputs: inputs, Err: err}
	}
}

// SetSize sets the component dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	visibleRows := height - 6 // Reserve space for header and footer
	if visibleRows < 1 {
		visibleRows = 1
	}
	m.scroller.SetVisibleRows(visibleRows)
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
		// Continue processing LoadedMsg even during loading
		if _, ok := msg.(LoadedMsg); !ok {
			if cmd != nil {
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {
	case LoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.inputs = msg.Inputs
		m.scroller.SetItemCount(len(m.inputs))
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
		m.scroller.CursorDown()
	case "k", "up":
		m.scroller.CursorUp()
	case "g":
		m.scroller.CursorToStart()
	case "G":
		m.scroller.CursorToEnd()
	case "ctrl+d", "pgdown":
		m.scroller.PageDown()
	case "ctrl+u", "pgup":
		m.scroller.PageUp()
	case "r":
		if !m.loading && m.device != "" {
			m.loading = true
			return m, tea.Batch(m.loader.Tick(), m.fetchInputs())
		}
	}

	return m, nil
}

// View renders the Inputs component.
func (m Model) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Inputs").
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

	if len(m.inputs) == 0 {
		r.SetContent(m.styles.Muted.Render("No inputs found"))
		return r.Render()
	}

	var content strings.Builder

	// Header
	content.WriteString(m.styles.Label.Render(fmt.Sprintf("Inputs (%d):\n\n", len(m.inputs))))

	// Input list
	start, end := m.scroller.VisibleRange()
	for i := start; i < end; i++ {
		input := m.inputs[i]
		isSelected := m.scroller.IsCursorAt(i)
		content.WriteString(m.renderInputLine(input, isSelected))
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render(m.scroller.ScrollInfo()))

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
	return m.scroller.Cursor()
}

// Refresh triggers a refresh of the inputs.
func (m Model) Refresh() (Model, tea.Cmd) {
	if m.device == "" {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.loader.Tick(), m.fetchInputs())
}

// FooterText returns keybinding hints for the footer.
func (m Model) FooterText() string {
	return "j/k:scroll g/G:top/bottom enter:details"
}

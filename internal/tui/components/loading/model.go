// Package loading provides a reusable loading/spinner component for the TUI.
package loading

import (
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Style represents different spinner animation styles.
type Style int

const (
	// StyleDot uses a rotating dot pattern.
	StyleDot Style = iota
	// StyleLine uses a line animation.
	StyleLine
	// StyleMiniDot uses a small rotating dot.
	StyleMiniDot
	// StylePulse uses a pulsing animation.
	StylePulse
	// StylePoints uses expanding dots.
	StylePoints
	// StyleGlobe uses a globe animation.
	StyleGlobe
	// StyleMoon uses moon phases.
	StyleMoon
	// StyleMonkey uses a monkey animation.
	StyleMonkey
)

// Spinners maps Style to spinner.Spinner configurations.
var Spinners = map[Style]spinner.Spinner{
	StyleDot: {
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		FPS:    time.Second / 10,
	},
	StyleLine: {
		Frames: []string{"-", "\\", "|", "/"},
		FPS:    time.Second / 8,
	},
	StyleMiniDot: {
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		FPS:    time.Second / 12,
	},
	StylePulse: {
		Frames: []string{"█", "▓", "▒", "░", "▒", "▓"},
		FPS:    time.Second / 6,
	},
	StylePoints: {
		Frames: []string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●", "∙∙∙"},
		FPS:    time.Second / 5,
	},
	StyleGlobe: {
		Frames: []string{"🌍", "🌎", "🌏"},
		FPS:    time.Second / 3,
	},
	StyleMoon: {
		Frames: []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"},
		FPS:    time.Second / 4,
	},
	StyleMonkey: {
		Frames: []string{"🙈", "🙉", "🙊"},
		FPS:    time.Second / 3,
	},
}

// Styles holds the visual styles for the loading component.
type Styles struct {
	Container lipgloss.Style
	Spinner   lipgloss.Style
	Message   lipgloss.Style
}

// DefaultStyles returns the default styles for the loading component.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Container: lipgloss.NewStyle().
			Padding(1, 2),
		Spinner: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Message: lipgloss.NewStyle().
			Foreground(colors.Text).
			MarginLeft(1),
	}
}

// Model holds the loading component state.
type Model struct {
	spinner spinner.Model
	message string
	visible bool
	width   int
	height  int
	styles  Styles
	centerH bool // Center horizontally
	centerV bool // Center vertically
}

// Option configures the loading model.
type Option func(*Model)

// WithMessage sets the loading message.
func WithMessage(msg string) Option {
	return func(m *Model) {
		m.message = msg
	}
}

// WithStyle sets the spinner style.
func WithStyle(style Style) Option {
	return func(m *Model) {
		if s, ok := Spinners[style]; ok {
			m.spinner.Spinner = s
		}
	}
}

// WithStyles sets custom visual styles.
func WithStyles(styles Styles) Option {
	return func(m *Model) {
		m.styles = styles
		m.spinner.Style = styles.Spinner
	}
}

// WithCentered enables centering in the container.
func WithCentered(horizontal, vertical bool) Option {
	return func(m *Model) {
		m.centerH = horizontal
		m.centerV = vertical
	}
}

// New creates a new loading model with the given options.
func New(opts ...Option) Model {
	styles := DefaultStyles()
	s := spinner.New(
		spinner.WithSpinner(Spinners[StyleDot]),
		spinner.WithStyle(styles.Spinner),
	)

	m := Model{
		spinner: s,
		message: "Loading...",
		visible: true,
		styles:  styles,
		centerH: true,
		centerV: true,
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// Init returns the initial command for the spinner.
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages for the loading component.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// View renders the loading component.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	content := m.spinner.View() + m.styles.Message.Render(m.message)
	styled := m.styles.Container.Render(content)

	// No dimensions set, return as-is
	if m.width == 0 && m.height == 0 {
		return styled
	}

	// Build container style with dimensions and alignment
	return m.buildContainerStyle().Render(styled)
}

// buildContainerStyle creates a style with dimensions and alignment.
func (m Model) buildContainerStyle() lipgloss.Style {
	style := lipgloss.NewStyle()

	if m.width > 0 {
		style = style.Width(m.width)
	}
	if m.height > 0 {
		style = style.Height(m.height)
	}
	if m.centerH {
		style = style.Align(lipgloss.Center)
	}
	if m.centerV {
		style = style.AlignVertical(lipgloss.Center)
	}

	return style
}

// SetSize sets the dimensions for the loading component.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// SetMessage updates the loading message.
func (m Model) SetMessage(msg string) Model {
	m.message = msg
	return m
}

// SetVisible sets whether the loading component is visible.
func (m Model) SetVisible(visible bool) Model {
	m.visible = visible
	return m
}

// Show makes the loading component visible and returns a tick command.
func (m Model) Show() (Model, tea.Cmd) {
	m.visible = true
	return m, m.spinner.Tick
}

// Hide hides the loading component.
func (m Model) Hide() Model {
	m.visible = false
	return m
}

// IsVisible returns whether the loading component is visible.
func (m Model) IsVisible() bool {
	return m.visible
}

// Message returns the current loading message.
func (m Model) Message() string {
	return m.message
}

// SetSpinnerStyle changes the spinner animation style.
func (m Model) SetSpinnerStyle(style Style) Model {
	if s, ok := Spinners[style]; ok {
		m.spinner.Spinner = s
	}
	return m
}

// Tick returns the spinner tick command for animation.
func (m Model) Tick() tea.Cmd {
	return m.spinner.Tick
}

// SpinnerFrame returns the current spinner character for embedding in headers/badges.
func (m Model) SpinnerFrame() string {
	return m.spinner.View()
}

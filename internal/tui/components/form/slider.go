package form

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// SliderStyles holds the visual styles for a slider.
type SliderStyles struct {
	Label   lipgloss.Style
	Track   lipgloss.Style
	Filled  lipgloss.Style
	Thumb   lipgloss.Style
	Value   lipgloss.Style
	Help    lipgloss.Style
	Focused lipgloss.Style
}

// DefaultSliderStyles returns the default styles for a slider.
func DefaultSliderStyles() SliderStyles {
	colors := theme.GetSemanticColors()
	return SliderStyles{
		Label: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		Track: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Filled: lipgloss.NewStyle().
			Foreground(colors.Highlight),
		Thumb: lipgloss.NewStyle().
			Foreground(colors.Online).
			Bold(true),
		Value: lipgloss.NewStyle().
			Foreground(colors.Text),
		Help: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Focused: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
	}
}

// SliderOption configures a slider.
type SliderOption func(*Slider)

// WithSliderLabel sets the slider label.
func WithSliderLabel(label string) SliderOption {
	return func(s *Slider) {
		s.label = label
	}
}

// WithSliderMin sets the minimum value.
func WithSliderMin(minVal float64) SliderOption {
	return func(s *Slider) {
		s.min = minVal
	}
}

// WithSliderMax sets the maximum value.
func WithSliderMax(maxVal float64) SliderOption {
	return func(s *Slider) {
		s.max = maxVal
	}
}

// WithSliderStep sets the step increment.
func WithSliderStep(step float64) SliderOption {
	return func(s *Slider) {
		s.step = step
	}
}

// WithSliderValue sets the initial value.
func WithSliderValue(value float64) SliderOption {
	return func(s *Slider) {
		s.value = value
	}
}

// WithSliderWidth sets the track width.
func WithSliderWidth(width int) SliderOption {
	return func(s *Slider) {
		s.width = width
	}
}

// WithSliderHelp sets the help text.
func WithSliderHelp(help string) SliderOption {
	return func(s *Slider) {
		s.help = help
	}
}

// WithSliderStyles sets custom styles.
func WithSliderStyles(styles SliderStyles) SliderOption {
	return func(s *Slider) {
		s.styles = styles
	}
}

// WithSliderFormat sets the value format string.
func WithSliderFormat(format string) SliderOption {
	return func(s *Slider) {
		s.format = format
	}
}

// WithSliderShowValue sets whether to show the value.
func WithSliderShowValue(show bool) SliderOption {
	return func(s *Slider) {
		s.showValue = show
	}
}

// Slider is a numeric slider component.
type Slider struct {
	label     string
	min       float64
	max       float64
	step      float64
	value     float64
	width     int
	help      string
	format    string
	showValue bool
	focused   bool
	styles    SliderStyles
}

// NewSlider creates a new slider with options.
func NewSlider(opts ...SliderOption) Slider {
	s := Slider{
		min:       0,
		max:       100,
		step:      1,
		value:     0,
		width:     20,
		format:    "%.0f",
		showValue: true,
		styles:    DefaultSliderStyles(),
	}

	for _, opt := range opts {
		opt(&s)
	}

	// Clamp initial value
	s.value = s.clamp(s.value)

	return s
}

func (s Slider) clamp(value float64) float64 {
	if value < s.min {
		return s.min
	}
	if value > s.max {
		return s.max
	}
	return value
}

// Init returns the initial command.
func (s Slider) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (s Slider) Update(msg tea.Msg) (Slider, tea.Cmd) {
	if !s.focused {
		return s, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "left", "h":
			s.value = s.clamp(s.value - s.step)
		case "right", "l":
			s.value = s.clamp(s.value + s.step)
		case "H", "home":
			s.value = s.min
		case "L", "end":
			s.value = s.max
		case "pagedown":
			bigStep := (s.max - s.min) / 10
			s.value = s.clamp(s.value - bigStep)
		case "pageup":
			bigStep := (s.max - s.min) / 10
			s.value = s.clamp(s.value + bigStep)
		}
	}

	return s, nil
}

// View renders the slider.
func (s Slider) View() string {
	var result string

	// Label
	if s.label != "" {
		result += s.styles.Label.Render(s.label) + "\n"
	}

	// Calculate fill percentage
	percentage := (s.value - s.min) / (s.max - s.min)
	if s.max == s.min {
		percentage = 0
	}

	filled := int(percentage * float64(s.width))
	if filled > s.width {
		filled = s.width
	}
	if filled < 0 {
		filled = 0
	}

	empty := s.width - filled

	// Build track
	track := s.styles.Filled.Render(repeatString("━", filled))
	track += s.styles.Thumb.Render("●")
	track += s.styles.Track.Render(repeatString("─", empty))

	// Min/max labels
	minLabel := s.styles.Help.Render(fmt.Sprintf(s.format, s.min))
	maxLabel := s.styles.Help.Render(fmt.Sprintf(s.format, s.max))

	sliderLine := minLabel + " " + track + " " + maxLabel

	if s.focused {
		result += s.styles.Focused.Render(sliderLine)
	} else {
		result += sliderLine
	}

	// Value display
	if s.showValue {
		valueStr := fmt.Sprintf(s.format, s.value)
		result += " " + s.styles.Value.Render("["+valueStr+"]")
	}

	// Help text
	if s.help != "" {
		result += "\n" + s.styles.Help.Render(s.help)
	}

	return result
}

func repeatString(s string, count int) string {
	if count <= 0 {
		return ""
	}
	result := ""
	for range count {
		result += s
	}
	return result
}

// Focus focuses the slider.
func (s Slider) Focus() Slider {
	s.focused = true
	return s
}

// Blur removes focus from the slider.
func (s Slider) Blur() Slider {
	s.focused = false
	return s
}

// Focused returns whether the slider is focused.
func (s Slider) Focused() bool {
	return s.focused
}

// Value returns the current value.
func (s Slider) Value() float64 {
	return s.value
}

// SetValue sets the slider value.
func (s Slider) SetValue(value float64) Slider {
	s.value = s.clamp(value)
	return s
}

// SetRange sets the min and max values.
func (s Slider) SetRange(minVal, maxVal float64) Slider {
	s.min = minVal
	s.max = maxVal
	s.value = s.clamp(s.value)
	return s
}

// Percentage returns the current value as a percentage (0-100).
func (s Slider) Percentage() float64 {
	if s.max == s.min {
		return 0
	}
	return (s.value - s.min) / (s.max - s.min) * 100
}

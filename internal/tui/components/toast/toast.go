// Package toast provides toast notification component for the TUI.
package toast

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Level indicates the toast notification level.
type Level int

const (
	// LevelInfo is an informational toast.
	LevelInfo Level = iota
	// LevelSuccess is a success toast.
	LevelSuccess
	// LevelWarning is a warning toast.
	LevelWarning
	// LevelError is an error toast.
	LevelError
)

// Position indicates where toasts appear on screen.
type Position int

const (
	// PositionTopRight places toasts at the top-right corner.
	PositionTopRight Position = iota
	// PositionTopLeft places toasts at the top-left corner.
	PositionTopLeft
	// PositionTopCenter places toasts at the top-center.
	PositionTopCenter
	// PositionBottomRight places toasts at the bottom-right corner.
	PositionBottomRight
	// PositionBottomLeft places toasts at the bottom-left corner.
	PositionBottomLeft
	// PositionBottomCenter places toasts at the bottom-center.
	PositionBottomCenter
)

// AnimationPhase tracks the animation state of a toast.
type AnimationPhase int

const (
	// PhaseEntering is when the toast is appearing.
	PhaseEntering AnimationPhase = iota
	// PhaseVisible is when the toast is fully visible.
	PhaseVisible
	// PhaseExiting is when the toast is disappearing.
	PhaseExiting
)

// DefaultDuration is the default toast display duration.
const DefaultDuration = 3 * time.Second

// Animation timing constants.
const (
	animationEnterDuration = 150 * time.Millisecond
	animationExitDuration  = 150 * time.Millisecond
)

// Toast represents a single toast notification.
type Toast struct {
	ID       int
	Message  string
	Level    Level
	Duration time.Duration
	Created  time.Time
	Phase    AnimationPhase
}

// dismissMsg is sent when a toast should be dismissed.
type dismissMsg struct {
	ID int
}

// phaseTransitionMsg signals a toast phase transition.
type phaseTransitionMsg struct {
	ID    int
	Phase AnimationPhase
}

// Model holds the toast notification state.
type Model struct {
	toasts   []Toast
	nextID   int
	width    int
	height   int
	visible  bool
	position Position
	animate  bool
	styles   Styles
}

// Styles for toast notifications.
type Styles struct {
	Container lipgloss.Style
	Info      lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Error     lipgloss.Style
}

// DefaultStyles returns default styles for toasts.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	baseStyle := lipgloss.NewStyle().
		Padding(0, 2).
		MarginBottom(1).
		Bold(true)

	return Styles{
		Container: lipgloss.NewStyle().
			Align(lipgloss.Right).
			Padding(1),
		Info: baseStyle.
			Foreground(colors.Text).
			Background(colors.Info),
		Success: baseStyle.
			Foreground(colors.Primary).
			Background(colors.Success),
		Warning: baseStyle.
			Foreground(colors.Primary).
			Background(colors.Warning),
		Error: baseStyle.
			Foreground(colors.Text).
			Background(colors.Error),
	}
}

// Option configures the toast model.
type Option func(*Model)

// WithPosition sets the toast position.
func WithPosition(pos Position) Option {
	return func(m *Model) {
		m.position = pos
	}
}

// WithAnimation enables or disables animation.
func WithAnimation(enabled bool) Option {
	return func(m *Model) {
		m.animate = enabled
	}
}

// WithStyles sets custom styles.
func WithStyles(styles Styles) Option {
	return func(m *Model) {
		m.styles = styles
	}
}

// New creates a new toast notification model.
func New(opts ...Option) Model {
	m := Model{
		toasts:   make([]Toast, 0),
		visible:  true,
		position: PositionTopRight,
		animate:  true,
		styles:   DefaultStyles(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the toast component.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ShowMsg:
		return m.handleShow(msg)

	case phaseTransitionMsg:
		return m.handlePhaseTransition(msg)

	case dismissMsg:
		return m.handleDismiss(msg)

	case ClearAllMsg:
		m.toasts = make([]Toast, 0)
		return m, nil
	}

	return m, nil
}

func (m Model) handleShow(msg ShowMsg) (Model, tea.Cmd) {
	toast := Toast{
		ID:       m.nextID,
		Message:  msg.Message,
		Level:    msg.Level,
		Duration: msg.Duration,
		Created:  time.Now(),
		Phase:    PhaseEntering,
	}
	if toast.Duration == 0 {
		toast.Duration = DefaultDuration
	}
	m.nextID++
	m.toasts = append(m.toasts, toast)

	var cmds []tea.Cmd

	if m.animate {
		// Schedule transition to visible phase
		cmds = append(cmds, schedulePhaseTransition(toast.ID, PhaseVisible, animationEnterDuration))
	} else {
		// Skip animation, set to visible immediately
		m.toasts[len(m.toasts)-1].Phase = PhaseVisible
	}

	// Schedule exit animation before dismissal
	exitTime := toast.Duration - animationExitDuration
	if exitTime < 0 {
		exitTime = toast.Duration / 2
	}
	if m.animate {
		cmds = append(cmds, schedulePhaseTransition(toast.ID, PhaseExiting, exitTime))
	}

	// Schedule final dismissal
	cmds = append(cmds, scheduleDismiss(toast.ID, toast.Duration))

	return m, tea.Batch(cmds...)
}

func (m Model) handlePhaseTransition(msg phaseTransitionMsg) (Model, tea.Cmd) {
	for i := range m.toasts {
		if m.toasts[i].ID == msg.ID {
			m.toasts[i].Phase = msg.Phase
			break
		}
	}
	return m, nil
}

func (m Model) handleDismiss(msg dismissMsg) (Model, tea.Cmd) {
	for i, t := range m.toasts {
		if t.ID == msg.ID {
			m.toasts = append(m.toasts[:i], m.toasts[i+1:]...)
			break
		}
	}
	return m, nil
}

// scheduleDismiss returns a command that dismisses a toast after duration.
func scheduleDismiss(id int, duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(_ time.Time) tea.Msg {
		return dismissMsg{ID: id}
	})
}

// schedulePhaseTransition returns a command that transitions a toast phase.
func schedulePhaseTransition(id int, phase AnimationPhase, delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(_ time.Time) tea.Msg {
		return phaseTransitionMsg{ID: id, Phase: phase}
	})
}

// ShowMsg shows a new toast notification.
type ShowMsg struct {
	Message  string
	Level    Level
	Duration time.Duration
}

// ClearAllMsg clears all toasts.
type ClearAllMsg struct{}

// Show returns a command to show a toast.
func Show(message string, level Level) tea.Cmd {
	return ShowWithDuration(message, level, DefaultDuration)
}

// ShowWithDuration returns a command to show a toast with custom duration.
func ShowWithDuration(message string, level Level, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		return ShowMsg{
			Message:  message,
			Level:    level,
			Duration: duration,
		}
	}
}

// Info shows an info toast.
func Info(message string) tea.Cmd {
	return Show(message, LevelInfo)
}

// Success shows a success toast.
func Success(message string) tea.Cmd {
	return Show(message, LevelSuccess)
}

// Warning shows a warning toast.
func Warning(message string) tea.Cmd {
	return Show(message, LevelWarning)
}

// Error shows an error toast.
func Error(message string) tea.Cmd {
	return Show(message, LevelError)
}

// ClearAll returns a command to clear all toasts.
func ClearAll() tea.Cmd {
	return func() tea.Msg {
		return ClearAllMsg{}
	}
}

// SetSize sets the toast container size.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// SetVisible sets whether toasts are visible.
func (m Model) SetVisible(visible bool) Model {
	m.visible = visible
	return m
}

// HasToasts returns true if there are any toasts to display.
func (m Model) HasToasts() bool {
	return len(m.toasts) > 0
}

// View renders the toast notifications.
func (m Model) View() string {
	if !m.visible || len(m.toasts) == 0 {
		return ""
	}

	// Render toasts (max 5 visible)
	maxVisible := 5
	startIdx := 0
	if len(m.toasts) > maxVisible {
		startIdx = len(m.toasts) - maxVisible
	}

	var rendered string
	for _, toast := range m.toasts[startIdx:] {
		rendered += m.renderToast(toast) + "\n"
	}

	return m.styles.Container.Width(m.width).Render(rendered)
}

func (m Model) renderToast(toast Toast) string {
	var style lipgloss.Style
	var icon string
	switch toast.Level {
	case LevelSuccess:
		style = m.styles.Success
		icon = "✓ "
	case LevelWarning:
		style = m.styles.Warning
		icon = "! "
	case LevelError:
		style = m.styles.Error
		icon = "✗ "
	default:
		style = m.styles.Info
		icon = "i "
	}

	content := icon + toast.Message

	// Apply animation effects based on phase
	switch toast.Phase {
	case PhaseEntering:
		// Fade in effect using dimmed style
		style = style.Faint(true)
	case PhaseVisible:
		// Normal visible state - no changes
	case PhaseExiting:
		// Fade out effect using dimmed style
		style = style.Faint(true)
	}

	return style.Render(content)
}

// Overlay renders the toasts as an overlay positioned at the configured location.
func (m Model) Overlay(base string) string {
	if !m.visible || len(m.toasts) == 0 {
		return base
	}

	toastView := m.View()
	if toastView == "" {
		return base
	}

	return m.positionOverlay(base, toastView)
}

func (m Model) positionOverlay(base, overlay string) string {
	if m.width == 0 || m.height == 0 {
		return base
	}

	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(strings.TrimSuffix(overlay, "\n"), "\n")

	overlayWidth := lipgloss.Width(overlay)
	overlayHeight := len(overlayLines)

	startRow, startCol := m.calculatePosition(overlayWidth, overlayHeight)

	return m.applyOverlay(baseLines, overlayLines, startRow, startCol, overlayWidth)
}

func (m Model) calculatePosition(overlayWidth, overlayHeight int) (row, col int) {
	switch m.position {
	case PositionTopRight:
		return 1, m.width - overlayWidth - 1
	case PositionTopLeft:
		return 1, 1
	case PositionTopCenter:
		return 1, (m.width - overlayWidth) / 2
	case PositionBottomRight:
		return m.height - overlayHeight - 2, m.width - overlayWidth - 1
	case PositionBottomLeft:
		return m.height - overlayHeight - 2, 1
	case PositionBottomCenter:
		return m.height - overlayHeight - 2, (m.width - overlayWidth) / 2
	default:
		return 1, m.width - overlayWidth - 1
	}
}

func (m Model) applyOverlay(baseLines, overlayLines []string, startRow, startCol, overlayWidth int) string {
	// Clamp positions
	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}

	// Overlay the toast on the base
	for i, overlayLine := range overlayLines {
		lineIdx := startRow + i
		if lineIdx >= len(baseLines) {
			break
		}
		baseLines[lineIdx] = m.overlayLine(baseLines[lineIdx], overlayLine, startCol, overlayWidth)
	}

	return strings.Join(baseLines, "\n")
}

func (m Model) overlayLine(baseLine, overlayLine string, startCol, overlayWidth int) string {
	baseRunes := []rune(baseLine)

	// Pad base line if needed
	for len(baseRunes) < startCol+overlayWidth {
		baseRunes = append(baseRunes, ' ')
	}

	// Insert overlay
	overlayRunes := []rune(overlayLine)
	for j, r := range overlayRunes {
		if startCol+j < len(baseRunes) {
			baseRunes[startCol+j] = r
		}
	}

	return string(baseRunes)
}

// SetPosition sets the toast position.
func (m Model) SetPosition(pos Position) Model {
	m.position = pos
	return m
}

// SetAnimate enables or disables animation.
func (m Model) SetAnimate(enabled bool) Model {
	m.animate = enabled
	return m
}

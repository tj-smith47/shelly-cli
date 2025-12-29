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
	toasts         []Toast
	nextID         int
	width          int
	height         int
	visible        bool
	position       Position
	animate        bool
	styles         Styles
	activeTimerID  int  // ID of toast with active dismiss timer (-1 if none)
	pendingDismiss bool // True if first Escape was pressed, waiting for second
}

// Styles for toast notifications.
type Styles struct {
	Container lipgloss.Style
	Info      lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Error     lipgloss.Style
	// InputBar styles - match search/cmdmode design
	InputBarInfo    lipgloss.Style
	InputBarSuccess lipgloss.Style
	InputBarWarning lipgloss.Style
	InputBarError   lipgloss.Style
}

// DefaultStyles returns default styles for toasts.
// Uses semantic colors for consistent theming.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	baseStyle := lipgloss.NewStyle().
		Padding(0, 2).
		MarginBottom(1).
		Bold(true)

	// Base input bar style matches search/cmdmode Container
	inputBarBase := lipgloss.NewStyle().
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder())

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
		// Input bar styles with colored borders
		InputBarInfo: inputBarBase.
			BorderForeground(colors.Info),
		InputBarSuccess: inputBarBase.
			BorderForeground(colors.Success),
		InputBarWarning: inputBarBase.
			BorderForeground(colors.Warning),
		InputBarError: inputBarBase.
			BorderForeground(colors.Error),
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
		toasts:        make([]Toast, 0),
		visible:       true,
		position:      PositionTopRight,
		animate:       true,
		styles:        DefaultStyles(),
		activeTimerID: -1,
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
		m.activeTimerID = -1
		m.pendingDismiss = false
		return m, nil

	case resetPendingDismissMsg:
		m.pendingDismiss = false
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey handles key press messages for Escape×2 dismiss.
func (m Model) handleKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	key := msg.String()

	// Only handle Esc/Ctrl+[ for dismissing toasts
	if key != "esc" && key != "ctrl+[" {
		return m, nil
	}

	if len(m.toasts) == 0 {
		return m, nil
	}

	if m.pendingDismiss {
		// Second Escape - clear all
		m.toasts = make([]Toast, 0)
		m.activeTimerID = -1
		m.pendingDismiss = false
		return m, nil
	}

	// First Escape - dismiss current toast, set pending state
	m.pendingDismiss = true

	// Remove first (currently displayed) toast
	m.toasts = m.toasts[1:]
	m.activeTimerID = -1

	var cmds []tea.Cmd

	// Start timer for next toast if any
	if len(m.toasts) > 0 {
		m.activeTimerID = m.toasts[0].ID
		cmds = append(cmds, m.startToastTimer(m.toasts[0]))
	}

	// Reset pending dismiss after 500ms
	cmds = append(cmds, tea.Tick(500*time.Millisecond, func(_ time.Time) tea.Msg {
		return resetPendingDismissMsg{}
	}))

	return m, tea.Batch(cmds...)
}

func (m Model) handleShow(msg ShowMsg) (Model, tea.Cmd) {
	t := Toast{
		ID:       m.nextID,
		Message:  msg.Message,
		Level:    msg.Level,
		Duration: msg.Duration,
		Created:  time.Now(),
		Phase:    PhaseVisible, // Start visible (no animation for queued toasts)
	}
	if t.Duration == 0 {
		t.Duration = DefaultDuration
	}
	m.nextID++
	m.toasts = append(m.toasts, t)

	// If this is the only toast, start its timer immediately
	if len(m.toasts) == 1 {
		m.activeTimerID = t.ID
		return m, m.startToastTimer(t)
	}

	// Otherwise, it's queued - timer will start when it becomes the first toast
	return m, nil
}

// startToastTimer starts the dismiss timer for a toast (timer starts when shown).
// Note: caller must set m.activeTimerID = t.ID before calling this method.
func (m Model) startToastTimer(t Toast) tea.Cmd {
	var cmds []tea.Cmd

	// Schedule exit animation before dismissal
	if m.animate {
		exitTime := t.Duration - animationExitDuration
		if exitTime < 0 {
			exitTime = t.Duration / 2
		}
		cmds = append(cmds, schedulePhaseTransition(t.ID, PhaseExiting, exitTime))
	}

	// Schedule final dismissal
	cmds = append(cmds, scheduleDismiss(t.ID, t.Duration))

	return tea.Batch(cmds...)
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
	// Only dismiss if it matches the current active timer (prevent stale dismissals)
	if msg.ID != m.activeTimerID && m.activeTimerID != -1 {
		return m, nil
	}

	for i, t := range m.toasts {
		if t.ID == msg.ID {
			m.toasts = append(m.toasts[:i], m.toasts[i+1:]...)
			m.activeTimerID = -1

			// Start timer for next toast if any
			if len(m.toasts) > 0 {
				m.activeTimerID = m.toasts[0].ID
				return m, m.startToastTimer(m.toasts[0])
			}
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

// resetPendingDismissMsg resets the pending dismiss state after timeout.
type resetPendingDismissMsg struct{}

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

// ViewAsInputBar renders the current (first) toast in the same style as search/cmdmode.
// This appears in the input bar area below the tab bar with a colored border.
// Shows a badge (e.g., "(+2)") if there are more toasts queued.
func (m Model) ViewAsInputBar() string {
	if !m.visible || len(m.toasts) == 0 {
		return ""
	}

	// Get the first toast (currently displayed)
	t := m.toasts[0]

	// Select style and icon based on level
	var style lipgloss.Style
	var icon string
	colors := theme.GetSemanticColors()

	switch t.Level {
	case LevelSuccess:
		style = m.styles.InputBarSuccess
		icon = "✓ "
	case LevelWarning:
		style = m.styles.InputBarWarning
		icon = "! "
	case LevelError:
		style = m.styles.InputBarError
		icon = "✗ "
	default:
		style = m.styles.InputBarInfo
		icon = "ℹ "
	}

	// Build the content with icon and styled message
	iconStyle := lipgloss.NewStyle().Bold(true)
	switch t.Level {
	case LevelSuccess:
		iconStyle = iconStyle.Foreground(colors.Success)
	case LevelWarning:
		iconStyle = iconStyle.Foreground(colors.Warning)
	case LevelError:
		iconStyle = iconStyle.Foreground(colors.Error)
	default:
		iconStyle = iconStyle.Foreground(colors.Info)
	}

	content := iconStyle.Render(icon) + t.Message

	// Add badge if there are more toasts queued
	if len(m.toasts) > 1 {
		badgeStyle := lipgloss.NewStyle().Foreground(colors.Muted)
		badge := badgeStyle.Render(" (+" + itoa(len(m.toasts)-1) + ")")
		content += badge
	}

	// Apply animation effects
	if t.Phase == PhaseEntering || t.Phase == PhaseExiting {
		content = lipgloss.NewStyle().Faint(true).Render(content)
	}

	return style.Width(m.width).Render(content)
}

// itoa converts an int to string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
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

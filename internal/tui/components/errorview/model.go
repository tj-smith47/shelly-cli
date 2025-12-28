// Package errorview provides consistent error display patterns for the TUI.
package errorview

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayMode determines how errors are displayed.
type DisplayMode int

const (
	// ModeBanner displays a full-width error banner.
	ModeBanner DisplayMode = iota
	// ModeInline displays an inline error message.
	ModeInline
	// ModeCompact displays a compact error icon and message.
	ModeCompact
	// ModeDetailed displays error with stack trace or details.
	ModeDetailed
)

// Severity indicates the error severity level.
type Severity int

const (
	// SeverityError is a standard error.
	SeverityError Severity = iota
	// SeverityWarning is a warning that doesn't block operation.
	SeverityWarning
	// SeverityCritical is a critical error requiring immediate attention.
	SeverityCritical
)

// Styles holds the visual styles for error display.
type Styles struct {
	Banner       lipgloss.Style
	BannerIcon   lipgloss.Style
	BannerText   lipgloss.Style
	Inline       lipgloss.Style
	InlineIcon   lipgloss.Style
	Compact      lipgloss.Style
	Detailed     lipgloss.Style
	DetailHeader lipgloss.Style
	DetailBody   lipgloss.Style
	Warning      lipgloss.Style
	WarningIcon  lipgloss.Style
	Critical     lipgloss.Style
	CriticalIcon lipgloss.Style
	Dismissible  lipgloss.Style
}

// DefaultStyles returns the default styles for error display.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Banner: lipgloss.NewStyle().
			Padding(0, 2).
			Width(80).
			Background(colors.Error).
			Foreground(colors.Text),
		BannerIcon: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		BannerText: lipgloss.NewStyle().
			Foreground(colors.Text),
		Inline: lipgloss.NewStyle().
			Foreground(colors.Error),
		InlineIcon: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true),
		Compact: lipgloss.NewStyle().
			Foreground(colors.Error).
			Italic(true),
		Detailed: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colors.Error).
			Padding(1, 2),
		DetailHeader: lipgloss.NewStyle().
			Foreground(colors.Error).
			Bold(true).
			MarginBottom(1),
		DetailBody: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Warning: lipgloss.NewStyle().
			Background(colors.Warning).
			Foreground(colors.Primary).
			Padding(0, 2),
		WarningIcon: lipgloss.NewStyle().
			Foreground(colors.Primary).
			Bold(true),
		Critical: lipgloss.NewStyle().
			Background(colors.Error).
			Foreground(colors.Text).
			Bold(true).
			Padding(1, 2),
		CriticalIcon: lipgloss.NewStyle().
			Foreground(colors.Text).
			Bold(true),
		Dismissible: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// Option configures the error view.
type Option func(*Model)

// WithMode sets the display mode.
func WithMode(mode DisplayMode) Option {
	return func(m *Model) {
		m.mode = mode
	}
}

// WithSeverity sets the error severity.
func WithSeverity(severity Severity) Option {
	return func(m *Model) {
		m.severity = severity
	}
}

// WithDismissible makes the error dismissible.
func WithDismissible(dismissible bool) Option {
	return func(m *Model) {
		m.dismissible = dismissible
	}
}

// WithDetails sets additional error details.
func WithDetails(details string) Option {
	return func(m *Model) {
		m.details = details
	}
}

// WithStyles sets custom styles.
func WithStyles(styles Styles) Option {
	return func(m *Model) {
		m.styles = styles
	}
}

// WithWidth sets the display width.
func WithWidth(width int) Option {
	return func(m *Model) {
		m.width = width
	}
}

// Model holds the error display state.
type Model struct {
	err         error
	message     string
	details     string
	mode        DisplayMode
	severity    Severity
	dismissible bool
	dismissed   bool
	width       int
	styles      Styles
}

// New creates a new error view with the given error and options.
func New(err error, opts ...Option) Model {
	m := Model{
		err:      err,
		mode:     ModeInline,
		severity: SeverityError,
		width:    80,
		styles:   DefaultStyles(),
	}

	if err != nil {
		m.message = err.Error()
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// NewFromMessage creates a new error view from a string message.
func NewFromMessage(message string, opts ...Option) Model {
	m := Model{
		message:  message,
		mode:     ModeInline,
		severity: SeverityError,
		width:    80,
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

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.dismissible || m.dismissed {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if keyMsg.String() == "esc" || keyMsg.String() == "enter" {
			m.dismissed = true
		}
	}

	return m, nil
}

// View renders the error display.
func (m Model) View() string {
	if m.dismissed || m.message == "" {
		return ""
	}

	switch m.mode {
	case ModeBanner:
		return m.viewBanner()
	case ModeInline:
		return m.viewInline()
	case ModeCompact:
		return m.viewCompact()
	case ModeDetailed:
		return m.viewDetailed()
	default:
		return m.viewInline()
	}
}

func (m Model) viewBanner() string {
	icon := m.getIcon()
	style := m.getBannerStyle()
	iconStyle := m.getIconStyle()

	content := iconStyle.Render(icon+" ") + m.styles.BannerText.Render(m.message)
	if m.dismissible {
		content += "\n" + m.styles.Dismissible.Render("Press Esc to dismiss")
	}

	return style.Width(m.width).Render(content)
}

func (m Model) viewInline() string {
	icon := m.getIcon()
	style := m.getInlineStyle()
	iconStyle := m.getIconStyle()

	return iconStyle.Render(icon+" ") + style.Render(m.message)
}

func (m Model) viewCompact() string {
	icon := m.getIcon()
	return m.styles.Compact.Render(icon + " " + m.message)
}

func (m Model) viewDetailed() string {
	icon := m.getIcon()
	header := m.styles.DetailHeader.Render(icon + " Error")

	var body strings.Builder
	body.WriteString(m.styles.BannerText.Render(m.message))

	if m.details != "" {
		body.WriteString("\n\n")
		body.WriteString(m.styles.DetailBody.Render(m.details))
	}

	if m.dismissible {
		body.WriteString("\n\n")
		body.WriteString(m.styles.Dismissible.Render("Press Esc to dismiss"))
	}

	content := header + "\n" + body.String()
	return m.styles.Detailed.Width(m.width - 4).Render(content)
}

func (m Model) getIcon() string {
	switch m.severity {
	case SeverityWarning:
		return "⚠"
	case SeverityCritical:
		return "🔴"
	default:
		return "✗"
	}
}

func (m Model) getBannerStyle() lipgloss.Style {
	switch m.severity {
	case SeverityWarning:
		return m.styles.Warning
	case SeverityCritical:
		return m.styles.Critical
	default:
		return m.styles.Banner
	}
}

func (m Model) getInlineStyle() lipgloss.Style {
	switch m.severity {
	case SeverityWarning:
		colors := theme.GetSemanticColors()
		return lipgloss.NewStyle().Foreground(colors.Warning)
	case SeverityCritical:
		return m.styles.Critical
	default:
		return m.styles.Inline
	}
}

func (m Model) getIconStyle() lipgloss.Style {
	switch m.severity {
	case SeverityWarning:
		return m.styles.WarningIcon
	case SeverityCritical:
		return m.styles.CriticalIcon
	default:
		return m.styles.InlineIcon
	}
}

// SetError sets the error to display.
func (m Model) SetError(err error) Model {
	m.err = err
	if err != nil {
		m.message = err.Error()
	} else {
		m.message = ""
	}
	m.dismissed = false
	return m
}

// SetMessage sets the error message directly.
func (m Model) SetMessage(message string) Model {
	m.message = message
	m.dismissed = false
	return m
}

// SetDetails sets additional error details.
func (m Model) SetDetails(details string) Model {
	m.details = details
	return m
}

// SetMode sets the display mode.
func (m Model) SetMode(mode DisplayMode) Model {
	m.mode = mode
	return m
}

// SetSeverity sets the error severity.
func (m Model) SetSeverity(severity Severity) Model {
	m.severity = severity
	return m
}

// SetWidth sets the display width.
func (m Model) SetWidth(width int) Model {
	m.width = width
	return m
}

// Dismiss dismisses the error.
func (m Model) Dismiss() Model {
	m.dismissed = true
	return m
}

// IsDismissed returns whether the error is dismissed.
func (m Model) IsDismissed() bool {
	return m.dismissed
}

// HasError returns whether there is an error to display.
func (m Model) HasError() bool {
	return m.message != "" && !m.dismissed
}

// Error returns the underlying error.
func (m Model) Error() error {
	return m.err
}

// Message returns the error message.
func (m Model) Message() string {
	return m.message
}

// Clear clears the error.
func (m Model) Clear() Model {
	m.err = nil
	m.message = ""
	m.details = ""
	m.dismissed = false
	return m
}

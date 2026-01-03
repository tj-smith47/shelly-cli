// Package modal provides a centered overlay dialog component for the TUI.
package modal

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/theme"
	tuistyles "github.com/tj-smith47/shelly-cli/internal/tui/styles"
)

// Size specifies how the modal dimensions are calculated.
type Size struct {
	Width     int // Fixed width in columns (0 = use percentage)
	Height    int // Fixed height in rows (0 = use percentage)
	WidthPct  int // Width as percentage of container (1-100)
	HeightPct int // Height as percentage of container (1-100)
	MaxWidth  int // Maximum width cap
	MaxHeight int // Maximum height cap
	MinWidth  int // Minimum width floor
	MinHeight int // Minimum height floor
}

// DefaultSize returns a default modal size (60% width, 50% height).
func DefaultSize() Size {
	return Size{
		WidthPct:  60,
		HeightPct: 50,
		MinWidth:  40,
		MinHeight: 10,
		MaxWidth:  120,
		MaxHeight: 40,
	}
}

// Styles holds the visual styles for the modal.
type Styles struct {
	Backdrop   lipgloss.Style
	Container  lipgloss.Style
	Title      lipgloss.Style
	TitleBar   lipgloss.Style
	Content    lipgloss.Style
	Footer     lipgloss.Style
	FooterHint lipgloss.Style
	CloseHint  lipgloss.Style
}

// DefaultStyles returns the default styles for the modal.
func DefaultStyles() Styles {
	colors := theme.GetSemanticColors()
	return Styles{
		Backdrop: lipgloss.NewStyle().
			Background(lipgloss.Color("#000000")),
		Container: tuistyles.ModalBorder().Padding(0, 1),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Text),
		TitleBar: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colors.TableBorder).
			MarginBottom(1),
		Content: lipgloss.NewStyle().
			Padding(0, 1),
		Footer: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(colors.TableBorder).
			MarginTop(1).
			Padding(0, 1),
		FooterHint: lipgloss.NewStyle().
			Foreground(colors.Muted),
		CloseHint: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
	}
}

// CloseMsg is sent when the modal should close.
type CloseMsg struct {
	Confirmed bool
}

// Model holds the modal state.
type Model struct {
	title          string
	content        string
	footer         string
	visible        bool
	size           Size
	width          int
	height         int
	scrollOffset   int
	contentHeight  int
	styles         Styles
	closeOnEsc     bool
	confirmOnEnter bool
}

// Option configures the modal model.
type Option func(*Model)

// WithTitle sets the modal title.
func WithTitle(title string) Option {
	return func(m *Model) {
		m.title = title
	}
}

// WithContent sets the modal content.
func WithContent(content string) Option {
	return func(m *Model) {
		m.content = content
		m.contentHeight = strings.Count(content, "\n") + 1
	}
}

// WithFooter sets the modal footer text.
func WithFooter(footer string) Option {
	return func(m *Model) {
		m.footer = footer
	}
}

// WithSize sets the modal size configuration.
func WithSize(size Size) Option {
	return func(m *Model) {
		m.size = size
	}
}

// WithStyles sets custom visual styles.
func WithStyles(styles Styles) Option {
	return func(m *Model) {
		m.styles = styles
	}
}

// WithCloseOnEsc enables/disables closing with Escape key.
func WithCloseOnEsc(enabled bool) Option {
	return func(m *Model) {
		m.closeOnEsc = enabled
	}
}

// WithConfirmOnEnter enables/disables confirming with Enter key.
func WithConfirmOnEnter(enabled bool) Option {
	return func(m *Model) {
		m.confirmOnEnter = enabled
	}
}

// New creates a new modal model with the given options.
func New(opts ...Option) Model {
	m := Model{
		title:          "Dialog",
		visible:        false,
		size:           DefaultSize(),
		styles:         DefaultStyles(),
		closeOnEsc:     true,
		confirmOnEnter: true,
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

// Update handles messages for the modal.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		return m.handleKeyPress(keyMsg)
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+[":
		if m.closeOnEsc {
			m.visible = false
			return m, func() tea.Msg { return CloseMsg{Confirmed: false} }
		}
	case "enter":
		if m.confirmOnEnter {
			m.visible = false
			return m, func() tea.Msg { return CloseMsg{Confirmed: true} }
		}
	case "j", "down":
		m = m.scrollDown()
	case "k", "up":
		m = m.scrollUp()
	case "g":
		m.scrollOffset = 0
	case "G":
		m = m.scrollToEnd()
	case "pgdown", "ctrl+d":
		m = m.pageDown()
	case "pgup", "ctrl+u":
		m = m.pageUp()
	}
	return m, nil
}

func (m Model) scrollDown() Model {
	maxScroll := m.maxScrollOffset()
	if m.scrollOffset < maxScroll {
		m.scrollOffset++
	}
	return m
}

func (m Model) scrollUp() Model {
	if m.scrollOffset > 0 {
		m.scrollOffset--
	}
	return m
}

func (m Model) scrollToEnd() Model {
	m.scrollOffset = m.maxScrollOffset()
	return m
}

func (m Model) pageDown() Model {
	visibleRows := m.visibleContentRows()
	m.scrollOffset += visibleRows / 2
	maxScroll := m.maxScrollOffset()
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
	return m
}

func (m Model) pageUp() Model {
	visibleRows := m.visibleContentRows() / 2
	m.scrollOffset -= visibleRows
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	return m
}

func (m Model) maxScrollOffset() int {
	visibleRows := m.visibleContentRows()
	if m.contentHeight <= visibleRows {
		return 0
	}
	return m.contentHeight - visibleRows
}

func (m Model) visibleContentRows() int {
	modalHeight := m.calculateHeight()
	// Account for: border (2), title bar (2), footer (2), padding (2)
	overhead := 8
	if m.title == "" {
		overhead -= 2
	}
	if m.footer == "" {
		overhead -= 2
	}
	visible := modalHeight - overhead
	if visible < 1 {
		visible = 1
	}
	return visible
}

// View renders the modal.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	modalWidth := m.calculateWidth()
	modalHeight := m.calculateHeight()

	// Build modal content
	var sections []string

	// Title bar
	if m.title != "" {
		titleContent := m.styles.Title.Render(m.title)
		closeHint := m.styles.CloseHint.Render("(Esc to close)")
		titleWidth := modalWidth - 4 // Account for border and padding
		spacing := titleWidth - lipgloss.Width(titleContent) - lipgloss.Width(closeHint)
		if spacing < 1 {
			spacing = 1
		}
		titleLine := titleContent + strings.Repeat(" ", spacing) + closeHint
		sections = append(sections, m.styles.TitleBar.Width(titleWidth).Render(titleLine))
	}

	// Content with scrolling
	if m.content != "" {
		contentWidth := modalWidth - 6 // Account for border, padding, and content padding
		visibleRows := m.visibleContentRows()
		scrolledContent := m.getScrolledContent(visibleRows)
		contentBlock := m.styles.Content.Width(contentWidth).Render(scrolledContent)
		sections = append(sections, contentBlock)
	}

	// Footer
	if m.footer != "" {
		footerWidth := modalWidth - 4
		footerContent := m.styles.FooterHint.Render(m.footer)
		sections = append(sections, m.styles.Footer.Width(footerWidth).Render(footerContent))
	}

	// Combine sections
	innerContent := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Apply container style
	modal := m.styles.Container.
		Width(modalWidth).
		Height(modalHeight).
		Render(innerContent)

	// Center the modal in the container
	return m.centerInContainer(modal, modalWidth, modalHeight)
}

func (m Model) getScrolledContent(visibleRows int) string {
	lines := strings.Split(m.content, "\n")
	if m.scrollOffset >= len(lines) {
		return ""
	}

	endIdx := m.scrollOffset + visibleRows
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	visibleLines := lines[m.scrollOffset:endIdx]

	// Add scroll indicator if needed
	result := strings.Join(visibleLines, "\n")
	if m.maxScrollOffset() > 0 {
		scrollInfo := m.scrollIndicator()
		result += "\n" + scrollInfo
	}

	return result
}

func (m Model) scrollIndicator() string {
	if m.contentHeight <= m.visibleContentRows() {
		return ""
	}
	position := m.scrollOffset + 1
	total := m.maxScrollOffset() + 1
	return m.styles.FooterHint.Render(
		strings.Repeat("─", 10) + " " +
			string(rune('0'+position%10)) + "/" + string(rune('0'+total%10)) + " " +
			strings.Repeat("─", 10),
	)
}

func (m Model) centerInContainer(modal string, modalWidth, modalHeight int) string {
	if m.width == 0 || m.height == 0 {
		return modal
	}

	// Calculate padding for centering
	leftPad := (m.width - modalWidth) / 2
	topPad := (m.height - modalHeight) / 2

	if leftPad < 0 {
		leftPad = 0
	}
	if topPad < 0 {
		topPad = 0
	}

	// Build centered output
	lines := strings.Split(modal, "\n")
	var result strings.Builder

	// Top padding
	for range topPad {
		result.WriteString(strings.Repeat(" ", m.width) + "\n")
	}

	// Modal lines with left padding
	leftPadStr := strings.Repeat(" ", leftPad)
	for _, line := range lines {
		result.WriteString(leftPadStr + line + "\n")
	}

	// Bottom padding (fill remaining space)
	bottomPad := m.height - topPad - len(lines)
	for range bottomPad {
		result.WriteString(strings.Repeat(" ", m.width) + "\n")
	}

	return result.String()
}

func (m Model) calculateWidth() int {
	if m.size.Width > 0 {
		return m.clampWidth(m.size.Width)
	}

	if m.width > 0 && m.size.WidthPct > 0 {
		calculated := m.width * m.size.WidthPct / 100
		return m.clampWidth(calculated)
	}

	return m.clampWidth(60) // Default fallback
}

func (m Model) calculateHeight() int {
	if m.size.Height > 0 {
		return m.clampHeight(m.size.Height)
	}

	if m.height > 0 && m.size.HeightPct > 0 {
		calculated := m.height * m.size.HeightPct / 100
		return m.clampHeight(calculated)
	}

	return m.clampHeight(20) // Default fallback
}

func (m Model) clampWidth(w int) int {
	if m.size.MinWidth > 0 && w < m.size.MinWidth {
		w = m.size.MinWidth
	}
	if m.size.MaxWidth > 0 && w > m.size.MaxWidth {
		w = m.size.MaxWidth
	}
	if m.width > 0 && w > m.width {
		w = m.width
	}
	return w
}

func (m Model) clampHeight(h int) int {
	if m.size.MinHeight > 0 && h < m.size.MinHeight {
		h = m.size.MinHeight
	}
	if m.size.MaxHeight > 0 && h > m.size.MaxHeight {
		h = m.size.MaxHeight
	}
	if m.height > 0 && h > m.height {
		h = m.height
	}
	return h
}

// SetSize sets the container dimensions for centering.
func (m Model) SetSize(w, h int) Model {
	m.width = w
	m.height = h
	return m
}

// SetTitle updates the modal title.
func (m Model) SetTitle(title string) Model {
	m.title = title
	return m
}

// SetContent updates the modal content.
func (m Model) SetContent(content string) Model {
	m.content = content
	m.contentHeight = strings.Count(content, "\n") + 1
	m.scrollOffset = 0
	return m
}

// SetFooter updates the modal footer.
func (m Model) SetFooter(footer string) Model {
	m.footer = footer
	return m
}

// Show makes the modal visible.
func (m Model) Show() Model {
	m.visible = true
	m.scrollOffset = 0
	return m
}

// Hide hides the modal.
func (m Model) Hide() Model {
	m.visible = false
	return m
}

// IsVisible returns whether the modal is visible.
func (m Model) IsVisible() bool {
	return m.visible
}

// Title returns the modal title.
func (m Model) Title() string {
	return m.title
}

// Content returns the modal content.
func (m Model) Content() string {
	return m.content
}

// Footer returns the modal footer.
func (m Model) Footer() string {
	return m.footer
}

// Overlay renders the modal on top of a base view.
// The base view is dimmed and the modal is centered on top.
func (m Model) Overlay(base string) string {
	if !m.visible {
		return base
	}

	if m.width == 0 || m.height == 0 {
		return m.View()
	}

	modalWidth := m.calculateWidth()
	modalHeight := m.calculateHeight()

	// Dim the base view
	dimmedLines := m.dimBaseView(base)

	// Render the modal box
	modalLines := m.renderModalBox(modalWidth, modalHeight)

	// Calculate centered position
	leftPad, topPad := m.calculateOverlayPosition(modalWidth, modalHeight)

	// Composite modal over dimmed base
	return m.compositeOverlay(dimmedLines, modalLines, leftPad, topPad)
}

func (m Model) dimBaseView(base string) []string {
	baseLines := strings.Split(base, "\n")
	dimmedLines := make([]string, len(baseLines))
	for i, line := range baseLines {
		dimmedLines[i] = m.styles.Backdrop.Render(line)
	}
	return dimmedLines
}

func (m Model) renderModalBox(modalWidth, modalHeight int) []string {
	var sections []string

	if m.title != "" {
		sections = append(sections, m.renderTitleBar(modalWidth))
	}

	if m.content != "" {
		contentWidth := modalWidth - 6
		visibleRows := m.visibleContentRows()
		scrolledContent := m.getScrolledContent(visibleRows)
		sections = append(sections, m.styles.Content.Width(contentWidth).Render(scrolledContent))
	}

	if m.footer != "" {
		footerWidth := modalWidth - 4
		footerContent := m.styles.FooterHint.Render(m.footer)
		sections = append(sections, m.styles.Footer.Width(footerWidth).Render(footerContent))
	}

	innerContent := lipgloss.JoinVertical(lipgloss.Left, sections...)
	modal := m.styles.Container.Width(modalWidth).Height(modalHeight).Render(innerContent)
	return strings.Split(modal, "\n")
}

func (m Model) renderTitleBar(modalWidth int) string {
	titleContent := m.styles.Title.Render(m.title)
	closeHint := m.styles.CloseHint.Render("(Esc to close)")
	titleWidth := modalWidth - 4
	spacing := titleWidth - lipgloss.Width(titleContent) - lipgloss.Width(closeHint)
	if spacing < 1 {
		spacing = 1
	}
	titleLine := titleContent + strings.Repeat(" ", spacing) + closeHint
	return m.styles.TitleBar.Width(titleWidth).Render(titleLine)
}

func (m Model) calculateOverlayPosition(modalWidth, modalHeight int) (leftPad, topPad int) {
	leftPad = (m.width - modalWidth) / 2
	topPad = (m.height - modalHeight) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	if topPad < 0 {
		topPad = 0
	}
	return leftPad, topPad
}

func (m Model) compositeOverlay(dimmedLines, modalLines []string, leftPad, topPad int) string {
	result := make([]string, m.height)
	for i := range m.height {
		if i < len(dimmedLines) {
			result[i] = dimmedLines[i]
		} else {
			result[i] = ""
		}
	}

	leftPadStr := strings.Repeat(" ", leftPad)
	for i, modalLine := range modalLines {
		row := topPad + i
		if row >= 0 && row < m.height {
			result[row] = leftPadStr + modalLine
		}
	}

	return strings.Join(result, "\n")
}

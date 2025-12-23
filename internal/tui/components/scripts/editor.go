package scripts

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

// CodeLoadedMsg signals that script code was loaded.
type CodeLoadedMsg struct {
	Code string
	Err  error
}

// StatusLoadedMsg signals that script status was loaded.
type StatusLoadedMsg struct {
	Status *shelly.ScriptStatus
	Err    error
}

// EditorDeps holds the dependencies for the script editor component.
type EditorDeps struct {
	Ctx context.Context
	Svc *shelly.Service
}

// Validate ensures all required dependencies are set.
func (d EditorDeps) Validate() error {
	if d.Ctx == nil {
		return fmt.Errorf("context is required")
	}
	if d.Svc == nil {
		return fmt.Errorf("service is required")
	}
	return nil
}

// EditorModel displays script code with syntax highlighting and status.
type EditorModel struct {
	ctx         context.Context
	svc         *shelly.Service
	device      string
	scriptID    int
	scriptName  string
	code        string
	codeLines   []string
	status      *shelly.ScriptStatus
	scroll      int
	loading     bool
	err         error
	width       int
	height      int
	focused     bool
	showNumbers bool
	styles      EditorStyles
}

// EditorStyles holds styles for the editor component.
type EditorStyles struct {
	LineNumber lipgloss.Style
	Code       lipgloss.Style
	Keyword    lipgloss.Style
	String     lipgloss.Style
	Comment    lipgloss.Style
	Header     lipgloss.Style
	Status     lipgloss.Style
	Running    lipgloss.Style
	Stopped    lipgloss.Style
	Error      lipgloss.Style
	Muted      lipgloss.Style
	Memory     lipgloss.Style
}

// DefaultEditorStyles returns the default styles for the script editor.
func DefaultEditorStyles() EditorStyles {
	colors := theme.GetSemanticColors()
	return EditorStyles{
		LineNumber: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Width(4).
			Align(lipgloss.Right),
		Code: lipgloss.NewStyle().
			Foreground(colors.Text),
		Keyword: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		String: lipgloss.NewStyle().
			Foreground(colors.Success),
		Comment: lipgloss.NewStyle().
			Foreground(colors.Muted).
			Italic(true),
		Header: lipgloss.NewStyle().
			Foreground(colors.Highlight).
			Bold(true),
		Status: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Running: lipgloss.NewStyle().
			Foreground(colors.Online),
		Stopped: lipgloss.NewStyle().
			Foreground(colors.Warning),
		Error: lipgloss.NewStyle().
			Foreground(colors.Error),
		Muted: lipgloss.NewStyle().
			Foreground(colors.Muted),
		Memory: lipgloss.NewStyle().
			Foreground(colors.Info),
	}
}

// NewEditor creates a new script editor model.
func NewEditor(deps EditorDeps) EditorModel {
	if err := deps.Validate(); err != nil {
		panic(fmt.Sprintf("scripts editor: invalid deps: %v", err))
	}

	return EditorModel{
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		showNumbers: true,
		styles:      DefaultEditorStyles(),
	}
}

// Init returns the initial command.
func (m EditorModel) Init() tea.Cmd {
	return nil
}

// SetScript sets the script to display and triggers code fetch.
func (m EditorModel) SetScript(device string, script Script) (EditorModel, tea.Cmd) {
	m.device = device
	m.scriptID = script.ID
	m.scriptName = script.Name
	m.code = ""
	m.codeLines = nil
	m.status = nil
	m.scroll = 0
	m.err = nil

	if device == "" || script.ID <= 0 {
		return m, nil
	}

	m.loading = true
	return m, tea.Batch(
		m.fetchCode(),
		m.fetchStatus(),
	)
}

// Clear clears the editor state.
func (m EditorModel) Clear() EditorModel {
	m.device = ""
	m.scriptID = 0
	m.scriptName = ""
	m.code = ""
	m.codeLines = nil
	m.status = nil
	m.scroll = 0
	m.loading = false
	m.err = nil
	return m
}

// fetchCode creates a command to fetch script code.
func (m EditorModel) fetchCode() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()

		code, err := m.svc.GetScriptCode(ctx, m.device, m.scriptID)
		return CodeLoadedMsg{Code: code, Err: err}
	}
}

// fetchStatus creates a command to fetch script status.
func (m EditorModel) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
		defer cancel()

		status, err := m.svc.GetScriptStatus(ctx, m.device, m.scriptID)
		return StatusLoadedMsg{Status: status, Err: err}
	}
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

// Update handles messages.
func (m EditorModel) Update(msg tea.Msg) (EditorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case CodeLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.loading = false
			return m, nil
		}
		m.code = msg.Code
		m.codeLines = strings.Split(msg.Code, "\n")
		m.loading = false
		return m, nil

	case StatusLoadedMsg:
		if msg.Err == nil {
			m.status = msg.Status
		}
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
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
	case "ctrl+d", "pgdown":
		m = m.pageDown()
	case "ctrl+u", "pgup":
		m = m.pageUp()
	case "n":
		m.showNumbers = !m.showNumbers
	case "r":
		m.loading = true
		return m, tea.Batch(m.fetchCode(), m.fetchStatus())
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

func (m EditorModel) pageDown() EditorModel {
	visible := m.visibleLines()
	m.scroll += visible
	maxScroll := m.maxScroll()
	if m.scroll > maxScroll {
		m.scroll = maxScroll
	}
	return m
}

func (m EditorModel) pageUp() EditorModel {
	visible := m.visibleLines()
	m.scroll -= visible
	if m.scroll < 0 {
		m.scroll = 0
	}
	return m
}

func (m EditorModel) visibleLines() int {
	lines := m.height - 6 // Account for borders, header, status
	if lines < 1 {
		return 1
	}
	return lines
}

func (m EditorModel) maxScroll() int {
	maxLines := len(m.codeLines) - m.visibleLines()
	if maxLines < 0 {
		return 0
	}
	return maxLines
}

// View renders the script editor.
func (m EditorModel) View() string {
	r := rendering.New(m.width, m.height).
		SetTitle("Code").
		SetFocused(m.focused)

	if m.scriptID == 0 {
		r.SetContent(m.styles.Muted.Render("No script selected"))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.styles.Muted.Render("Loading script code..."))
		return r.Render()
	}

	if m.err != nil {
		r.SetContent(m.styles.Error.Render("Error: " + m.err.Error()))
		return r.Render()
	}

	var content strings.Builder

	// Script header
	name := m.scriptName
	if name == "" {
		name = fmt.Sprintf("script_%d", m.scriptID)
	}
	content.WriteString(m.styles.Header.Render(name))

	// Status info
	if m.status != nil {
		content.WriteString(" ")
		if m.status.Running {
			content.WriteString(m.styles.Running.Render("(running)"))
		} else {
			content.WriteString(m.styles.Stopped.Render("(stopped)"))
		}

		// Memory info
		if m.status.MemUsage > 0 {
			memStr := fmt.Sprintf(" [mem: %d/%d KB]",
				m.status.MemUsage/1024,
				(m.status.MemUsage+m.status.MemFree)/1024,
			)
			content.WriteString(m.styles.Memory.Render(memStr))
		}
	}
	content.WriteString("\n\n")

	// Code with line numbers
	content.WriteString(m.renderCodeLines())

	r.SetContent(content.String())
	return r.Render()
}

// renderCodeLines renders the code with line numbers and scroll indicator.
func (m EditorModel) renderCodeLines() string {
	if len(m.codeLines) == 0 {
		return m.styles.Muted.Render("(empty script)")
	}

	var content strings.Builder
	visible := m.visibleLines()
	endIdx := m.scroll + visible
	if endIdx > len(m.codeLines) {
		endIdx = len(m.codeLines)
	}

	for i := m.scroll; i < endIdx; i++ {
		line := m.codeLines[i]
		if m.showNumbers {
			lineNum := m.styles.LineNumber.Render(fmt.Sprintf("%3d", i+1))
			content.WriteString(lineNum + " ")
		}
		content.WriteString(m.highlightLine(line))
		if i < endIdx-1 {
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(m.codeLines) > visible {
		content.WriteString(m.styles.Muted.Render(
			fmt.Sprintf("\n\n[%d-%d/%d lines]",
				m.scroll+1, endIdx, len(m.codeLines)),
		))
	}

	return content.String()
}

// highlightLine applies basic syntax highlighting to a line.
func (m EditorModel) highlightLine(line string) string {
	// Basic highlighting: comments only for now
	trimmed := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(trimmed, "//") {
		return m.styles.Comment.Render(line)
	}
	return m.styles.Code.Render(line)
}

// ScriptID returns the current script ID.
func (m EditorModel) ScriptID() int {
	return m.scriptID
}

// ScriptName returns the current script name.
func (m EditorModel) ScriptName() string {
	return m.scriptName
}

// Code returns the current script code.
func (m EditorModel) Code() string {
	return m.code
}

// LineCount returns the number of lines in the code.
func (m EditorModel) LineCount() int {
	return len(m.codeLines)
}

// Loading returns whether the component is loading.
func (m EditorModel) Loading() bool {
	return m.loading
}

// Error returns any error that occurred.
func (m EditorModel) Error() error {
	return m.err
}

// Status returns the current script status.
func (m EditorModel) Status() *shelly.ScriptStatus {
	return m.status
}

// Refresh triggers a refresh of the script code and status.
func (m EditorModel) Refresh() (EditorModel, tea.Cmd) {
	if m.device == "" || m.scriptID == 0 {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.fetchCode(), m.fetchStatus())
}

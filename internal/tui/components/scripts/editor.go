package scripts

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/generics"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
	"github.com/tj-smith47/shelly-cli/internal/tui/rendering"
	"github.com/tj-smith47/shelly-cli/internal/tui/styles"
	"github.com/tj-smith47/shelly-cli/internal/tui/tuierrors"
)

// CodeLoadedMsg signals that script code was loaded.
type CodeLoadedMsg struct {
	Code string
	Err  error
}

// StatusLoadedMsg signals that script status was loaded.
type StatusLoadedMsg struct {
	Status *automation.ScriptStatus
	Err    error
}

// EditorFinishedMsg signals that external editor closed.
type EditorFinishedMsg struct {
	Device   string
	ScriptID int
	Code     string
	Err      error
}

// CodeUploadedMsg signals that code was uploaded to device.
type CodeUploadedMsg struct {
	Device   string
	ScriptID int
	Err      error
}

// EditorDeps holds the dependencies for the script editor component.
type EditorDeps struct {
	Ctx context.Context
	Svc *automation.Service
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
	helpers.Sizable
	ctx              context.Context
	svc              *automation.Service
	device           string
	scriptID         int
	scriptName       string
	code             string
	codeLines        []string
	highlightedLines []string // Chroma-highlighted lines (from theme)
	status           *automation.ScriptStatus
	scroll           int
	loading          bool
	err              error
	focused          bool
	panelIndex       int // 1-based panel index for Shift+N hotkey hint
	showNumbers      bool
	styles           EditorStyles
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
		iostreams.DebugErr("scripts editor component init", err)
		panic(fmt.Sprintf("scripts editor: invalid deps: %v", err))
	}

	m := EditorModel{
		Sizable:     helpers.NewSizableLoaderOnly(),
		ctx:         deps.Ctx,
		svc:         deps.Svc,
		showNumbers: true,
		styles:      DefaultEditorStyles(),
	}
	m.Loader = m.Loader.SetMessage("Loading script code...")
	return m
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
		m.Loader.Tick(),
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
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		code, err := m.svc.GetScriptCode(ctx, m.device, m.scriptID)
		return CodeLoadedMsg{Code: code, Err: err}
	}
}

// fetchStatus creates a command to fetch script status.
func (m EditorModel) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()

		status, err := m.svc.GetScriptStatus(ctx, m.device, m.scriptID)
		return StatusLoadedMsg{Status: status, Err: err}
	}
}

// SetSize sets the component dimensions.
func (m EditorModel) SetSize(width, height int) EditorModel {
	m.ApplySize(width, height)
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
	// Forward tick messages to loader when loading
	if m.loading {
		result := generics.UpdateLoader(m.Loader, msg, func(msg tea.Msg) bool {
			switch msg.(type) {
			case CodeLoadedMsg, StatusLoadedMsg:
				return true
			}
			return false
		})
		m.Loader = result.Loader
		if result.Consumed {
			return m, result.Cmd
		}
	}

	switch msg := msg.(type) {
	case CodeLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.loading = false
			return m, nil
		}
		m.code = msg.Code
		m.codeLines = strings.Split(msg.Code, "\n")
		// Generate syntax-highlighted lines using theme colors
		highlighted := theme.HighlightJavaScript(msg.Code)
		m.highlightedLines = strings.Split(highlighted, "\n")
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
		return m, tea.Batch(m.Loader.Tick(), m.fetchCode(), m.fetchStatus())
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
	lines := m.Height - 6 // Account for borders, header, status
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
	r := rendering.New(m.Width, m.Height).
		SetTitle("Code").
		SetFocused(m.focused).
		SetPanelIndex(m.panelIndex)

	if m.scriptID == 0 {
		r.SetContent(styles.EmptyStateWithBorder("No script selected", m.Width, m.Height))
		return r.Render()
	}

	if m.loading {
		r.SetContent(m.Loader.View())
		return r.Render()
	}

	if m.err != nil {
		if tuierrors.IsUnsupportedFeature(m.err) {
			r.SetContent(styles.EmptyStateWithBorder(tuierrors.UnsupportedMessage("Scripts"), m.Width, m.Height))
		} else {
			msg, _ := tuierrors.FormatError(m.err)
			r.SetContent(m.styles.Error.Render(msg))
		}
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
		if m.showNumbers {
			lineNum := m.styles.LineNumber.Render(fmt.Sprintf("%3d", i+1))
			content.WriteString(lineNum + " ")
		}
		// Use pre-highlighted lines from chroma (theme-aware)
		if i < len(m.highlightedLines) {
			content.WriteString(m.highlightedLines[i])
		} else {
			// Fallback to plain code if highlighting failed
			content.WriteString(m.codeLines[i])
		}
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
func (m EditorModel) Status() *automation.ScriptStatus {
	return m.status
}

// Refresh triggers a refresh of the script code and status.
func (m EditorModel) Refresh() (EditorModel, tea.Cmd) {
	if m.device == "" || m.scriptID == 0 {
		return m, nil
	}
	m.loading = true
	return m, tea.Batch(m.Loader.Tick(), m.fetchCode(), m.fetchStatus())
}

// Edit opens the script in an external editor.
// Following the superfile pattern: saves code to temp file, launches $EDITOR or nano,
// then reads the modified code back and signals completion.
func (m EditorModel) Edit() tea.Cmd {
	if m.device == "" || m.scriptID == 0 {
		return nil
	}

	device := m.device
	scriptID := m.scriptID
	code := m.code

	// Create temp file with script code
	tmpFile, err := os.CreateTemp("", "shelly-script-*.js")
	if err != nil {
		return func() tea.Msg {
			return EditorFinishedMsg{Device: device, ScriptID: scriptID, Err: err}
		}
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.WriteString(code); err != nil {
		iostreams.CloseWithDebug("closing temp file on error", tmpFile)
		if rmErr := os.Remove(tmpPath); rmErr != nil {
			iostreams.DebugErr("cleanup temp file on error", rmErr)
		}
		return func() tea.Msg {
			return EditorFinishedMsg{Device: device, ScriptID: scriptID, Err: err}
		}
	}
	iostreams.CloseWithDebug("closing temp file before editor", tmpFile)

	// Get editor: config setting > EDITOR env > VISUAL env > nano
	editor := config.GetEditor()
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "nano" // fallback to nano
	}

	// Split editor command into parts (handles "vim -u NONE" etc.)
	parts := strings.Fields(editor)
	editorCmd := parts[0]
	editorArgs := make([]string, len(parts)-1, len(parts))
	copy(editorArgs, parts[1:])
	editorArgs = append(editorArgs, tmpPath)

	//nolint:gosec,noctx // G204: User's EDITOR env var; context N/A for tea.ExecProcess
	c := exec.Command(editorCmd, editorArgs...)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		//nolint:gosec // G304: Reading temp file we created - safe and expected
		modifiedCode, readErr := os.ReadFile(tmpPath)
		if rmErr := os.Remove(tmpPath); rmErr != nil {
			iostreams.DebugErr("cleanup temp file after editor", rmErr)
		}

		if err != nil {
			return EditorFinishedMsg{Device: device, ScriptID: scriptID, Err: err}
		}
		if readErr != nil {
			return EditorFinishedMsg{Device: device, ScriptID: scriptID, Err: readErr}
		}

		return EditorFinishedMsg{Device: device, ScriptID: scriptID, Code: string(modifiedCode)}
	})
}

// Device returns the current device address.
func (m EditorModel) Device() string {
	return m.device
}

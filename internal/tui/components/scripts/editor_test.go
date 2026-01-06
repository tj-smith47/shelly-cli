package scripts

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/tui/helpers"
)

func TestEditorDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    EditorDeps
		wantErr bool
	}{
		{
			name:    "nil context",
			deps:    EditorDeps{Ctx: nil, Svc: nil},
			wantErr: true,
		},
		{
			name:    "nil service",
			deps:    EditorDeps{Ctx: context.Background(), Svc: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.deps.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEditorModel_SetSize(t *testing.T) {
	t.Parallel()
	m := EditorModel{}
	m = m.SetSize(100, 50)

	if m.Width != 100 {
		t.Errorf("width = %d, want 100", m.Width)
	}
	if m.Height != 50 {
		t.Errorf("height = %d, want 50", m.Height)
	}
}

func TestEditorModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := EditorModel{}

	m = m.SetFocused(true)
	if !m.focused {
		t.Error("focused = false, want true")
	}

	m = m.SetFocused(false)
	if m.focused {
		t.Error("focused = true, want false")
	}
}

func TestEditorModel_Clear(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		device:     "192.168.1.100",
		scriptID:   1,
		scriptName: "test",
		code:       "console.log('hello');",
		codeLines:  []string{"console.log('hello');"},
		loading:    true,
		err:        context.DeadlineExceeded,
	}

	m = m.Clear()

	if m.device != "" {
		t.Errorf("device = %q, want empty", m.device)
	}
	if m.scriptID != 0 {
		t.Errorf("scriptID = %d, want 0", m.scriptID)
	}
	if m.scriptName != "" {
		t.Errorf("scriptName = %q, want empty", m.scriptName)
	}
	if m.code != "" {
		t.Errorf("code = %q, want empty", m.code)
	}
	if len(m.codeLines) != 0 {
		t.Errorf("codeLines length = %d, want 0", len(m.codeLines))
	}
	if m.loading {
		t.Error("loading = true, want false")
	}
	if m.err != nil {
		t.Errorf("err = %v, want nil", m.err)
	}
}

func TestEditorModel_VisibleLines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{"zero height", 0, 1},
		{"small height", 7, 1},
		{"normal height", 20, 14},
		{"large height", 50, 44},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := EditorModel{Sizable: helpers.NewSizableLoaderOnly()}
			m = m.SetSize(80, tt.height)
			got := m.visibleLines()
			if got != tt.want {
				t.Errorf("visibleLines() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEditorModel_MaxScroll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		codeLines int
		height    int
		want      int
	}{
		{"no code", 0, 20, 0},
		{"fits in view", 5, 20, 0},
		{"needs scroll", 30, 20, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lines := make([]string, tt.codeLines)
			m := EditorModel{Sizable: helpers.NewSizableLoaderOnly(), codeLines: lines}
			m = m.SetSize(80, tt.height)
			got := m.maxScroll()
			if got != tt.want {
				t.Errorf("maxScroll() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEditorModel_ScrollNavigation(t *testing.T) {
	t.Parallel()
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "code"
	}

	m := EditorModel{
		Sizable:   helpers.NewSizableLoaderOnly(),
		codeLines: lines,
		scroll:    0,
	}
	m = m.SetSize(80, 20)

	// Scroll down
	m = m.scrollDown()
	if m.scroll != 1 {
		t.Errorf("after scrollDown: scroll = %d, want 1", m.scroll)
	}

	// Scroll up
	m = m.scrollUp()
	if m.scroll != 0 {
		t.Errorf("after scrollUp: scroll = %d, want 0", m.scroll)
	}

	// Don't scroll below 0
	m = m.scrollUp()
	if m.scroll != 0 {
		t.Errorf("scroll below 0: scroll = %d, want 0", m.scroll)
	}

	// Scroll to end
	m = m.scrollToEnd()
	maxScroll := m.maxScroll()
	if m.scroll != maxScroll {
		t.Errorf("scrollToEnd: scroll = %d, want %d", m.scroll, maxScroll)
	}

	// Don't scroll past max
	m = m.scrollDown()
	if m.scroll != maxScroll {
		t.Errorf("scroll past max: scroll = %d, want %d", m.scroll, maxScroll)
	}
}

func TestEditorModel_PageNavigation(t *testing.T) {
	t.Parallel()
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "code"
	}

	m := EditorModel{
		Sizable:   helpers.NewSizableLoaderOnly(),
		codeLines: lines,
		scroll:    50,
	}
	m = m.SetSize(80, 20)

	visible := m.visibleLines()

	// Page up
	m = m.pageUp()
	expected := 50 - visible
	if m.scroll != expected {
		t.Errorf("after pageUp: scroll = %d, want %d", m.scroll, expected)
	}

	// Page up from near start doesn't go negative
	m.scroll = 5
	m = m.pageUp()
	if m.scroll != 0 {
		t.Errorf("pageUp from 5: scroll = %d, want 0", m.scroll)
	}

	// Page down
	m.scroll = 0
	m = m.pageDown()
	if m.scroll != visible {
		t.Errorf("after pageDown: scroll = %d, want %d", m.scroll, visible)
	}

	// Page down from near end doesn't exceed max
	maxScroll := m.maxScroll()
	m.scroll = maxScroll - 5
	m = m.pageDown()
	if m.scroll != maxScroll {
		t.Errorf("pageDown near end: scroll = %d, want %d", m.scroll, maxScroll)
	}
}

func TestEditorModel_Getters(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		scriptID:   5,
		scriptName: "myScript",
		code:       "line1\nline2\nline3",
		codeLines:  []string{"line1", "line2", "line3"},
		loading:    true,
		err:        context.DeadlineExceeded,
		status:     &automation.ScriptStatus{Running: true},
	}

	if m.ScriptID() != 5 {
		t.Errorf("ScriptID() = %d, want 5", m.ScriptID())
	}
	if m.ScriptName() != "myScript" {
		t.Errorf("ScriptName() = %q, want %q", m.ScriptName(), "myScript")
	}
	if m.Code() != "line1\nline2\nline3" {
		t.Errorf("Code() = %q, want code", m.Code())
	}
	if m.LineCount() != 3 {
		t.Errorf("LineCount() = %d, want 3", m.LineCount())
	}
	if !m.Loading() {
		t.Error("Loading() = false, want true")
	}
	if !errors.Is(m.Error(), context.DeadlineExceeded) {
		t.Errorf("Error() = %v, want %v", m.Error(), context.DeadlineExceeded)
	}
	if m.Status() == nil || !m.Status().Running {
		t.Error("Status() should return running status")
	}
}

func TestEditorModel_View_NoScript(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		Sizable:  helpers.NewSizableLoaderOnly(),
		scriptID: 0,
		styles:   DefaultEditorStyles(),
	}
	m = m.SetSize(50, 20)

	view := m.View()
	if !strings.Contains(view, "No script selected") {
		t.Errorf("View() should show 'No script selected', got:\n%s", view)
	}
}

func TestEditorModel_View_Loading(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		Sizable:  helpers.NewSizableLoaderOnly(),
		scriptID: 1,
		loading:  true,
		styles:   DefaultEditorStyles(),
	}
	m = m.SetSize(50, 20)

	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("View() should show 'Loading', got:\n%s", view)
	}
}

func TestEditorModel_View_Error(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		Sizable:  helpers.NewSizableLoaderOnly(),
		scriptID: 1,
		err:      context.DeadlineExceeded,
		styles:   DefaultEditorStyles(),
	}
	m = m.SetSize(50, 20)

	view := m.View()
	if !strings.Contains(view, "Error") {
		t.Errorf("View() should show 'Error', got:\n%s", view)
	}
}

func TestEditorModel_View_EmptyCode(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		Sizable:    helpers.NewSizableLoaderOnly(),
		scriptID:   1,
		scriptName: "empty_script",
		codeLines:  []string{},
		styles:     DefaultEditorStyles(),
	}
	m = m.SetSize(50, 20)

	view := m.View()
	if !strings.Contains(view, "empty script") {
		t.Errorf("View() should show 'empty script', got:\n%s", view)
	}
}

func TestEditorModel_View_WithCode(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		Sizable:     helpers.NewSizableLoaderOnly(),
		scriptID:    1,
		scriptName:  "test_script",
		codeLines:   []string{"let x = 1;", "let y = 2;", "// comment"},
		showNumbers: true,
		styles:      DefaultEditorStyles(),
	}
	m = m.SetSize(60, 20)

	view := m.View()
	if !strings.Contains(view, "test_script") {
		t.Errorf("View() should show script name, got:\n%s", view)
	}
}

func TestEditorModel_View_WithStatus(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		Sizable:    helpers.NewSizableLoaderOnly(),
		scriptID:   1,
		scriptName: "running_script",
		codeLines:  []string{"code"},
		status: &automation.ScriptStatus{
			Running:  true,
			MemUsage: 8192,
			MemFree:  16384,
		},
		styles: DefaultEditorStyles(),
	}
	m = m.SetSize(60, 20)

	view := m.View()
	if !strings.Contains(view, "running") {
		t.Errorf("View() should show 'running', got:\n%s", view)
	}
}

func TestEditorModel_SyntaxHighlighting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
	}{
		{"regular code", "let x = 1;"},
		{"comment", "// this is a comment"},
		{"indented comment", "    // indented comment"},
		{"empty", ""},
		{"function", "function foo() { return 42; }"},
		{"string", `const msg = "hello world";`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := theme.HighlightJavaScript(tt.code)
			// Just verify it returns something (highlighting may add ANSI codes)
			if result == "" && tt.code != "" {
				t.Error("HighlightJavaScript() returned empty for non-empty code")
			}
		})
	}
}

func TestEditorModel_RenderCodeLines(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		Sizable:     helpers.NewSizableLoaderOnly(),
		codeLines:   []string{"line1", "line2", "line3"},
		showNumbers: true,
		scroll:      0,
		styles:      DefaultEditorStyles(),
	}
	m = m.SetSize(80, 20)

	result := m.renderCodeLines()
	if result == "" {
		t.Error("renderCodeLines() returned empty string")
	}
}

func TestEditorModel_RenderCodeLines_Empty(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		Sizable:   helpers.NewSizableLoaderOnly(),
		codeLines: []string{},
		styles:    DefaultEditorStyles(),
	}
	m = m.SetSize(80, 20)

	result := m.renderCodeLines()
	if !strings.Contains(result, "empty") {
		t.Errorf("renderCodeLines() should indicate empty, got: %s", result)
	}
}

func TestEditorModel_Update_CodeLoadedMsg(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		loading: true,
		styles:  DefaultEditorStyles(),
	}

	code := "let x = 1;\nlet y = 2;"
	m, _ = m.Update(CodeLoadedMsg{Code: code})

	if m.loading {
		t.Error("loading should be false after CodeLoadedMsg")
	}
	if m.code != code {
		t.Errorf("code = %q, want %q", m.code, code)
	}
	if len(m.codeLines) != 2 {
		t.Errorf("codeLines length = %d, want 2", len(m.codeLines))
	}
}

func TestEditorModel_Update_CodeLoadedMsg_Error(t *testing.T) {
	t.Parallel()
	m := EditorModel{
		loading: true,
		styles:  DefaultEditorStyles(),
	}

	m, _ = m.Update(CodeLoadedMsg{Err: context.DeadlineExceeded})

	if m.loading {
		t.Error("loading should be false after error")
	}
	if !errors.Is(m.err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want %v", m.err, context.DeadlineExceeded)
	}
}

func TestEditorModel_Update_StatusLoadedMsg(t *testing.T) {
	t.Parallel()
	m := EditorModel{styles: DefaultEditorStyles()}

	status := &automation.ScriptStatus{Running: true, MemUsage: 1024}
	m, _ = m.Update(StatusLoadedMsg{Status: status})

	if m.status == nil {
		t.Fatal("status should not be nil")
	}
	if !m.status.Running {
		t.Error("status.Running = false, want true")
	}
}

func TestEditorModel_Init(t *testing.T) {
	t.Parallel()
	m := EditorModel{}
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestDefaultEditorStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultEditorStyles()

	// Just verify styles are created without panic
	_ = styles.LineNumber.Render("1")
	_ = styles.Code.Render("code")
	_ = styles.Keyword.Render("let")
	_ = styles.String.Render("hello")
	_ = styles.Comment.Render("// comment")
	_ = styles.Header.Render("Script")
	_ = styles.Status.Render("status")
	_ = styles.Running.Render("running")
	_ = styles.Stopped.Render("stopped")
	_ = styles.Error.Render("error")
	_ = styles.Muted.Render("muted")
	_ = styles.Memory.Render("1KB")
}

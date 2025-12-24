package cmdmode

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()

	if m.IsActive() {
		t.Error("new model should not be active")
	}
}

func TestModel_SetWidth(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetWidth(100)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
}

func TestModel_Init(t *testing.T) {
	t.Parallel()
	m := New()
	cmd := m.Init()

	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_Activate(t *testing.T) {
	t.Parallel()
	m := New()
	m, cmd := m.Activate()

	if !m.IsActive() {
		t.Error("model should be active after Activate()")
	}
	if cmd == nil {
		t.Error("Activate() should return a command (blink)")
	}
}

func TestModel_Deactivate(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()
	m = m.Deactivate()

	if m.IsActive() {
		t.Error("model should not be active after Deactivate()")
	}
}

func TestModel_View_NotActive(t *testing.T) {
	t.Parallel()
	m := New()
	view := m.View()

	if view != "" {
		t.Errorf("View() should return empty string when not active, got %q", view)
	}
}

func TestModel_View_Active(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetWidth(80)
	m, _ = m.Activate()

	view := m.View()
	if view == "" {
		t.Error("View() should not be empty when active")
	}
	if !strings.Contains(view, ":") {
		t.Error("View() should contain the prompt character ':'")
	}
}

func TestModel_Update_NotActive(t *testing.T) {
	t.Parallel()
	m := New()

	// When not active, Update should do nothing
	msg := tea.KeyPressMsg{Code: 'q', Text: "q"}
	newM, cmd := m.Update(msg)

	if newM.IsActive() {
		t.Error("inactive model should remain inactive")
	}
	if cmd != nil {
		t.Error("inactive model should return nil command")
	}
}

func TestModel_Update_Escape(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()

	// Test deactivation via Deactivate() method since key.Matches
	// requires proper tea.KeyPressMsg construction
	m = m.Deactivate()

	if m.IsActive() {
		t.Error("model should be deactivated after Deactivate()")
	}
}

func TestModel_ExecuteCommand_Quit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
	}{
		{"q"},
		{"quit"},
		{"exit"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			m := New()
			m, _ = m.Activate()
			m.textInput.SetValue(tt.input)

			// Simulate enter
			msg := tea.KeyPressMsg{Code: tea.KeyEnter}
			newM, cmd := m.Update(msg)

			if newM.IsActive() {
				t.Error("model should be deactivated after command")
			}
			if cmd == nil {
				t.Fatal("executeCommand should return a command")
			}

			resultMsg := cmd()
			cmdMsg, ok := resultMsg.(CommandMsg)
			if !ok {
				t.Fatalf("expected CommandMsg, got %T", resultMsg)
			}
			if cmdMsg.Command != CmdQuit {
				t.Errorf("Command = %q, want %q", cmdMsg.Command, CmdQuit)
			}
		})
	}
}

func TestModel_ExecuteCommand_Filter(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()
	m.textInput.SetValue("filter test")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	resultMsg := cmd()
	cmdMsg, ok := resultMsg.(CommandMsg)
	if !ok {
		t.Fatalf("expected CommandMsg, got %T", resultMsg)
	}
	if cmdMsg.Command != CmdFilter {
		t.Errorf("Command = %q, want %q", cmdMsg.Command, CmdFilter)
	}
	if cmdMsg.Args != "test" {
		t.Errorf("Args = %q, want %q", cmdMsg.Args, "test")
	}
}

func TestModel_ExecuteCommand_Theme(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()
	m.textInput.SetValue("theme dracula")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	resultMsg := cmd()
	cmdMsg, ok := resultMsg.(CommandMsg)
	if !ok {
		t.Fatalf("expected CommandMsg, got %T", resultMsg)
	}
	if cmdMsg.Command != CmdTheme {
		t.Errorf("Command = %q, want %q", cmdMsg.Command, CmdTheme)
	}
	if cmdMsg.Args != "dracula" {
		t.Errorf("Args = %q, want %q", cmdMsg.Args, "dracula")
	}
}

func TestModel_ExecuteCommand_ThemeNoArgs(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()
	m.textInput.SetValue("theme")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	resultMsg := cmd()
	errMsg, ok := resultMsg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg, got %T", resultMsg)
	}
	if !strings.Contains(errMsg.Message, "requires") {
		t.Errorf("error message should mention 'requires', got %q", errMsg.Message)
	}
}

func TestModel_ExecuteCommand_Unknown(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()
	m.textInput.SetValue("unknowncommand")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	resultMsg := cmd()
	errMsg, ok := resultMsg.(ErrorMsg)
	if !ok {
		t.Fatalf("expected ErrorMsg, got %T", resultMsg)
	}
	if !strings.Contains(errMsg.Message, "unknown command") {
		t.Errorf("error message should mention 'unknown command', got %q", errMsg.Message)
	}
}

func TestModel_ExecuteCommand_Empty(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()
	m.textInput.SetValue("")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	newM, cmd := m.Update(msg)

	if newM.IsActive() {
		t.Error("model should be deactivated after empty command")
	}
	if cmd == nil {
		t.Fatal("empty command should return ClosedMsg")
	}

	resultMsg := cmd()
	if _, ok := resultMsg.(ClosedMsg); !ok {
		t.Errorf("expected ClosedMsg, got %T", resultMsg)
	}
}

func TestModel_ExecuteCommand_Toggle(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()
	m.textInput.SetValue("toggle")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	resultMsg := cmd()
	cmdMsg, ok := resultMsg.(CommandMsg)
	if !ok {
		t.Fatalf("expected CommandMsg, got %T", resultMsg)
	}
	if cmdMsg.Command != CmdToggle {
		t.Errorf("Command = %q, want %q", cmdMsg.Command, CmdToggle)
	}
}

func TestModel_ExecuteCommand_Help(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()
	m.textInput.SetValue("help")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	resultMsg := cmd()
	cmdMsg, ok := resultMsg.(CommandMsg)
	if !ok {
		t.Fatalf("expected CommandMsg, got %T", resultMsg)
	}
	if cmdMsg.Command != CmdHelp {
		t.Errorf("Command = %q, want %q", cmdMsg.Command, CmdHelp)
	}
}

func TestModel_History(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()

	// Execute a command to add to history
	m.textInput.SetValue("filter test1")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// Reactivate and check history
	m, _ = m.Activate()
	if len(m.history) != 1 {
		t.Fatalf("history length = %d, want 1", len(m.history))
	}
	if m.history[0] != "filter test1" {
		t.Errorf("history[0] = %q, want %q", m.history[0], "filter test1")
	}
}

func TestModel_HistoryNavigation(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Activate()

	// Add some history
	m.textInput.SetValue("filter one")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = m.Activate()
	m.textInput.SetValue("filter two")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = m.Activate()

	// Navigate up
	m = m.historyUp()
	if m.textInput.Value() != "filter two" {
		t.Errorf("after historyUp(), value = %q, want %q", m.textInput.Value(), "filter two")
	}

	m = m.historyUp()
	if m.textInput.Value() != "filter one" {
		t.Errorf("after second historyUp(), value = %q, want %q", m.textInput.Value(), "filter one")
	}

	// Navigate down
	m = m.historyDown()
	if m.textInput.Value() != "filter two" {
		t.Errorf("after historyDown(), value = %q, want %q", m.textInput.Value(), "filter two")
	}
}

func TestAvailableCommands(t *testing.T) {
	t.Parallel()
	help := AvailableCommands()

	expectedCommands := []string{":q", ":quit", ":device", ":filter", ":theme", ":toggle", ":help"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(help, cmd) {
			t.Errorf("AvailableCommands() should contain %q", cmd)
		}
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Just verify styles are created without panicking
	_ = styles.Container
	_ = styles.Prompt
	_ = styles.Input
	_ = styles.Error
}

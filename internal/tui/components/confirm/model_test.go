package confirm

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

const (
	testOp      = "factory-reset"
	testTitle   = "Factory Reset"
	testMessage = "This will erase all device settings."
	testPhrase  = "CONFIRM"
)

func TestNew(t *testing.T) {
	t.Parallel()

	m := New()

	if m.visible {
		t.Error("should not be visible initially")
	}
	if m.input != "" {
		t.Error("input should be empty")
	}
}

func TestModel_Init(t *testing.T) {
	t.Parallel()
	m := New()

	cmd := m.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestModel_Show(t *testing.T) {
	t.Parallel()
	m := New()

	updated := m.Show(testOp, testTitle, testMessage, testPhrase)

	if !updated.visible {
		t.Error("should be visible after Show")
	}
	if updated.operation != testOp {
		t.Errorf("operation = %q, want %q", updated.operation, testOp)
	}
	if updated.title != testTitle {
		t.Errorf("title = %q, want %q", updated.title, testTitle)
	}
	if updated.message != testMessage {
		t.Errorf("message = %q, want %q", updated.message, testMessage)
	}
	if updated.confirmPhrase != testPhrase {
		t.Errorf("confirmPhrase = %q, want %q", updated.confirmPhrase, testPhrase)
	}
	if updated.input != "" {
		t.Error("input should be empty after Show")
	}
}

func TestModel_Show_ClearsInput(t *testing.T) {
	t.Parallel()
	m := New()
	m.input = "previous"

	updated := m.Show(testOp, testTitle, testMessage, testPhrase)

	if updated.input != "" {
		t.Error("input should be cleared after Show")
	}
}

func TestModel_Hide(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m.input = "some input"

	updated := m.Hide()

	if updated.visible {
		t.Error("should not be visible after Hide")
	}
	if updated.input != "" {
		t.Error("input should be cleared after Hide")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := New()

	updated := m.SetSize(100, 50)

	if updated.width != 100 {
		t.Errorf("width = %d, want 100", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height = %d, want 50", updated.height)
	}
}

func TestModel_Update_NotVisible(t *testing.T) {
	t.Parallel()
	m := New()

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'a'})

	if cmd != nil {
		t.Error("should not return command when not visible")
	}
	if updated.input != "" {
		t.Error("should not accept input when not visible")
	}
}

func TestModel_Update_TypeInput(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)

	// Type each character of "CONFIRM"
	for _, c := range testPhrase {
		m, _ = m.Update(tea.KeyPressMsg{Code: c})
	}

	if m.input != testPhrase {
		t.Errorf("input = %q, want %q", m.input, testPhrase)
	}
}

func TestModel_Update_Backspace(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m.input = "CON"

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if updated.input != "CO" {
		t.Errorf("input = %q, want %q", updated.input, "CO")
	}
}

func TestModel_Update_BackspaceEmpty(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if updated.input != "" {
		t.Error("input should remain empty")
	}
}

func TestModel_Update_Escape(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m.input = "partial"

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if updated.visible {
		t.Error("should be hidden after Escape")
	}
	if updated.input != "" {
		t.Error("input should be cleared after Escape")
	}
	if cmd == nil {
		t.Error("should return CancelledMsg command")
	}
	// Execute the command to check the message
	msg := cmd()
	if cancelMsg, ok := msg.(CancelledMsg); !ok {
		t.Error("expected CancelledMsg")
	} else if cancelMsg.Operation != testOp {
		t.Errorf("Operation = %q, want %q", cancelMsg.Operation, testOp)
	}
}

func TestModel_Update_EnterNoMatch(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m.input = "wrong"

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !updated.visible {
		t.Error("should remain visible when input doesn't match")
	}
	if cmd != nil {
		t.Error("should not return command when input doesn't match")
	}
}

func TestModel_Update_EnterMatch(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m.input = testPhrase

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if updated.visible {
		t.Error("should be hidden after confirm")
	}
	if updated.input != "" {
		t.Error("input should be cleared after confirm")
	}
	if cmd == nil {
		t.Error("should return ConfirmedMsg command")
	}
	// Execute the command to check the message
	msg := cmd()
	if confirmMsg, ok := msg.(ConfirmedMsg); !ok {
		t.Error("expected ConfirmedMsg")
	} else if confirmMsg.Operation != testOp {
		t.Errorf("Operation = %q, want %q", confirmMsg.Operation, testOp)
	}
}

func TestModel_Update_EnterMatchCaseInsensitive(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m.input = strings.ToLower(testPhrase)

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if updated.visible {
		t.Error("should be hidden - case insensitive match")
	}
	if cmd == nil {
		t.Error("should return ConfirmedMsg command - case insensitive match")
	}
}

func TestModel_InputMatches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  bool
	}{
		{"CONFIRM", true},
		{"confirm", true},
		{"Confirm", true},
		{"CONFI", false},
		{"CONFIRMX", false},
		{"", false},
	}

	for _, tt := range tests {
		m := New()
		m = m.Show(testOp, testTitle, testMessage, testPhrase)
		m.input = tt.input

		if got := m.InputMatches(); got != tt.want {
			t.Errorf("InputMatches() with input=%q = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestModel_View_NotVisible(t *testing.T) {
	t.Parallel()
	m := New()

	view := m.View()

	if view != "" {
		t.Error("View() should be empty when not visible")
	}
}

func TestModel_View_Visible(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not be empty when visible")
	}
}

func TestModel_View_WithInput(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m.input = "CON"
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not be empty")
	}
}

func TestModel_View_InputMatches(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m.input = testPhrase
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not be empty")
	}
}

func TestModel_Accessors(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m.input = "test"

	if !m.Visible() {
		t.Error("Visible() should be true")
	}
	if m.Operation() != testOp {
		t.Errorf("Operation() = %q, want %q", m.Operation(), testOp)
	}
	if m.Input() != "test" {
		t.Errorf("Input() = %q, want %q", m.Input(), "test")
	}
	if m.ConfirmPhrase() != testPhrase {
		t.Errorf("ConfirmPhrase() = %q, want %q", m.ConfirmPhrase(), testPhrase)
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Verify styles are created without panic
	_ = styles.Border.Render("test")
	_ = styles.Title.Render("test")
	_ = styles.Message.Render("test")
	_ = styles.Warning.Render("test")
	_ = styles.Input.Render("test")
	_ = styles.Prompt.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.Danger.Render("test")
	_ = styles.Highlight.Render("test")
}

func TestWithModalOverlay(t *testing.T) {
	t.Parallel()
	m := New(WithModalOverlay())

	if !m.useModal {
		t.Error("WithModalOverlay should set useModal to true")
	}
}

func TestWithStyles(t *testing.T) {
	t.Parallel()
	customStyles := DefaultStyles()
	m := New(WithStyles(customStyles))

	// Styles should be applied
	_ = m.styles.Border.Render("test")
}

func TestModel_Show_WithModal(t *testing.T) {
	t.Parallel()
	m := New(WithModalOverlay())

	updated := m.Show(testOp, testTitle, testMessage, testPhrase)

	if !updated.visible {
		t.Error("should be visible after Show")
	}
	// Modal should also be visible when useModal is true
	if !updated.modal.IsVisible() {
		t.Error("modal should be visible when useModal is true")
	}
}

func TestModel_Hide_WithModal(t *testing.T) {
	t.Parallel()
	m := New(WithModalOverlay())
	m = m.Show(testOp, testTitle, testMessage, testPhrase)

	updated := m.Hide()

	if updated.visible {
		t.Error("should not be visible after Hide")
	}
	if updated.modal.IsVisible() {
		t.Error("modal should be hidden when Hide is called")
	}
}

func TestModel_SetSize_WithModal(t *testing.T) {
	t.Parallel()
	m := New(WithModalOverlay())

	updated := m.SetSize(100, 50)

	if updated.width != 100 {
		t.Errorf("width = %d, want 100", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height = %d, want 50", updated.height)
	}
}

func TestModel_View_WithModal(t *testing.T) {
	t.Parallel()
	m := New(WithModalOverlay())
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not be empty when visible with modal")
	}
}

func TestModel_Overlay_NotVisible(t *testing.T) {
	t.Parallel()
	m := New()
	base := "base content"

	result := m.Overlay(base)

	if result != base {
		t.Error("Overlay should return base when not visible")
	}
}

func TestModel_Overlay_WithModal(t *testing.T) {
	t.Parallel()
	m := New(WithModalOverlay())
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m = m.SetSize(80, 40)
	base := "base content\nline 2\nline 3"

	result := m.Overlay(base)

	// Result should contain the modal overlay
	if result == base {
		t.Error("Overlay should modify base when modal is visible")
	}
}

func TestModel_Overlay_NoModal(t *testing.T) {
	t.Parallel()
	m := New() // No modal
	m = m.Show(testOp, testTitle, testMessage, testPhrase)
	m = m.SetSize(80, 40)
	base := "base content"

	result := m.Overlay(base)

	// Without modal, Overlay returns the View
	if result == "" {
		t.Error("Overlay should return View when not using modal")
	}
}

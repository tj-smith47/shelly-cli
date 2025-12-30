package toast

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()

	if m.HasToasts() {
		t.Error("new model should not have toasts")
	}
	if !m.visible {
		t.Error("new model should be visible")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestModel_SetVisible(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetVisible(false)

	if m.visible {
		t.Error("visible should be false after SetVisible(false)")
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

func TestModel_Update_ShowMsg(t *testing.T) {
	t.Parallel()
	m := New()

	msg := ShowMsg{
		Message:  "Test toast",
		Level:    LevelSuccess,
		Duration: time.Second,
	}
	newM, cmd := m.Update(msg)

	if !newM.HasToasts() {
		t.Error("model should have toasts after ShowMsg")
	}
	if len(newM.toasts) != 1 {
		t.Errorf("toasts count = %d, want 1", len(newM.toasts))
	}
	if newM.toasts[0].Message != "Test toast" {
		t.Errorf("toast message = %q, want %q", newM.toasts[0].Message, "Test toast")
	}
	if newM.toasts[0].Level != LevelSuccess {
		t.Errorf("toast level = %v, want LevelSuccess", newM.toasts[0].Level)
	}
	if cmd == nil {
		t.Error("Update(ShowMsg) should return a dismiss command")
	}
}

func TestModel_Update_ShowMsg_DefaultDuration(t *testing.T) {
	t.Parallel()
	m := New()

	msg := ShowMsg{
		Message: "Test toast",
		Level:   LevelInfo,
		// Duration not set
	}
	newM, _ := m.Update(msg)

	if newM.toasts[0].Duration != DefaultDuration {
		t.Errorf("duration = %v, want %v", newM.toasts[0].Duration, DefaultDuration)
	}
}

func TestModel_Update_DismissMsg(t *testing.T) {
	t.Parallel()
	m := New()

	// Add a toast
	showMsg := ShowMsg{Message: "Test", Level: LevelInfo, Duration: time.Second}
	m, _ = m.Update(showMsg)

	if len(m.toasts) != 1 {
		t.Fatalf("expected 1 toast, got %d", len(m.toasts))
	}

	// Dismiss it
	toastID := m.toasts[0].ID
	dismissM := dismissMsg{ID: toastID}
	m, _ = m.Update(dismissM)

	if m.HasToasts() {
		t.Error("toast should be dismissed")
	}
}

func TestModel_Update_ClearAllMsg(t *testing.T) {
	t.Parallel()
	m := New()

	// Add multiple toasts
	m, _ = m.Update(ShowMsg{Message: "Toast 1", Level: LevelInfo})
	m, _ = m.Update(ShowMsg{Message: "Toast 2", Level: LevelSuccess})
	m, _ = m.Update(ShowMsg{Message: "Toast 3", Level: LevelError})

	if len(m.toasts) != 3 {
		t.Fatalf("expected 3 toasts, got %d", len(m.toasts))
	}

	// Clear all
	m, _ = m.Update(ClearAllMsg{})

	if m.HasToasts() {
		t.Error("all toasts should be cleared")
	}
}

func TestModel_View_Empty(t *testing.T) {
	t.Parallel()
	m := New()

	view := m.View()
	if view != "" {
		t.Errorf("View() should return empty string when no toasts, got %q", view)
	}
}

func TestModel_View_NotVisible(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Update(ShowMsg{Message: "Test", Level: LevelInfo})
	m = m.SetVisible(false)

	view := m.View()
	if view != "" {
		t.Error("View() should return empty string when not visible")
	}
}

func TestModel_View_WithToasts(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(80, 40)
	m, _ = m.Update(ShowMsg{Message: "Test toast", Level: LevelSuccess})

	view := m.View()
	if view == "" {
		t.Error("View() should not be empty when has toasts")
	}
}

func TestShow(t *testing.T) {
	t.Parallel()
	cmd := Show("test message", LevelError)

	if cmd == nil {
		t.Fatal("Show() should return a command")
	}

	msg := cmd()
	showMsg, ok := msg.(ShowMsg)
	if !ok {
		t.Fatalf("expected ShowMsg, got %T", msg)
	}
	if showMsg.Message != "test message" {
		t.Errorf("Message = %q, want %q", showMsg.Message, "test message")
	}
	if showMsg.Level != LevelError {
		t.Errorf("Level = %v, want LevelError", showMsg.Level)
	}
	if showMsg.Duration != DefaultDuration {
		t.Errorf("Duration = %v, want %v", showMsg.Duration, DefaultDuration)
	}
}

func TestShowWithDuration(t *testing.T) {
	t.Parallel()
	duration := 5 * time.Second
	cmd := ShowWithDuration("test", LevelInfo, duration)

	msg := cmd()
	showMsg, ok := msg.(ShowMsg)
	if !ok {
		t.Fatalf("expected ShowMsg, got %T", msg)
	}
	if showMsg.Duration != duration {
		t.Errorf("Duration = %v, want %v", showMsg.Duration, duration)
	}
}

func TestConvenienceFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(string) tea.Cmd
		want Level
	}{
		{"Info", Info, LevelInfo},
		{"Success", Success, LevelSuccess},
		{"Warning", Warning, LevelWarning},
		{"Error", Error, LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := tt.fn("test")
			msg := cmd()
			showMsg, ok := msg.(ShowMsg)
			if !ok {
				t.Fatalf("expected ShowMsg, got %T", msg)
			}
			if showMsg.Level != tt.want {
				t.Errorf("Level = %v, want %v", showMsg.Level, tt.want)
			}
		})
	}
}

func TestClearAll(t *testing.T) {
	t.Parallel()
	cmd := ClearAll()

	if cmd == nil {
		t.Fatal("ClearAll() should return a command")
	}

	msg := cmd()
	_, ok := msg.(ClearAllMsg)
	if !ok {
		t.Fatalf("expected ClearAllMsg, got %T", msg)
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Just verify styles are created without panicking
	_ = styles.Info
	_ = styles.Success
	_ = styles.Warning
	_ = styles.Error
}

func TestModel_QueueBehavior_TimerStartsOnFirstToastOnly(t *testing.T) {
	t.Parallel()
	m := New()

	// Add first toast - should return a command (timer starts)
	m, cmd1 := m.Update(ShowMsg{Message: "First", Level: LevelInfo, Duration: time.Second})
	if cmd1 == nil {
		t.Error("first toast should return a command (timer)")
	}
	if m.activeTimerID != m.toasts[0].ID {
		t.Errorf("activeTimerID = %d, want %d", m.activeTimerID, m.toasts[0].ID)
	}

	// Add second toast - should NOT return a command (queued, no timer)
	m, cmd2 := m.Update(ShowMsg{Message: "Second", Level: LevelInfo, Duration: time.Second})
	if cmd2 != nil {
		t.Error("queued toast should not return a command")
	}
	if len(m.toasts) != 2 {
		t.Errorf("toasts count = %d, want 2", len(m.toasts))
	}
	// activeTimerID should still be first toast
	if m.activeTimerID != m.toasts[0].ID {
		t.Errorf("activeTimerID should still be first toast")
	}
}

func TestModel_DismissStartsNextToastTimer(t *testing.T) {
	t.Parallel()
	m := New()

	// Add two toasts
	m, _ = m.Update(ShowMsg{Message: "First", Level: LevelInfo, Duration: time.Second})
	m, _ = m.Update(ShowMsg{Message: "Second", Level: LevelInfo, Duration: time.Second})

	firstID := m.toasts[0].ID
	secondID := m.toasts[1].ID

	// Dismiss first toast
	m, cmd := m.Update(dismissMsg{ID: firstID})

	// Should have one toast remaining
	if len(m.toasts) != 1 {
		t.Fatalf("expected 1 toast after dismiss, got %d", len(m.toasts))
	}

	// activeTimerID should now be second toast
	if m.activeTimerID != secondID {
		t.Errorf("activeTimerID = %d, want %d", m.activeTimerID, secondID)
	}

	// Should return a command for next toast's timer
	if cmd == nil {
		t.Error("dismiss should return a command to start next toast's timer")
	}
}

func TestModel_StaleDismissIgnored(t *testing.T) {
	t.Parallel()
	m := New()

	// Add toast
	m, _ = m.Update(ShowMsg{Message: "Test", Level: LevelInfo, Duration: time.Second})
	currentID := m.toasts[0].ID

	// Try to dismiss with wrong ID (stale dismiss message)
	staleID := currentID + 100
	m, _ = m.Update(dismissMsg{ID: staleID})

	// Toast should NOT be dismissed
	if len(m.toasts) != 1 {
		t.Error("stale dismiss should be ignored")
	}
}

func TestModel_HandleKey_FirstEscapeDismissesCurrentToast(t *testing.T) {
	t.Parallel()
	m := New()

	// Add two toasts
	m, _ = m.Update(ShowMsg{Message: "First", Level: LevelInfo})
	m, _ = m.Update(ShowMsg{Message: "Second", Level: LevelInfo})

	// Press Escape
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// First toast should be dismissed
	if len(m.toasts) != 1 {
		t.Errorf("expected 1 toast after first Esc, got %d", len(m.toasts))
	}
	if m.toasts[0].Message != "Second" {
		t.Errorf("remaining toast should be Second, got %q", m.toasts[0].Message)
	}

	// pendingDismiss should be set
	if !m.pendingDismiss {
		t.Error("pendingDismiss should be true after first Esc")
	}
}

func TestModel_HandleKey_SecondEscapeClearsAll(t *testing.T) {
	t.Parallel()
	m := New()

	// Add three toasts
	m, _ = m.Update(ShowMsg{Message: "First", Level: LevelInfo})
	m, _ = m.Update(ShowMsg{Message: "Second", Level: LevelInfo})
	m, _ = m.Update(ShowMsg{Message: "Third", Level: LevelInfo})

	// Simulate first Escape - dismiss current, set pendingDismiss
	m.pendingDismiss = true
	m.toasts = m.toasts[1:] // Remove first toast

	// Second Escape (within 500ms window)
	m, _ = m.handleKey(tea.KeyPressMsg{Code: tea.KeyEscape})

	// All toasts should be cleared
	if len(m.toasts) != 0 {
		t.Errorf("expected 0 toasts after second Esc, got %d", len(m.toasts))
	}
	if m.pendingDismiss {
		t.Error("pendingDismiss should be false after clearing all")
	}
	if m.activeTimerID != -1 {
		t.Errorf("activeTimerID should be -1 after clearing all, got %d", m.activeTimerID)
	}
}

func TestModel_HandleKey_NoToastsNoOp(t *testing.T) {
	t.Parallel()
	m := New()

	// Press Escape with no toasts
	newM, cmd := m.handleKey(tea.KeyPressMsg{Code: tea.KeyEscape})

	if cmd != nil {
		t.Error("Esc with no toasts should return nil cmd")
	}
	if newM.pendingDismiss {
		t.Error("pendingDismiss should not be set with no toasts")
	}
}

func TestModel_HandleKey_NonEscapeKeyIgnored(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Update(ShowMsg{Message: "Test", Level: LevelInfo})

	// Press a different key
	m, _ = m.handleKey(tea.KeyPressMsg{Code: 'a'})

	// Toast should still be there
	if len(m.toasts) != 1 {
		t.Error("non-Esc key should not dismiss toast")
	}
}

func TestModel_ViewAsInputBar_Empty(t *testing.T) {
	t.Parallel()
	m := New()

	view := m.ViewAsInputBar()
	if view != "" {
		t.Error("ViewAsInputBar should return empty string when no toasts")
	}
}

func TestModel_ViewAsInputBar_SingleToast(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(80, 40)
	m, _ = m.Update(ShowMsg{Message: "Success message", Level: LevelSuccess})

	view := m.ViewAsInputBar()
	if view == "" {
		t.Error("ViewAsInputBar should not be empty with a toast")
	}
	// Should not have badge since only one toast
	if contains(view, "(+") {
		t.Error("ViewAsInputBar should not show badge for single toast")
	}
}

func TestModel_ViewAsInputBar_WithBadge(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetSize(80, 40)

	// Add multiple toasts
	m, _ = m.Update(ShowMsg{Message: "First", Level: LevelInfo})
	m, _ = m.Update(ShowMsg{Message: "Second", Level: LevelInfo})
	m, _ = m.Update(ShowMsg{Message: "Third", Level: LevelInfo})

	view := m.ViewAsInputBar()

	// Should show badge "(+2)" for 2 remaining queued toasts
	if !contains(view, "(+2)") {
		t.Errorf("ViewAsInputBar should show badge (+2), got: %s", view)
	}
}

func TestModel_ViewAsInputBar_LevelIcons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level Level
		icon  string
	}{
		{LevelSuccess, "✓"},
		{LevelWarning, "!"},
		{LevelError, "✗"},
		{LevelInfo, "ℹ"},
	}

	for _, tt := range tests {
		m := New()
		m = m.SetSize(80, 40)
		m, _ = m.Update(ShowMsg{Message: "Test", Level: tt.level})

		view := m.ViewAsInputBar()
		if !contains(view, tt.icon) {
			t.Errorf("ViewAsInputBar for level %v should contain icon %s", tt.level, tt.icon)
		}
	}
}

func TestModel_ResetPendingDismissMsg(t *testing.T) {
	t.Parallel()
	m := New()
	m.pendingDismiss = true

	m, _ = m.Update(resetPendingDismissMsg{})

	if m.pendingDismiss {
		t.Error("pendingDismiss should be false after resetPendingDismissMsg")
	}
}

func TestModel_ClearAllResetsPendingDismiss(t *testing.T) {
	t.Parallel()
	m := New()
	m, _ = m.Update(ShowMsg{Message: "Test", Level: LevelInfo})
	m.pendingDismiss = true
	m.activeTimerID = 5

	m, _ = m.Update(ClearAllMsg{})

	if m.pendingDismiss {
		t.Error("ClearAllMsg should reset pendingDismiss")
	}
	if m.activeTimerID != -1 {
		t.Errorf("ClearAllMsg should reset activeTimerID to -1, got %d", m.activeTimerID)
	}
}

func TestModel_New_InitializesActiveTimerID(t *testing.T) {
	t.Parallel()
	m := New()

	if m.activeTimerID != -1 {
		t.Errorf("new model should have activeTimerID = -1, got %d", m.activeTimerID)
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(s != "" && substr != "" && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

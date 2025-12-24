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

func TestModel_Overlay(t *testing.T) {
	t.Parallel()
	m := New()

	base := "base content"
	result := m.Overlay(base)

	// Overlay with no toasts should return base unchanged
	if result != base {
		t.Errorf("Overlay() = %q, want %q", result, base)
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Just verify styles are created without panicking
	_ = styles.Container
	_ = styles.Info
	_ = styles.Success
	_ = styles.Warning
	_ = styles.Error
}

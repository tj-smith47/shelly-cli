package statusbar

import (
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()

	if m.message != "Ready" {
		t.Errorf("message = %q, want %q", m.message, "Ready")
	}
	if m.messageType != MessageNormal {
		t.Errorf("messageType = %v, want MessageNormal", m.messageType)
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

func TestModel_Update_StatusMsg(t *testing.T) {
	t.Parallel()
	m := New()

	msg := StatusMsg{Message: "Test message", Type: MessageSuccess}
	newM, _ := m.Update(msg)

	if newM.message != "Test message" {
		t.Errorf("message = %q, want %q", newM.message, "Test message")
	}
	if newM.messageType != MessageSuccess {
		t.Errorf("messageType = %v, want MessageSuccess", newM.messageType)
	}
}

func TestModel_Update_TickMsg(t *testing.T) {
	t.Parallel()
	m := New()
	originalTime := m.lastUpdate

	// Give it time to be different
	time.Sleep(time.Millisecond)

	msg := tickMsg(time.Now())
	newM, cmd := m.Update(msg)

	if !newM.lastUpdate.After(originalTime) {
		t.Error("lastUpdate should be updated after tick")
	}
	if cmd == nil {
		t.Error("Update(tickMsg) should return a command for next tick")
	}
}

func TestModel_Init(t *testing.T) {
	t.Parallel()
	m := New()
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return a tick command")
	}
}

func TestModel_View(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetWidth(80)

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
	if !strings.Contains(view, "Ready") {
		t.Errorf("View() should contain initial message 'Ready', got: %q", view)
	}
}

func TestModel_View_MessageTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		msgType MessageType
		message string
	}{
		{"normal", MessageNormal, "Normal message"},
		{"success", MessageSuccess, "Success message"},
		{"error", MessageError, "Error message"},
		{"warning", MessageWarning, "Warning message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := New()
			m = m.SetWidth(80)

			msg := StatusMsg{Message: tt.message, Type: tt.msgType}
			m, _ = m.Update(msg)

			view := m.View()
			if !strings.Contains(view, tt.message) {
				t.Errorf("View() should contain message %q", tt.message)
			}
		})
	}
}

func TestSetMessage(t *testing.T) {
	t.Parallel()
	cmd := SetMessage("test message", MessageError)

	if cmd == nil {
		t.Fatal("SetMessage() should return a command")
	}

	msg := cmd()
	statusMsg, ok := msg.(StatusMsg)
	if !ok {
		t.Fatalf("expected StatusMsg, got %T", msg)
	}
	if statusMsg.Message != "test message" {
		t.Errorf("Message = %q, want %q", statusMsg.Message, "test message")
	}
	if statusMsg.Type != MessageError {
		t.Errorf("Type = %v, want MessageError", statusMsg.Type)
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Just verify styles are created without panicking
	_ = styles.Bar
	_ = styles.Normal
	_ = styles.Success
	_ = styles.Error
	_ = styles.Warning
	_ = styles.Debug
}

func TestModel_SetDebugActive(t *testing.T) {
	t.Parallel()
	m := New()

	if m.IsDebugActive() {
		t.Error("debug should be inactive by default")
	}

	m = m.SetDebugActive(true)
	if !m.IsDebugActive() {
		t.Error("debug should be active after SetDebugActive(true)")
	}

	m = m.SetDebugActive(false)
	if m.IsDebugActive() {
		t.Error("debug should be inactive after SetDebugActive(false)")
	}
}

func TestModel_View_DebugIndicator(t *testing.T) {
	t.Parallel()
	m := New()
	m = m.SetWidth(120)
	m = m.SetDebugActive(true)

	view := m.View()
	if !strings.Contains(view, "Debug active") {
		t.Errorf("View() should contain 'Debug active' when debug is enabled in full tier, got: %q", view)
	}

	// Test compact tier
	m = m.SetWidth(80)
	view = m.View()
	if !strings.Contains(view, "REC") {
		t.Errorf("View() should contain 'REC' when debug is enabled in compact tier, got: %q", view)
	}
}

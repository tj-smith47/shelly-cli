package control

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// mockPluginService implements PluginService for testing.
type mockPluginService struct {
	lastAction    string
	lastComponent string
	lastID        int
	err           error
}

func (m *mockPluginService) PluginControl(_ context.Context, _, action, comp string, id int) error {
	m.lastAction = action
	m.lastComponent = comp
	m.lastID = id
	return m.err
}

func TestNewPlugin(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{}
	components := []PluginComponent{
		{Type: "switch", ID: 0, Name: "Relay", On: true},
		{Type: "light", ID: 1, Name: "Dimmer", On: false},
	}

	m := NewPlugin(context.Background(), svc, "device1", "tasmota", components)

	if len(m.Components()) != 2 {
		t.Fatalf("expected 2 components, got %d", len(m.Components()))
	}
	if m.Cursor() != 0 {
		t.Fatalf("expected cursor at 0, got %d", m.Cursor())
	}
	if !m.Focused() {
		t.Fatal("expected focused by default")
	}
	if m.platform != "tasmota" {
		t.Fatalf("expected platform tasmota, got %s", m.platform)
	}
}

func TestPluginNavigation(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{}
	components := []PluginComponent{
		{Type: "switch", ID: 0, Name: "Relay 0"},
		{Type: "switch", ID: 1, Name: "Relay 1"},
		{Type: "cover", ID: 0, Name: "Shutter"},
	}

	m := NewPlugin(context.Background(), svc, "device1", "tasmota", components)

	// Navigate down
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.Cursor() != 1 {
		t.Fatalf("expected cursor 1 after j, got %d", m.Cursor())
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.Cursor() != 2 {
		t.Fatalf("expected cursor 2 after j, got %d", m.Cursor())
	}

	// Should not go past last item
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})
	if m.Cursor() != 2 {
		t.Fatalf("expected cursor 2 at boundary, got %d", m.Cursor())
	}

	// Navigate up
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if m.Cursor() != 1 {
		t.Fatalf("expected cursor 1 after k, got %d", m.Cursor())
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if m.Cursor() != 0 {
		t.Fatalf("expected cursor 0 after k, got %d", m.Cursor())
	}

	// Should not go below 0
	m, _ = m.Update(tea.KeyPressMsg{Code: 'k'})
	if m.Cursor() != 0 {
		t.Fatalf("expected cursor 0 at boundary, got %d", m.Cursor())
	}
}

func TestPluginToggle(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{}
	components := []PluginComponent{
		{Type: "switch", ID: 0, Name: "Relay", On: false},
	}

	m := NewPlugin(context.Background(), svc, "device1", "tasmota", components)

	// Toggle with 't' key
	var cmd tea.Cmd
	m, cmd = m.Update(tea.KeyPressMsg{Code: 't'})
	if cmd == nil {
		t.Fatal("expected command from toggle")
	}
	if !m.loading {
		t.Fatal("expected loading state after toggle")
	}

	// Execute the command to get ActionMsg
	msg := cmd()
	actionMsg, ok := msg.(ActionMsg)
	if !ok {
		t.Fatalf("expected ActionMsg, got %T", msg)
	}
	if actionMsg.Action != actionToggle {
		t.Fatalf("expected toggle action, got %s", actionMsg.Action)
	}

	// Verify service was called
	if svc.lastAction != "toggle" {
		t.Fatalf("expected service action toggle, got %s", svc.lastAction)
	}
	if svc.lastComponent != "switch" {
		t.Fatalf("expected service component switch, got %s", svc.lastComponent)
	}
	if svc.lastID != 0 {
		t.Fatalf("expected service ID 0, got %d", svc.lastID)
	}

	// Apply result
	m, _ = m.Update(ActionMsg{Action: actionToggle})
	if m.loading {
		t.Fatal("expected loading cleared")
	}
	if !m.components[0].On {
		t.Fatal("expected switch toggled on")
	}
}

func TestPluginOnOff(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{}
	components := []PluginComponent{
		{Type: "light", ID: 0, Name: "LED", On: false},
	}

	m := NewPlugin(context.Background(), svc, "device1", "esphome", components)

	// On with 'o'
	m, cmd := m.Update(tea.KeyPressMsg{Code: 'o'})
	if cmd == nil {
		t.Fatal("expected command from on")
	}
	msg := cmd()
	actionMsg, ok := msg.(ActionMsg)
	if !ok {
		t.Fatalf("expected ActionMsg, got %T", msg)
	}
	if actionMsg.Action != actionOn {
		t.Fatalf("expected on action, got %s", actionMsg.Action)
	}

	m, _ = m.Update(ActionMsg{Action: actionOn})
	if !m.components[0].On {
		t.Fatal("expected light turned on")
	}

	// Off with 'O'
	m, cmd = m.Update(tea.KeyPressMsg{Code: 'O'})
	if cmd == nil {
		t.Fatal("expected command from off")
	}
	msg = cmd()
	actionMsg, ok = msg.(ActionMsg)
	if !ok {
		t.Fatalf("expected ActionMsg, got %T", msg)
	}
	if actionMsg.Action != actionOff {
		t.Fatalf("expected off action, got %s", actionMsg.Action)
	}

	m, _ = m.Update(ActionMsg{Action: actionOff})
	if m.components[0].On {
		t.Fatal("expected light turned off")
	}
}

func TestPluginError(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{err: fmt.Errorf("plugin hook failed")}
	components := []PluginComponent{
		{Type: "switch", ID: 0, Name: "Relay", On: false},
	}

	m := NewPlugin(context.Background(), svc, "device1", "tasmota", components)

	m, cmd := m.Update(tea.KeyPressMsg{Code: 't'})
	msg := cmd()
	m, _ = m.Update(msg)

	if m.errorMsg == "" {
		t.Fatal("expected error message set")
	}
	if m.errorMsg != "plugin hook failed" {
		t.Fatalf("expected 'plugin hook failed', got %s", m.errorMsg)
	}
}

func TestPluginViewNoComponents(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{}
	m := NewPlugin(context.Background(), svc, "device1", "tasmota", nil)

	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}
	if !strings.Contains(view, "No controllable components") {
		t.Fatal("expected 'No controllable components' in view")
	}
}

func TestPluginViewWithComponents(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{}
	components := []PluginComponent{
		{Type: "switch", ID: 0, Name: "Relay", On: true},
		{Type: "cover", ID: 0, Name: "Shutter", State: "open"},
	}

	m := NewPlugin(context.Background(), svc, "device1", "tasmota", components)
	view := m.View()

	if !strings.Contains(view, "tasmota") {
		t.Fatal("expected platform name in view")
	}
	if !strings.Contains(view, "Relay") {
		t.Fatal("expected component name 'Relay' in view")
	}
	if !strings.Contains(view, "Shutter") {
		t.Fatal("expected component name 'Shutter' in view")
	}
}

func TestPluginSetSize(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{}
	m := NewPlugin(context.Background(), svc, "device1", "tasmota", nil)

	m = m.SetSize(80, 40)
	if m.width != 80 {
		t.Fatalf("expected width 80, got %d", m.width)
	}
	if m.height != 40 {
		t.Fatalf("expected height 40, got %d", m.height)
	}
}

func TestPluginSetFocused(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{}
	m := NewPlugin(context.Background(), svc, "device1", "tasmota", nil)

	m = m.SetFocused(false)
	if m.Focused() {
		t.Fatal("expected unfocused")
	}

	// Keys should be ignored when unfocused
	m2, cmd := m.Update(tea.KeyPressMsg{Code: 't'})
	if cmd != nil {
		t.Fatal("expected no command when unfocused")
	}
	_ = m2
}

func TestPluginLoadingBlocksInput(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{}
	components := []PluginComponent{
		{Type: "switch", ID: 0, Name: "Relay", On: false},
	}

	m := NewPlugin(context.Background(), svc, "device1", "tasmota", components)
	m.loading = true

	// Keys should be ignored when loading
	_, cmd := m.Update(tea.KeyPressMsg{Code: 't'})
	if cmd != nil {
		t.Fatal("expected no command when loading")
	}
}

func TestPluginCoverState(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{}
	components := []PluginComponent{
		{Type: "cover", ID: 0, Name: "Shutter", State: "open"},
	}

	m := NewPlugin(context.Background(), svc, "device1", "tasmota", components)
	view := m.View()

	if !strings.Contains(view, "OPEN") {
		t.Fatal("expected OPEN state in view")
	}
}

func TestPluginComponentNameFallback(t *testing.T) {
	t.Parallel()
	svc := &mockPluginService{}
	components := []PluginComponent{
		{Type: "switch", ID: 2, Name: ""},
	}

	m := NewPlugin(context.Background(), svc, "device1", "tasmota", components)
	view := m.View()

	if !strings.Contains(view, "switch:2") {
		t.Fatal("expected fallback name 'switch:2' in view")
	}
}

package views

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// mockView implements View for testing.
type mockView struct {
	id     ViewID
	width  int
	height int
}

func (v mockView) Init() tea.Cmd                    { return nil }
func (v mockView) Update(_ tea.Msg) (View, tea.Cmd) { return v, nil }
func (v mockView) View() string                     { return "mock-" + v.id.String() }
func (v mockView) SetSize(w, h int) View            { v.width = w; v.height = h; return v }
func (v mockView) ID() ViewID                       { return v.id }

func TestNew(t *testing.T) {
	t.Parallel()
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
	if m.Active() != ViewDashboard {
		t.Errorf("Active() = %v, want %v", m.Active(), ViewDashboard)
	}
}

func TestManager_Register(t *testing.T) {
	t.Parallel()
	m := New()
	m.Register(mockView{id: ViewDashboard})
	m.Register(mockView{id: ViewConfig})

	if m.ViewCount() != 2 {
		t.Errorf("ViewCount() = %d, want 2", m.ViewCount())
	}
}

func TestManager_SetActive(t *testing.T) {
	t.Parallel()
	m := New()
	m.Register(mockView{id: ViewDashboard})
	m.Register(mockView{id: ViewConfig})

	cmd := m.SetActive(ViewConfig)
	if cmd == nil {
		t.Fatal("SetActive() should return a command")
	}

	msg := cmd()
	vcMsg, ok := msg.(ViewChangedMsg)
	if !ok {
		t.Fatalf("expected ViewChangedMsg, got %T", msg)
	}
	if vcMsg.Previous != ViewDashboard {
		t.Errorf("Previous = %v, want %v", vcMsg.Previous, ViewDashboard)
	}
	if vcMsg.Current != ViewConfig {
		t.Errorf("Current = %v, want %v", vcMsg.Current, ViewConfig)
	}
}

func TestManager_SetActive_NoChange(t *testing.T) {
	t.Parallel()
	m := New()

	cmd := m.SetActive(ViewDashboard)
	if cmd != nil {
		t.Error("SetActive() should return nil when view doesn't change")
	}
}

func TestManager_Back(t *testing.T) {
	t.Parallel()
	m := New()
	m.Register(mockView{id: ViewDashboard})
	m.Register(mockView{id: ViewConfig})
	m.Register(mockView{id: ViewManage})

	// Navigate: Dashboard -> Config -> Manage
	m.SetActive(ViewConfig)
	m.SetActive(ViewManage)

	if m.Active() != ViewManage {
		t.Errorf("Active() = %v, want %v", m.Active(), ViewManage)
	}

	// Go back
	m.Back()
	if m.Active() != ViewConfig {
		t.Errorf("After Back(), Active() = %v, want %v", m.Active(), ViewConfig)
	}

	// Go back again
	m.Back()
	if m.Active() != ViewDashboard {
		t.Errorf("After second Back(), Active() = %v, want %v", m.Active(), ViewDashboard)
	}
}

func TestManager_Back_NoHistory(t *testing.T) {
	t.Parallel()
	m := New()

	cmd := m.Back()
	if cmd != nil {
		t.Error("Back() should return nil when no history")
	}
}

func TestManager_HasHistory(t *testing.T) {
	t.Parallel()
	m := New()

	if m.HasHistory() {
		t.Error("HasHistory() should be false initially")
	}

	m.SetActive(ViewConfig)

	if !m.HasHistory() {
		t.Error("HasHistory() should be true after navigation")
	}
}

func TestManager_Get(t *testing.T) {
	t.Parallel()
	m := New()
	v := mockView{id: ViewDashboard}
	m.Register(v)

	got := m.Get(ViewDashboard)
	if got == nil {
		t.Error("Get() returned nil for registered view")
	}

	notFound := m.Get(ViewManage)
	if notFound != nil {
		t.Error("Get() should return nil for unregistered view")
	}
}

func TestManager_ActiveView(t *testing.T) {
	t.Parallel()
	m := New()
	v := mockView{id: ViewDashboard}
	m.Register(v)

	av := m.ActiveView()
	if av == nil {
		t.Error("ActiveView() returned nil")
	}
	if av.ID() != ViewDashboard {
		t.Errorf("ActiveView().ID() = %v, want %v", av.ID(), ViewDashboard)
	}
}

func TestManager_View(t *testing.T) {
	t.Parallel()
	m := New()
	m.Register(mockView{id: ViewDashboard})

	view := m.View()
	if view != "mock-Dashboard" {
		t.Errorf("View() = %q, want %q", view, "mock-Dashboard")
	}
}

func TestManager_SetSize(t *testing.T) {
	t.Parallel()
	m := New()
	m.Register(mockView{id: ViewDashboard})

	m.SetSize(100, 50)

	if m.Width() != 100 {
		t.Errorf("Width() = %d, want 100", m.Width())
	}
	if m.Height() != 50 {
		t.Errorf("Height() = %d, want 50", m.Height())
	}
}

// mockSettableView is a mock view that implements DeviceSettable.
type mockSettableView struct {
	mockView
	device string
}

func (v *mockSettableView) SetDevice(device string) tea.Cmd {
	if v.device == device {
		return nil
	}
	v.device = device
	return func() tea.Msg { return struct{}{} }
}

func TestManager_PropagateDevice(t *testing.T) {
	t.Parallel()
	m := New()

	// Register a settable view
	settable := &mockSettableView{mockView: mockView{id: ViewAutomation}}
	m.Register(settable)

	// Register a non-settable view
	m.Register(mockView{id: ViewDashboard})

	cmd := m.PropagateDevice("kitchen-plug")
	if cmd == nil {
		t.Fatal("PropagateDevice() should return a command when settable views exist")
	}

	// Verify the device was set on the settable view
	if settable.device != "kitchen-plug" {
		t.Errorf("device = %q, want %q", settable.device, "kitchen-plug")
	}
}

func TestManager_PropagateDevice_NoSettable(t *testing.T) {
	t.Parallel()
	m := New()
	m.Register(mockView{id: ViewDashboard})

	cmd := m.PropagateDevice("device")
	if cmd != nil {
		t.Error("PropagateDevice() should return nil when no settable views")
	}
}

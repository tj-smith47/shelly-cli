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
	if m.Active() != ViewDevices {
		t.Errorf("Active() = %v, want %v", m.Active(), ViewDevices)
	}
}

func TestManager_Register(t *testing.T) {
	t.Parallel()
	m := New()
	m.Register(mockView{id: ViewDevices})
	m.Register(mockView{id: ViewEvents})

	if m.ViewCount() != 2 {
		t.Errorf("ViewCount() = %d, want 2", m.ViewCount())
	}
}

func TestManager_SetActive(t *testing.T) {
	t.Parallel()
	m := New()
	m.Register(mockView{id: ViewDevices})
	m.Register(mockView{id: ViewEvents})

	cmd := m.SetActive(ViewEvents)
	if cmd == nil {
		t.Fatal("SetActive() should return a command")
	}

	msg := cmd()
	vcMsg, ok := msg.(ViewChangedMsg)
	if !ok {
		t.Fatalf("expected ViewChangedMsg, got %T", msg)
	}
	if vcMsg.Previous != ViewDevices {
		t.Errorf("Previous = %v, want %v", vcMsg.Previous, ViewDevices)
	}
	if vcMsg.Current != ViewEvents {
		t.Errorf("Current = %v, want %v", vcMsg.Current, ViewEvents)
	}
}

func TestManager_SetActive_NoChange(t *testing.T) {
	t.Parallel()
	m := New()

	cmd := m.SetActive(ViewDevices)
	if cmd != nil {
		t.Error("SetActive() should return nil when view doesn't change")
	}
}

func TestManager_Back(t *testing.T) {
	t.Parallel()
	m := New()
	m.Register(mockView{id: ViewDevices})
	m.Register(mockView{id: ViewEvents})
	m.Register(mockView{id: ViewEnergy})

	// Navigate: Devices -> Events -> Energy
	m.SetActive(ViewEvents)
	m.SetActive(ViewEnergy)

	if m.Active() != ViewEnergy {
		t.Errorf("Active() = %v, want %v", m.Active(), ViewEnergy)
	}

	// Go back
	m.Back()
	if m.Active() != ViewEvents {
		t.Errorf("After Back(), Active() = %v, want %v", m.Active(), ViewEvents)
	}

	// Go back again
	m.Back()
	if m.Active() != ViewDevices {
		t.Errorf("After second Back(), Active() = %v, want %v", m.Active(), ViewDevices)
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

	m.SetActive(ViewEvents)

	if !m.HasHistory() {
		t.Error("HasHistory() should be true after navigation")
	}
}

func TestManager_Get(t *testing.T) {
	t.Parallel()
	m := New()
	v := mockView{id: ViewDevices}
	m.Register(v)

	got := m.Get(ViewDevices)
	if got == nil {
		t.Error("Get() returned nil for registered view")
	}

	notFound := m.Get(ViewEnergy)
	if notFound != nil {
		t.Error("Get() should return nil for unregistered view")
	}
}

func TestManager_ActiveView(t *testing.T) {
	t.Parallel()
	m := New()
	v := mockView{id: ViewDevices}
	m.Register(v)

	av := m.ActiveView()
	if av == nil {
		t.Error("ActiveView() returned nil")
	}
	if av.ID() != ViewDevices {
		t.Errorf("ActiveView().ID() = %v, want %v", av.ID(), ViewDevices)
	}
}

func TestManager_View(t *testing.T) {
	t.Parallel()
	m := New()
	m.Register(mockView{id: ViewDevices})

	view := m.View()
	if view != "mock-Devices" {
		t.Errorf("View() = %q, want %q", view, "mock-Devices")
	}
}

func TestManager_SetSize(t *testing.T) {
	t.Parallel()
	m := New()
	m.Register(mockView{id: ViewDevices})

	m.SetSize(100, 50)

	if m.Width() != 100 {
		t.Errorf("Width() = %d, want 100", m.Width())
	}
	if m.Height() != 50 {
		t.Errorf("Height() = %d, want 50", m.Height())
	}
}

package focus

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New(PanelDeviceList, PanelDeviceInfo, PanelEvents)
	if m == nil {
		t.Fatal("New() returned nil")
	}
	if m.Current() != PanelDeviceList {
		t.Errorf("Current() = %v, want %v", m.Current(), PanelDeviceList)
	}
}

func TestNew_Empty(t *testing.T) {
	t.Parallel()
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
	if m.Current() != PanelNone {
		t.Errorf("Current() = %v, want %v", m.Current(), PanelNone)
	}
}

func TestManager_IsFocused(t *testing.T) {
	t.Parallel()
	m := New(PanelDeviceList, PanelDeviceInfo)

	if !m.IsFocused(PanelDeviceList) {
		t.Error("IsFocused(PanelDeviceList) should be true")
	}
	if m.IsFocused(PanelDeviceInfo) {
		t.Error("IsFocused(PanelDeviceInfo) should be false")
	}
}

func TestManager_SetFocus(t *testing.T) {
	t.Parallel()
	m := New(PanelDeviceList, PanelDeviceInfo, PanelEvents)

	cmd := m.SetFocus(PanelDeviceInfo)
	if cmd == nil {
		t.Fatal("SetFocus() should return a command")
	}

	msg := cmd()
	fcMsg, ok := msg.(ChangedMsg)
	if !ok {
		t.Fatalf("expected ChangedMsg, got %T", msg)
	}
	if fcMsg.Previous != PanelDeviceList {
		t.Errorf("Previous = %v, want %v", fcMsg.Previous, PanelDeviceList)
	}
	if fcMsg.Current != PanelDeviceInfo {
		t.Errorf("Current = %v, want %v", fcMsg.Current, PanelDeviceInfo)
	}
}

func TestManager_SetFocus_NoChange(t *testing.T) {
	t.Parallel()
	m := New(PanelDeviceList, PanelDeviceInfo)

	cmd := m.SetFocus(PanelDeviceList)
	if cmd != nil {
		t.Error("SetFocus() should return nil when focus doesn't change")
	}
}

func TestManager_Next(t *testing.T) {
	t.Parallel()
	m := New(PanelDeviceList, PanelDeviceInfo, PanelEvents)

	// First Next: DeviceList -> DeviceInfo
	m.Next()
	if m.Current() != PanelDeviceInfo {
		t.Errorf("After first Next(), Current() = %v, want %v", m.Current(), PanelDeviceInfo)
	}

	// Second Next: DeviceInfo -> Events
	m.Next()
	if m.Current() != PanelEvents {
		t.Errorf("After second Next(), Current() = %v, want %v", m.Current(), PanelEvents)
	}

	// Third Next: Events -> DeviceList (wrap around)
	m.Next()
	if m.Current() != PanelDeviceList {
		t.Errorf("After third Next(), Current() = %v, want %v", m.Current(), PanelDeviceList)
	}
}

func TestManager_Prev(t *testing.T) {
	t.Parallel()
	m := New(PanelDeviceList, PanelDeviceInfo, PanelEvents)

	// First Prev: DeviceList -> Events (wrap around)
	m.Prev()
	if m.Current() != PanelEvents {
		t.Errorf("After first Prev(), Current() = %v, want %v", m.Current(), PanelEvents)
	}

	// Second Prev: Events -> DeviceInfo
	m.Prev()
	if m.Current() != PanelDeviceInfo {
		t.Errorf("After second Prev(), Current() = %v, want %v", m.Current(), PanelDeviceInfo)
	}
}

func TestManager_BorderColor(t *testing.T) {
	t.Parallel()
	m := New(PanelDeviceList, PanelDeviceInfo)

	focusColor := m.BorderColor(PanelDeviceList)
	blurColor := m.BorderColor(PanelDeviceInfo)

	if focusColor == blurColor {
		t.Error("Focus and blur colors should be different")
	}
}

func TestManager_Reset(t *testing.T) {
	t.Parallel()
	m := New(PanelDeviceList, PanelDeviceInfo, PanelEvents)

	// Move to last panel
	m.SetFocus(PanelEvents)
	if m.Current() != PanelEvents {
		t.Errorf("Current() = %v, want %v", m.Current(), PanelEvents)
	}

	// Reset should go back to first
	m.Reset()
	if m.Current() != PanelDeviceList {
		t.Errorf("After Reset(), Current() = %v, want %v", m.Current(), PanelDeviceList)
	}
}

func TestManager_SetPanels(t *testing.T) {
	t.Parallel()
	m := New(PanelDeviceList, PanelDeviceInfo)

	// Add a new panel
	m.SetPanels(PanelDeviceList, PanelDeviceInfo, PanelEvents)

	if m.PanelCount() != 3 {
		t.Errorf("PanelCount() = %d, want 3", m.PanelCount())
	}

	// Current should still be DeviceList
	if m.Current() != PanelDeviceList {
		t.Errorf("Current() = %v, want %v", m.Current(), PanelDeviceList)
	}
}

func TestPanelID_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		panel PanelID
		want  string
	}{
		{PanelNone, "none"},
		{PanelEvents, "events"},
		{PanelDeviceList, "devices"},
		{PanelDeviceInfo, "info"},
		{PanelJSON, "json"},
		{PanelEnergy, "energy"},
		{PanelMonitor, "monitor"},
		{PanelID(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.panel.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

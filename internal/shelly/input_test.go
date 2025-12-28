package shelly

import (
	"testing"
)

func TestInputInfo_Fields(t *testing.T) {
	t.Parallel()

	info := InputInfo{
		ID:    0,
		Name:  "Front Door",
		Type:  "button",
		State: true,
	}

	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if info.Name != "Front Door" {
		t.Errorf("Name = %q, want %q", info.Name, "Front Door")
	}
	if info.Type != "button" {
		t.Errorf("Type = %q, want %q", info.Type, "button")
	}
	if !info.State {
		t.Error("State = false, want true")
	}
}

func TestInputInfo_ZeroValues(t *testing.T) {
	t.Parallel()

	var info InputInfo

	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if info.Name != "" {
		t.Errorf("Name = %q, want empty", info.Name)
	}
	if info.Type != "" {
		t.Errorf("Type = %q, want empty", info.Type)
	}
	if info.State {
		t.Error("State = true, want false")
	}
}

func TestInputInfo_Types(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeVal  string
		expected string
	}{
		{"button type", "button", "button"},
		{"switch type", "switch", "switch"},
		{"analog type", "analog", "analog"},
		{"count type", "count", "count"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			info := InputInfo{Type: tt.typeVal}
			if info.Type != tt.expected {
				t.Errorf("Type = %q, want %q", info.Type, tt.expected)
			}
		})
	}
}

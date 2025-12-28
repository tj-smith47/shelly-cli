package shelly

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestActionConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"ActionOn", ActionOn, "on"},
		{"ActionOff", ActionOff, "off"},
		{"ActionToggle", ActionToggle, "toggle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.constant != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.want)
			}
		})
	}
}

func TestQuickResult_Fields(t *testing.T) {
	t.Parallel()

	result := QuickResult{
		Count:        3,
		PluginResult: nil,
	}

	if result.Count != 3 {
		t.Errorf("Count = %d, want 3", result.Count)
	}
	if result.PluginResult != nil {
		t.Error("PluginResult should be nil")
	}
}

func TestQuickResult_WithPluginResult(t *testing.T) {
	t.Parallel()

	pluginResult := &PluginQuickResult{
		State: "on",
	}
	result := QuickResult{
		Count:        1,
		PluginResult: pluginResult,
	}

	if result.Count != 1 {
		t.Errorf("Count = %d, want 1", result.Count)
	}
	if result.PluginResult == nil {
		t.Error("PluginResult should not be nil")
	}
	if result.PluginResult.State != "on" {
		t.Errorf("PluginResult.State = %q, want %q", result.PluginResult.State, "on")
	}
}

func TestIsControllable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		t    model.ComponentType
		want bool
	}{
		{"switch is controllable", model.ComponentSwitch, true},
		{"light is controllable", model.ComponentLight, true},
		{"rgb is controllable", model.ComponentRGB, true},
		{"cover is controllable", model.ComponentCover, true},
		{"input is not controllable", model.ComponentInput, false},
		{"em is not controllable", model.ComponentType("em"), false},
		{"sys is not controllable", model.ComponentType("sys"), false},
		{"unknown is not controllable", model.ComponentType("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isControllable(tt.t); got != tt.want {
				t.Errorf("isControllable(%q) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}

func TestSelectComponents(t *testing.T) {
	t.Parallel()

	controllable := []model.Component{
		{Type: model.ComponentSwitch, ID: 0},
		{Type: model.ComponentSwitch, ID: 1},
		{Type: model.ComponentLight, ID: 0},
	}

	t.Run("nil componentID returns all", func(t *testing.T) {
		t.Parallel()
		result := selectComponents(controllable, nil)
		if len(result) != 3 {
			t.Errorf("len(result) = %d, want 3", len(result))
		}
	})

	t.Run("specific componentID returns matching", func(t *testing.T) {
		t.Parallel()
		id := 1
		result := selectComponents(controllable, &id)
		if len(result) != 1 {
			t.Errorf("len(result) = %d, want 1", len(result))
		}
		if result[0].ID != 1 {
			t.Errorf("result[0].ID = %d, want 1", result[0].ID)
		}
	})

	t.Run("non-existent componentID returns nil", func(t *testing.T) {
		t.Parallel()
		id := 99
		result := selectComponents(controllable, &id)
		if result != nil {
			t.Errorf("result = %v, want nil", result)
		}
	})

	t.Run("empty controllable returns nil", func(t *testing.T) {
		t.Parallel()
		result := selectComponents([]model.Component{}, nil)
		if len(result) != 0 {
			t.Errorf("len(result) = %d, want 0", len(result))
		}
	})
}

func TestComponentControlResult_Fields(t *testing.T) {
	t.Parallel()

	result := ComponentControlResult{
		Type:    model.ComponentSwitch,
		ID:      0,
		Success: true,
		Err:     nil,
	}

	if result.Type != model.ComponentSwitch {
		t.Errorf("Type = %q, want %q", result.Type, model.ComponentSwitch)
	}
	if result.ID != 0 {
		t.Errorf("ID = %d, want 0", result.ID)
	}
	if !result.Success {
		t.Error("Success = false, want true")
	}
	if result.Err != nil {
		t.Errorf("Err = %v, want nil", result.Err)
	}
}

func TestPartyColors(t *testing.T) {
	t.Parallel()

	// Verify we have expected colors
	if len(PartyColors) < 5 {
		t.Errorf("len(PartyColors) = %d, want at least 5", len(PartyColors))
	}

	// Verify all colors are valid RGB values
	for i, color := range PartyColors {
		if color.R < 0 || color.R > 255 {
			t.Errorf("PartyColors[%d].R = %d, want 0-255", i, color.R)
		}
		if color.G < 0 || color.G > 255 {
			t.Errorf("PartyColors[%d].G = %d, want 0-255", i, color.G)
		}
		if color.B < 0 || color.B > 255 {
			t.Errorf("PartyColors[%d].B = %d, want 0-255", i, color.B)
		}
	}

	// Verify red is first
	if PartyColors[0].R != 255 || PartyColors[0].G != 0 || PartyColors[0].B != 0 {
		t.Errorf("PartyColors[0] = %+v, want red (255,0,0)", PartyColors[0])
	}
}

func TestIOStreamsDebugger_Interface(t *testing.T) {
	t.Parallel()

	// Verify the interface is properly defined
	var _ IOStreamsDebugger = (*mockDebugger)(nil)
}

// mockDebugger is a mock for IOStreamsDebugger interface.
type mockDebugger struct {
	called bool
}

func (m *mockDebugger) DebugErr(_ string, _ error) {
	m.called = true
}

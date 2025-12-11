package config

import (
	"testing"
)

func TestValidateSceneName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "movie-night", false},
		{"valid with underscore", "wake_up", false},
		{"valid alphanumeric", "scene123", false},
		{"empty name", "", true},
		{"too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true},
		{"starts with hyphen", "-invalid", true},
		{"contains space", "movie night", true},
		{"contains special char", "movie@night", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateSceneName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSceneName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestSceneStruct(t *testing.T) {
	t.Parallel()

	scene := Scene{
		Name:        "movie-night",
		Description: "Dim lights for movie watching",
		Actions: []SceneAction{
			{Device: "living-room-light", Method: "Light.Set", Params: map[string]any{"brightness": 20}},
			{Device: "tv-backlight", Method: "Switch.Set", Params: map[string]any{"on": true}},
		},
	}

	if scene.Name != "movie-night" {
		t.Errorf("Name = %q, want movie-night", scene.Name)
	}
	if len(scene.Actions) != 2 {
		t.Errorf("Actions length = %d, want 2", len(scene.Actions))
	}
	if scene.Actions[0].Device != "living-room-light" {
		t.Errorf("Action[0].Device = %q, want living-room-light", scene.Actions[0].Device)
	}
	if scene.Actions[0].Method != "Light.Set" {
		t.Errorf("Action[0].Method = %q, want Light.Set", scene.Actions[0].Method)
	}
}

func TestSceneActionStruct(t *testing.T) {
	t.Parallel()

	action := SceneAction{
		Device: "kitchen-switch",
		Method: "Switch.Toggle",
		Params: map[string]any{"id": 0},
	}

	if action.Device != "kitchen-switch" {
		t.Errorf("Device = %q, want kitchen-switch", action.Device)
	}
	if action.Method != "Switch.Toggle" {
		t.Errorf("Method = %q, want Switch.Toggle", action.Method)
	}
	if action.Params["id"] != 0 {
		t.Errorf("Params[id] = %v, want 0", action.Params["id"])
	}
}

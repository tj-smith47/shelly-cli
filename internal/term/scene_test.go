package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

func TestDisplaySceneDetails(t *testing.T) {
	t.Parallel()

	t.Run("with actions", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		scene := config.Scene{
			Name:        "Morning Routine",
			Description: "Turn on lights and open blinds",
			Actions: []config.SceneAction{
				{Device: "kitchen", Method: "Switch.Set", Params: map[string]any{"id": 0, "on": true}},
				{Device: "blinds", Method: "Cover.Open", Params: nil},
			},
		}

		DisplaySceneDetails(ios, scene)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "Morning Routine") {
			t.Error("output should contain scene name")
		}
		if !strings.Contains(allOutput, "Turn on lights") {
			t.Error("output should contain description")
		}
		if !strings.Contains(allOutput, "kitchen") {
			t.Error("output should contain device name")
		}
		if !strings.Contains(allOutput, "Switch.Set") {
			t.Error("output should contain method name")
		}
		if !strings.Contains(allOutput, "Cover.Open") {
			t.Error("output should contain second method name")
		}
	})

	t.Run("no actions", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		scene := config.Scene{
			Name:    "Empty Scene",
			Actions: []config.SceneAction{},
		}

		DisplaySceneDetails(ios, scene)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "Empty Scene") {
			t.Error("output should contain scene name")
		}
		if !strings.Contains(allOutput, "No actions defined") {
			t.Errorf("output should contain 'No actions defined', got %q", allOutput)
		}
		if !strings.Contains(allOutput, "shelly scene add-action") {
			t.Error("output should contain add-action hint")
		}
	})

	t.Run("without description", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		scene := config.Scene{
			Name:        "Simple Scene",
			Description: "",
			Actions: []config.SceneAction{
				{Device: "light", Method: "Light.Set", Params: map[string]any{"brightness": 50}},
			},
		}

		DisplaySceneDetails(ios, scene)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "Simple Scene") {
			t.Error("output should contain scene name")
		}
		if !strings.Contains(allOutput, "Actions (1):") {
			t.Error("output should contain action count")
		}
	})

	t.Run("action without params", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		scene := config.Scene{
			Name: "Toggle Scene",
			Actions: []config.SceneAction{
				{Device: "switch1", Method: "Switch.Toggle", Params: nil},
			},
		}

		DisplaySceneDetails(ios, scene)

		output := out.String()
		if !strings.Contains(output, "Switch.Toggle") {
			t.Error("output should contain method name")
		}
	})

	t.Run("multiple actions with various params", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		scene := config.Scene{
			Name: "Complex Scene",
			Actions: []config.SceneAction{
				{Device: "dev1", Method: "Method1", Params: map[string]any{"key1": "value1"}},
				{Device: "dev2", Method: "Method2", Params: map[string]any{"key2": 42}},
				{Device: "dev3", Method: "Method3", Params: map[string]any{}},
			},
		}

		DisplaySceneDetails(ios, scene)

		output := out.String()
		if !strings.Contains(output, "dev1") {
			t.Error("output should contain first device")
		}
		if !strings.Contains(output, "dev2") {
			t.Error("output should contain second device")
		}
		if !strings.Contains(output, "dev3") {
			t.Error("output should contain third device")
		}
	})
}

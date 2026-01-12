package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayConfigTable(t *testing.T) {
	t.Parallel()

	t.Run("map config", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		configData := map[string]any{
			"switch:0": map[string]any{
				"name": "Living Room Switch",
			},
		}

		err := DisplayConfigTable(ios, configData)
		if err != nil {
			t.Errorf("DisplayConfigTable returned error: %v", err)
		}

		output := out.String()
		if !strings.Contains(output, "switch:0") {
			t.Error("output should contain 'switch:0'")
		}
	})

	t.Run("non-map value", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		configData := map[string]any{
			"component": []string{"value1", "value2"},
		}

		err := DisplayConfigTable(ios, configData)
		if err != nil {
			t.Errorf("DisplayConfigTable returned error: %v", err)
		}

		output := out.String()
		if output == "" {
			t.Error("DisplayConfigTable should produce output")
		}
	})

	t.Run("non-map config data", func(t *testing.T) {
		t.Parallel()
		ios, _, _ := testIOStreams()
		configData := "not a map"

		_ = DisplayConfigTable(ios, configData) //nolint:errcheck // testing non-panic behavior
		// Should not panic, may print JSON
	})
}

func TestDisplaySceneList(t *testing.T) {
	t.Parallel()

	t.Run("with scenes", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		scenes := []config.Scene{
			{Name: "movie-time", Description: "Dim lights for movie", Actions: []config.SceneAction{{Device: "light1", Method: "Switch.Off"}}},
			{Name: "all-off", Description: "", Actions: []config.SceneAction{{Device: "switch1", Method: "Switch.Off"}}},
		}

		DisplaySceneList(ios, scenes)

		output := out.String()
		if !strings.Contains(output, "movie-time") {
			t.Error("output should contain 'movie-time'")
		}
		if !strings.Contains(output, "all-off") {
			t.Error("output should contain 'all-off'")
		}
	})

	t.Run("empty scenes", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()

		DisplaySceneList(ios, []config.Scene{})

		output := out.String()
		if output == "" {
			t.Error("DisplaySceneList should produce output even for empty list")
		}
	})
}

func TestDisplayAliasList(t *testing.T) {
	t.Parallel()

	t.Run("with aliases", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		aliases := []config.NamedAlias{
			{Name: "ls", Alias: config.Alias{Command: "device list", Shell: false}},
			{Name: "reboot-all", Alias: config.Alias{Command: "for d in $(shelly device list -q); do shelly device reboot $d; done", Shell: true}},
		}

		DisplayAliasList(ios, aliases)

		output := out.String()
		if !strings.Contains(output, "ls") {
			t.Error("output should contain 'ls'")
		}
		if !strings.Contains(output, "shell") {
			t.Error("output should contain 'shell' for shell aliases")
		}
	})

	t.Run("empty aliases", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()

		DisplayAliasList(ios, []config.NamedAlias{})

		output := out.String()
		if output == "" {
			t.Error("DisplayAliasList should produce output even for empty list")
		}
	})
}

func TestDisplayResetableComponents(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	configKeys := []string{"switch:0", "light:0", "input:0"}

	DisplayResetableComponents(ios, "device1", configKeys)

	output := out.String()
	if !strings.Contains(output, "Available components") {
		t.Error("output should contain 'Available components'")
	}
	if !strings.Contains(output, "switch:0") {
		t.Error("output should contain 'switch:0'")
	}
	if !strings.Contains(output, "device1") {
		t.Error("output should contain device name")
	}
}

func TestDisplayTemplateDiffs(t *testing.T) {
	t.Parallel()

	t.Run("with differences", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		diffs := []model.ConfigDiff{
			{Path: "switch:0.name", DiffType: "changed", OldValue: "Switch", NewValue: "Living Room"},
			{Path: "light:0.brightness", DiffType: "added", OldValue: nil, NewValue: 80},
		}

		DisplayTemplateDiffs(ios, "my-template", "device1", diffs)

		output := out.String()
		if !strings.Contains(output, "Configuration Differences") {
			t.Error("output should contain 'Configuration Differences'")
		}
		if !strings.Contains(output, "my-template") {
			t.Error("output should contain template name")
		}
		if !strings.Contains(output, "device1") {
			t.Error("output should contain device name")
		}
		if !strings.Contains(output, "2 difference(s)") {
			t.Error("output should contain difference count")
		}
	})

	t.Run("no differences", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()

		DisplayTemplateDiffs(ios, "my-template", "device1", []model.ConfigDiff{})

		output := out.String()
		if !strings.Contains(output, "No differences") {
			t.Error("output should contain 'No differences'")
		}
	})
}

func TestDisplayDeviceTemplateList(t *testing.T) {
	t.Parallel()

	t.Run("with templates", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		templates := []config.DeviceTemplate{
			{Name: "pro1pm-switch", Model: "SNSW-001P16EU", Generation: 2, SourceDevice: "device1", CreatedAt: "2024-01-15T10:30:00Z"},
			{Name: "dimmer-default", Model: "SNDM-0013US", Generation: 2, SourceDevice: "", CreatedAt: "2024-01-20"},
		}

		DisplayDeviceTemplateList(ios, templates)

		output := out.String()
		if !strings.Contains(output, "Configuration Templates") {
			t.Error("output should contain 'Configuration Templates'")
		}
		if !strings.Contains(output, "pro1pm-switch") {
			t.Error("output should contain 'pro1pm-switch'")
		}
		if !strings.Contains(output, "2 template(s)") {
			t.Error("output should contain template count")
		}
	})

	t.Run("template without source device", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		templates := []config.DeviceTemplate{
			{Name: "manual-template", Model: "SNSW-001X16EU", Generation: 2, SourceDevice: ""},
		}

		DisplayDeviceTemplateList(ios, templates)

		output := out.String()
		if !strings.Contains(output, "-") {
			t.Error("output should contain '-' for empty source device")
		}
	})

	t.Run("empty templates", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()

		DisplayDeviceTemplateList(ios, []config.DeviceTemplate{})

		output := out.String()
		if !strings.Contains(output, "0 template(s)") {
			t.Error("output should contain '0 template(s)'")
		}
	})
}

package term

import (
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
)

// DisplayConfigTable prints a configuration map as formatted tables.
// Each top-level key becomes a titled section with a settings table.
func DisplayConfigTable(ios *iostreams.IOStreams, configData any) error {
	configMap, ok := configData.(map[string]any)
	if !ok {
		return output.PrintJSON(configData)
	}

	for component, cfg := range configMap {
		ios.Title("%s", component)

		tbl := output.FormatConfigTable(cfg)
		if tbl == nil {
			// If it's not a map, print as JSON
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				ios.DebugErr("marshaling config component", err)
			} else {
				ios.Printf("%s\n", data)
			}
			ios.Printf("\n")
			continue
		}

		if ios.IsPlainMode() {
			tbl.SetStyle(table.PlainStyle())
		}
		if err := tbl.PrintTo(ios.Out); err != nil {
			ios.DebugErr("print config table", err)
		}
		ios.Printf("\n")
	}

	return nil
}

// DisplaySceneList prints a table of scenes.
func DisplaySceneList(ios *iostreams.IOStreams, scenes []config.Scene) {
	builder := table.NewBuilder("Name", "Actions", "Description")
	for _, scene := range scenes {
		actions := output.FormatActionCount(len(scene.Actions))
		description := scene.Description
		if description == "" {
			description = "-"
		}
		builder.AddRow(scene.Name, actions, description)
	}

	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print scenes table", err)
	}
	ios.Println()
	ios.Count("scene", len(scenes))
}

// DisplayAliasList prints a table of aliases.
func DisplayAliasList(ios *iostreams.IOStreams, aliases []config.Alias) {
	builder := table.NewBuilder("Name", "Command", "Type")

	for _, alias := range aliases {
		aliasType := "command"
		if alias.Shell {
			aliasType = "shell"
		}
		builder.AddRow(alias.Name, alias.Command, aliasType)
	}

	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print aliases table", err)
	}
	ios.Println()
	ios.Count("alias", len(aliases))
}

// DisplayResetableComponents lists available components that can be reset.
func DisplayResetableComponents(ios *iostreams.IOStreams, device string, configKeys []string) {
	ios.Title("Available components")
	ios.Printf("Specify a component to reset its configuration:\n")
	ios.Printf("\n")

	for _, key := range configKeys {
		ios.Printf("  shelly config reset %s %s\n", device, key)
	}
}

// DisplayTemplateDiffs prints a table of template comparison diffs.
func DisplayTemplateDiffs(ios *iostreams.IOStreams, templateName, deviceName string, diffs []model.ConfigDiff) {
	if len(diffs) == 0 {
		ios.Info("No differences - device matches template")
		return
	}

	ios.Title("Configuration Differences")
	ios.Printf("Template: %s  Device: %s\n\n", templateName, deviceName)

	builder := table.NewBuilder("Path", "Type", "Device Value", "Template Value")
	for _, d := range diffs {
		builder.AddRow(d.Path, d.DiffType, output.FormatDisplayValue(d.OldValue), output.FormatDisplayValue(d.NewValue))
	}
	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print template diffs table", err)
	}
	ios.Printf("\n%d difference(s) found\n", len(diffs))
}

// DisplayDeviceTemplateList prints a table of device configuration templates.
func DisplayDeviceTemplateList(ios *iostreams.IOStreams, templates []config.DeviceTemplate) {
	ios.Title("Configuration Templates")
	ios.Println()

	builder := table.NewBuilder("Name", "Model", "Gen", "Source", "Created")
	for _, t := range templates {
		source := t.SourceDevice
		if source == "" {
			source = "-"
		}
		created := t.CreatedAt
		if len(created) > 10 {
			created = created[:10] // Just the date part
		}
		builder.AddRow(t.Name, t.Model, fmt.Sprintf("Gen%d", t.Generation), source, created)
	}

	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print templates table", err)
	}
	ios.Printf("\n%d template(s)\n", len(templates))
}

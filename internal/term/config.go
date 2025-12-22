package term

import (
	"encoding/json"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
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

		table := output.FormatConfigTable(cfg)
		if table == nil {
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
			table.SetStyle(output.PlainTableStyle())
		}
		if err := table.PrintTo(ios.Out); err != nil {
			ios.DebugErr("print config table", err)
		}
		ios.Printf("\n")
	}

	return nil
}

// DisplaySceneList prints a table of scenes.
func DisplaySceneList(ios *iostreams.IOStreams, scenes []config.Scene) {
	table := output.NewTable("Name", "Actions", "Description")
	for _, scene := range scenes {
		actions := output.FormatActionCount(len(scene.Actions))
		description := scene.Description
		if description == "" {
			description = "-"
		}
		table.AddRow(scene.Name, actions, description)
	}

	printTable(ios, table)
	ios.Println()
	ios.Count("scene", len(scenes))
}

// DisplayAliasList prints a table of aliases.
func DisplayAliasList(ios *iostreams.IOStreams, aliases []config.Alias) {
	table := output.NewTable("Name", "Command", "Type")

	for _, alias := range aliases {
		aliasType := "command"
		if alias.Shell {
			aliasType = "shell"
		}
		table.AddRow(alias.Name, alias.Command, aliasType)
	}

	printTable(ios, table)
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

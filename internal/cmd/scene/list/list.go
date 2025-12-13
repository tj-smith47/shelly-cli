// Package list provides the scene list subcommand.
package list

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the scene list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List saved scenes",
		Long:    `List all saved scenes with their action counts and descriptions.`,
		Example: `  # List all scenes
  shelly scene list

  # Output as JSON
  shelly scene list -o json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")

	return cmd
}

func run(cmd *cobra.Command) error {
	scenes := config.ListScenes()

	if len(scenes) == 0 {
		iostreams.NoResults("scenes")
		return nil
	}

	return outputList(cmd, scenes)
}

func outputList(cmd *cobra.Command, scenes map[string]config.Scene) error {
	// Convert to slice for consistent ordering
	sceneList := make([]config.Scene, 0, len(scenes))
	for _, scene := range scenes {
		sceneList = append(sceneList, scene)
	}
	sort.Slice(sceneList, func(i, j int) bool {
		return sceneList[i].Name < sceneList[j].Name
	})

	switch viper.GetString("output") {
	case string(output.FormatJSON):
		return output.JSON(cmd.OutOrStdout(), sceneList)
	case string(output.FormatYAML):
		return output.YAML(cmd.OutOrStdout(), sceneList)
	default:
		printTable(sceneList)
		return nil
	}
}

func printTable(scenes []config.Scene) {
	t := output.NewTable("Name", "Actions", "Description")
	for _, scene := range scenes {
		actions := formatActionCount(len(scene.Actions))
		description := scene.Description
		if description == "" {
			description = "-"
		}
		t.AddRow(scene.Name, actions, description)
	}
	t.Print()

	iostreams.Count("scene", len(scenes))
}

func formatActionCount(count int) string {
	if count == 0 {
		return theme.StatusWarn().Render("0 (empty)")
	}
	if count == 1 {
		return theme.StatusOK().Render("1 action")
	}
	return theme.StatusOK().Render(fmt.Sprintf("%d actions", count))
}

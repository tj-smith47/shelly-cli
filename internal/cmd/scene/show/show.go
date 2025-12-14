// Package show provides the scene show subcommand.
package show

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the scene show command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "show <name>",
		Aliases: []string{"info", "get"},
		Short:   "Show scene details",
		Long:    `Display detailed information about a scene including all its actions.`,
		Example: `  # Show scene details
  shelly scene show movie-night

  # Output as JSON
  shelly scene show movie-night --output json

  # Using alias
  shelly scene info bedtime

  # Short form
  shelly sc show morning-routine`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.SceneNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args[0], outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")

	return cmd
}

func run(name, outputFormat string) error {
	scene, exists := config.GetScene(name)
	if !exists {
		return fmt.Errorf("scene %q not found", name)
	}

	switch outputFormat {
	case "json":
		return output.PrintJSON(scene)
	case "yaml":
		return output.PrintYAML(scene)
	default:
		return printDetails(scene)
	}
}

func printDetails(scene config.Scene) error {
	// Header
	iostreams.Info("Scene: %s", theme.Bold().Render(scene.Name))
	if scene.Description != "" {
		iostreams.Info("Description: %s", scene.Description)
	}

	if len(scene.Actions) == 0 {
		iostreams.Info("")
		iostreams.Info("%s", theme.Dim().Render("No actions defined"))
		iostreams.Info("Add actions with: shelly scene add-action %s <device> <method> [params]", scene.Name)
		return nil
	}

	iostreams.Info("")
	iostreams.Info("Actions (%d):", len(scene.Actions))

	table := output.NewTable("#", "Device", "Method", "Parameters")

	for i, action := range scene.Actions {
		params := "-"
		if len(action.Params) > 0 {
			params = formatParams(action.Params)
		}
		table.AddRow(
			theme.Dim().Render(fmt.Sprintf("%d", i+1)),
			theme.Bold().Render(action.Device),
			theme.Highlight().Render(action.Method),
			params,
		)
	}

	table.Print()
	return nil
}

func formatParams(params map[string]any) string {
	if len(params) == 0 {
		return "-"
	}
	// Simple formatting for display
	result := ""
	for k, v := range params {
		if result != "" {
			result += ", "
		}
		result += fmt.Sprintf("%s=%v", k, v)
	}
	return result
}

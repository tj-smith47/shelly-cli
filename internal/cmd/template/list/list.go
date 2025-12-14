// Package list provides the template list subcommand.
package list

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the template list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List configuration templates",
		Long: `List all saved configuration templates.

Templates are snapshots of device configuration that can be applied to
other devices of the same model. They're useful for provisioning multiple
devices with identical settings or backing up specific configurations.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: Name, Model, Gen, Source (device), Created`,
		Example: `  # List all templates
  shelly template list

  # Output as JSON
  shelly template list -o json

  # Find templates for a specific model
  shelly template list -o json | jq '.[] | select(.model | contains("Plus"))'

  # Get template names only
  shelly template list -o json | jq -r '.[].name'

  # Export templates list to backup file
  shelly template list -o yaml > templates-list.yaml

  # Short form
  shelly template ls`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(f)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory) error {
	ios := f.IOStreams()
	templates := config.ListTemplates()

	if len(templates) == 0 {
		ios.NoResults("No templates configured")
		return nil
	}

	// Handle different output formats
	switch viper.GetString("output") {
	case "json", "yaml":
		return cmdutil.PrintListResult(ios, templateList(templates), nil)
	default:
		displayTemplates(ios, templates)
		return nil
	}
}

// templateInfo is used for JSON/YAML output.
type templateInfo struct {
	Name         string `json:"name" yaml:"name"`
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	Model        string `json:"model" yaml:"model"`
	Generation   int    `json:"generation" yaml:"generation"`
	CreatedAt    string `json:"created_at" yaml:"created_at"`
	SourceDevice string `json:"source_device,omitempty" yaml:"source_device,omitempty"`
}

func templateList(templates map[string]config.Template) []templateInfo {
	result := make([]templateInfo, 0, len(templates))
	for _, t := range templates {
		result = append(result, templateInfo{
			Name:         t.Name,
			Description:  t.Description,
			Model:        t.Model,
			Generation:   t.Generation,
			CreatedAt:    t.CreatedAt,
			SourceDevice: t.SourceDevice,
		})
	}
	return result
}

func displayTemplates(ios *iostreams.IOStreams, templates map[string]config.Template) {
	ios.Title("Configuration Templates")
	ios.Println()

	table := output.NewTable("Name", "Model", "Gen", "Source", "Created")
	for _, t := range templates {
		source := t.SourceDevice
		if source == "" {
			source = "-"
		}
		created := t.CreatedAt
		if len(created) > 10 {
			created = created[:10] // Just the date part
		}
		table.AddRow(t.Name, t.Model, genStr(t.Generation), source, created)
	}
	table.Print()

	ios.Printf("\n%d template(s)\n", len(templates))
}

func genStr(gen int) string {
	if gen == 1 {
		return "Gen1"
	}
	return "Gen2+"
}

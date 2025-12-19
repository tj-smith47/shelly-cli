// Package list provides the template list subcommand.
package list

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// templateInfo is used for JSON/YAML output.
type templateInfo struct {
	Name         string `json:"name" yaml:"name"`
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	Model        string `json:"model" yaml:"model"`
	Generation   int    `json:"generation" yaml:"generation"`
	CreatedAt    string `json:"created_at" yaml:"created_at"`
	SourceDevice string `json:"source_device,omitempty" yaml:"source_device,omitempty"`
}

// NewCommand creates the template list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewConfigListCommand(f, factories.ConfigListOpts[templateInfo]{
		Resource: "template",
		FetchFunc: func() []templateInfo {
			templates := config.ListTemplates()
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
			sort.Slice(result, func(i, j int) bool {
				return result[i].Name < result[j].Name
			})
			return result
		},
		DisplayFunc: displayTemplates,
		EmptyMsg:    "No templates configured",
	})
}

func displayTemplates(ios *iostreams.IOStreams, templates []templateInfo) {
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
		table.AddRow(t.Name, t.Model, fmt.Sprintf("Gen%d", t.Generation), source, created)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	ios.Printf("\n%d template(s)\n", len(templates))
}

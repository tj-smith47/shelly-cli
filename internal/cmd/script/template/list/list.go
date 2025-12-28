// Package list provides the script template list subcommand.
package list

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Factory *cmdutil.Factory
}

// NewCommand creates the script template list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List available script templates",
		Long: `List all available script templates.

Shows both built-in templates (bundled with the CLI) and user-defined
templates from your configuration.`,
		Example: `  # List all templates
  shelly script template list

  # Output as JSON
  shelly script template list -o json`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	flags.AddOutputFlags(cmd, &opts.OutputFlags)

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Get all templates (built-in + user-defined)
	templates := automation.ListAllScriptTemplates()

	if len(templates) == 0 {
		ios.NoResults("script templates")
		return nil
	}

	// Convert to slice and sort
	list := make([]config.ScriptTemplate, 0, len(templates))
	for _, tpl := range templates {
		list = append(list, tpl)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	// Handle output formats
	if output.WantsStructured() {
		return cmdutil.PrintListResult(ios, list, nil)
	}

	term.DisplayScriptTemplateList(ios, list)
	return nil
}

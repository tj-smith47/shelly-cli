// Package importcmd provides the template import subcommand.
package importcmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds command options.
type Options struct {
	File    string
	Name    string
	Force   bool
	Factory *cmdutil.Factory
}

// NewCommand creates the template import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "import <file> [name]",
		Aliases: []string{"load"},
		Short:   "Import a template from a file",
		Long: `Import a configuration template from a JSON or YAML file.

If no name is specified, the template name from the file is used.
Use --force to overwrite an existing template with the same name.`,
		Example: `  # Import a template
  shelly template import template.yaml

  # Import with a different name
  shelly template import template.yaml my-new-config

  # Overwrite existing template
  shelly template import template.yaml --force`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completion.FileThenNoComplete(),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.File = args[0]
			if len(args) > 1 {
				opts.Name = args[1]
			}
			return run(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing template")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Read file
	data, err := os.ReadFile(opts.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse template
	tpl, err := config.ParseTemplateFile(opts.File, data)
	if err != nil {
		return err
	}

	// Override name if specified
	if opts.Name != "" {
		tpl.Name = opts.Name
	}

	// Validate name
	if err := config.ValidateTemplateName(tpl.Name); err != nil {
		return err
	}

	// Check if exists
	if _, exists := config.GetTemplate(tpl.Name); exists && !opts.Force {
		return fmt.Errorf("template %q already exists (use --force to overwrite)", tpl.Name)
	}

	// Save template
	if err := config.SaveTemplate(tpl); err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	ios.Success("Template %q imported from %s", tpl.Name, opts.File)
	return nil
}

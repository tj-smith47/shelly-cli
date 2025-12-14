// Package export provides the template export subcommand.
package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds command options.
type Options struct {
	Template string
	File     string
	Format   string
	Factory  *cmdutil.Factory
}

// NewCommand creates the template export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "export <template> [file]",
		Aliases: []string{"save", "dump"},
		Short:   "Export a template to a file",
		Long: `Export a configuration template to a JSON or YAML file.

If no file is specified, outputs to stdout.
Format is auto-detected from file extension, or can be specified with --format.`,
		Example: `  # Export to YAML file
  shelly template export my-config template.yaml

  # Export to JSON file
  shelly template export my-config template.json

  # Export to stdout as JSON
  shelly template export my-config --format json

  # Export to stdout as YAML
  shelly template export my-config`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completion.TemplateThenFile(),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.Template = args[0]
			if len(args) > 1 {
				opts.File = args[1]
			}
			return run(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Format, "format", "yaml", "Output format (json, yaml)")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Get template
	tpl, exists := config.GetTemplate(opts.Template)
	if !exists {
		return fmt.Errorf("template %q not found", opts.Template)
	}

	// Determine format
	format := opts.Format
	if opts.File != "" {
		switch strings.ToLower(filepath.Ext(opts.File)) {
		case ".json":
			format = "json"
		case ".yaml", ".yml":
			format = "yaml"
		}
	}

	// Serialize template
	var data []byte
	var err error
	switch format {
	case "json":
		data, err = json.MarshalIndent(tpl, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(tpl)
	default:
		return fmt.Errorf("unsupported format: %s (use json or yaml)", format)
	}
	if err != nil {
		return fmt.Errorf("failed to serialize template: %w", err)
	}

	// Output
	if opts.File == "" {
		// Output to stdout
		ios.Printf("%s\n", string(data))
		return nil
	}

	// Write to file
	if err := os.WriteFile(opts.File, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	ios.Success("Template %q exported to %s", opts.Template, opts.File)
	return nil
}

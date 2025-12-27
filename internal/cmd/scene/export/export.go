// Package export provides the scene export subcommand.
package export

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the scene export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "export <name> [file]",
		Aliases: []string{"save", "backup"},
		Short:   "Export a scene to file",
		Long: `Export a scene definition to a file.

If no file is specified, outputs to stdout.
Format is auto-detected from file extension (.json, .yaml, .yml).`,
		Example: `  # Export to YAML file
  shelly scene export movie-night scene.yaml

  # Export to JSON file
  shelly scene export movie-night scene.json

  # Export to stdout as YAML
  shelly scene export movie-night

  # Export to stdout as JSON
  shelly scene export movie-night --output json`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completion.SceneNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			file := ""
			if len(args) > 1 {
				file = args[1]
			}
			return run(f, name, file, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "yaml", "Output format: json, yaml")

	return cmd
}

func run(f *cmdutil.Factory, name, file, format string) error {
	ios := f.IOStreams()

	scene, exists := config.GetScene(name)
	if !exists {
		return fmt.Errorf("scene %q not found", name)
	}

	// Use config.Scene directly for export (has JSON/YAML tags)
	export := scene

	// Determine format from file extension if file specified
	if file != "" {
		ext := strings.ToLower(filepath.Ext(file))
		switch ext {
		case ".json":
			format = "json"
		case ".yaml", ".yml":
			format = "yaml"
		}
	}

	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(export, "", "  ")
	default:
		data, err = yaml.Marshal(export)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal scene: %w", err)
	}

	// Write to file or stdout
	if file == "" {
		ios.Printf("%s\n", data)
		return nil
	}

	if err := output.WriteFile(file, data); err != nil {
		return err
	}

	ios.Success("Exported scene %q to %s", name, file)
	return nil
}

// Package export provides the scene export subcommand.
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
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// NewCommand creates the scene export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "export <name> [file]",
		Short: "Export a scene to file",
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
		Args: cobra.RangeArgs(1, 2),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			file := ""
			if len(args) > 1 {
				file = args[1]
			}
			return run(name, file, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "yaml", "Output format: json, yaml")

	return cmd
}

// Scene represents a scene for export (same as config.Scene but explicit).
type Scene struct {
	Name        string               `json:"name" yaml:"name"`
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Actions     []config.SceneAction `json:"actions" yaml:"actions"`
}

func run(name, file, format string) error {
	scene, exists := config.GetScene(name)
	if !exists {
		return fmt.Errorf("scene %q not found", name)
	}

	export := Scene{
		Name:        scene.Name,
		Description: scene.Description,
		Actions:     scene.Actions,
	}

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
		fmt.Println(string(data))
		return nil
	}

	if err := writeToFile(file, data); err != nil {
		return err
	}

	iostreams.Success("Exported scene %q to %s", name, file)
	return nil
}

func writeToFile(file string, data []byte) error {
	// Ensure parent directory exists
	dir := filepath.Dir(file)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(file, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

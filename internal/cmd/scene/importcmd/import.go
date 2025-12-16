// Package importcmd provides the scene import subcommand.
package importcmd

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
)

// NewCommand creates the scene import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		name      string
		overwrite bool
	)

	cmd := &cobra.Command{
		Use:     "import <file>",
		Aliases: []string{"load", "restore"},
		Short:   "Import a scene from file",
		Long: `Import a scene definition from a file.

Format is auto-detected from file extension (.json, .yaml, .yml).
Use --name to override the scene name from the file.`,
		Example: `  # Import from YAML file
  shelly scene import scene.yaml

  # Import from JSON file
  shelly scene import scene.json

  # Import with different name
  shelly scene import scene.yaml --name my-scene

  # Overwrite existing scene
  shelly scene import scene.yaml --overwrite`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(f, args[0], name, overwrite)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Override scene name from file")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing scene")

	return cmd
}

// Scene represents an imported scene definition.
type Scene struct {
	Name        string               `json:"name" yaml:"name"`
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Actions     []config.SceneAction `json:"actions" yaml:"actions"`
}

func run(f *cmdutil.Factory, file, nameOverride string, overwrite bool) error {
	ios := f.IOStreams()

	scene, err := parseSceneFile(file)
	if err != nil {
		return err
	}

	// Override name if specified
	if nameOverride != "" {
		scene.Name = nameOverride
	}

	if scene.Name == "" {
		return fmt.Errorf("scene name is required (use --name to specify)")
	}

	if err := saveScene(scene, overwrite); err != nil {
		return err
	}

	ios.Success("Imported scene %q with %d action(s)", scene.Name, len(scene.Actions))
	return nil
}

func parseSceneFile(file string) (*Scene, error) {
	// #nosec G304 -- file path comes from user CLI argument
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var scene Scene

	// Determine format from file extension
	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &scene); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &scene); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		if err := parseUnknownFormat(data, &scene); err != nil {
			return nil, err
		}
	}

	return &scene, nil
}

func parseUnknownFormat(data []byte, scene *Scene) error {
	// Try YAML first, then JSON
	if err := yaml.Unmarshal(data, scene); err != nil {
		if jsonErr := json.Unmarshal(data, scene); jsonErr != nil {
			return fmt.Errorf("failed to parse file (tried YAML and JSON)")
		}
	}
	return nil
}

func saveScene(scene *Scene, overwrite bool) error {
	// Check if scene exists
	if _, exists := config.GetScene(scene.Name); exists {
		if !overwrite {
			return fmt.Errorf("scene %q already exists (use --overwrite to replace)", scene.Name)
		}
		// Delete existing scene first
		if err := config.DeleteScene(scene.Name); err != nil {
			return fmt.Errorf("failed to delete existing scene: %w", err)
		}
	}

	// Create scene
	if err := config.CreateScene(scene.Name, scene.Description); err != nil {
		return fmt.Errorf("failed to create scene: %w", err)
	}

	// Add actions
	if len(scene.Actions) > 0 {
		if err := config.SetSceneActions(scene.Name, scene.Actions); err != nil {
			return fmt.Errorf("failed to set scene actions: %w", err)
		}
	}

	return nil
}

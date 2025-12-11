// Package configimport provides the config import subcommand.
// Named configimport to avoid conflict with Go's import keyword.
package configimport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	dryRunFlag    bool
	mergeFlag     bool
	overwriteFlag bool
)

// NewCommand creates the config import command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <device> <file>",
		Short: "Import configuration from a file",
		Long: `Import device configuration from a JSON or YAML file.

By default, only specified keys are updated (merge mode). Use --overwrite
to replace the entire configuration.`,
		Example: `  # Import configuration (merge mode)
  shelly config import living-room config-backup.json

  # Dry run - show what would change without applying
  shelly config import living-room config.json --dry-run

  # Overwrite entire configuration
  shelly config import living-room config.json --overwrite`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], args[1])
		},
	}

	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what would be changed without applying")
	cmd.Flags().BoolVar(&mergeFlag, "merge", true, "Merge with existing configuration (default)")
	cmd.Flags().BoolVar(&overwriteFlag, "overwrite", false, "Overwrite entire configuration")

	return cmd
}

func run(ctx context.Context, device, filePath string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	// Read and parse file
	fileData, err := os.ReadFile(filePath) //nolint:gosec // G304: filePath is user-provided CLI argument, intentional
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var config map[string]any

	// Try JSON first, then YAML
	if err := json.Unmarshal(fileData, &config); err != nil {
		if err := yaml.Unmarshal(fileData, &config); err != nil {
			return fmt.Errorf("failed to parse file as JSON or YAML: %w", err)
		}
	}

	svc := shelly.NewService()

	if dryRunFlag {
		// Get current config and show diff
		spin := iostreams.NewSpinner("Getting current configuration...")
		spin.Start()

		currentConfig, err := svc.GetConfig(ctx, device)
		spin.Stop()

		if err != nil {
			return fmt.Errorf("failed to get current configuration: %w", err)
		}

		iostreams.Title("Dry run - changes that would be applied")
		showDiff(currentConfig, config)
		return nil
	}

	spin := iostreams.NewSpinner("Importing configuration...")
	spin.Start()

	err = svc.SetConfig(ctx, device, config)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to import configuration: %w", err)
	}

	iostreams.Success("Configuration imported to %s", device)
	return nil
}

// showDiff displays the differences between current and incoming config.
func showDiff(current, incoming map[string]any) {
	for key, incomingVal := range incoming {
		currentVal, exists := current[key]
		if !exists {
			fmt.Printf("  + %s: %v (new)\n", key, formatValue(incomingVal))
		} else if !deepEqual(currentVal, incomingVal) {
			fmt.Printf("  ~ %s: %v -> %v\n", key, formatValue(currentVal), formatValue(incomingVal))
		}
	}
}

// formatValue formats a value for display.
func formatValue(v any) string {
	if v == nil {
		return "<null>"
	}
	switch val := v.(type) {
	case string:
		if val == "" {
			return `""`
		}
		return fmt.Sprintf("%q", val)
	case map[string]any, []any:
		data, err := json.Marshal(val)
		if err != nil {
			iostreams.DebugErr("marshaling config value", err)
			return fmt.Sprintf("%v", val)
		}
		s := string(data)
		if len(s) > 50 {
			return s[:47] + "..."
		}
		return s
	default:
		return fmt.Sprintf("%v", val)
	}
}

// deepEqual compares two values for equality.
func deepEqual(a, b any) bool {
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)
	if aErr != nil {
		iostreams.DebugErr("marshaling value for comparison", aErr)
		return false
	}
	if bErr != nil {
		iostreams.DebugErr("marshaling value for comparison", bErr)
		return false
	}
	return bytes.Equal(aJSON, bJSON)
}

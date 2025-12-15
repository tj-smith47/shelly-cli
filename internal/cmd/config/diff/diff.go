// Package diff provides the config diff subcommand.
package diff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the config diff command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "diff <device> <file>",
		Aliases: []string{"compare", "cmp"},
		Short:   "Compare device configuration with a file",
		Long: `Compare the current device configuration with a saved configuration file.

Shows differences between the device's current configuration and the file.`,
		Example: `  # Compare config with a backup file
  shelly config diff living-room config-backup.json

  # Compare after making changes
  shelly config diff office-switch original-config.json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], args[1])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device, filePath string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	// Read file config
	fileData, err := os.ReadFile(filePath) //nolint:gosec // G304: filePath is user-provided CLI argument, intentional
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var fileConfig map[string]any
	if err := json.Unmarshal(fileData, &fileConfig); err != nil {
		return fmt.Errorf("failed to parse file as JSON: %w", err)
	}

	svc := f.ShellyService()
	ios := f.IOStreams()

	ios.StartProgress("Getting device configuration...")

	deviceConfig, err := svc.GetConfig(ctx, device)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to get device configuration: %w", err)
	}

	// Compare configurations
	diffs := compareConfigs("", deviceConfig, fileConfig)

	if len(diffs) == 0 {
		ios.Success("Configurations are identical")
		return nil
	}

	ios.Title("Configuration differences")
	ios.Printf("Comparing device %s with file %s\n\n", device, filePath)

	for _, d := range diffs {
		ios.Printf("%s\n", d)
	}

	ios.Printf("\n%d difference(s) found\n", len(diffs))
	return nil
}

// compareConfigs recursively compares two configuration maps.
func compareConfigs(prefix string, device, file map[string]any) []string {
	var diffs []string

	// Get all keys from both maps
	keys := make(map[string]bool)
	for k := range device {
		keys[k] = true
	}
	for k := range file {
		keys[k] = true
	}

	// Sort keys for consistent output
	sortedKeys := make([]string, 0, len(keys))
	for k := range keys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		deviceVal, deviceHas := device[key]
		fileVal, fileHas := file[key]

		if !deviceHas {
			diffs = append(diffs, fmt.Sprintf("  + %s: %v (in file only)", path, formatValue(fileVal)))
			continue
		}

		if !fileHas {
			diffs = append(diffs, fmt.Sprintf("  - %s: %v (in device only)", path, formatValue(deviceVal)))
			continue
		}

		// Both have the key, compare values
		deviceMap, deviceIsMap := deviceVal.(map[string]any)
		fileMap, fileIsMap := fileVal.(map[string]any)

		if deviceIsMap && fileIsMap {
			// Recursively compare nested maps
			nested := compareConfigs(path, deviceMap, fileMap)
			diffs = append(diffs, nested...)
		} else if !reflect.DeepEqual(deviceVal, fileVal) {
			diffs = append(diffs, fmt.Sprintf("  ~ %s: %v -> %v", path, formatValue(deviceVal), formatValue(fileVal)))
		}
	}

	return diffs
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
			return fmt.Sprintf("%v", val)
		}
		return string(data)
	default:
		return fmt.Sprintf("%v", val)
	}
}

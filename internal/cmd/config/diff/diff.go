// Package diff provides the config diff subcommand.
package diff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
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

	// Compare configurations using shared comparison function
	diffs := output.CompareConfigs(deviceConfig, fileConfig)

	if len(diffs) == 0 {
		ios.Success("Configurations are identical")
		return nil
	}

	ios.Title("Configuration differences")
	ios.Printf("Comparing device %s with file %s\n\n", device, filePath)

	for _, d := range diffs {
		switch d.DiffType {
		case model.DiffAdded:
			ios.Printf("  + %s: %v (in file only)\n", d.Path, output.FormatDisplayValue(d.NewValue))
		case model.DiffRemoved:
			ios.Printf("  - %s: %v (in device only)\n", d.Path, output.FormatDisplayValue(d.OldValue))
		case model.DiffChanged:
			ios.Printf("  ~ %s: %v -> %v\n", d.Path, output.FormatDisplayValue(d.OldValue), output.FormatDisplayValue(d.NewValue))
		}
	}

	ios.Printf("\n%d difference(s) found\n", len(diffs))
	return nil
}

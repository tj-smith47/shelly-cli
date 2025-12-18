// Package diff provides the diff command for comparing device configurations.
package diff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	OnlyDiffs bool
}

// NewCommand creates the diff command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "diff <source> <target>",
		Aliases: []string{"compare", "cmp"},
		Short:   "Compare device configurations",
		Long: `Compare configurations between two devices or a device and a backup file.

This command shows differences in configuration between:
  - Two live devices: shelly diff device1 device2
  - Device and backup: shelly diff device backup.json
  - Two backup files: shelly diff backup1.json backup2.json

Differences are shown with:
  - Added values (only in target)
  - Removed values (only in source)
  - Changed values (different between source and target)`,
		Example: `  # Compare two devices
  shelly diff kitchen-light bedroom-light

  # Compare device with backup
  shelly diff kitchen-light kitchen-backup.json

  # Show only differences (hide identical values)
  shelly diff device1 device2 --only-diff

  # JSON output
  shelly diff device1 device2 --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], args[1], opts)
		},
	}

	cmd.Flags().BoolVar(&opts.OnlyDiffs, "only-diff", false, "Show only differences")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, source, target string, _ *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	ios.StartProgress("Fetching configurations...")

	// Get source config
	sourceConfig, sourceName, err := getConfig(ctx, svc, source)
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to get source config: %w", err)
	}

	// Get target config
	targetConfig, targetName, err := getConfig(ctx, svc, target)
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to get target config: %w", err)
	}

	ios.StopProgress()

	// Calculate diff using shared comparison function
	diffs := output.CompareConfigs(sourceConfig, targetConfig)

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, diffs)
	}

	// Display results
	ios.Printf("Comparing: %s ↔ %s\n", sourceName, targetName)
	ios.Println("")

	if len(diffs) == 0 {
		ios.Success("Configurations are identical")
		return nil
	}

	// Group by type
	var added, removed, changed []model.ConfigDiff
	for _, d := range diffs {
		switch d.DiffType {
		case model.DiffAdded:
			added = append(added, d)
		case model.DiffRemoved:
			removed = append(removed, d)
		case model.DiffChanged:
			changed = append(changed, d)
		}
	}

	// Display differences
	if len(removed) > 0 {
		ios.Println(output.RenderDiffRemoved())
		for _, d := range removed {
			ios.Printf("  - %s: %v\n", d.Path, d.OldValue)
		}
		ios.Println("")
	}

	if len(added) > 0 {
		ios.Println(output.RenderDiffAdded())
		for _, d := range added {
			ios.Printf("  + %s: %v\n", d.Path, d.NewValue)
		}
		ios.Println("")
	}

	if len(changed) > 0 {
		ios.Println(output.RenderDiffChanged())
		for _, d := range changed {
			ios.Printf("  ~ %s:\n", d.Path)
			ios.Printf("    - %v\n", d.OldValue)
			ios.Printf("    + %v\n", d.NewValue)
		}
		ios.Println("")
	}

	ios.Printf("Summary: %d added, %d removed, %d changed\n",
		len(added), len(removed), len(changed))

	return nil
}

func getConfig(ctx context.Context, svc *shelly.Service, source string) (cfg map[string]any, name string, err error) {
	// Check if it's a file
	if isFile(source) {
		data, err := os.ReadFile(source) //nolint:gosec // User-provided file path is intentional
		if err != nil {
			return nil, "", fmt.Errorf("failed to read file: %w", err)
		}

		var config map[string]any
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, "", fmt.Errorf("failed to parse JSON: %w", err)
		}

		return config, source, nil
	}

	// It's a device - get config via API
	conn, err := svc.Connect(ctx, source)
	if err != nil {
		return nil, "", err
	}
	defer iostreams.CloseWithDebug("closing diff connection", conn)

	rawResult, err := conn.Call(ctx, "Shelly.GetConfig", nil)
	if err != nil {
		return nil, "", err
	}

	// Convert to map
	jsonBytes, err := json.Marshal(rawResult)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal config: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, "", fmt.Errorf("failed to parse config: %w", err)
	}

	return result, source, nil
}

func isFile(path string) bool {
	// Check if path looks like a file (contains .json or .yaml extension)
	if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		return true
	}
	// Check if file exists
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

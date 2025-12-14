// Package importcmd provides the kvs import subcommand.
// Named importcmd to avoid conflict with the "import" keyword.
package importcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device    string
	File      string
	Overwrite bool
	DryRun    bool
	Yes       bool
	Factory   *cmdutil.Factory
}

// NewCommand creates the kvs import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "import <device> <file>",
		Aliases: []string{"load", "restore"},
		Short:   "Import KVS data from file",
		Long: `Import key-value pairs from a file to the device.

By default, existing keys are skipped. Use --overwrite to replace them.
Use --dry-run to see what would be imported without making changes.`,
		Example: `  # Import from JSON file
  shelly kvs import living-room kvs-backup.json

  # Import with overwrite
  shelly kvs import living-room kvs-backup.json --overwrite

  # Dry run to see what would be imported
  shelly kvs import living-room kvs-backup.json --dry-run

  # Import without confirmation
  shelly kvs import living-room kvs-backup.json --yes`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.File = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Overwrite, "overwrite", false, "Overwrite existing keys")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be imported without making changes")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*2) // Allow more time for imports
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Parse the import file
	data, err := parseImportFile(opts.File)
	if err != nil {
		return err
	}

	if len(data.Items) == 0 {
		ios.NoResults("No keys to import")
		return nil
	}

	// Show preview
	showPreview(ios, data)

	// Handle dry run
	if opts.DryRun {
		handleDryRun(ios, len(data.Items), opts.Overwrite)
		return nil
	}

	// Confirm import
	if !opts.Yes {
		if !confirmImport(ios, len(data.Items), opts.Device, opts.Overwrite) {
			ios.Info("Aborted")
			return nil
		}
	}

	// Execute import
	return executeImport(ctx, ios, svc, opts.Device, data, opts.Overwrite)
}

// parseImportFile reads and parses the import file (JSON or YAML).
func parseImportFile(file string) (*shelly.KVSExport, error) {
	//nolint:gosec // G304: file path is from user command line argument
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data shelly.KVSExport
	if err := json.Unmarshal(content, &data); err != nil {
		// Try YAML
		if yamlErr := yaml.Unmarshal(content, &data); yamlErr != nil {
			return nil, fmt.Errorf("failed to parse file (tried JSON and YAML): %w", err)
		}
	}

	return &data, nil
}

// showPreview displays the items that will be imported.
func showPreview(ios *iostreams.IOStreams, data *shelly.KVSExport) {
	ios.Printf("Found %d key(s) to import:\n", len(data.Items))
	for _, item := range data.Items {
		ios.Printf("  - %s = %s\n", item.Key, formatValue(item.Value))
	}
	ios.Println()
}

// handleDryRun handles the dry run output.
func handleDryRun(ios *iostreams.IOStreams, count int, overwrite bool) {
	if overwrite {
		ios.Info("Would import %d key(s) (overwrite enabled)", count)
	} else {
		ios.Info("Would import up to %d key(s) (existing keys skipped)", count)
	}
}

// confirmImport prompts for user confirmation.
func confirmImport(ios *iostreams.IOStreams, count int, device string, overwrite bool) bool {
	action := "Import"
	if overwrite {
		action = "Import and overwrite"
	}
	msg := fmt.Sprintf("%s %d key(s) to %s?", action, count, device)
	confirmed, err := ios.Confirm(msg, false)
	if err != nil {
		return false
	}
	return confirmed
}

// executeImport performs the actual import operation.
func executeImport(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string, data *shelly.KVSExport, overwrite bool) error {
	var imported, skipped int
	err := cmdutil.RunWithSpinner(ctx, ios, "Importing KVS data...", func(ctx context.Context) error {
		var importErr error
		imported, skipped, importErr = svc.ImportKVS(ctx, device, data, overwrite)
		return importErr
	})
	if err != nil {
		return err
	}

	// Report results
	reportResults(ios, imported, skipped)
	return nil
}

// reportResults outputs the import results.
func reportResults(ios *iostreams.IOStreams, imported, skipped int) {
	var msgs []string
	if imported > 0 {
		msgs = append(msgs, fmt.Sprintf("%d imported", imported))
	}
	if skipped > 0 {
		msgs = append(msgs, fmt.Sprintf("%d skipped (already exist)", skipped))
	}

	result := strings.Join(msgs, ", ")
	if imported > 0 {
		ios.Success("%s", result)
	} else {
		ios.Info("%s", result)
	}
}

func formatValue(v any) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case string:
		if len(val) > 30 {
			return fmt.Sprintf("%q...", val[:27])
		}
		return fmt.Sprintf("%q", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%g", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
